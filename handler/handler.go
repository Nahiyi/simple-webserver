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
// 处理流程：解析请求 -> 安全检查 -> 文件服务 -> 记录日志
func HandleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(30 * time.Second)) // 30秒读超时

	// 解析请求行（格式：METHOD URL HTTP/VERSION）
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
	method, url := parts[0], parts[1]
	clientIP := conn.RemoteAddr().String()

	// 解析请求头部（逐行读取直到空行）
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

	rootDir := config.GetRootDir()

	// 方法检查：仅支持GET
	if method != "GET" {
		logger.Add(clientIP, method, url, 501)
		sendErrorResponse(conn, 501, "Not Implemented")
		return
	}

	// 安全检查：防止路径遍历（检测".."）
	if strings.Contains(url, "..") {
		logger.Add(clientIP, method, url, 403)
		sendErrorResponse(conn, 403, "Forbidden")
		return
	}

	// 解析URL为文件路径
	if url == "/" {
		url = "/index.html"
	}
	filePath := filepath.Join(rootDir, filepath.FromSlash(url))

	// 二次安全检查：确保文件在rootDir内（绝对路径比对）
	absRoot, _ := filepath.Abs(rootDir)
	absPath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absPath, absRoot) {
		logger.Add(clientIP, method, url, 403)
		sendErrorResponse(conn, 403, "Forbidden")
		return
	}

	// 文件存在性检查
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Add(clientIP, method, url, 404)
		sendErrorResponse(conn, 404, "Not Found")
		return
	}

	// 读取并发送文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		logger.Add(clientIP, method, url, 500)
		sendErrorResponse(conn, 500, "Internal Server Error")
		return
	}

	logger.Add(clientIP, method, url, 200)

	// 构造HTTP响应：状态行 + 头部 + 空行 + body
	contentType := config.GetContentType(filePath)
	disposition := config.GetContentDisposition(filePath)
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
		"Content-Type: %s\r\n"+
		"Content-Length: %d\r\n"+
		"Content-Disposition: %s\r\n"+
		"Connection: close\r\n"+
		"\r\n", contentType, len(content), disposition)

	conn.Write([]byte(response))
	conn.Write(content)
}

// sendErrorResponse 发送错误响应（带美化HTML页面）
func sendErrorResponse(conn net.Conn, statusCode int, statusMsg string) {
	var title, message string
	var bgColor, textColor string

	switch statusCode {
	case 404:
		title = "404"
		message = "抱歉，您访问的页面不存在"
		bgColor = "#fff3cd"
		textColor = "#856404"
	case 501:
		title = "501"
		message = "抱歉，请求方法暂未实现"
		bgColor = "#f8d7da"
		textColor = "#721c24"
	case 403:
		title = "403"
		message = "抱歉，您没有访问权限"
		bgColor = "#f8d7da"
		textColor = "#721c24"
	case 400:
		title = "400"
		message = "抱歉，请求格式错误"
		bgColor = "#fff3cd"
		textColor = "#856404"
	default:
		title = "500"
		message = "抱歉，服务器内部错误"
		bgColor = "#f8d7da"
		textColor = "#721c24"
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%d %s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #f5f7f9 0%%, #e4e7eb 100%%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .error-container {
            text-align: center;
            padding: 48px;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.1);
            max-width: 480px;
            width: 90%%;
        }
        .error-code {
            font-size: 96px;
            font-weight: 800;
            color: #18A058;
            line-height: 1;
            margin-bottom: 16px;
        }
        .error-title {
            font-size: 24px;
            font-weight: 600;
            color: #1a1a1a;
            margin-bottom: 12px;
        }
        .error-message {
            font-size: 16px;
            color: #666;
            margin-bottom: 32px;
        }
        .error-box {
            background: %s;
            color: %s;
            padding: 16px 24px;
            border-radius: 10px;
            display: inline-block;
        }
        .error-box .status {
            font-size: 20px;
            font-weight: 700;
        }
        .error-box .desc {
            font-size: 13px;
            margin-top: 4px;
        }
        .footer {
            margin-top: 32px;
            font-size: 13px;
            color: #999;
        }
    </style>
</head>
<body>
    <div class="error-container">
        <div class="error-code">%s</div>
        <h1 class="error-title">%s</h1>
        <p class="error-message">%s</p>
        <div class="error-box">
            <div class="status">%d %s</div>
            <div class="desc">Web服务器</div>
        </div>
        <div class="footer">请检查URL是否正确，或联系管理员</div>
    </div>
</body>
</html>`, statusCode, statusMsg, bgColor, textColor, title, title, message, statusCode, statusMsg)

	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n"+
		"Content-Type: text/html; charset=utf-8\r\n"+
		"Content-Length: %d\r\n"+
		"Connection: close\r\n"+
		"\r\n%s", statusCode, statusMsg, len(body), body)

	conn.Write([]byte(response))
}
