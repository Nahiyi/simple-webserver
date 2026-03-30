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

	cancel()

	listener.Close()
	isRunning.Store(false)

	fmt.Println("Web服务器已停止")
	return nil
}

// run 运行服务器主循环
func run() {
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		connections.Add(1)

		go func() {
			defer connections.Add(-1)
			handler.HandleConnection(conn)
		}()
	}
}
