package ai

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

// const ollamaUrl = "http://localhost:11434/api/chat"
// const ollamaUrl = "http://8.138.143.14:6781/api/chat"

func injectAndCopy(w http.ResponseWriter, body io.Reader, fileinfos []string) string {
	flusher, _ := w.(http.Flusher)
	var capturedData strings.Builder

	logx.Info("injectAndCopy started")

	buf := make([]byte, 1024)
	for {
		n, err := body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			if flusher != nil {
				flusher.Flush()
			}
			fmt.Print(".")
			capturedData.Write(buf[:n])
		}
		if err != nil {
			if err != io.EOF {
				logx.Error("Error reading from response body:", err)
			} else {
				logx.Info("reading from response body EOF")
			}
			break
		}
	}

	// After stream ends, inject additional data
	extra := ""
	if len(fileinfos) > 0 {
		for _, fileinfo := range fileinfos {
			extra += fileinfo + "\n"
		}
	}
	w.Write([]byte(extra))
	if flusher != nil {
		flusher.Flush()
	}
	capturedData.WriteString(extra)

	return capturedData.String()
}

func AIChatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientId, _ := r.Context().Value("clientId").(string)
		baseURL, _, completionApiKey := utils.ResolveCompletionRuntime(&svcCtx.Config, clientId)
		apiUrl := ""
		if baseURL != "" {
			apiUrl = strings.TrimSuffix(baseURL, "/") + "/v1/chat/completions"
		}
		if apiUrl == "" {
			http.Error(w, "未配置 LLM 地址（Llm.CompletionUrl）", http.StatusServiceUnavailable)
			return
		}
		req, err := http.NewRequest("POST", apiUrl, r.Body)
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}
		req.Header = r.Header
		if strings.TrimSpace(completionApiKey) != "" {
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(completionApiKey))
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Failed to contact Ollama API", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy Ollama response headers
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		w.WriteHeader(resp.StatusCode)
		var fileinfos []string
		fileinfos = append(fileinfos, "https://coolpeople.com.cn/api/v1/static/kb/%E5%8F%98%E6%A1%A81.jpg")
		fileinfos = append(fileinfos, "https://coolpeople.com.cn/api/v1/static/kb/v2.pdf")

		// Stream response to client, with possible injection
		capturedData := injectAndCopy(w, resp.Body, fileinfos)

		logx.Infof("Captured data: %s", capturedData)
	}
}
