package main

import (
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// FileServer 文件服务器
type FileServer struct {
	port     string
	filePath string
}

// NewFileServer 创建文件服务器
func NewFileServer(port, filePath string) *FileServer {
	return &FileServer{
		port:     port,
		filePath: filePath,
	}
}

// Start 启动文件服务器
func (fs *FileServer) Start() error {
	router := gin.Default()

	// 配置 CORS 中间件
	config := cors.DefaultConfig()
	// 允许所有来源（生产环境建议配置具体域名）
	config.AllowAllOrigins = true
	// 允许的 HTTP 方法
	config.AllowMethods = []string{"GET", "POST", "OPTIONS"}
	// 允许的请求头
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Cache-Control", "X-Requested-With"}
	// 允许暴露的响应头
	config.ExposeHeaders = []string{"Content-Length", "Content-Type"}
	// 允许携带凭证（如果需要）
	config.AllowCredentials = true
	// 预检请求缓存时间（秒）
	config.MaxAge = 12 * 60 * 60 // 12小时

	router.Use(cors.New(config))

	// 静态文件服务
	router.Static("/files", fs.filePath)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "file_server",
		})
	})

	log.Printf("文件服务器启动在端口 %s，文件路径: %s", fs.port, fs.filePath)
	log.Printf("访问文件: http://localhost:%s/files/<filename>", fs.port)

	return router.Run(":" + fs.port)
}

// GetFileURL 获取文件URL
func (fs *FileServer) GetFileURL(filename string, host string) string {
	if host == "" {
		host = "localhost:" + fs.port
	}
	return "http://" + host + "/files/" + filepath.Base(filename)
}
