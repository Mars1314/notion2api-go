# notion-2api-go

> 一个将 Notion AI 转换为兼容 OpenAI 格式 API 的高性能代理服务（Go 实现）

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)

## ✨ 特性

- 🚀 **高性能**：基于 Go 语言开发，内存占用低，响应速度快
- 🔄 **流式响应**：支持 SSE（Server-Sent Events）流式输出
- 🤖 **多模型支持**：支持 Claude、GPT、Gemini 系列模型
- 🔌 **OpenAI 兼容**：完全兼容 OpenAI API 格式，无缝集成
- 🐳 **Docker 支持**：提供 Docker 和 docker-compose 部署方案
- 📝 **详细日志**：完善的日志记录，便于调试和监控
- 🛡️ **安全可靠**：支持 API Key 认证，保护服务安全

## 📋 支持的模型

### ✅ 可用模型（4/6）- 成功率 67%

| 模型名称 | 状态 | 说明 | 推荐场景 |
|---------|------|------|---------|
| `claude-sonnet-4.5` | ✅ 可用 | Claude Sonnet 4.5 | 通用任务（**强烈推荐**） |
| `gpt-5` | ✅ 可用 | GPT-5 | 高级推理 |
| `claude-opus-4.1` | ✅ 可用 | Claude Opus 4.1 | 复杂任务 |
| `gpt-4.1` | ✅ 可用 | GPT-4.1 | 快速响应 |

### ⚠️ 暂不可用（2/6）

| 模型名称 | 状态 | 说明 |
|---------|------|------|
| `gemini-2.5-flash` | ⚠️ 不可用 | Notion AI 返回空响应 |
| `gemini-2.5-pro` | ⚠️ 不可用 | Notion AI 返回空响应 |

> **注意**: Gemini 系列模型目前从 Notion AI 获取响应时返回空内容，可能是 Notion AI 对这些模型的支持问题。建议使用 Claude 或 GPT 系列模型。

## 🚀 快速开始

### 方式一：使用启动脚本（推荐）

```bash
# 1. 配置环境变量
cp .env.example .env
vim .env  # 填入你的 Notion 凭证

# 2. 赋予执行权限
chmod +x start.sh stop.sh

# 3. 启动服务（自动检测 Go 或 Docker 环境）
./start.sh

# 4. 停止服务
./stop.sh
```

### 方式二：使用 Docker Compose

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 方式三：本地 Go 运行

```bash
# 安装依赖
go mod download

# 构建
go build -o notion-2api main.go

# 运行
./notion-2api
```

## 🔧 配置说明

### 环境变量

在 `.env` 文件中配置以下变量：

| 变量名 | 默认值 | 说明 | 必需 |
|--------|--------|------|------|
| `NGINX_PORT` | 8004 | 服务端口 | 否 |
| `API_MASTER_KEY` | - | API 认证密钥 | 是 |
| `NOTION_COOKIE` | - | Notion Cookie（token_v2） | 是 |
| `NOTION_SPACE_ID` | - | Notion 空间 ID | 是 |
| `NOTION_USER_ID` | - | Notion 用户 ID | 是 |
| `NOTION_USER_NAME` | - | Notion 用户名称 | 否 |
| `NOTION_USER_EMAIL` | - | Notion 用户邮箱 | 否 |
| `NOTION_BLOCK_ID` | - | Notion 块 ID（可选） | 否 |
| `DEFAULT_MODEL` | claude-sonnet-4.5 | 默认使用的模型 | 否 |
| `API_REQUEST_TIMEOUT` | 180 | API 请求超时时间（秒） | 否 |

### 获取 Notion 凭证

1. **获取 Cookie (token_v2)**
   - 登录 [Notion](https://www.notion.so/)
   - 打开浏览器开发者工具（F12）
   - 进入 Application/Storage → Cookies
   - 复制 `token_v2` 的值

2. **获取 Space ID**
   - 访问任意 Notion 页面
   - 在开发者工具的 Network 标签页中查找请求
   - 在请求头中找到 `x-notion-space-id`

3. **获取 User ID**
   - 在 Network 标签页中找到任意 API 请求
   - 在请求头中找到 `x-notion-active-user-header`

## 📖 API 使用

### 在线文档

访问 [http://localhost:8004/docs](http://localhost:8004/docs) 查看完整的交互式文档。

### 健康检查

```bash
curl http://localhost:8004/
```

### 获取模型列表

```bash
curl http://localhost:8004/v1/models \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### 聊天补全（流式）

```bash
curl -X POST http://localhost:8004/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4.5",
    "messages": [
      {"role": "user", "content": "你好，介绍一下你自己"}
    ],
    "stream": true
  }'
```

### 聊天补全（非流式）

```bash
curl -X POST http://localhost:8004/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4.5",
    "messages": [
      {"role": "user", "content": "什么是量子计算？"}
    ],
    "stream": false
  }'
```

## 🔌 集成示例

### Python (OpenAI SDK)

```python
from openai import OpenAI

client = OpenAI(
    api_key="YOUR_API_KEY",
    base_url="http://localhost:8004/v1"
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
        print(chunk.choices[0].delta.content, end="")
```

### Node.js

```javascript
import OpenAI from 'openai';

const client = new OpenAI({
    apiKey: 'YOUR_API_KEY',
    baseURL: 'http://localhost:8004/v1'
});

const response = await client.chat.completions.create({
    model: 'claude-sonnet-4.5',
    messages: [
        { role: 'user', content: '你好' }
    ],
    stream: true
});

for await (const chunk of response) {
    process.stdout.write(chunk.choices[0]?.delta?.content || '');
}
```

### cURL

```bash
curl -X POST http://localhost:8004/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [
      {"role": "user", "content": "解释一下人工智能"}
    ],
    "stream": false
  }'
```

## 🏗️ 项目结构

```
notion-2api-go/
├── cmd/                    # 命令行工具
├── internal/              # 内部包
│   ├── config/           # 配置管理
│   ├── providers/        # AI 提供者实现
│   └── utils/            # 工具函数
├── main.go               # 主程序入口
├── go.mod                # Go 模块定义
├── go.sum                # 依赖版本锁定
├── Dockerfile            # Docker 镜像构建
├── docker-compose.yml    # Docker Compose 配置
├── start.sh              # 启动脚本
├── stop.sh               # 停止脚本
├── test_api.sh           # API 测试脚本
├── .env.example          # 环境变量模板
└── README.md             # 项目文档
```

## 🛠️ 开发指南

### 本地开发

```bash
# 克隆仓库
git clone <repository-url>
cd notion-2api-go

# 安装依赖
go mod download

# 运行（开发模式）
go run main.go

# 构建
go build -o notion-2api main.go

# 运行测试
go test ./...
```

### 构建 Docker 镜像

```bash
# 构建镜像
docker build -t notion-2api-go:latest .

# 运行容器
docker run -d \
  --name notion-2api \
  -p 8004:8004 \
  --env-file .env \
  notion-2api-go:latest
```

## 🔍 故障排查

### 服务无法启动

1. 检查环境变量配置是否正确
```bash
cat .env
```

2. 查看服务日志
```bash
tail -f notion-2api.log
# 或
docker-compose logs -f
```

3. 检查端口占用
```bash
lsof -i :8004
```

### API 请求失败

1. 验证 API Key 是否正确
2. 检查 Notion Cookie 是否过期
3. 确认模型名称是否正确

### Notion Cookie 过期

Notion Cookie 会定期过期，需要重新获取并更新 `.env` 文件中的 `NOTION_COOKIE`。

### 性能问题

- 增加 `API_REQUEST_TIMEOUT` 值
- 检查网络连接
- 查看系统资源使用情况

## 📊 性能指标

- **内存占用**：约 20-30 MB
- **响应时间**：平均 5-8 秒（取决于 Notion API 和模型）
- **并发支持**：支持多用户同时访问
- **稳定性**：7x24 小时运行
- **模型成功率**：67% (4/6 模型可用)

## ✅ 测试结果

最新测试时间：2025-10-13

| 测试项 | 状态 | 说明 |
|--------|------|------|
| 服务启动 | ✅ 通过 | 服务正常启动在端口 8004 |
| 健康检查 | ✅ 通过 | 根路径响应正常 |
| 模型列表 | ✅ 通过 | 成功返回 6 个模型 |
| claude-sonnet-4.5 | ✅ 通过 | 流式响应正常 |
| gpt-5 | ✅ 通过 | 流式响应正常 |
| claude-opus-4.1 | ✅ 通过 | 流式响应正常 |
| gpt-4.1 | ✅ 通过 | 流式响应正常 |
| gemini-2.5-flash | ⚠️ 失败 | 返回空响应 |
| gemini-2.5-pro | ⚠️ 失败 | 返回空响应 |

### 推荐使用模型

基于测试结果，推荐使用以下模型：
1. **claude-sonnet-4.5** - 最稳定，响应质量高
2. **gpt-5** - 高级推理能力强
3. **claude-opus-4.1** - 适合复杂任务
4. **gpt-4.1** - 快速响应

## 🔒 安全建议

1. **使用强 API Key**：设置复杂的 `API_MASTER_KEY`
2. **限制访问**：通过防火墙限制访问 IP
3. **HTTPS 部署**：生产环境使用 HTTPS
4. **定期更新**：及时更新 Notion Cookie
5. **日志监控**：定期检查日志文件

## 🤝 贡献

我们欢迎所有形式的贡献！

### 如何贡献

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个 Pull Request

### 贡献类型

- 🐛 报告 Bug
- 💡 提出新功能建议
- 📝 改进文档
- 🔧 提交代码修复
- ⭐ Star 支持项目

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源协议。

Copyright (c) 2025 [libaxuan](https://github.com/libaxuan)

## 🔗 相关链接

- **GitHub 仓库**: https://github.com/libaxuan/notion2api-go
- **在线文档**: http://localhost:8004/docs (启动服务后访问)
- **问题反馈**: https://github.com/libaxuan/notion2api-go/issues
- **贡献代码**: https://github.com/libaxuan/notion2api-go/pulls

欢迎 Star ⭐ 和 Fork 🍴 支持项目！

## 🙏 致谢

- [Notion AI](https://www.notion.so/product/ai) - 提供强大的 AI 能力
- [OpenAI API](https://platform.openai.com/) - API 格式标准
- [Gin](https://github.com/gin-gonic/gin) - Web 框架
- [Logrus](https://github.com/sirupsen/logrus) - 日志库

## 👨‍💻 作者

**libaxuan**

- GitHub: [@libaxuan](https://github.com/libaxuan)
- 项目主页: [notion2api-go](https://github.com/libaxuan/notion2api-go)

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 [Issue](https://github.com/libaxuan/notion2api-go/issues)
- 发起 [Pull Request](https://github.com/libaxuan/notion2api-go/pulls)
- GitHub 讨论: [Discussions](https://github.com/libaxuan/notion2api-go/discussions)

---

**⚠️ 免责声明**：本项目仅供学习和研究使用，请遵守 Notion 的服务条款。