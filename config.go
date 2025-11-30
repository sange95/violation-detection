package main

import (
	"fmt"
	"os"
)

// Config 配置结构体
type Config struct {
	// 文件服务器配置
	FileServerPort string
	FileServerPath string // 文件存储路径

	// API服务器配置
	APIServerPort string

	// VLLM配置
	VLLMBaseURL string // VLLM服务地址，例如: http://localhost:8001
	VLLMModel   string // 模型名称（可选）
	VLLMPrompt  string // 提示词（可选）
}

// LoadConfig 从环境变量加载配置
func LoadConfig() *Config {
	config := &Config{
		// 文件服务器配置
		FileServerPort: getEnv("FILE_SERVER_PORT", "8080"),
		FileServerPath: getEnv("FILE_SERVER_PATH", "./uploads"),

		// API服务器配置
		APIServerPort: getEnv("API_SERVER_PORT", "8081"),

		// VLLM配置
		VLLMBaseURL: getEnv("VLLM_BASE_URL", "http://localhost:8001"),
		VLLMModel:   getEnv("VLLM_MODEL", ""),
		VLLMPrompt:  getEnv("VLLM_PROMPT", "请检测该图片中内容。"),
	}

	return config
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.FileServerPort == "" {
		return fmt.Errorf("FILE_SERVER_PORT 环境变量未设置")
	}
	if c.APIServerPort == "" {
		return fmt.Errorf("API_SERVER_PORT 环境变量未设置")
	}
	if c.VLLMBaseURL == "" {
		return fmt.Errorf("VLLM_BASE_URL 环境变量未设置")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

