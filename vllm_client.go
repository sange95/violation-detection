package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// VLLMClient VLLM客户端
type VLLMClient struct {
	baseURL string
	model   string
	prompt  string
	client  *http.Client
}

// NewVLLMClient 创建VLLM客户端
func NewVLLMClient(baseURL, model, prompt string) *VLLMClient {
	return &VLLMClient{
		baseURL: baseURL,
		model:   model,
		prompt:  prompt,
		client: &http.Client{
			Timeout: 60 * time.Second, // 推理可能需要较长时间
		},
	}
}

// ChatMessage 聊天消息
type ChatMessage struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

// ImageURL 图片URL结构
type ImageURL struct {
	Type     string `json:"type"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

// TextContent 文本内容结构
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model    string         `json:"model"`
	Messages []ChatMessage  `json:"messages"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// InferenceRequest 推理请求（简化版，用于返回给调用者）
type InferenceRequest struct {
	ImageURL string `json:"image_url"`
	Prompt   string `json:"prompt,omitempty"`
}

// InferenceResponse 推理响应（简化版，用于返回给调用者）
type InferenceResponse struct {
	Success   bool   `json:"success"`
	Content   string `json:"content,omitempty"`
	Error     string `json:"error,omitempty"`
	Usage     *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
}

// InferImage 推理图片
func (c *VLLMClient) InferImage(imageURL string, customPrompt string) (*InferenceResponse, error) {
	// 使用自定义提示词或默认提示词
	prompt := customPrompt
	if prompt == "" {
		prompt = c.prompt
	}

	// 构建请求消息
	messageContent := []interface{}{
		ImageURL{
			Type: "image_url",
			ImageURL: struct {
				URL string `json:"url"`
			}{URL: imageURL},
		},
		TextContent{
			Type: "text",
			Text: prompt,
		},
	}

	// 构建请求
	reqBody := ChatRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: messageContent,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return &InferenceResponse{
			Success: false,
			Error:   fmt.Sprintf("序列化请求失败: %v", err),
		}, err
	}

	// 发送请求
	url := fmt.Sprintf("%s/v1/chat/completions", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return &InferenceResponse{
			Success: false,
			Error:   fmt.Sprintf("创建请求失败: %v", err),
		}, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return &InferenceResponse{
			Success: false,
			Error:   fmt.Sprintf("请求失败: %v", err),
		}, err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &InferenceResponse{
			Success: false,
			Error:   fmt.Sprintf("读取响应失败: %v", err),
		}, err
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &InferenceResponse{
			Success: false,
			Error:   fmt.Sprintf("VLLM服务返回错误: %d, %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return &InferenceResponse{
			Success: false,
			Error:   fmt.Sprintf("解析响应失败: %v, 原始响应: %s", err, string(body)),
		}, err
	}

	// 提取内容
	if len(chatResp.Choices) == 0 {
		return &InferenceResponse{
			Success: false,
			Error:   "响应中没有choices",
		}, fmt.Errorf("响应中没有choices")
	}

	content := chatResp.Choices[0].Message.Content

	return &InferenceResponse{
		Success: true,
		Content: content,
		Usage: &struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     chatResp.Usage.PromptTokens,
			CompletionTokens: chatResp.Usage.CompletionTokens,
			TotalTokens:      chatResp.Usage.TotalTokens,
		},
	}, nil
}

