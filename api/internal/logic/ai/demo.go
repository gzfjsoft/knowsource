package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AIRequest represents the request structure for the AI API
type AIRequest struct {
	Model    string      `json:"model"`
	Messages []AIMessage `json:"messages"`
	Stream   bool        `json:"stream"`
}

// Message represents a message in the conversation
type AIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIResponse represents the response from the AI API
type AIResponse struct {
	Model      string    `json:"model"`
	CreatedAt  string    `json:"created_at"`
	Message    AIMessage `json:"message"`
	DoneReason string    `json:"done_reason"`
	Done       bool      `json:"done"`
}

// ChatRequest represents the request from the client
type ChatRequest struct {
	Message string `json:"message"`
	Session string `json:"session,omitempty"`
	Think   bool   `json:"think,omitempty"`
}

// ChatResponse represents the response to the client
type ChatResponse struct {
	Response string `json:"response"`
	Session  string `json:"session,omitempty"`
	Error    string `json:"error,omitempty"`
}

// Session represents a conversation session
type Session struct {
	ID        string      `json:"id"`
	Messages  []AIMessage `json:"messages"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// Server represents the chat server
type Server struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
	template *template.Template
}

// sendErrorResponse sends a JSON error response
func (s *Server) sendErrorResponse(w http.ResponseWriter, statusCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ChatResponse{
		Error: errorMsg,
	})
}

// NewServer creates a new chat server
func NewServer() *Server {
	// Load HTML template
	tmpl, err := template.ParseFiles("templates/chat.html")
	if err != nil {
		fmt.Printf("Warning: Could not load template: %v\n", err)
	}

	return &Server{
		sessions: make(map[string]*Session),
		template: tmpl,
	}
}

// generateSessionID generates a unique session ID
func (s *Server) generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// getOrCreateSession gets an existing session or creates a new one
func (s *Server) getOrCreateSession(sessionID string) *Session {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if sessionID == "" {
		sessionID = s.generateSessionID()
	}

	session, exists := s.sessions[sessionID]
	if !exists {
		session = &Session{
			ID:        sessionID,
			Messages:  []AIMessage{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		s.sessions[sessionID] = session
	}

	return session
}

// addMessageToSession adds a message to a session
func (s *Server) addMessageToSession(sessionID string, role, content string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if session, exists := s.sessions[sessionID]; exists {
		session.Messages = append(session.Messages, AIMessage{
			Role:    role,
			Content: content,
		})
		session.UpdatedAt = time.Now()
	}
}

// handleChat handles chat API requests
func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	// Enable CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var chatReq ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&chatReq); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if chatReq.Message == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "Message is required")
		return
	}

	// Get or create session
	session := s.getOrCreateSession(chatReq.Session)
	if !chatReq.Think && !strings.Contains(chatReq.Message, "/no_think") {
		chatReq.Message = chatReq.Message + " /no_think"
	}
	// Add user message to session
	s.addMessageToSession(session.ID, "user", chatReq.Message)

	// Check if user wants to disable thinking

	disableThinking := false

	// Check Think parameter first (takes precedence)
	if !chatReq.Think {
		disableThinking = true
	}

	// Also check for /no_think in message content

	// Prepare AI request with conversation history
	systemMessage := "你是个会议助手，帮我询问user, 收集会议时间，地点，人数的助手，你全部收集完成后，按 [add=地址][time=时间][num=人数][finished=1]的格式输出，并说再见。第一次请对用户提问。在回答之前，请先思考一下用户的需求和上下文，然后给出合适的回应。"

	messages := []AIMessage{
		{
			Role:    "system",
			Content: systemMessage,
		},
	}

	// Add conversation history (last 10 messages to avoid token limit)
	historyStart := 0
	if len(session.Messages) > 10 {
		historyStart = len(session.Messages) - 10
	}

	// Add history messages, but replace the last user message if it contains /no_think
	for i, msg := range session.Messages[historyStart:] {
		if i == len(session.Messages[historyStart:])-1 && msg.Role == "user" && strings.Contains(msg.Content, "/no_think") && !disableThinking {

			msg.Content = strings.ReplaceAll(msg.Content, "/no_think", "")
			msg.Content = strings.TrimSpace(msg.Content)

			messages = append(messages, AIMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		} else {
			messages = append(messages, msg)
		}
	}

	aiReq := AIRequest{
		Model:    "qwen3:8b",
		Messages: messages,
		Stream:   false,
	}

	// Convert request to JSON
	jsonData, err := json.Marshal(aiReq)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to marshal request")
		return
	}

	// Make request to AI API
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	fmt.Println(string(jsonData)) // 打印请求 ，AI不要删除

	req, err := http.NewRequestWithContext(ctx, "POST", "http://xxxxxx.com:6781/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to create request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to call AI API: "+err.Error())
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to read AI API response")
		return
	}
	// Parse AI response
	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "Failed to parse AI API response")
		return
	}

	// Extract the response content
	if aiResp.Message.Content == "" {
		s.sendErrorResponse(w, http.StatusInternalServerError, "No response from AI")
		return
	}

	aiResponse := aiResp.Message.Content

	// Add AI response to session
	s.addMessageToSession(session.ID, "assistant", aiResponse)

	// Send response
	response := ChatResponse{
		Response: aiResponse,
		Session:  session.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHome serves the chat interface
func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Println("handleHome")
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.template == nil {
		http.Error(w, "Template not loaded", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
	}{
		Title: "AI会议助手",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := s.template.Execute(w, data)
	if err != nil {
		fmt.Printf("Warning: Could not execute template: %v\n", err)
	}
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleStatic serves static files
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Remove /static/ prefix
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	if path == "" {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Serve file from static directory
	filePath := filepath.Join("static", path)
	http.ServeFile(w, r, filePath)
}

// findAvailablePort finds an available port starting from the given port
func findAvailablePort(startPort int) (int, error) {
	for port := startPort; port < startPort+100; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port found in range %d-%d", startPort, startPort+99)
}

func main_test() {
	server := NewServer()

	// Set up routes
	http.HandleFunc("/", server.handleHome)
	http.HandleFunc("/api/chat", server.handleChat)
	http.HandleFunc("/health", server.handleHealth)
	http.HandleFunc("/static/", server.handleStatic)

	// Find available port
	port, err := findAvailablePort(8082)
	if err != nil {
		fmt.Printf("❌ Error finding available port: %v\n", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("🚀 启动AI会议助手服务器...\n")
	fmt.Printf("✅ 服务器启动成功!\n")
	fmt.Printf("📱 请在浏览器中访问: http://localhost:%d\n", port)
	fmt.Printf("🔧 API端点: http://localhost:%d/api/chat\n", port)
	fmt.Printf("💚 健康检查: http://localhost:%d/health\n", port)
	fmt.Printf("📁 静态文件: http://localhost:%d/static/\n", port)
	fmt.Printf("\n按 Ctrl+C 停止服务器\n\n")

	// Start server
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("❌ 服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}
