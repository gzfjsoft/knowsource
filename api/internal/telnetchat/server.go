package telnetchat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"knowsource/consts"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type Options struct {
	ListenAddr string
	ChatURL    string
	RouterURL  string // resolved LLM chat/completions url (no ai/session/chat)
	AdminJWT   string
	ClientID   string
	FilesRoot  string // root directory for list_files tool
}

func (o Options) withDefaults() Options {
	if strings.TrimSpace(o.ListenAddr) == "" {
		o.ListenAddr = ":2323"
	}
	if strings.TrimSpace(o.ClientID) == "" {
		o.ClientID = consts.ONLY_ADMIN
	}
	if strings.TrimSpace(o.FilesRoot) == "" {
		o.FilesRoot = "."
	}
	return o
}

// Start starts a TCP telnet-like chat server and blocks until ctx is canceled or listener fails.
func Start(ctx context.Context, opt Options) error {
	opt = opt.withDefaults()
	if strings.TrimSpace(opt.AdminJWT) == "" {
		return fmt.Errorf("missing AdminJWT")
	}
	if strings.TrimSpace(opt.ChatURL) == "" {
		return fmt.Errorf("missing ChatURL")
	}
	if strings.TrimSpace(opt.RouterURL) == "" {
		// fallback to ChatURL; router will still work but via ai/session/chat (legacy)
		opt.RouterURL = opt.ChatURL
	}

	ln, err := net.Listen("tcp", opt.ListenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	go func() {
		<-ctx.Done()
		_ = ln.Close()
	}()

	for {
		c, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		go handleConn(c, opt.ChatURL, opt.RouterURL, opt.AdminJWT, opt.ClientID, opt.FilesRoot)
	}
}

type chatRequest struct {
	Message      string   `json:"message"`
	Session      string   `json:"session,omitempty"`
	Think        bool     `json:"think,omitempty"`
	Prompt       string   `json:"prompt,omitempty"`
	Model        string   `json:"model,omitempty"`
	DocumentCode string   `json:"documentCode,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Skiprag      bool     `json:"skiprag,omitempty"`
}

type chatResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Info    string `json:"info,omitempty"`
	Data    struct {
		Session  string `json:"session,omitempty"`
		Response string `json:"response"`
	} `json:"data"`
}

func handleConn(conn net.Conn, chatURL, routerURL, adminJWT, clientID string, filesRoot string) {
	defer conn.Close()

	_ = conn.SetDeadline(time.Time{})
	bw := bufio.NewWriter(conn)
	br := bufio.NewReader(conn)

	write := func(s string) {
		_, _ = bw.WriteString(s)
		_ = bw.Flush()
	}
	writeBytes := func(b []byte) {
		_, _ = bw.Write(b)
		_ = bw.Flush()
	}

	session := ""
	think := true
	skiprag := false
	model := ""
	// Prompt is always included in ai/session/chat requests.
	// Can be overridden via /prompt.
	prompt := "用中文回答用户"
	documentCode := ""
	planEnabled := false

	write("Knowsource Telnet Chat\n")
	write("Controls:\n")
	write("- Ctrl+Enter: send (CRLF)\n")
	write("- Enter: newline while composing (LF)\n")
	write("- /docs : list knowledge bases\n")
	write("- /doc <code> : set knowledge base\n")
	write("- /ls [path] : list local files under root\n")
	write("- /plan on|off : show router decision stream\n")
	write("- /new : new session\n")
	write("- /think on|off\n")
	write("- /skiprag on|off\n")
	write("- /model <name>\n")
	write("- /prompt <text>\n")
	write("- /quit\n\n")

	// Select knowledge base (documentCode) at entry.
	docTypes, _ := fetchDocumentTypes(context.Background(), chatURL, adminJWT, clientID)
	if len(docTypes) > 0 {
		write("Knowledge bases:\n")
		for i, d := range docTypes {
			code := strings.TrimSpace(d.Code)
			name := strings.TrimSpace(d.Name)
			if code == "" {
				code = "(empty)"
			}
			if name == "" {
				name = "-"
			}
			write(fmt.Sprintf("  %d) %s  %s\n", i+1, code, name))
		}
		write("Choose documentCode by number or code (empty=none): ")
		if s, err := readLine(br); err == nil {
			s = strings.TrimSpace(s)
			if s != "" {
				if idx := parseIndex(s); idx >= 1 && idx <= len(docTypes) {
					documentCode = strings.TrimSpace(docTypes[idx-1].Code)
				} else {
					documentCode = s
				}
			}
		}
	} else {
		write("Input documentCode (knowledge base), empty=none: ")
		if s, err := readLine(br); err == nil {
			documentCode = strings.TrimSpace(s)
		}
	}
	if documentCode != "" {
		write(fmt.Sprintf("[ok] documentCode=%s\n\n", documentCode))
	} else {
		write("[ok] documentCode=(none)\n\n")
	}
	write("> ")

	var compose bytes.Buffer
	for {
		msg, send, quit, err := readCompose(br, &compose)
		if err != nil {
			if err == io.EOF {
				return
			}
			write("\n[error] " + err.Error() + "\n> ")
			continue
		}
		if quit {
			return
		}
		if !send {
			continue
		}

		trimmed := strings.TrimSpace(msg)
		if trimmed == "" {
			write("> ")
			continue
		}

		if strings.HasPrefix(trimmed, "/") {
			switch {
			case trimmed == "/quit":
				return
			case trimmed == "/docs":
				docTypes, err := fetchDocumentTypes(context.Background(), chatURL, adminJWT, clientID)
				if err != nil {
					write("[error] " + err.Error() + "\n> ")
					continue
				}
				if len(docTypes) == 0 {
					write("[ok] no knowledge bases\n> ")
					continue
				}
				for i, d := range docTypes {
					write(fmt.Sprintf("%d) %s  %s\n", i+1, strings.TrimSpace(d.Code), strings.TrimSpace(d.Name)))
				}
				write("> ")
				continue
			case strings.HasPrefix(trimmed, "/doc "):
				documentCode = strings.TrimSpace(strings.TrimPrefix(trimmed, "/doc "))
				write(fmt.Sprintf("[ok] documentCode=%s\n> ", documentCode))
				continue
			case trimmed == "/ls" || strings.HasPrefix(trimmed, "/ls "):
				p := strings.TrimSpace(strings.TrimPrefix(trimmed, "/ls"))
				out, err := listLocalFiles(filesRoot, p)
				if err != nil {
					write("[error] " + err.Error() + "\n> ")
					continue
				}
				write(out + "\n> ")
				continue
			case strings.HasPrefix(trimmed, "/plan "):
				v := strings.TrimSpace(strings.TrimPrefix(trimmed, "/plan "))
				planEnabled = (v == "on" || v == "1" || strings.EqualFold(v, "true"))
				write(fmt.Sprintf("[ok] plan=%v\n> ", planEnabled))
				continue
			case trimmed == "/new":
				session = ""
				write("[ok] new session\n> ")
				continue
			case strings.HasPrefix(trimmed, "/think "):
				v := strings.TrimSpace(strings.TrimPrefix(trimmed, "/think "))
				think = (v == "on" || v == "1" || strings.EqualFold(v, "true"))
				write(fmt.Sprintf("[ok] think=%v\n> ", think))
				continue
			case strings.HasPrefix(trimmed, "/skiprag "):
				v := strings.TrimSpace(strings.TrimPrefix(trimmed, "/skiprag "))
				skiprag = (v == "on" || v == "1" || strings.EqualFold(v, "true"))
				write(fmt.Sprintf("[ok] skiprag=%v\n> ", skiprag))
				continue
			case strings.HasPrefix(trimmed, "/model "):
				model = strings.TrimSpace(strings.TrimPrefix(trimmed, "/model "))
				write(fmt.Sprintf("[ok] model=%s\n> ", model))
				continue
			case strings.HasPrefix(trimmed, "/prompt "):
				prompt = strings.TrimSpace(strings.TrimPrefix(trimmed, "/prompt "))
				write("[ok] prompt set\n> ")
				continue
			default:
				write("[error] unknown command\n> ")
				continue
			}
		}

		// Ask AI router which tool to use for this user input.
		tool, arg, routeErr := routeByAI(context.Background(), routerURL, model, prompt, msg, planEnabled, writeBytes)
		if routeErr == nil && tool == "list_files" {
			out, err := listLocalFiles(filesRoot, arg)
			if err != nil {
				write("\n[error] " + err.Error() + "\n> ")
				continue
			}
			write("\n" + out + "\n> ")
			continue
		}

		req := chatRequest{
			Message:      msg,
			Session:      session,
			Think:        think,
			Skiprag:      skiprag,
			Model:        model,
			Prompt:       prompt,
			DocumentCode: documentCode,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		// Stream output to telnet as chunks arrive.
		write("\n")
		thinkOpened := false
		_, newSession, err := callChat(
			ctx, chatURL, adminJWT, clientID, req,
			func(contentChunk string) {
				if contentChunk == "" {
					return
				}
				if thinkOpened {
					writeBytes([]byte("\n</think>\n"))
					thinkOpened = false
				}
				writeBytes([]byte(filterThinkTags(contentChunk)))
			},
			func(reasonChunk string) {
				if reasonChunk == "" {
					return
				}
				if !thinkOpened {
					writeBytes([]byte("<think>"))
					thinkOpened = true
				}
				writeBytes([]byte(reasonChunk))
			},
		)
		cancel()
		if err != nil {
			write("\n[error] " + err.Error() + "\n> ")
			continue
		}
		if thinkOpened {
			writeBytes([]byte("</think>\n"))
		}
		if newSession != "" {
			session = newSession
		}

		write("\n> ")
	}
}

const routerPrompt = `You are a tool router. Output STRICT JSON only, no extra text.

You have exactly two tools:
1) list_files: list local files/directories (for user intent like "list files", "show directory", "what files exist", "ls")
2) rag_chat: knowledge base RAG chat/search (all other cases)

Output format:
{"tool":"list_files","path":"<relative path or empty>"} OR {"tool":"rag_chat"}

Rules:
- Use list_files ONLY when the user clearly wants to list local directory contents
- "path" MUST be a relative path. Never output absolute paths. Never output "..". If unsure, output empty string.
`

// routeByAI calls RouterURL (LLM chat/completions) directly, not /ai/session/chat.
// When planEnabled, it streams the router output to user (telnet) prefixed by "[plan] ".
func routeByAI(ctx context.Context, routerURL string, model string, userPrompt string, userInput string, planEnabled bool, writeBytes func([]byte)) (tool string, path string, err error) {
	var buf strings.Builder
	streamPrefix := func(b []byte) {
		if !planEnabled || writeBytes == nil || len(b) == 0 {
			return
		}
		writeBytes([]byte("\n[plan] "))
		writeBytes(b)
	}

	combined := strings.TrimSpace(userPrompt)
	if combined != "" {
		combined = combined + "\n\n" + routerPrompt
	} else {
		combined = routerPrompt
	}

	content, callErr := callRouterLLMStream(ctx, routerURL, model, combined, userInput, func(chunk string) {
		if chunk == "" {
			return
		}
		buf.WriteString(chunk)
		streamPrefix([]byte(chunk))
	})
	_ = content
	if callErr != nil {
		return "", "", callErr
	}

	raw := strings.TrimSpace(buf.String())
	if raw == "" {
		return "", "", fmt.Errorf("router empty response")
	}
	j := extractFirstJSON(raw)
	if j == "" {
		return "", "", fmt.Errorf("router non-json response: %s", raw)
	}
	var parsed struct {
		Tool string `json:"tool"`
		Path string `json:"path"`
	}
	if e := json.Unmarshal([]byte(j), &parsed); e != nil {
		return "", "", e
	}
	parsed.Tool = strings.TrimSpace(parsed.Tool)
	if parsed.Tool == "" {
		return "", "", fmt.Errorf("router missing tool")
	}
	if parsed.Tool == "list_files" {
		p := strings.TrimSpace(parsed.Path)
		p = strings.TrimPrefix(p, "/")
		if strings.Contains(p, "..") {
			p = ""
		}
		return parsed.Tool, p, nil
	}
	return parsed.Tool, "", nil
}

type openAIStreamRequest struct {
	Model    string `json:"model,omitempty"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Stream bool `json:"stream"`
}

// callRouterLLMStream calls an OpenAI-compatible /v1/chat/completions (or Ollama-compatible stream) endpoint
// and extracts only the incremental content chunks.
func callRouterLLMStream(ctx context.Context, url string, model string, systemPrompt string, userInput string, onContentChunk func(string)) (string, error) {
	url = strings.TrimSpace(url)
	if url == "" {
		return "", fmt.Errorf("router url is empty")
	}

	reqBody := openAIStreamRequest{
		Model:  strings.TrimSpace(model),
		Stream: true,
	}
	reqBody.Messages = append(reqBody.Messages, struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{Role: "system", Content: systemPrompt})
	reqBody.Messages = append(reqBody.Messages, struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{Role: "user", Content: strings.TrimSpace(userInput)})

	bts, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", strings.TrimSuffix(url, "/"), bytes.NewReader(bts))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{Timeout: 60 * time.Second}).Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("router http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// parse like session chat stream: either "data: {...}" OpenAI SSE, or line json (ollama)
	content, err := parseLLMStream(resp.Body, onContentChunk)
	if err != nil {
		return content, err
	}
	return content, nil
}

func parseLLMStream(r io.Reader, onContentChunk func(string)) (string, error) {
	var full strings.Builder
	sc := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 8*1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		jsonStr := line
		if strings.HasPrefix(line, "data: ") {
			jsonStr = strings.TrimSpace(strings.TrimPrefix(line, "data: "))
			if jsonStr == "[DONE]" {
				continue
			}
		}
		if !strings.HasPrefix(jsonStr, "{") {
			continue
		}
		var obj map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
			continue
		}
		// Ollama style
		if msgAny, ok := obj["message"]; ok {
			if msgMap, ok := msgAny.(map[string]any); ok {
				if cAny, ok := msgMap["content"]; ok {
					if s, ok := cAny.(string); ok && s != "" {
						full.WriteString(s)
						if onContentChunk != nil {
							onContentChunk(s)
						}
					}
				}
			}
			continue
		}
		// OpenAI style
		if choicesAny, ok := obj["choices"]; ok {
			if choices, ok := choicesAny.([]any); ok {
				for _, chAny := range choices {
					chMap, ok := chAny.(map[string]any)
					if !ok {
						continue
					}
					deltaAny := chMap["delta"]
					if deltaAny == nil {
						deltaAny = chMap["message"]
					}
					if deltaMap, ok := deltaAny.(map[string]any); ok {
						if cAny, ok := deltaMap["content"]; ok {
							if s, ok := cAny.(string); ok && s != "" {
								full.WriteString(s)
								if onContentChunk != nil {
									onContentChunk(s)
								}
							}
						}
					}
				}
			}
		}
	}
	if err := sc.Err(); err != nil {
		return full.String(), err
	}
	return full.String(), nil
}

func extractFirstJSON(s string) string {
	start := strings.Index(s, "{")
	if start == -1 {
		return ""
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

func listLocalFiles(root, rel string) (string, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		root = "."
	}
	rel = strings.TrimSpace(rel)
	rel = strings.TrimPrefix(rel, "/")
	if strings.Contains(rel, "..") {
		return "", fmt.Errorf("invalid path")
	}

	full := filepath.Join(root, filepath.FromSlash(rel))
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	sep := string(filepath.Separator)
	if absFull != absRoot && !strings.HasPrefix(absFull, absRoot+sep) {
		return "", fmt.Errorf("invalid path")
	}

	ents, err := os.ReadDir(absFull)
	if err != nil {
		return "", err
	}

	type item struct {
		name string
		dir  bool
	}
	items := make([]item, 0, len(ents))
	for _, e := range ents {
		n := e.Name()
		if n == "" {
			continue
		}
		items = append(items, item{name: n, dir: e.IsDir()})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].dir != items[j].dir {
			return items[i].dir
		}
		return strings.ToLower(items[i].name) < strings.ToLower(items[j].name)
	})

	var b strings.Builder
	b.WriteString("root: " + absRoot + "\n")
	b.WriteString("path: " + strings.TrimPrefix(absFull, absRoot) + "\n")
	for _, it := range items {
		if it.dir {
			b.WriteString("[D] " + it.name + "/\n")
		} else {
			b.WriteString("[F] " + it.name + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

// readCompose reads raw bytes and supports:
// - LF (\n): newline while composing
// - CRLF (\r\n): send (Ctrl+Enter in many terminals)
// - backspace (0x7f or 0x08): delete last byte
func readCompose(br *bufio.Reader, compose *bytes.Buffer) (msg string, send bool, quit bool, err error) {
	b, err := br.ReadByte()
	if err != nil {
		return "", false, false, err
	}

	switch b {
	case 0x04: // Ctrl+D
		return "", false, true, nil
	case 0x08, 0x7f: // backspace
		deleteLastRune(compose)
		return "", false, false, nil
	case '\n': // newline (compose)
		compose.WriteByte('\n')
		return "", false, false, nil
	case '\r': // maybe CRLF -> send
		peek, pErr := br.Peek(1)
		if pErr == nil && len(peek) == 1 && peek[0] == '\n' {
			_, _ = br.ReadByte()
			msg = compose.String()
			compose.Reset()
			return msg, true, false, nil
		}
		// bare CR -> treat as send too
		msg = compose.String()
		compose.Reset()
		return msg, true, false, nil
	default:
		compose.WriteByte(b)
		return "", false, false, nil
	}
}

func deleteLastRune(buf *bytes.Buffer) {
	if buf == nil || buf.Len() == 0 {
		return
	}
	b := buf.Bytes()
	_, size := utf8.DecodeLastRune(b)
	if size <= 0 || size > len(b) {
		size = 1
	}
	buf.Reset()
	buf.Write(b[:len(b)-size])
}

func readLine(br *bufio.Reader) (string, error) {
	s, err := br.ReadString('\n')
	if err != nil && s == "" {
		return "", err
	}
	return strings.TrimRight(s, "\r\n"), nil
}

func parseIndex(s string) int {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
		if n > 10000 {
			return 0
		}
	}
	return n
}

type documentTypeItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func fetchDocumentTypes(ctx context.Context, chatURL, adminJWT, clientID string) ([]documentTypeItem, error) {
	// chatURL is like http://host:port/api/ai/session/chat
	base := chatURL
	if i := strings.Index(base, "/api/"); i >= 0 {
		base = base[:i]
	}
	url := strings.TrimRight(base, "/") + "/api/documents/type/list"

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte(`{}`)))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+adminJWT)
	httpReq.Header.Set("Client-Id", clientID)
	httpReq.Header.Set("App-Platform", "web")

	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bts, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(bts)))
	}

	var parsed struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Info    string `json:"info,omitempty"`
		Data    *struct {
			List []documentTypeItem `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(bts, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 200 {
		msg := parsed.Message
		if msg == "" {
			msg = "api error"
		}
		if strings.TrimSpace(parsed.Info) != "" {
			return nil, fmt.Errorf("%s (%d): %s", msg, parsed.Code, parsed.Info)
		}
		return nil, fmt.Errorf("%s (%d)", msg, parsed.Code)
	}
	if parsed.Data == nil || len(parsed.Data.List) == 0 {
		return nil, nil
	}
	return parsed.Data.List, nil
}

func callChat(
	ctx context.Context,
	url, adminJWT, clientID string,
	req chatRequest,
	onContentChunk func(string),
	onReasoningChunk func(string),
) (answer string, session string, err error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+adminJWT)
	httpReq.Header.Set("Client-Id", clientID)
	httpReq.Header.Set("App-Platform", "web")

	resp, err := (&http.Client{Timeout: 180 * time.Second}).Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	session = resp.Header.Get("SessionUuid")

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bts, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(bts)))
	}

	// Stream (SSE) response: extract content like Vue `sessionChat`.
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/event-stream") {
		content, parseErr := parseSessionChatStream(resp.Body, onContentChunk, onReasoningChunk)
		if parseErr != nil {
			return "", session, parseErr
		}
		return filterThinkTags(content), session, nil
	}

	// Non-stream response: try JSON envelope first.
	bts, _ := io.ReadAll(resp.Body)
	var cr chatResponse
	if jErr := json.Unmarshal(bts, &cr); jErr == nil && cr.Code != 0 {
		if cr.Code != 200 {
			msg := cr.Message
			if msg == "" {
				msg = "api error"
			}
			if strings.TrimSpace(cr.Info) != "" {
				return "", session, fmt.Errorf("%s (%d): %s", msg, cr.Code, cr.Info)
			}
			return "", session, fmt.Errorf("%s (%d)", msg, cr.Code)
		}
		if cr.Data.Session != "" {
			session = cr.Data.Session
		}
		return cr.Data.Response, session, nil
	}

	return strings.TrimSpace(string(bts)), session, nil
}

// parseSessionChatStream parses the server's streaming response and extracts only assistant content.
// It supports both:
// - vLLM/OpenAI-style SSE: lines like "data: {...}" with choices[].delta.content
// - Ollama-style line-by-line JSON with message.content
// It ignores meta lines like {"filesinfos":[...], "stats": {...}}.
func parseSessionChatStream(r io.Reader, onContentChunk func(string), onReasoningChunk func(string)) (string, error) {
	var full strings.Builder

	sc := bufio.NewScanner(r)
	// allow long lines/chunks
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 8*1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		jsonStr := line
		if strings.HasPrefix(line, "data: ") {
			jsonStr = strings.TrimSpace(strings.TrimPrefix(line, "data: "))
			if jsonStr == "[DONE]" {
				continue
			}
		}
		if !strings.HasPrefix(jsonStr, "{") {
			continue
		}

		var obj map[string]any
		if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
			continue
		}

		// ignore meta
		if _, ok := obj["filesinfos"]; ok {
			continue
		}
		if _, ok := obj["stats"]; ok {
			continue
		}

		// Ollama style: {"message":{"content":"..."...}}
		if msgAny, ok := obj["message"]; ok {
			if msgMap, ok := msgAny.(map[string]any); ok {
				// reasoning/thinking
				if onReasoningChunk != nil {
					if rAny, ok := msgMap["reasoning"]; ok {
						if s, ok := rAny.(string); ok && s != "" {
							onReasoningChunk(s)
						}
					}
					if tAny, ok := msgMap["thinking"]; ok {
						if s, ok := tAny.(string); ok && s != "" {
							onReasoningChunk(s)
						}
					}
				}
				if cAny, ok := msgMap["content"]; ok {
					if s, ok := cAny.(string); ok && s != "" {
						full.WriteString(s)
						if onContentChunk != nil {
							onContentChunk(s)
						}
					}
				}
			}
			continue
		}

		// OpenAI/vLLM style: {"choices":[{"delta":{"content":"..."}}]}
		if choicesAny, ok := obj["choices"]; ok {
			if choices, ok := choicesAny.([]any); ok {
				for _, chAny := range choices {
					chMap, ok := chAny.(map[string]any)
					if !ok {
						continue
					}
					deltaAny := chMap["delta"]
					if deltaAny == nil {
						deltaAny = chMap["message"]
					}
					if deltaMap, ok := deltaAny.(map[string]any); ok {
						// reasoning/thinking
						if onReasoningChunk != nil {
							if rAny, ok := deltaMap["reasoning"]; ok {
								if s, ok := rAny.(string); ok && s != "" {
									onReasoningChunk(s)
								}
							}
							if tAny, ok := deltaMap["thinking"]; ok {
								if s, ok := tAny.(string); ok && s != "" {
									onReasoningChunk(s)
								}
							}
						}
						if cAny, ok := deltaMap["content"]; ok {
							if s, ok := cAny.(string); ok && s != "" {
								full.WriteString(s)
								if onContentChunk != nil {
									onContentChunk(s)
								}
							}
						}
					}
				}
			}
			continue
		}
	}
	if err := sc.Err(); err != nil {
		return full.String(), err
	}
	return full.String(), nil
}

// filterThinkTags mirrors the Vue-side behavior:
// - remove <think>...</think> blocks that contain only whitespace
// - trim leading whitespace/newlines inside think blocks
func filterThinkTags(s string) string {
	if s == "" {
		return s
	}
	// Lightweight handling: if <think></think> exists and contains only whitespace, drop it.
	// Also trim whitespace right after <think>.
	out := s
	for {
		start := strings.Index(out, "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(out[start:], "</think>")
		if end == -1 {
			break
		}
		end = start + end + len("</think>")
		block := out[start:end]
		inner := strings.TrimPrefix(block, "<think>")
		inner = strings.TrimSuffix(inner, "</think>")
		trimLeading := strings.TrimLeft(inner, " \t\r\n")
		if strings.TrimSpace(trimLeading) == "" {
			out = out[:start] + out[end:]
			continue
		}
		repl := "<think>" + trimLeading + "</think>"
		out = out[:start] + repl + out[end:]
	}
	return out
}
