package main

import (
	"fmt"
	"notion-2api-go/internal/config"
	"notion-2api-go/internal/providers"
	"notion-2api-go/internal/utils"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var provider providers.BaseProvider

func main() {
	// 设置日志格式
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// 加载配置
	cfg := config.LoadConfig()
	log.Infof("应用启动中... %s v%s", cfg.AppName, cfg.AppVersion)
	log.Info("服务已配置为 Notion AI 代理模式。")
	log.Infof("服务将在 http://localhost:%d 上可用", cfg.NginxPort)

	// 初始化 Provider
	var err error
	provider, err = providers.NewNotionAIProvider(cfg)
	if err != nil {
		log.Fatalf("初始化 Notion Provider 失败: %v", err)
	}

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建 Gin 路由
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 根路径
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("欢迎来到 %s v%s. 服务运行正常。", cfg.AppName, cfg.AppVersion),
		})
	})

	// 文档页面
	r.GET("/docs", func(c *gin.Context) {
		log.Infof("访问文档页面 - 来源: %s, User-Agent: %s", c.ClientIP(), c.Request.UserAgent())
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, utils.GetDocsHTML(cfg.AppName, cfg.AppVersion, cfg.NginxPort))
	})

	// API 路由
	api := r.Group("/v1")
	{
		// 聊天补全
		api.POST("/chat/completions", authMiddleware(cfg), chatCompletionsHandler)
		
		// 模型列表
		api.GET("/models", authMiddleware(cfg), listModelsHandler)
	}

	// 启动服务器
	port := fmt.Sprintf(":%d", cfg.NginxPort)
	log.Infof("服务器启动在端口 %s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// authMiddleware API 认证中间件
func authMiddleware(cfg *config.Settings) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.APIMasterKey != "" && cfg.APIMasterKey != "1" {
			authorization := c.GetHeader("Authorization")
			if authorization == "" || !strings.Contains(strings.ToLower(authorization), "bearer") {
				c.JSON(401, gin.H{
					"error": "需要 Bearer Token 认证。",
				})
				c.Abort()
				return
			}

			parts := strings.Split(authorization, " ")
			if len(parts) != 2 {
				c.JSON(401, gin.H{
					"error": "无效的认证格式。",
				})
				c.Abort()
				return
			}

			token := parts[1]
			if token != cfg.APIMasterKey {
				c.JSON(403, gin.H{
					"error": "无效的 API Key。",
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// chatCompletionsHandler 处理聊天补全请求
func chatCompletionsHandler(c *gin.Context) {
	var requestData map[string]interface{}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(400, gin.H{
			"error": fmt.Sprintf("无效的请求数据: %v", err),
		})
		return
	}

	if err := provider.ChatCompletion(c, requestData); err != nil {
		log.Errorf("处理聊天请求时发生错误: %v", err)
		// 错误已在 provider 中处理并发送给客户端
	}
}

// listModelsHandler 处理模型列表请求
func listModelsHandler(c *gin.Context) {
	if err := provider.GetModels(c); err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("获取模型列表失败: %v", err),
		})
		return
	}
}