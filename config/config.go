package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// Config 服务器配置
type Config struct {
	Port      int    // 服务器监听端口
	AdminPort int    // 管理界面端口
	RootDir   string // 资源文件根目录
}

var (
	cfg  *Config
	once sync.Once
	mu   sync.RWMutex
)

// GetInstance 获取单例配置实例
func GetInstance() *Config {
	once.Do(func() {
		cfg = &Config{
			Port:      8080,
			AdminPort: 8081,
			RootDir:   "/root",
		}
	})
	return cfg
}

// GetPort 获取服务器端口
func GetPort() int {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.Port
}

// SetPort 设置服务器端口
func SetPort(port int) {
	mu.Lock()
	defer mu.Unlock()
	cfg.Port = port
}

// GetAdminPort 获取管理端口
func GetAdminPort() int {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.AdminPort
}

// SetAdminPort 设置管理端口
func SetAdminPort(port int) {
	mu.Lock()
	defer mu.Unlock()
	cfg.AdminPort = port
}

// GetRootDir 获取根目录
func GetRootDir() string {
	mu.RLock()
	defer mu.RUnlock()
	return cfg.RootDir
}

// SetRootDir 设置根目录
func SetRootDir(dir string) {
	mu.Lock()
	defer mu.Unlock()
	cfg.RootDir = dir
}

// GetContentType 根据文件扩展名返回MIME类型
func GetContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
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
	case ".txt", ".md", ".csv":
		return "text/plain; charset=utf-8"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".pdf":
		return "application/pdf"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}

// GetContentDisposition 返回Content-Disposition头
// inline: 浏览器直接显示，attachment: 下载
func GetContentDisposition(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	inlineExtensions := []string{
		".html", ".htm", ".css", ".js", ".json",
		".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".bmp", ".ico",
		".txt", ".pdf", ".xml", ".md", ".csv",
	}
	for _, inlineExt := range inlineExtensions {
		if ext == inlineExt {
			return fmt.Sprintf("inline; filename=\"%s\"", filepath.Base(filePath))
		}
	}
	return fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(filePath))
}
