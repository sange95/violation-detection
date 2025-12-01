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

	// 自定义 CORS 中间件 - 允许所有方法和请求头
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// 动态返回请求中指定的方法和header（允许所有）
		requestMethod := c.Request.Header.Get("Access-Control-Request-Method")
		if requestMethod != "" {
			c.Header("Access-Control-Allow-Methods", requestMethod)
		} else {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD, CONNECT, TRACE")
		}

		requestHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
		if requestHeaders != "" {
			// 预检请求时，返回请求中指定的所有header（允许所有）
			c.Header("Access-Control-Allow-Headers", requestHeaders)
		} else {
			// 当没有预检请求时，返回常用header列表（不包含*，因为AllowCredentials为true）
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Cache-Control, X-Requested-With, X-Custom-Header, Access-Control-Request-Method, Access-Control-Request-Headers")
		}

		c.Header("Access-Control-Expose-Headers", "*")
		c.Header("Access-Control-Max-Age", "43200") // 12小时

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

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
