package main

import (
	"log"
	"net/http"
	"path/filepath"

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
