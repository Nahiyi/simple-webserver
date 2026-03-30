package logger

import (
	"fmt"
	"sync"
	"time"
)

// AccessLog 访问日志结构
type AccessLog struct {
	ClientIP string
	Method   string
	URL      string
	Status   int
	Time     string
}

var (
	logs    []AccessLog
	logsMu  sync.Mutex
	logChan chan AccessLog
)

// Init 初始化日志系统
func Init() {
	logChan = make(chan AccessLog, 100)
	go processLogs()
}

// processLogs 处理日志
func processLogs() {
	for log := range logChan {
		logsMu.Lock()
		logs = append(logs, log)
		// 最多保留1000条
		if len(logs) > 1000 {
			logs = logs[len(logs)-1000:]
		}
		logsMu.Unlock()
		// 打印到控制台
		fmt.Printf("[%s] %s %s -> %d\n", log.Time, log.Method, log.URL, log.Status)
	}
}

// Add 添加日志
func Add(clientIP, method, url string, status int) {
	log := AccessLog{
		ClientIP: clientIP,
		Method:   method,
		URL:      url,
		Status:   status,
		Time:     time.Now().Format("2006-01-02 15:04:05"),
	}
	select {
	case logChan <- log:
	default:
	}
}

// GetAll 获取所有日志
func GetAll() []AccessLog {
	logsMu.Lock()
	defer logsMu.Unlock()
	result := make([]AccessLog, len(logs))
	copy(result, logs)
	return result
}

// GetRecent 获取最近N条日志
func GetRecent(n int) []AccessLog {
	logsMu.Lock()
	defer logsMu.Unlock()
	if n > len(logs) {
		n = len(logs)
	}
	result := make([]AccessLog, n)
	copy(result, logs[len(logs)-n:])
	return result
}

// GetCount 获取日志总数
func GetCount() int {
	logsMu.Lock()
	defer logsMu.Unlock()
	return len(logs)
}

// Clear 清空日志
func Clear() {
	logsMu.Lock()
	defer logsMu.Unlock()
	logs = logs[:0]
}
