package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// APIServer API服务器
type APIServer struct {
	port         string
	filePath     string
	fileServer   *FileServer
	vllmClient   *VLLMClient
	fileServerHost string // 文件服务器的主机地址，用于生成文件URL
}

// NewAPIServer 创建API服务器
func NewAPIServer(port, filePath string, fileServer *FileServer, vllmClient *VLLMClient, fileServerHost string) *APIServer {
	return &APIServer{
		port:           port,
		filePath:       filePath,
		fileServer:     fileServer,
		vllmClient:     vllmClient,
		fileServerHost: fileServerHost,
	}
}

// UploadRequest 上传请求
type UploadRequest struct {
	Prompt string `json:"prompt,omitempty"` // 可选的提示词
}

// UploadResponse 上传响应
type UploadResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`
	Inference *struct {
		Content string `json:"content"`
		Usage   *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage,omitempty"`
	} `json:"inference,omitempty"`
	Error string `json:"error,omitempty"`
}

// Start 启动API服务器
func (as *APIServer) Start() error {
	router := gin.Default()

	// 配置 CORS 中间件
	config := cors.DefaultConfig()
	// 允许所有来源（生产环境建议配置具体域名）
	config.AllowAllOrigins = true
	// 允许的 HTTP 方法
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	// 允许的请求头
	config.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Cache-Control", "X-Requested-With"}
	// 允许暴露的响应头
	config.ExposeHeaders = []string{"Content-Length", "Content-Type"}
	// 允许携带凭证（如果需要）
	config.AllowCredentials = true
	// 预检请求缓存时间（秒）
	config.MaxAge = 12 * 60 * 60 // 12小时

	router.Use(cors.New(config))

	// 设置最大上传大小（32MB）
	router.MaxMultipartMemory = 32 << 20

	// 上传图片接口
	router.POST("/api/upload", as.handleUpload)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "api_server",
		})
	})

	log.Printf("API服务器启动在端口 %s", as.port)

	return router.Run(":" + as.port)
}

// handleUpload 处理图片上传
func (as *APIServer) handleUpload(c *gin.Context) {
	// 获取文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("获取文件失败: %v", err),
		})
		return
	}

	// 验证文件类型（只允许图片）
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("不支持的文件类型: %s，仅支持: jpg, jpeg, png, gif, webp", ext),
		})
		return
	}

	// 确保上传目录存在
	if err := os.MkdirAll(as.filePath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("创建上传目录失败: %v", err),
		})
		return
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	savePath := filepath.Join(as.filePath, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, UploadResponse{
			Success: false,
			Error:   fmt.Sprintf("保存文件失败: %v", err),
		})
		return
	}

	log.Printf("文件上传成功: %s", filename)

	// 生成文件URL
	imageURL := as.fileServer.GetFileURL(filename, as.fileServerHost)

	// 获取可选的提示词
	var prompt string
	if c.PostForm("prompt") != "" {
		prompt = c.PostForm("prompt")
	}

	// 调用VLLM推理接口
	inferenceResp, err := as.vllmClient.InferImage(imageURL, prompt)
	if err != nil {
		log.Printf("VLLM推理失败: %v", err)
		// 即使推理失败，也返回上传成功的信息
		c.JSON(http.StatusOK, UploadResponse{
			Success:  true,
			Message:  "文件上传成功，但推理失败",
			ImageURL: imageURL,
			Error:    fmt.Sprintf("推理失败: %v", err),
		})
		return
	}

	// 构建响应
	response := UploadResponse{
		Success:  true,
		Message:  "文件上传成功并完成推理",
		ImageURL: imageURL,
	}

	if inferenceResp.Success {
		response.Inference = &struct {
			Content string `json:"content"`
			Usage   *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage,omitempty"`
		}{
			Content: inferenceResp.Content,
			Usage:   inferenceResp.Usage,
		}
	} else {
		response.Error = inferenceResp.Error
	}

	c.JSON(http.StatusOK, response)
}

