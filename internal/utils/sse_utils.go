package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// SSE 相关常量
var DoneChunk = []byte("data: [DONE]\n\n")

// CreateSSEData 将数据转换为 SSE 格式
func CreateSSEData(data interface{}) []byte {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return []byte{}
	}
	return []byte(fmt.Sprintf("data: %s\n\n", string(jsonData)))
}

// ChatCompletionChunk OpenAI 聊天补全响应块结构
type ChatCompletionChunk struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []CompletionChoice     `json:"choices"`
}

// CompletionChoice 选择结构
type CompletionChoice struct {
	Index        int                    `json:"index"`
	Delta        map[string]interface{} `json:"delta"`
	FinishReason *string                `json:"finish_reason"`
}

// CreateChatCompletionChunk 创建聊天补全响应块
func CreateChatCompletionChunk(requestID, model string, content *string, finishReason *string, role *string) ChatCompletionChunk {
	delta := make(map[string]interface{})
	
	if role != nil {
		delta["role"] = *role
	}
	if content != nil {
		delta["content"] = *content
	}

	return ChatCompletionChunk{
		ID:      requestID,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []CompletionChoice{
			{
				Index:        0,
				Delta:        delta,
				FinishReason: finishReason,
			},
		},
	}
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// CreateErrorSSE 创建错误 SSE 响应
func CreateErrorSSE(message string) []byte {
	errorResp := ErrorResponse{
		Error: ErrorDetail{
			Message: message,
			Type:    "internal_server_error",
		},
	}
	return CreateSSEData(errorResp)
}

// StringPtr 返回字符串指针
func StringPtr(s string) *string {
	return &s
}