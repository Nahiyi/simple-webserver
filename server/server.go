package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"simple-webserver/config"
	"simple-webserver/handler"
)

var (
	isRunning   bool
	stopChan    chan struct{}
	listener    net.Listener
	mu          sync.RWMutex
	connections int32 // 当前连接数
)

// IsRunning 返回服务器运行状态
func IsRunning() bool {
	mu.RLock()
	defer mu.RUnlock()
	return isRunning
}

// GetConnections 返回当前连接数
func GetConnections() int32 {
	return connections
}

// Start 启动服务器
func Start() error {
	mu.Lock()
	if isRunning {
		mu.Unlock()
		return fmt.Errorf("服务器已经在运行")
	}

	stopChan = make(chan struct{})
	port := config.GetPort()

	addr := fmt.Sprintf(":%d", port)
	var err error
	listener, err = net.Listen("tcp", addr)
	if err != nil {
		mu.Unlock()
		return fmt.Errorf("启动失败: %v", err)
	}

	isRunning = true
	mu.Unlock()

	fmt.Printf("Web服务器已启动，监听端口: %d\n", port)
	fmt.Printf("资源根目录: %s\n", config.GetRootDir())

	go run()

	return nil
}

// Stop 停止服务器
func Stop() error {
	mu.Lock()
	if !isRunning {
		mu.Unlock()
		return fmt.Errorf("服务器未运行")
	}
	mu.Unlock()

	close(stopChan)

	mu.Lock()
	listener.Close()
	isRunning = false
	mu.Unlock()

	fmt.Println("Web服务器已停止")
	return nil
}

// run 运行服务器主循环
func run() {
	for {
		conn, err := listener.Accept()
		if err != nil {
			mu.RLock()
			closed := stopChan == nil
			mu.RUnlock()
			if closed {
				return
			}
			log.Printf("接受连接失败: %v\n", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 增加连接计数
		atomicAdd(&connections, 1)

		// 每个客户端连接由独立goroutine处理
		go func() {
			defer func() {
				atomicAdd(&connections, -1)
				handler.HandleConnection(conn)
			}()
		}()
	}
}

// atomicAdd 原子增加
func atomicAdd(val *int32, delta int32) {
	mu.Lock()
	*val += delta
	mu.Unlock()
}
