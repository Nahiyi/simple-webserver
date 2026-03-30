package main

import (
	"fmt"
	"os"

	"simple-webserver/config"
	"simple-webserver/logger"
	"simple-webserver/router"
)

func main() {
	// 初始化日志系统
	logger.Init()

	// 初始化配置
	cfg := config.GetInstance()
	_ = cfg

	fmt.Println("========================================")
	fmt.Println("       简单Web服务器")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("管理控制面板: http://localhost:8081/static/")
	fmt.Println()
	fmt.Println("初始配置:")
	fmt.Printf("  - Web服务端口: %d\n", config.GetPort())
	fmt.Printf("  - 管理端口: %d\n", config.GetAdminPort())
	fmt.Printf("  - 资源目录: %s\n", config.GetRootDir())
	fmt.Println()
	fmt.Println("提示: 请确保 /root 目录存在并包含要服务的文件")
	fmt.Println()
	fmt.Println("按 Ctrl+C 退出")
	fmt.Println("========================================")
	fmt.Println()

	// 确保/root目录存在(仅用于演示)
	if _, err := os.Stat("/root"); os.IsNotExist(err) {
		fmt.Println("警告: /root 目录不存在，将创建...")
		if err := os.MkdirAll("/root", 0755); err != nil {
			fmt.Printf("创建目录失败: %v\n", err)
		} else {
			fmt.Println("/root 目录已创建")
		}
	}

	// 启动管理服务器(包含Web控制面板)
	if err := router.StartAdminServer(); err != nil {
		fmt.Printf("管理服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}
