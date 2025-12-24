package providers

import "github.com/gin-gonic/gin"

// BaseProvider 定义了所有 AI Provider 必须实现的接口
type BaseProvider interface {
	// ChatCompletion 处理聊天补全请求 (OpenAI 格式)
	ChatCompletion(c *gin.Context, requestData map[string]interface{}) error

	// ChatCompletionAnthropic 处理聊天补全请求 (Anthropic 格式响应)
	ChatCompletionAnthropic(c *gin.Context, convertedData map[string]interface{}, originalData map[string]interface{}) error

	// GetModels 获取可用模型列表
	GetModels(c *gin.Context) error
}

// ChatMessage 聊天消息结构
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求结构
type ChatRequest struct {
	Model          string        `json:"model"`
	Messages       []ChatMessage `json:"messages"`
	Stream         bool          `json:"stream"`
	NotionBlockID  string        `json:"notion_block_id,omitempty"`
}

// ModelResponse 模型响应结构
type ModelResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// ModelInfo 模型信息
type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}