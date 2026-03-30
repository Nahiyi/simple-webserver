package server

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"

	"simple-webserver/config"
	"simple-webserver/handler"
)

var (
	isRunning   atomic.Bool        // 服务器运行状态（true=运行中，false=已停止）
	listener    net.Listener       // TCP监听器，接收客户端连接，Stop时关闭以中断Accept
	connections atomic.Int32       // 当前活跃连接数，用于监控
	ctx         context.Context    // 取消上下文，停止时触发Done()，用于优雅关闭主循环
	cancel      context.CancelFunc // 取消函数，调用时设置ctx.Done()
)

// IsRunning 返回服务器运行状态
func IsRunning() bool {
	return isRunning.Load()
}

// GetConnections 返回当前连接数
func GetConnections() int32 {
	return connections.Load()
}

// Start 启动Web服务器
// 流程：创建TCP监听 -> 初始化上下文 -> 启动主循环goroutine
func Start() error {
	// 防止重复启动（并发安全）
	if isRunning.Load() {
		return fmt.Errorf("服务器已经在运行")
	}

	// 获取配置的端口号，构造地址字符串（格式 ":8080"）
	port := config.GetPort()
	addr := fmt.Sprintf(":%d", port)

	// 创建TCP监听器
	var err error
	listener, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("启动失败: %v", err)
	}

	// 创建可取消的上下文，用于优雅关闭
	// ctx在cancel()被调用时，其Done()通道会关闭
	ctx, cancel = context.WithCancel(context.Background())

	// 设置运行状态为true（必须在线程启动前设置）
	isRunning.Store(true)

	fmt.Printf("Web服务器已启动，监听端口: %d\n", port)
	fmt.Printf("资源根目录: %s\n", config.GetRootDir())

	// 启动主循环（在独立goroutine中运行，不阻塞主线程）
	go run()

	return nil
}

// Stop 停止Web服务器
// 停止流程（三步顺序不能乱）：
//  1. isRunning.Store(false)  - 标记状态为停止
//  2. cancel()                 - 触发ctx.Done()，让主循环感知停止信号
//  3. listener.Close()         - 关闭监听器，中断Accept阻塞
func Stop() error {
	if !isRunning.Load() {
		return fmt.Errorf("服务器未运行")
	}

	// ① 先设置停止标志（可选，但提供双保险）
	isRunning.Store(false)

	// ② 取消上下文，触发ctx.Done()通道关闭
	//    run()中的select会立即感知到并退出循环
	cancel()

	// ③ 关闭TCP监听器
	//    这会导致listener.Accept()返回错误，从而退出循环
	listener.Close()

	fmt.Println("Web服务器已停止")
	return nil
}

// run 服务器主循环
// 核心设计：使用 select + context 实现优雅退出
// 优雅退出机制：
//
//	Stop()调用时：
//	1. cancel()触发ctx.Done()关闭
//	2. 下次Accept失败时，select命中ctx.Done()分支
//	3. run()函数返回，goroutine结束
func run() {
	for {
		// 阻塞等待客户端连接
		conn, err := listener.Accept()

		// 如果Accept返回错误（非阻塞返回）
		if err != nil {
			// 使用select检查：是因为Stop()取消的，还是其他错误
			select {
			case <-ctx.Done():
				// ctx已被取消（Stop()时调用了cancel()）
				// 这是停止信号，正常退出
				return

			default:
				// ctx未取消，说明是其他错误（如端口被占用）
				// 继续尝试Accept，不退出
				continue
			}
		}

		// Accept成功，增加连接计数
		connections.Add(1)

		// 启动独立goroutine处理该连接
		// defer确保连接关闭时减少计数
		go func() {
			defer connections.Add(-1) // 连接处理完毕，计数-1
			handler.HandleConnection(conn)
		}()
	}
}
