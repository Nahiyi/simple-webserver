package handler

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"simple-webserver/config"
	"simple-webserver/logger"
)

// HandleConnection 处理客户端连接
func HandleConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	parts := strings.Split(strings.TrimSpace(line), " ")
	if len(parts) < 3 {
		sendErrorResponse(conn, 400, "Bad Request")
		return
	}

	method := parts[0]
	url := parts[1]
	clientIP := conn.RemoteAddr().String()

	// 解析头部
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil || strings.TrimSpace(line) == "" {
			break
		}
		if idx := strings.Index(line, ":"); idx != -1 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			headers[key] = value
		}
	}

	// 处理请求
	rootDir := config.GetRootDir()

	if method != "GET" {
		logger.Add(clientIP, method, url, 501)
		sendErrorResponse(conn, 501, "Not Implemented")
		return
	}

	// 防止路径遍历
	if strings.Contains(url, "..") {
		logger.Add(clientIP, method, url, 403)
		sendErrorResponse(conn, 403, "Forbidden")
		return
	}

	// 解析URL获取文件路径
	if url == "/" {
		url = "/index.html"
	}
	filePath := filepath.Join(rootDir, filepath.FromSlash(url))

	// 安全检查：确保文件在根目录内
	absRoot, _ := filepath.Abs(rootDir)
	absPath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absPath, absRoot) {
		logger.Add(clientIP, method, url, 403)
		sendErrorResponse(conn, 403, "Forbidden")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Add(clientIP, method, url, 404)
		sendErrorResponse(conn, 404, "Not Found")
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		logger.Add(clientIP, method, url, 500)
		sendErrorResponse(conn, 500, "Internal Server Error")
		return
	}

	logger.Add(clientIP, method, url, 200)

	// 发送成功响应
	contentType := getContentType(filePath)
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
		"Content-Type: %s\r\n"+
		"Content-Length: %d\r\n"+
		"Connection: close\r\n"+
		"\r\n", contentType, len(content))

	conn.Write([]byte(response))
	conn.Write(content)
}

func sendErrorResponse(conn net.Conn, statusCode int, statusMsg string) {
	var body string
	switch statusCode {
	case 404:
		body = `<!DOCTYPE html>
<html>
<head><title>404 Not Found</title></head>
<body>
<h1>404 Not Found</h1>
<p>The requested file was not found on this server.</p>
<hr>
<small>Simple Web Server</small>
</body>
</html>`
	case 501:
		body = `<!DOCTYPE html>
<html>
<head><title>501 Not Implemented</title></head>
<body>
<h1>501 Not Implemented</h1>
<p>The requested method is not implemented by this server.</p>
<hr>
<small>Simple Web Server</small>
</body>
</html>`
	case 403:
		body = `<!DOCTYPE html>
<html>
<head><title>403 Forbidden</title></head>
<body>
<h1>403 Forbidden</h1>
<p>Access denied.</p>
<hr>
<small>Simple Web Server</small>
</body>
</html>`
	case 400:
		body = `<!DOCTYPE html>
<html>
<head><title>400 Bad Request</title></head>
<body>
<h1>400 Bad Request</h1>
<p>Invalid request.</p>
<hr>
<small>Simple Web Server</small>
</body>
</html>`
	default:
		body = `<!DOCTYPE html>
<html>
<head><title>500 Internal Server Error</title></head>
<body>
<h1>500 Internal Server Error</h1>
<p>Internal server error.</p>
<hr>
<small>Simple Web Server</small>
</body>
</html>`
	}

	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n"+
		"Content-Type: text/html\r\n"+
		"Content-Length: %d\r\n"+
		"Connection: close\r\n"+
		"\r\n%s", statusCode, statusMsg, len(body), body)

	conn.Write([]byte(response))
}

func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
