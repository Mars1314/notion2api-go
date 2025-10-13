package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Settings 应用配置结构
type Settings struct {
	AppName          string
	AppVersion       string
	Description      string
	APIMasterKey     string
	NotionCookie     string
	NotionSpaceID    string
	NotionUserID     string
	NotionUserName   string
	NotionUserEmail  string
	NotionBlockID    string
	NotionClientVersion string
	APIRequestTimeout int
	NginxPort        int
	DefaultModel     string
	KnownModels      []string
	ModelMap         map[string]string
}

var Config *Settings

// LoadConfig 从环境变量加载配置
func LoadConfig() *Settings {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，将使用系统环境变量")
	}

	config := &Settings{
		AppName:    getEnv("APP_NAME", "notion-2api-go"),
		AppVersion: getEnv("APP_VERSION", "1.0.0"),
		Description: getEnv("DESCRIPTION", "一个将 Notion AI 转换为兼容 OpenAI 格式 API 的高性能代理 (Go 版本)。"),

		APIMasterKey:    getEnv("API_MASTER_KEY", ""),
		NotionCookie:    getEnv("NOTION_COOKIE", ""),
		NotionSpaceID:   getEnv("NOTION_SPACE_ID", ""),
		NotionUserID:    getEnv("NOTION_USER_ID", ""),
		NotionUserName:  getEnv("NOTION_USER_NAME", ""),
		NotionUserEmail: getEnv("NOTION_USER_EMAIL", ""),
		NotionBlockID:   getEnv("NOTION_BLOCK_ID", ""),
		NotionClientVersion: getEnv("NOTION_CLIENT_VERSION", "23.13.20251011.2037"),

		APIRequestTimeout: getEnvAsInt("API_REQUEST_TIMEOUT", 180),
		NginxPort:        getEnvAsInt("NGINX_PORT", 8004),
		DefaultModel:     getEnv("DEFAULT_MODEL", "claude-sonnet-4.5"),

		KnownModels: []string{
			"claude-sonnet-4.5",
			"gpt-5",
			"claude-opus-4.1",
			"gpt-4.1",
			"gemini-2.5-flash",
			"gemini-2.5-pro",
		},

		ModelMap: map[string]string{
			"claude-sonnet-4.5":  "anthropic-sonnet-alt",
			"gpt-5":              "openai-turbo",
			"claude-opus-4.1":    "anthropic-opus-4.1",
			"gpt-4.1":            "openai-gpt-4.1",
			"gemini-2.5-flash":   "vertex-gemini-2.5-flash",
			"gemini-2.5-pro":     "vertex-gemini-2.5-pro",
		},
	}

	// 验证必需的配置
	if config.NotionCookie == "" || config.NotionSpaceID == "" || config.NotionUserID == "" {
		log.Fatal("配置错误: NOTION_COOKIE, NOTION_SPACE_ID 和 NOTION_USER_ID 必须在 .env 文件中全部设置。")
	}

	Config = config
	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt 获取环境变量作为整数，如果不存在或解析失败则返回默认值
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetCookieHeader 获取格式化的 Cookie 头
func (s *Settings) GetCookieHeader() string {
	cookie := strings.TrimSpace(s.NotionCookie)
	if strings.Contains(cookie, "=") {
		return cookie
	}
	return "token_v2=" + cookie
}