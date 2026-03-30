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
	logs    []AccessLog    // 日志切片，存储所有的日志记录
	logsMu  sync.Mutex     // 互斥锁，用来对上述切片进行并发安全读写
	logChan chan AccessLog // 日志发送、消费管道，经典的生产者-消费者模型
)

// Init 初始化日志系统
func Init() {
	logChan = make(chan AccessLog, 100)
	go processLogs()
}

// processLogs 处理日志
func processLogs() {
	// 基于channel控制日志处理：从ch中读取日志
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
	// 非阻塞发送常见模式
	select {
	case logChan <- log: // 日志通道可写，则发送日志到通道
	default: // 通道已满，则丢弃日志，不阻塞
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
	// 返回副本
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
