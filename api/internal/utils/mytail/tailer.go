package mytail

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

// LogTailer 日志监控器
type LogTailer struct {
	filePath string
	file     *os.File
	offset   int64
	watcher  *fsnotify.Watcher
	clients  map[*websocket.Conn]bool
	mutex    sync.RWMutex
	upgrader websocket.Upgrader
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Raw       string      `json:"raw"`
}

// NewLogTailer 创建新的日志监控器
func NewLogTailer(filePath string) (*LogTailer, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	t := &LogTailer{
		filePath: filePath,
		watcher:  watcher,
		clients:  make(map[*websocket.Conn]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许跨域
			},
		},
	}

	return t, nil
}

// Start 开始监控日志文件
func (t *LogTailer) Start(ctx context.Context) error {
	// 确保日志文件存在
	if _, err := os.Stat(t.filePath); os.IsNotExist(err) {
		// 创建日志文件目录
		dir := filepath.Dir(t.filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %v", err)
		}
		// 创建空的日志文件
		file, err := os.Create(t.filePath)
		if err != nil {
			return fmt.Errorf("创建日志文件失败: %v", err)
		}
		file.Close()
	}

	// 打开日志文件
	file, err := os.Open(t.filePath)
	if err != nil {
		return err
	}
	t.file = file

	// 移动到文件末尾
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	t.offset = stat.Size()
	file.Seek(t.offset, io.SeekStart)

	// 监控文件变化
	err = t.watcher.Add(filepath.Dir(t.filePath))
	if err != nil {
		return err
	}

	go t.watchFile(ctx)
	return nil
}

// watchFile 监控文件变化
func (t *LogTailer) watchFile(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-t.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				if strings.Contains(event.Name, filepath.Base(t.filePath)) {
					t.readNewLines()
				}
			}
		case err, ok := <-t.watcher.Errors:
			if !ok {
				return
			}
			logx.Infof("文件监控错误: %v", err)
		}
	}
}

// readNewLines 读取新的日志行
func (t *LogTailer) readNewLines() {
	stat, err := t.file.Stat()
	if err != nil {
		return
	}

	if stat.Size() < t.offset {
		// 文件被截断，重新打开
		t.file.Close()
		file, err := os.Open(t.filePath)
		if err != nil {
			return
		}
		t.file = file
		t.offset = 0
	}

	t.file.Seek(t.offset, io.SeekStart)
	scanner := bufio.NewScanner(t.file)

	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			entry := t.parseLogLine(line)
			t.broadcastToClients(entry)
		}
	}

	t.offset, _ = t.file.Seek(0, io.SeekCurrent)
}

// parseLogLine 解析日志行
func (t *LogTailer) parseLogLine(line string) *LogEntry {
	entry := &LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Raw:       line,
	}

	// 尝试解析JSON格式的日志
	if strings.HasPrefix(strings.TrimSpace(line), "{") {
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonData); err == nil {
			// 提取常见字段
			if ts, ok := jsonData["timestamp"].(string); ok {
				entry.Timestamp = ts
			} else if ts, ok := jsonData["time"].(string); ok {
				entry.Timestamp = ts
			} else if ts, ok := jsonData["@timestamp"].(string); ok {
				entry.Timestamp = ts
			}

			if level, ok := jsonData["level"].(string); ok {
				entry.Level = level
			} else if level, ok := jsonData["severity"].(string); ok {
				entry.Level = level
			}

			if msg, ok := jsonData["message"].(string); ok {
				entry.Message = msg
			} else if msg, ok := jsonData["msg"].(string); ok {
				entry.Message = msg
			}

			entry.Data = jsonData
		}
	} else {
		// 普通文本日志，尝试提取级别
		entry.Message = line
		lower := strings.ToLower(line)
		if strings.Contains(lower, "error") {
			entry.Level = "ERROR"
		} else if strings.Contains(lower, "warn") {
			entry.Level = "WARN"
		} else if strings.Contains(lower, "info") {
			entry.Level = "INFO"
		} else if strings.Contains(lower, "debug") {
			entry.Level = "DEBUG"
		} else {
			entry.Level = "INFO"
		}
	}

	return entry
}

// broadcastToClients 广播给所有客户端
func (t *LogTailer) broadcastToClients(entry *LogEntry) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	data, _ := json.Marshal(entry)
	for client := range t.clients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			client.Close()
			delete(t.clients, client)
		}
	}
}

// HandleWebSocket 处理WebSocket连接
func (t *LogTailer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := t.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Infof("WebSocket升级失败: %v", err)
		return
	}
	defer conn.Close()

	t.mutex.Lock()
	t.clients[conn] = true
	t.mutex.Unlock()

	logx.Infof("新的WebSocket客户端连接")

	// 发送最近的日志行
	t.sendRecentLines(conn)

	// 保持连接
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			logx.Infof("WebSocket连接断开: %v", err)
			break
		}
	}

	t.mutex.Lock()
	delete(t.clients, conn)
	t.mutex.Unlock()
}

// sendRecentLines 发送最近的日志行
func (t *LogTailer) sendRecentLines(conn *websocket.Conn) {
	// 读取最后100行
	file, err := os.Open(t.filePath)
	if err != nil {
		return
	}
	defer file.Close()

	lines := make([]string, 0, 100)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > 100 {
			lines = lines[1:]
		}
	}

	// 发送最近的行
	for _, line := range lines {
		if line != "" {
			entry := t.parseLogLine(line)
			data, _ := json.Marshal(entry)
			conn.WriteMessage(websocket.TextMessage, data)
		}
	}
}

// HandleLogPage 处理日志查看页面
func (t *LogTailer) HandleLogPage(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>实时日志查看器</title>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: 'Courier New', monospace;
            margin: 0;
            padding: 20px;
            background-color: #1e1e1e;
            color: #d4d4d4;
        }
        .header {
            background-color: #2d2d30;
            padding: 15px;
            border-radius: 5px;
            margin-bottom: 20px;
        }
        .log-container {
            height: 70vh;
            overflow-y: auto;
            background-color: #0d1117;
            border: 1px solid #30363d;
            border-radius: 5px;
            padding: 10px;
        }
        .log-entry {
            margin-bottom: 5px;
            padding: 5px;
            border-left: 3px solid #666;
            background-color: #161b22;
        }
        .log-entry.ERROR {
            border-left-color: #f85149;
            background-color: #2d1b1b;
        }
        .log-entry.WARN {
            border-left-color: #f0ad4e;
            background-color: #2d2a1b;
        }
        .log-entry.INFO {
            border-left-color: #5bc0de;
            background-color: #1b2a2d;
        }
        .log-entry.DEBUG {
            border-left-color: #999;
            background-color: #1e1e1e;
        }
        .timestamp {
            color: #7d8590;
            font-size: 0.9em;
        }
        .level {
            font-weight: bold;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 0.8em;
        }
        .level.ERROR { background-color: #f85149; color: white; }
        .level.WARN { background-color: #f0ad4e; color: white; }
        .level.INFO { background-color: #5bc0de; color: white; }
        .level.DEBUG { background-color: #999; color: white; }
        .message {
            margin-top: 5px;
        }
        .json-data {
            background-color: #0d1117;
            border: 1px solid #30363d;
            border-radius: 3px;
            padding: 10px;
            margin-top: 5px;
            font-size: 0.9em;
            overflow-x: auto;
        }
        .controls {
            margin-bottom: 20px;
        }
        .btn {
            background-color: #238636;
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 3px;
            cursor: pointer;
            margin-right: 10px;
        }
        .btn:hover {
            background-color: #2ea043;
        }
        .status {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 3px;
            font-size: 0.9em;
        }
        .status.connected {
            background-color: #238636;
            color: white;
        }
        .status.disconnected {
            background-color: #da3633;
            color: white;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>实时日志查看器</h1>
        <div class="controls">
            <button class="btn" onclick="clearLogs()">清空日志</button>
            <button class="btn" onclick="toggleAutoScroll()">自动滚动: <span id="autoScrollStatus">开启</span></button>
            <span class="status" id="connectionStatus">连接中...</span>
        </div>
    </div>
    <div class="log-container" id="logContainer"></div>

    <script>
        let ws;
        let autoScroll = true;
        const logContainer = document.getElementById('logContainer');
        const connectionStatus = document.getElementById('connectionStatus');
        const autoScrollStatus = document.getElementById('autoScrollStatus');

        function connect() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = protocol + '//' + window.location.host + '/api/v1/mytail/ws';
            
            ws = new WebSocket(wsUrl);
            
            ws.onopen = function() {
                connectionStatus.textContent = '已连接';
                connectionStatus.className = 'status connected';
            };
            
            ws.onmessage = function(event) {
                const entry = JSON.parse(event.data);
                addLogEntry(entry);
            };
            
            ws.onclose = function() {
                connectionStatus.textContent = '连接断开';
                connectionStatus.className = 'status disconnected';
                // 3秒后重连
                setTimeout(connect, 3000);
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket错误:', error);
                connectionStatus.textContent = '连接错误';
                connectionStatus.className = 'status disconnected';
            };
        }

        function addLogEntry(entry) {
            const div = document.createElement('div');
            div.className = 'log-entry ' + (entry.level || 'INFO');
            
            let html = '<div class="timestamp">' + entry.timestamp + '</div>';
            if (entry.level) {
                html += '<span class="level ' + entry.level + '">' + entry.level + '</span> ';
            }
            html += '<div class="message">' + escapeHtml(entry.message || entry.raw) + '</div>';
            
            if (entry.data && typeof entry.data === 'object') {
                html += '<div class="json-data"><pre>' + JSON.stringify(entry.data, null, 2) + '</pre></div>';
            }
            
            div.innerHTML = html;
            logContainer.appendChild(div);
            
            // 限制日志条数
            while (logContainer.children.length > 1000) {
                logContainer.removeChild(logContainer.firstChild);
            }
            
            if (autoScroll) {
                logContainer.scrollTop = logContainer.scrollHeight;
            }
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function clearLogs() {
            logContainer.innerHTML = '';
        }

        function toggleAutoScroll() {
            autoScroll = !autoScroll;
            autoScrollStatus.textContent = autoScroll ? '开启' : '关闭';
        }

        // 启动连接
        connect();
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// RegisterHandlers 注册HTTP处理器到现有的服务器
func (t *LogTailer) RegisterHandlers(server interface{}) {
	// 这里需要根据实际的服务器类型来注册路由
	// 由于go-zero的特殊性，我们需要在主程序中手动注册
}

// Close 关闭监控器
func (t *LogTailer) Close() {
	if t.file != nil {
		t.file.Close()
	}
	if t.watcher != nil {
		t.watcher.Close()
	}

	t.mutex.Lock()
	for client := range t.clients {
		client.Close()
	}
	t.mutex.Unlock()
}
