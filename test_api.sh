#!/bin/bash

# notion-2api-go API 测试脚本

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 配置
API_URL="${API_URL:-http://localhost:8004}"
API_KEY="${API_KEY:-your_secret_key_here}"

echo "======================================"
echo "notion-2api-go API 测试脚本"
echo "======================================"
echo ""

# 测试1: 健康检查
echo -e "${YELLOW}测试 1: 健康检查${NC}"
echo "GET $API_URL/"
response=$(curl -s "$API_URL/")
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 服务运行正常${NC}"
    echo "响应: $response"
else
    echo -e "${RED}✗ 服务无响应${NC}"
    exit 1
fi
echo ""

# 测试2: 获取模型列表
echo -e "${YELLOW}测试 2: 获取模型列表${NC}"
echo "GET $API_URL/v1/models"
response=$(curl -s -H "Authorization: Bearer $API_KEY" "$API_URL/v1/models")
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 成功获取模型列表${NC}"
    echo "$response" | jq . 2>/dev/null || echo "$response"
else
    echo -e "${RED}✗ 获取模型列表失败${NC}"
fi
echo ""

# 测试3: 文档页面
echo -e "${YELLOW}测试 3: 文档页面${NC}"
echo "GET $API_URL/docs"
response=$(curl -s "$API_URL/docs" | head -5)
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 文档页面访问正常${NC}"
    echo "访问地址: $API_URL/docs"
else
    echo -e "${RED}✗ 文档页面访问失败${NC}"
fi
echo ""

# 模型测试数组
declare -a MODELS=(
    "claude-sonnet-4.5:推荐"
    "gpt-5:高级"
    "claude-opus-4.1:复杂"
    "gpt-4.1:快速"
    "gemini-2.5-flash:快速"
    "gemini-2.5-pro:高质量"
)

success_count=0
fail_count=0

echo -e "${YELLOW}开始测试所有模型...${NC}"
echo ""

for model_info in "${MODELS[@]}"; do
    IFS=':' read -r model desc <<< "$model_info"
    
    echo -e "${YELLOW}测试模型: ${model} (${desc})${NC}"
    
    # 发送请求并捕获响应
    response=$(curl -s -N -H "Content-Type: application/json" \
         -H "Authorization: Bearer $API_KEY" \
         -d "{
           \"model\": \"${model}\",
           \"messages\": [{\"role\": \"user\", \"content\": \"你好，请用一句话介绍你自己\"}],
           \"stream\": true
         }" \
         "$API_URL/v1/chat/completions" 2>/dev/null)
    
    # 检查是否包含错误
    if echo "$response" | grep -q "error"; then
        echo -e "${RED}✗ ${model} - 失败${NC}"
        echo "  错误: $(echo "$response" | grep -o '"message":"[^"]*"' | cut -d'"' -f4)"
        ((fail_count++))
    elif [ -z "$response" ]; then
        echo -e "${RED}✗ ${model} - 无响应${NC}"
        ((fail_count++))
    else
        echo -e "${GREEN}✓ ${model} - 成功${NC}"
        # 提取并显示响应内容（简化显示）
        content=$(echo "$response" | grep -o '"content":"[^"]*"' | head -1 | cut -d'"' -f4)
        if [ ! -z "$content" ]; then
            echo "  响应: ${content:0:50}..."
        fi
        ((success_count++))
    fi
    echo ""
    
    # 短暂延迟，避免请求过快
    sleep 1
done

echo "======================================"
echo -e "${GREEN}测试完成！${NC}"
echo "======================================"
echo ""
echo "测试统计："
echo -e "  成功: ${GREEN}${success_count}/6${NC}"
echo -e "  失败: ${RED}${fail_count}/6${NC}"
echo -e "  成功率: $(( success_count * 100 / 6 ))%"
echo ""
echo "提示："
echo "- 如果测试失败，请检查 .env 配置是否正确"
echo "- 查看详细日志: tail -f notion-2api.log"
echo "- 访问文档: $API_URL/docs"
echo "- 确保 Notion 凭证有效且未过期"