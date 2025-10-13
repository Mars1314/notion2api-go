
package utils

import "fmt"

// GetDocsHTML 返回文档页面 HTML
func GetDocsHTML(appName, appVersion string, port int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - API 文档</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 15px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 40px;
            text-align: center;
        }
        .header h1 { font-size: 2.5em; margin-bottom: 10px; }
        .header p { font-size: 1.2em; opacity: 0.9; }
        .content { padding: 40px; }
        .section { margin-bottom: 40px; }
        .section h2 {
            color: #667eea;
            font-size: 1.8em;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 3px solid #667eea;
        }
        .section h3 { color: #764ba2; font-size: 1.3em; margin: 20px 0 10px 0; }
        .model-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .model-card {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 20px;
            border-left: 4px solid #667eea;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .model-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.2);
        }
        .model-card.success { border-left-color: #28a745; background: #f0f9f4; }
        .model-card.warning { border-left-color: #ffc107; background: #fff9e6; }
        .model-name {
            font-size: 1.2em;
            font-weight: bold;
            color: #333;
            margin-bottom: 8px;
        }
        .model-status {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: 600;
            margin-bottom: 10px;
        }
        .status-success { background: #28a745; color: white; }
        .status-warning { background: #ffc107; color: #333; }
        .model-desc { color: #666; font-size: 0.95em; line-height: 1.5; }
        .code-block {
            background: #2d2d2d;
            color: #f8f8f2;
            padding: 20px;
            border-radius: 8px;
            overflow-x: auto;
            margin: 15px 0;
            font-family: "Courier New", monospace;
            font-size: 0.9em;
            line-height: 1.5;
        }
        .endpoint {
            background: #e3f2fd;
            padding: 15px;
            border-radius: 8px;
            margin: 10px 0;
            border-left: 4px solid #2196f3;
        }
        .endpoint .method {
            display: inline-block;
            background: #2196f3;
            color: white;
            padding: 4px 12px;
            border-radius: 4px;
            font-weight: bold;
            margin-right: 10px;
        }
        .endpoint .path {
            font-family: monospace;
            font-size: 1.1em;
            color: #1976d2;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .stat-card {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            text-align: center;
        }
        .stat-value { font-size: 2.5em; font-weight: bold; margin: 10px 0; }
        .stat-label { opacity: 0.9; }
        .footer {
            background: #f8f9fa;
            padding: 20px;
            text-align: center;
            color: #666;
            border-top: 1px solid #dee2e6;
        }
        .info-box {
            background: #fff3cd;
            border: 1px solid #ffc107;
            border-radius: 8px;
            padding: 15px;
            margin: 15px 0;
        }
        .info-box strong { color: #856404; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 %s</h1>
            <p>版本 %s | OpenAI 兼容 API</p>
        </div>
        
        <div class="content">
            <div class="section">
                <h2>📊 服务状态</h2>
                <div class="stats">
                    <div class="stat-card">
                        <div class="stat-label">服务端口</div>
                        <div class="stat-value">%d</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">可用模型</div>
                        <div class="stat-value">6</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">成功率</div>
                        <div class="stat-value">67%%</div>
                    </div>
                </div>
            </div>

            <div class="section">
                <h2>🤖 支持的模型</h2>
                
                <h3>✅ 可用模型 (4/6)</h3>
                <div class="model-grid">
                    <div class="model-card success">
                        <div class="model-name">claude-sonnet-4.5</div>
                        <span class="model-status status-success">✓ 可用</span>
                        <div class="model-desc">
                            <strong>推荐使用</strong><br>
                            Claude Sonnet 4.5 - 通用任务的最佳选择，平衡性能与质量
                        </div>
                    </div>
                    
                    <div class="model-card success">
                        <div class="model-name">gpt-5</div>
                        <span class="model-status status-success">✓ 可用</span>
                        <div class="model-desc">
                            GPT-5 - 最新一代模型，适合高级推理和复杂任务
                        </div>
                    </div>
                    
                    <div class="model-card success">
                        <div class="model-name">claude-opus-4.1</div>
                        <span class="model-status status-success">✓ 可用</span>
                        <div class="model-desc">
                            Claude Opus 4.1 - 适合需要深度思考的复杂任务
                        </div>
                    </div>
                    
                    <div class="model-card success">
                        <div class="model-name">gpt-4.1</div>
                        <span class="model-status status-success">✓ 可用</span>
                        <div class="model-desc">
                            GPT-4.1 - 快速响应，适合常规对话和任务
                        </div>
                    </div>
                </div>
                
                <h3>⚠️ 暂不可用 (2/6)</h3>
                <div class="info-box">
                    <strong>注意：</strong> 以下 Gemini 模型目前从 Notion AI 
返回空响应，可能是 Notion AI 的支持问题。
                </div>
                <div class="model-grid">
                    <div class="model-card warning">
                        <div class="model-name">gemini-2.5-flash</div>
                        <span class="model-status status-warning">⚠ 暂不可用</span>
                        <div class="model-desc">
                            Gemini 2.5 Flash - 快速对话模型（目前返回空响应）
                        </div>
                    </div>
                    
                    <div class="model-card warning">
                        <div class="model-name">gemini-2.5-pro</div>
                        <span class="model-status status-warning">⚠ 暂不可用</span>
                        <div class="model-desc">
                            Gemini 2.5 Pro - 高质量输出模型（目前返回空响应）
                        </div>
                    </div>
                </div>
            </div>

            <div class="section">
                <h2>📡 API 端点</h2>
                
                <div class="endpoint">
                    <span class="method">GET</span>
                    <span class="path">/</span>
                    <div style="margin-top: 10px;">健康检查端点</div>
                </div>
                
                <div class="endpoint">
                    <span class="method">GET</span>
                    <span class="path">/v1/models</span>
                    <div style="margin-top: 10px;">获取可用模型列表</div>
                </div>
                
                <div class="endpoint">
                    <span class="method">POST</span>
                    <span class="path">/v1/chat/completions</span>
                    <div style="margin-top: 10px;">聊天补全接口（支持流式和非流式）</div>
                </div>
            </div>

            <div class="section">
                <h2>💻 使用示例</h2>
                
                <h3>1. 获取模型列表</h3>
                <div class="code-block">curl http://localhost:%d/v1/models \
  -H "Authorization: Bearer YOUR_API_KEY"</div>

                <h3>2. 流式聊天（推荐）</h3>
                <div class="code-block">curl -N -X POST http://localhost:%d/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4.5",
    "messages": [
      {"role": "user", "content": "你好"}
    ],
    "stream": true
  }'</div>

                <h3>3. 非流式聊天</h3>
                <div class="code-block">curl -X POST http://localhost:%d/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5",
    "messages": [
      {"role": "user", "content": "解释量子计算"}
    ],
    "stream": false
  }'</div>

                <h3>4. Python 示例</h3>
                <div class="code-block">from openai import OpenAI

client = OpenAI(
    api_key="YOUR_API_KEY",
    base_url="http://localhost:%d/v1"
)

response = client.chat.completions.create(
    model="claude-sonnet-4.5",
    messages=[
        {"role": "user", "content": "你好"}
    ],
    stream=True
)

for chunk in response:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")</div>
            </div>

            <div class="section">
                <h2>🔑 认证</h2>
                <p>所有 API 请求都需要在请求头中包含 Authorization:</p>
                <div class="code-block">Authorization: Bearer YOUR_API_KEY</div>
                <p style="margin-top: 15px;">API Key 在 .env 文件的 <code>API_MASTER_KEY</code> 中配置。</p>
            </div>

            <div class="section">
                <h2>📝 测试建议</h2>
                <ul style="padding-left: 20px; line-height: 2;">
                    <li><strong>推荐使用：</strong> claude-sonnet-4.5, gpt-5, claude-opus-4.1, gpt-4.1</li>
                    <li><strong>避免使用：</strong> gemini-2.5-flash, gemini-2.5-pro（目前不可用）</li>
                    <li><strong>流式响应：</strong> 建议使用流式模式以获得更好的用户体验</li>
                    <li><strong>超时设置：</strong> 建议设置 180 秒以上的超时时间</li>
                </ul>
            </div>
        </div>
        
        <div class="footer">
            <p>%s v%s | Powered by Notion AI</p>
            <p style="margin-top: 10px;">📚 <a href="https://github.com/libaxuan/notion2api-go" style="color: #667eea; text-decoration: none;">查看 GitHub 仓库</a></p>
        </div>
    </div>
</body>
</html>`, appName, appName, appVersion, port, port, port, port, port, appName, appVersion)
}