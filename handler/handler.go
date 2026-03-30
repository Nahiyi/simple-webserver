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
	disposition := getContentDisposition(filePath)

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
		"Content-Type: %s\r\n"+
		"Content-Length: %d\r\n"+
		"Content-Disposition: %s\r\n"+
		"Connection: close\r\n"+
		"\r\n", contentType, len(content), disposition)

	conn.Write([]byte(response))
	conn.Write(content)
}

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
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".pdf":
		return "application/pdf"
	case ".xml":
		return "application/xml"
	case ".md":
		return "text/plain; charset=utf-8"
	case ".csv":
		return "text/csv; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

// getContentDisposition 返回 Content-Disposition 头
// inline: 浏览器直接显示
// attachment: 下载
func getContentDisposition(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	// 浏览器可以直接显示的文件类型
	inlineExtensions := []string{
		".html", ".htm", ".css", ".js", ".json",
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".bmp", ".ico",
		".txt", ".pdf", ".xml", ".md", ".csv",
	}

	for _, inlineExt := range inlineExtensions {
		if ext == inlineExt {
			// 提取文件名用于 inline 显示
			filename := filepath.Base(filePath)
			return fmt.Sprintf("inline; filename=\"%s\"", filename)
		}
	}

	// 其他文件类型强制下载
	filename := filepath.Base(filePath)
	return fmt.Sprintf("attachment; filename=\"%s\"", filename)
}
