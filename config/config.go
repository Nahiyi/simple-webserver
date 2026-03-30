package config

import (
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
