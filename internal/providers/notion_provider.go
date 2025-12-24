package providers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"notion-2api-go/internal/config"
	"notion-2api-go/internal/utils"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// NotionAIProvider Notion AI 提供者实现
type NotionAIProvider struct {
	client       *http.Client
	apiEndpoints map[string]string
	config       *config.Settings
}

// NewNotionAIProvider 创建新的 Notion AI 提供者
func NewNotionAIProvider(cfg *config.Settings) (*NotionAIProvider, error) {
	if cfg.NotionCookie == "" || cfg.NotionSpaceID == "" || cfg.NotionUserID == "" {
		return nil, fmt.Errorf("配置错误: NOTION_COOKIE, NOTION_SPACE_ID 和 NOTION_USER_ID 必须在 .env 文件中全部设置")
	}

	// 配置 Transport 以更好地模拟浏览器行为
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	provider := &NotionAIProvider{
		client: &http.Client{
			Timeout:   time.Duration(cfg.APIRequestTimeout) * time.Second,
			Transport: transport,
		},
		apiEndpoints: map[string]string{
			"runInference":     "https://www.notion.so/api/v3/runInferenceTranscript",
			"saveTransactions": "https://www.notion.so/api/v3/saveTransactionsFanout",
		},
		config: cfg,
	}

	// 会话预热
	provider.warmupSession()
	return provider, nil
}

// warmupSession 会话预热
func (p *NotionAIProvider) warmupSession() {
	log.Info("正在进行会话预热 (Session Warm-up)...")
	req, err := http.NewRequest("GET", "https://www.notion.so/", nil)
	if err != nil {
		log.Errorf("会话预热失败: %v", err)
		return
	}

	headers := p.prepareHeaders()
	delete(headers, "Accept")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		log.Errorf("会话预热失败: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Info("会话预热成功。")
}

// prepareHeaders 准备请求头
func (p *NotionAIProvider) prepareHeaders() map[string]string {
	return map[string]string{
		"Content-Type":                "application/json",
		"Accept":                      "application/x-ndjson",
		"Accept-Language":             "zh-CN,zh;q=0.9,en;q=0.8",
		"Cookie":                      p.config.GetCookieHeader(),
		"x-notion-space-id":           p.config.NotionSpaceID,
		"x-notion-active-user-header": p.config.NotionUserID,
		"x-notion-client-version":     p.config.NotionClientVersion,
		"notion-audit-log-platform":   "web",
		"Origin":                      "https://www.notion.so",
		"Referer":                     "https://www.notion.so/",
		"User-Agent":                  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
		"sec-ch-ua":                   `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`,
		"sec-ch-ua-mobile":            "?0",
		"sec-ch-ua-platform":          `"Windows"`,
		"sec-fetch-dest":              "empty",
		"sec-fetch-mode":              "cors",
		"sec-fetch-site":              "same-origin",
	}
}

// createThread 创建对话线程
func (p *NotionAIProvider) createThread(threadType string) (string, error) {
	threadID := uuid.New().String()
	payload := map[string]interface{}{
		"requestId": uuid.New().String(),
		"transactions": []map[string]interface{}{
			{
				"id":      uuid.New().String(),
				"spaceId": p.config.NotionSpaceID,
				"operations": []map[string]interface{}{
					{
						"pointer": map[string]interface{}{
							"table":   "thread",
							"id":      threadID,
							"spaceId": p.config.NotionSpaceID,
						},
						"path":    []string{},
						"command": "set",
						"args": map[string]interface{}{
							"id":               threadID,
							"version":          1,
							"parent_id":        p.config.NotionSpaceID,
							"parent_table":     "space",
							"space_id":         p.config.NotionSpaceID,
							"created_time":     time.Now().UnixMilli(),
							"created_by_id":    p.config.NotionUserID,
							"created_by_table": "notion_user",
							"messages":         []interface{}{},
							"data":             map[string]interface{}{},
							"alive":            true,
							"type":             threadType,
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", p.apiEndpoints["saveTransactions"], bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	for key, value := range p.prepareHeaders() {
		req.Header.Set(key, value)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("创建对话线程失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("创建对话线程失败，状态码: %d", resp.StatusCode)
	}

	log.Infof("对话线程创建成功, Thread ID: %s", threadID)
	return threadID, nil
}

// normalizeBlockID 规范化 Block ID
func (p *NotionAIProvider) normalizeBlockID(blockID string) string {
	if blockID == "" {
		return blockID
	}
	b := strings.ReplaceAll(strings.TrimSpace(blockID), "-", "")
	if len(b) == 32 && regexp.MustCompile(`^[0-9a-fA-F]{32}$`).MatchString(b) {
		return fmt.Sprintf("%s-%s-%s-%s-%s", b[0:8], b[8:12], b[12:16], b[16:20], b[20:])
	}
	return blockID
}

// preparePayload 准备请求载荷
func (p *NotionAIProvider) preparePayload(requestData map[string]interface{}, threadID, mappedModel, threadType string) map[string]interface{} {
	// 准备 config - 使用与浏览器一致的完整配置
	configValue := map[string]interface{}{
		"type":                            threadType,
		"model":                           mappedModel,
		"modelFromUser":                   true,
		"useWebSearch":                    true,
		"useReadOnlyMode":                 false,
		"writerMode":                      false,
		"isCustomAgent":                   false,
		"isCustomAgentBuilder":            false,
		"useCustomAgentDraft":             false,
		"enableAgentAutomations":          false,
		"enableAgentIntegrations":         false,
		"enableBackgroundAgents":          false,
		"enableCustomAgents":              false,
		"enableExperimentalIntegrations":  false,
		"enableAgentViewNotificationsTool": false,
		"enableAgentRevertTool":           false,
		"enableAgentDiffs":                false,
		"enableAgentCreateDbTemplate":     false,
		"enableCsvAttachmentSupport":      true,
		"enableDatabaseAgents":            false,
		"enableAgentThreadTools":          false,
		"enableRunAgentTool":              false,
		"enableAgentDashboards":           false,
		"enableAgentCardCustomization":    true,
		"enableSystemPromptAsPage":        false,
		"enableUserSessionContext":        false,
		"enableComputer":                  false,
		"enableScriptAgent":               false,
		"enableAgentGenerateImage":        false,
		"enableAgentTodos":                false,
		"enableSpeculativeSearch":         false,
		"enableQueryCalendar":             false,
		"enableQueryMail":                 false,
		"enableUpdatePageV2Tool":          true,
		"enableUpdatePageAutofixer":       true,
		"enableUpdateAgentV2Tools":        true,
		"enableUpdatePageMarkdownTree":    false,
		"enableUpdatePageTreeDiff":        false,
		"enableUpdatePageOrderUpdates":    true,
		"enableUpdatePageTreeDiffMetrics": false,
		"availableConnectors":             []interface{}{},
		"searchScopes":                    []map[string]interface{}{{"type": "everything"}},
	}

	// 准备 context
	contextValue := map[string]interface{}{
		"timezone":        "Asia/Shanghai",
		"userName":        p.config.NotionUserName,
		"userId":          p.config.NotionUserID,
		"userEmail":       p.config.NotionUserEmail,
		"spaceName":       p.config.NotionUserName + "的工作空间",
		"spaceId":         p.config.NotionSpaceID,
		"currentDatetime": time.Now().Format(time.RFC3339Nano),
		"surface":         "ai_module",
	}

	// 构建 transcript
	transcript := []map[string]interface{}{
		{
			"id":    uuid.New().String(),
			"type":  "config",
			"value": configValue,
		},
		{
			"id":    uuid.New().String(),
			"type":  "context",
			"value": contextValue,
		},
	}

	// 添加消息
	log.Infof("开始处理消息，类型: %T", requestData["messages"])
	if messages, ok := requestData["messages"].([]interface{}); ok {
		log.Infof("消息数量: %d", len(messages))
		for i, msg := range messages {
			if msgMap, ok := msg.(map[string]interface{}); ok {
				role, _ := msgMap["role"].(string)
				content, _ := msgMap["content"].(string)
				log.Infof("消息 %d: role=%s, content长度=%d", i, role, len(content))

				// 跳过 system 消息（Notion 不支持）
				if role == "system" {
					continue
				}

				if role == "user" {
					transcript = append(transcript, map[string]interface{}{
						"id":        uuid.New().String(),
						"type":      "user",
						"value":     []interface{}{[]interface{}{content}},
						"userId":    p.config.NotionUserID,
						"createdAt": time.Now().Format(time.RFC3339),
					})
				} else if role == "assistant" {
					transcript = append(transcript, map[string]interface{}{
						"id":   uuid.New().String(),
						"type": "agent-inference",
						"value": []interface{}{
							map[string]interface{}{
								"type":    "text",
								"content": content,
							},
						},
					})
				}
			}
		}
	} else if messages, ok := requestData["messages"].([]map[string]interface{}); ok {
		// 处理 []map[string]interface{} 类型
		log.Infof("消息数量 (map类型): %d", len(messages))
		for i, msgMap := range messages {
			role, _ := msgMap["role"].(string)
			content, _ := msgMap["content"].(string)
			log.Infof("消息 %d: role=%s, content长度=%d", i, role, len(content))

			// 跳过 system 消息
			if role == "system" {
				continue
			}

			if role == "user" {
				transcript = append(transcript, map[string]interface{}{
					"id":        uuid.New().String(),
					"type":      "user",
					"value":     []interface{}{[]interface{}{content}},
					"userId":    p.config.NotionUserID,
					"createdAt": time.Now().Format(time.RFC3339),
				})
			} else if role == "assistant" {
				transcript = append(transcript, map[string]interface{}{
					"id":   uuid.New().String(),
					"type": "agent-inference",
					"value": []interface{}{
						map[string]interface{}{
							"type":    "text",
							"content": content,
						},
					},
				})
			}
		}
	} else {
		log.Warnf("无法解析消息，类型: %T", requestData["messages"])
	}
	
	log.Infof("最终 transcript 长度: %d", len(transcript))

	payload := map[string]interface{}{
		"traceId":                 uuid.New().String(),
		"spaceId":                 p.config.NotionSpaceID,
		"transcript":              transcript,
		"threadId":                threadID,
		"threadParentPointer": map[string]interface{}{
			"table":   "space",
			"id":      p.config.NotionSpaceID,
			"spaceId": p.config.NotionSpaceID,
		},
		"createThread":            true,
		"isPartialTranscript":     false,
		"asPatchResponse":         false,
		"generateTitle":           true,
		"saveAllThreadOperations": true,
		"threadType":              threadType,
		"isUserInAnySalesAssistedSpace": false,
		"isSpaceSalesAssisted":          false,
	}

	// 添加 debugOverrides
	payload["debugOverrides"] = map[string]interface{}{
		"emitAgentSearchExtractedResults": true,
		"cachedInferences":                map[string]interface{}{},
		"annotationInferences":            map[string]interface{}{},
		"emitInferences":                  false,
	}

	return payload
}

// cleanContent 清理响应内容
func (p *NotionAIProvider) cleanContent(content string) string {
	if content == "" {
		return ""
	}

	// 移除各种标记和噪音文本
	patterns := []string{
		`<lang primary="[^"]*"\s*/>\n*`,
		`<thinking>[\s\S]*?</thinking>\s*`,
		`<thought>[\s\S]*?</thought>\s*`,
		`(?i)^.*?Chinese whatmodel I am.*?Theyspecifically.*?requested.*?me.*?to.*?reply.*?in.*?Chinese\.\s*`,
		`(?i)^.*?This.*?is.*?a.*?straightforward.*?question.*?about.*?my.*?identity.*?asan.*?AI.*?assistant\.\s*`,
		`(?i)^.*?Idon't.*?need.*?to.*?use.*?any.*?tools.*?for.*?this.*?-\s*it's.*?asimple.*?informational.*?response.*?aboutwhat.*?I.*?am\.\s*`,
		`(?i)^.*?Sincethe.*?user.*?asked.*?in.*?Chinese.*?and.*?specifically.*?requested.*?a.*?Chinese.*?response.*?I.*?should.*?respond.*?in.*?Chinese\.\s*`,
		`(?i)^.*?What model are you.*?in Chinese and specifically requesting.*?me.*?to.*?reply.*?in.*?Chinese\.\s*`,
		`(?i)^.*?This.*?is.*?a.*?question.*?about.*?my.*?identity.*?not requiring.*?any.*?tool.*?use.*?I.*?should.*?respond.*?directly.*?to.*?the.*?user.*?in.*?Chinese.*?as.*?requested\.\s*`,
		`(?i)^.*?I.*?should.*?identify.*?myself.*?as.*?Notion.*?AI.*?as.*?mentioned.*?in.*?the.*?system.*?prompt.*?\s*`,
		`(?i)^.*?I.*?should.*?not.*?make.*?specific.*?claims.*?about.*?the.*?underlying.*?model.*?architecture.*?since.*?that.*?information.*?is.*?not.*?provided.*?in.*?my.*?context\.\s*`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		content = re.ReplaceAllString(content, "")
	}

	return strings.TrimSpace(content)
}

// parseNDJSONLine 解析 NDJSON 行
func (p *NotionAIProvider) parseNDJSONLine(line string) []map[string]interface{} {
	results := []map[string]interface{}{}
	
	if strings.TrimSpace(line) == "" {
		return results
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		log.Warnf("解析NDJSON行失败: %v - Line: %s", err, line)
		return results
	}

	log.Debugf("原始响应数据: %v", data)

	// 检查是否是额度用尽错误
	if dataType, ok := data["type"].(string); ok && dataType == "premium-feature-unavailable" {
		if featureAvailability, ok := data["featureAvailability"].(map[string]interface{}); ok {
			if limit, ok := featureAvailability["limit"].(map[string]interface{}); ok {
				current, _ := limit["current"].(float64)
				total, _ := limit["total"].(float64)
				errorMsg := fmt.Sprintf("Notion AI 额度已用尽 (%d/%d)，请升级到 Business 计划或等待额度重置", int(current), int(total))
				log.Errorf(errorMsg)
				results = append(results, map[string]interface{}{
					"type":  "error",
					"error": errorMsg,
				})
				return results
			}
		}
		results = append(results, map[string]interface{}{
			"type":  "error",
			"error": "Notion AI 功能不可用，可能是额度用尽或需要升级计划",
		})
		return results
	}

	// 格式1: Gemini 返回的 markdown-chat 事件
	if dataType, ok := data["type"].(string); ok && dataType == "markdown-chat" {
		if content, ok := data["value"].(string); ok && content != "" {
			log.Info("从 'markdown-chat' 直接事件中提取到内容。")
			results = append(results, map[string]interface{}{
				"type":    "final",
				"content": content,
			})
		}
	}

	// 格式2: Claude 和 GPT 返回的补丁流，以及 Gemini 的 patch 格式
	if dataType, ok := data["type"].(string); ok && dataType == "patch" {
		if v, ok := data["v"].([]interface{}); ok {
			for _, operation := range v {
				if op, ok := operation.(map[string]interface{}); ok {
					opType, _ := op["o"].(string)
					path, _ := op["p"].(string)
					value := op["v"]

					// Gemini 的完整内容 patch 格式
					if opType == "a" && strings.HasSuffix(path, "/s/-") {
						if valueMap, ok := value.(map[string]interface{}); ok {
							if valueMap["type"] == "markdown-chat" {
								if content, ok := valueMap["value"].(string); ok && content != "" {
									log.Info("从 'patch' (Gemini-style) 中提取到完整内容。")
									results = append(results, map[string]interface{}{
										"type":    "final",
										"content": content,
									})
								}
							}
						}
					}

					// Gemini 的增量内容 patch 格式
					if opType == "x" && strings.Contains(path, "/s/") && strings.HasSuffix(path, "/value") {
						if content, ok := value.(string); ok && content != "" {
							log.Infof("从 'patch' (Gemini增量) 中提取到内容: %s", content)
							results = append(results, map[string]interface{}{
								"type":    "incremental",
								"content": content,
							})
						}
					}

					// Claude 和 GPT 的增量内容 patch 格式
					if opType == "x" && strings.Contains(path, "/value/") {
						if content, ok := value.(string); ok && content != "" {
							log.Infof("从 'patch' (Claude/GPT增量) 中提取到内容: %s", content)
							results = append(results, map[string]interface{}{
								"type":    "incremental",
								"content": content,
							})
						}
					}

					// Claude 和 GPT 的完整内容 patch 格式
					if opType == "a" && strings.HasSuffix(path, "/value/-") {
						if valueMap, ok := value.(map[string]interface{}); ok {
							if valueMap["type"] == "text" {
								if content, ok := valueMap["content"].(string); ok && content != "" {
									log.Info("从 'patch' (Claude/GPT-style) 中提取到完整内容。")
									results = append(results, map[string]interface{}{
										"type":    "final",
										"content": content,
									})
								}
							}
						}
					}
				}
			}
		}
	}

	// 格式3: 处理record-map类型的数据
	if dataType, ok := data["type"].(string); ok && dataType == "record-map" {
		if recordMap, ok := data["recordMap"].(map[string]interface{}); ok {
			if threadMessage, ok := recordMap["thread_message"].(map[string]interface{}); ok {
				// 找到最新的 agent-inference 消息（created_time 最大的）
				var latestContent string
				var latestTime float64 = 0

				for _, msgData := range threadMessage {
					if msgMap, ok := msgData.(map[string]interface{}); ok {
						if valueData, ok := msgMap["value"].(map[string]interface{}); ok {
							if valueValue, ok := valueData["value"].(map[string]interface{}); ok {
								if step, ok := valueValue["step"].(map[string]interface{}); ok {
									stepType, _ := step["type"].(string)

									// 只处理 agent-inference 或 markdown-chat 类型
									if stepType != "agent-inference" && stepType != "markdown-chat" {
										continue
									}

									// 获取创建时间
									createdTime, _ := valueValue["created_time"].(float64)

									var content string
									if stepType == "markdown-chat" {
										content, _ = step["value"].(string)
									} else if stepType == "agent-inference" {
										if agentValues, ok := step["value"].([]interface{}); ok {
											for _, item := range agentValues {
												if itemMap, ok := item.(map[string]interface{}); ok {
													if itemMap["type"] == "text" {
														content, _ = itemMap["content"].(string)
														break
													}
												}
											}
										}
									}

									// 保留最新的消息
									if content != "" && createdTime >= latestTime {
										latestTime = createdTime
										latestContent = content
									}
								}
							}
						}
					}
				}

				if latestContent != "" {
					log.Infof("从 record-map 提取到最终内容 (created_time: %.0f)", latestTime)
					results = append(results, map[string]interface{}{
						"type":    "final",
						"content": latestContent,
					})
				}
			}
		}
	}

	return results
}

// ChatCompletion 处理聊天补全请求（支持流式和非流式）
func (p *NotionAIProvider) ChatCompletion(c *gin.Context, requestData map[string]interface{}) error {
	// 解析 stream 参数，默认为 true
	stream := true
	if streamVal, ok := requestData["stream"].(bool); ok {
		stream = streamVal
	}

	// 获取模型
	modelName := p.config.DefaultModel
	if model, ok := requestData["model"].(string); ok && model != "" {
		modelName = model
	}

	mappedModel := p.config.ModelMap[modelName]
	if mappedModel == "" {
		mappedModel = "anthropic-sonnet-alt-thinking"
	}

	// 确定线程类型
	threadType := "workflow"
	if strings.HasPrefix(mappedModel, "vertex-") {
		threadType = "markdown-chat"
	}

	// 生成新的 thread ID，让 Notion 自动创建
	threadID := uuid.New().String()

	// 准备请求载荷
	payload := p.preparePayload(requestData, threadID, mappedModel, threadType)
	// 设置 createThread 为 true，让 Notion 自动创建线程
	payload["createThread"] = true

	// 发送请求到 Notion
	jsonData, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("序列化请求失败: %v", err),
		})
		return err
	}

	log.Infof("请求 Notion AI URL: %s", p.apiEndpoints["runInference"])
	log.Debugf("请求体: %s", string(jsonData))

	req, err := http.NewRequest("POST", p.apiEndpoints["runInference"], bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("创建请求失败: %v", err),
		})
		return err
	}

	for key, value := range p.prepareHeaders() {
		req.Header.Set(key, value)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("请求 Notion AI 失败: %v", err),
		})
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Errorf("Notion AI 返回错误，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Notion AI 返回错误状态码: %d", resp.StatusCode),
		})
		return fmt.Errorf("状态码: %d", resp.StatusCode)
	}

	// 处理响应 - 先收集所有数据
	var incrementalFragments []string
	var finalMessage string

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// 调试：打印原始响应行
		log.Infof("收到响应行: %s", line)

		parsedResults := p.parseNDJSONLine(line)
		for _, result := range parsedResults {
			textType, _ := result["type"].(string)
			content, _ := result["content"].(string)

			// 处理错误类型
			if textType == "error" {
				errorMsg, _ := result["error"].(string)
				if stream {
					errorData := utils.CreateErrorSSE(errorMsg)
					c.Writer.Header().Set("Content-Type", "text/event-stream")
					c.Writer.Write(errorData)
					c.Writer.Write(utils.DoneChunk)
				} else {
					c.JSON(http.StatusPaymentRequired, gin.H{"error": errorMsg})
				}
				return fmt.Errorf(errorMsg)
			}

			if textType == "final" {
				finalMessage = content
			} else if textType == "incremental" {
				incrementalFragments = append(incrementalFragments, content)
			}
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Errorf("读取响应流时出错: %v", err)
	}

	// 确定最终响应
	var fullResponse string
	if finalMessage != "" {
		fullResponse = finalMessage
		log.Info("成功从 record-map 或 Gemini patch/event 中提取到最终消息。")
	} else if len(incrementalFragments) > 0 {
		fullResponse = strings.Join(incrementalFragments, "")
		log.Info("使用拼接所有增量片段的方式获得最终消息。")
	}

	if fullResponse == "" {
		errorMsg := "未能从 Notion 获取有效响应"
		if stream {
			errorData := utils.CreateErrorSSE(errorMsg)
			c.Writer.Header().Set("Content-Type", "text/event-stream")
			c.Writer.Write(errorData)
			c.Writer.Write(utils.DoneChunk)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errorMsg})
		}
		return fmt.Errorf("空响应")
	}

	cleanedResponse := p.cleanContent(fullResponse)
	log.Infof("清洗后的最终响应: %s", cleanedResponse)

	if stream {
		// 流式响应
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		requestID := fmt.Sprintf("chatcmpl-%s", uuid.New().String())

		// 发送角色块
		role := "assistant"
		roleChunk := utils.CreateChatCompletionChunk(requestID, modelName, nil, nil, &role)
		c.Writer.Write(utils.CreateSSEData(roleChunk))
		c.Writer.Flush()

		// 发送内容
		chunk := utils.CreateChatCompletionChunk(requestID, modelName, &cleanedResponse, nil, nil)
		c.Writer.Write(utils.CreateSSEData(chunk))
		c.Writer.Flush()

		// 发送完成标记
		finishReason := "stop"
		finalChunk := utils.CreateChatCompletionChunk(requestID, modelName, nil, &finishReason, nil)
		c.Writer.Write(utils.CreateSSEData(finalChunk))
		c.Writer.Write(utils.DoneChunk)
		c.Writer.Flush()
	} else {
		// 非流式响应（OpenAI 格式）
		response := map[string]interface{}{
			"id":      fmt.Sprintf("chatcmpl-%s", uuid.New().String()),
			"object":  "chat.completion",
			"created": time.Now().Unix(),
			"model":   modelName,
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": cleanedResponse,
					},
					"finish_reason": "stop",
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     0,
				"completion_tokens": 0,
				"total_tokens":      0,
			},
		}
		c.JSON(http.StatusOK, response)
	}

	return nil
}

// GetModels 获取模型列表
func (p *NotionAIProvider) GetModels(c *gin.Context) error {
	models := []ModelInfo{}
	created := time.Now().Unix()
	
	for _, modelName := range p.config.KnownModels {
		models = append(models, ModelInfo{
			ID:      modelName,
			Object:  "model",
			Created: created,
			OwnedBy: "lzA6",
		})
	}

	c.JSON(http.StatusOK, ModelResponse{
		Object: "list",
		Data:   models,
	})
	
	return nil
}


// ChatCompletionAnthropic 处理 Anthropic Messages API 请求
func (p *NotionAIProvider) ChatCompletionAnthropic(c *gin.Context, convertedData map[string]interface{}, originalData map[string]interface{}) error {
	// 解析 stream 参数
	stream := false
	if streamVal, ok := originalData["stream"].(bool); ok {
		stream = streamVal
	}

	// 获取模型
	modelName := p.config.DefaultModel
	if model, ok := originalData["model"].(string); ok && model != "" {
		modelName = model
	}

	mappedModel := p.config.ModelMap[modelName]
	if mappedModel == "" {
		mappedModel = "anthropic-sonnet-alt-thinking"
	}

	// 确定线程类型
	threadType := "workflow"
	if strings.HasPrefix(mappedModel, "vertex-") {
		threadType = "markdown-chat"
	}

	// 生成新的 thread ID
	threadID := uuid.New().String()

	// 准备请求载荷
	payload := p.preparePayload(convertedData, threadID, mappedModel, threadType)
	payload["createThread"] = true

	// 发送请求到 Notion
	jsonData, err := json.Marshal(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":  "error",
			"error": map[string]string{"type": "api_error", "message": fmt.Sprintf("序列化请求失败: %v", err)},
		})
		return err
	}

	log.Infof("请求 Notion AI URL: %s", p.apiEndpoints["runInference"])

	req, err := http.NewRequest("POST", p.apiEndpoints["runInference"], bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":  "error",
			"error": map[string]string{"type": "api_error", "message": fmt.Sprintf("创建请求失败: %v", err)},
		})
		return err
	}

	for key, value := range p.prepareHeaders() {
		req.Header.Set(key, value)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":  "error",
			"error": map[string]string{"type": "api_error", "message": fmt.Sprintf("请求 Notion AI 失败: %v", err)},
		})
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Errorf("Notion AI 返回错误，状态码: %d, 响应: %s", resp.StatusCode, string(bodyBytes))
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":  "error",
			"error": map[string]string{"type": "api_error", "message": fmt.Sprintf("Notion AI 返回错误状态码: %d", resp.StatusCode)},
		})
		return fmt.Errorf("状态码: %d", resp.StatusCode)
	}

	// 处理响应
	var incrementalFragments []string
	var finalMessage string

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parsedResults := p.parseNDJSONLine(line)
		for _, result := range parsedResults {
			textType, _ := result["type"].(string)
			content, _ := result["content"].(string)

			if textType == "error" {
				errorMsg, _ := result["error"].(string)
				c.JSON(http.StatusPaymentRequired, gin.H{
					"type":  "error",
					"error": map[string]string{"type": "api_error", "message": errorMsg},
				})
				return fmt.Errorf(errorMsg)
			}

			if textType == "final" {
				finalMessage = content
			} else if textType == "incremental" {
				incrementalFragments = append(incrementalFragments, content)
			}
		}
	}

	// 确定最终响应
	var fullResponse string
	if finalMessage != "" {
		fullResponse = finalMessage
	} else if len(incrementalFragments) > 0 {
		fullResponse = strings.Join(incrementalFragments, "")
	}

	if fullResponse == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"type":  "error",
			"error": map[string]string{"type": "api_error", "message": "未能从 Notion 获取有效响应"},
		})
		return fmt.Errorf("空响应")
	}

	cleanedResponse := p.cleanContent(fullResponse)
	log.Infof("清洗后的最终响应: %s", cleanedResponse)

	messageID := fmt.Sprintf("msg_%s", uuid.New().String())

	if stream {
		// 流式响应 (Anthropic SSE 格式)
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")

		// message_start 事件
		messageStart := map[string]interface{}{
			"type": "message_start",
			"message": map[string]interface{}{
				"id":           messageID,
				"type":         "message",
				"role":         "assistant",
				"content":      []interface{}{},
				"model":        modelName,
				"stop_reason":  nil,
				"stop_sequence": nil,
				"usage":        map[string]int{"input_tokens": 0, "output_tokens": 0},
			},
		}
		c.Writer.Write([]byte(fmt.Sprintf("event: message_start\ndata: %s\n\n", mustMarshal(messageStart))))

		// content_block_start 事件
		contentBlockStart := map[string]interface{}{
			"type":  "content_block_start",
			"index": 0,
			"content_block": map[string]interface{}{
				"type": "text",
				"text": "",
			},
		}
		c.Writer.Write([]byte(fmt.Sprintf("event: content_block_start\ndata: %s\n\n", mustMarshal(contentBlockStart))))

		// content_block_delta 事件
		contentBlockDelta := map[string]interface{}{
			"type":  "content_block_delta",
			"index": 0,
			"delta": map[string]interface{}{
				"type": "text_delta",
				"text": cleanedResponse,
			},
		}
		c.Writer.Write([]byte(fmt.Sprintf("event: content_block_delta\ndata: %s\n\n", mustMarshal(contentBlockDelta))))

		// content_block_stop 事件
		contentBlockStop := map[string]interface{}{
			"type":  "content_block_stop",
			"index": 0,
		}
		c.Writer.Write([]byte(fmt.Sprintf("event: content_block_stop\ndata: %s\n\n", mustMarshal(contentBlockStop))))

		// message_delta 事件
		messageDelta := map[string]interface{}{
			"type": "message_delta",
			"delta": map[string]interface{}{
				"stop_reason":   "end_turn",
				"stop_sequence": nil,
			},
			"usage": map[string]int{"output_tokens": len(cleanedResponse)},
		}
		c.Writer.Write([]byte(fmt.Sprintf("event: message_delta\ndata: %s\n\n", mustMarshal(messageDelta))))

		// message_stop 事件
		messageStop := map[string]interface{}{
			"type": "message_stop",
		}
		c.Writer.Write([]byte(fmt.Sprintf("event: message_stop\ndata: %s\n\n", mustMarshal(messageStop))))

		c.Writer.Flush()
	} else {
		// 非流式响应 (Anthropic 格式)
		response := map[string]interface{}{
			"id":   messageID,
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": cleanedResponse,
				},
			},
			"model":         modelName,
			"stop_reason":   "end_turn",
			"stop_sequence": nil,
			"usage": map[string]int{
				"input_tokens":  0,
				"output_tokens": len(cleanedResponse),
			},
		}
		c.JSON(http.StatusOK, response)
	}

	return nil
}

// mustMarshal JSON 序列化，忽略错误
func mustMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
