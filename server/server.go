package server

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"simple-webserver/config"
	"simple-webserver/handler"
)

var (
	isRunning   atomic.Bool
	listener    net.Listener
	connections atomic.Int32
	ctx         context.Context
	cancel      context.CancelFunc
)

// IsRunning 返回服务器运行状态
func IsRunning() bool {
	return isRunning.Load()
}

// GetConnections 返回当前连接数
func GetConnections() int32 {
	return connections.Load()
}

// Start 启动服务器
func Start() error {
	if isRunning.Load() {
		return fmt.Errorf("服务器已经在运行")
	}

	port := config.GetPort()
	addr := fmt.Sprintf(":%d", port)
	var err error
	listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("启动失败: %v", err)
	}

	ctx, cancel = context.WithCancel(context.Background())
	isRunning.Store(true)

	fmt.Printf("Web服务器已启动，监听端口: %d\n", port)
	fmt.Printf("资源根目录: %s\n", config.GetRootDir())

	go run()

	return nil
}

// Stop 停止服务器
func Stop() error {
	if !isRunning.Load() {
		return fmt.Errorf("服务器未运行")
	}

	// 先设置停止标志，再关闭listener
	// 这样Accept()返回错误时，isRunning已经是false，能正确退出
	isRunning.Store(false)
	cancel()
	listener.Close()

	fmt.Println("Web服务器已停止")
	return nil
}

// run 运行服务器主循环
func run() {
	for isRunning.Load() {
		conn, err := listener.Accept()
		if err != nil {
			// 如果服务器已停止，就退出循环
			if !isRunning.Load() {
				return
			}
			// 短暂等待后重试
			time.Sleep(100 * time.Millisecond)
			continue
		}

		connections.Add(1)

		go func() {
			defer connections.Add(-1)
			handler.HandleConnection(conn)
		}()
	}
}
