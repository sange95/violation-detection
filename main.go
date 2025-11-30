package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"sync"
)

func main() {
	log.Println("VLLM Show 服务启动中...")

	// 加载配置
	config := LoadConfig()
	if err := config.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	// 创建VLLM客户端
	vllmClient := NewVLLMClient(config.VLLMBaseURL, config.VLLMModel, config.VLLMPrompt)

	// 创建文件服务器
	fileServer := NewFileServer(config.FileServerPort, config.FileServerPath)

	// 获取文件服务器主机地址（用于生成文件URL）
	fileServerHost := os.Getenv("FILE_SERVER_HOST")
	if fileServerHost == "" {
		fileServerHost = "localhost:" + config.FileServerPort
	}

	// 创建API服务器
	apiServer := NewAPIServer(
		config.APIServerPort,
		config.FileServerPath,
		fileServer,
		vllmClient,
		fileServerHost,
	)

	// 使用WaitGroup等待两个服务
	var wg sync.WaitGroup
	wg.Add(2)

	// 启动文件服务器（协程）
	go func() {
		defer wg.Done()
		if err := fileServer.Start(); err != nil {
			log.Fatalf("文件服务器启动失败: %v", err)
		}
	}()

	// 启动API服务器（协程）
	go func() {
		defer wg.Done()
		if err := apiServer.Start(); err != nil {
			log.Fatalf("API服务器启动失败: %v", err)
		}
	}()

	log.Println("所有服务已启动")
	log.Printf("文件服务器: http://localhost:%s", config.FileServerPort)
	log.Printf("API服务器: http://localhost:%s", config.APIServerPort)
	log.Printf("上传接口: http://localhost:%s/api/upload", config.APIServerPort)

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("收到停止信号，正在关闭服务...")
	// 服务会在收到信号后自动关闭
}

