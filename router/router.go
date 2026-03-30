package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"simple-webserver/config"
	"simple-webserver/logger"
	"simple-webserver/server"
)

// StartAdminServer 启动管理服务器
func StartAdminServer() error {
	port := config.GetAdminPort()
	addr := fmt.Sprintf(":%d", port)

	mux := http.NewServeMux()

	// 向mux注册路由和对应的处理器

	// 静态文件服务（控制面板HTML）
	mux.HandleFunc("/static/", serveStatic)

	// API路由
	mux.HandleFunc("/api/status", apiStatus)
	mux.HandleFunc("/api/start", apiStart)
	mux.HandleFunc("/api/stop", apiStop)
	mux.HandleFunc("/api/config", apiConfig)
	mux.HandleFunc("/api/logs", apiLogs)
	mux.HandleFunc("/api/logs/clear", apiLogsClear)

	// 首页重定向到控制面板
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/static/index.html", 302)
			return
		}
		http.NotFound(w, r)
	})

	fmt.Printf("管理控制面板已启动: http://localhost:%d/static/\n", port)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}

func serveStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/static/"):]
	if path == "" {
		path = "index.html"
	}

	filePath := "./static/" + path
	// 读取本地文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
		} else {
			http.Error(w, "Internal Server Error", 500)
		}
		return
	}

	// 设置Content-Type
	contentType := config.GetContentType(path)
	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

// APIResponse API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func apiStatus(w http.ResponseWriter, r *http.Request) {
	resp := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"running":     server.IsRunning(),
			"port":        config.GetPort(),
			"adminPort":   config.GetAdminPort(),
			"rootDir":     config.GetRootDir(),
			"connections": server.GetConnections(),
			"logCount":    logger.GetCount(),
		},
	}
	writeJSON(w, 200, resp)
}

func apiStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, APIResponse{Success: false, Message: "Method not allowed"})
		return
	}

	err := server.Start()
	if err != nil {
		writeJSON(w, 400, APIResponse{Success: false, Message: err.Error()})
		return
	}
	writeJSON(w, 200, APIResponse{Success: true, Message: "服务器已启动"})
}

func apiStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, APIResponse{Success: false, Message: "Method not allowed"})
		return
	}

	err := server.Stop()
	if err != nil {
		writeJSON(w, 400, APIResponse{Success: false, Message: err.Error()})
		return
	}
	writeJSON(w, 200, APIResponse{Success: true, Message: "服务器已停止"})
}

func apiConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 获取配置
		resp := APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"port":      config.GetPort(),
				"adminPort": config.GetAdminPort(),
				"rootDir":   config.GetRootDir(),
			},
		}
		writeJSON(w, 200, resp)
		return
	}

	if r.Method == "POST" {
		// 设置配置
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, 400, APIResponse{Success: false, Message: "Invalid request body"})
			return
		}
		defer r.Body.Close()

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			writeJSON(w, 400, APIResponse{Success: false, Message: "Invalid JSON"})
			return
		}

		var errors []string

		// 处理端口配置
		if v, ok := data["port"]; ok {
			var port int
			switch val := v.(type) {
			case float64:
				port = int(val)
			case int:
				port = val
			case string:
				if val == "" {
					port = 0
				} else {
					port = 8080 // 默认
				}
			}
			if port < 1 || port > 65535 {
				errors = append(errors, "端口号无效 (1-65535)")
			} else {
				config.SetPort(port)
			}
		}

		// 处理根目录配置
		if v, ok := data["rootDir"]; ok {
			if rootDir, ok := v.(string); ok {
				if rootDir != "" {
					// 目录不存在时自动创建
					if _, err := os.Stat(rootDir); os.IsNotExist(err) {
						if err := os.MkdirAll(rootDir, 0755); err != nil {
							errors = append(errors, "无法创建目录: "+rootDir)
						}
					}
					config.SetRootDir(rootDir)
				}
			}
		}

		if len(errors) > 0 {
			writeJSON(w, 400, APIResponse{Success: false, Message: errors[0]})
			return
		}

		resp := APIResponse{
			Success: true,
			Message: "配置已更新",
			Data: map[string]interface{}{
				"port":    config.GetPort(),
				"rootDir": config.GetRootDir(),
			},
		}
		writeJSON(w, 200, resp)
		return
	}

	writeJSON(w, 405, APIResponse{Success: false, Message: "Method not allowed"})
}

func apiLogs(w http.ResponseWriter, r *http.Request) {
	logs := logger.GetRecent(100)
	resp := APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"count": len(logs),
			"logs":  logs,
		},
	}
	writeJSON(w, 200, resp)
}

func apiLogsClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, APIResponse{Success: false, Message: "Method not allowed"})
		return
	}

	logger.Clear()
	writeJSON(w, 200, APIResponse{Success: true, Message: "日志已清空"})
}
