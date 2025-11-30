#!/bin/bash

# 图片检测 API 测试脚本
# 使用方法: ./test_upload.sh <图片文件路径> [提示词]

# 配置
API_URL="${API_URL:-http://localhost:8081/api/upload}"
IMAGE_FILE="${1:-}"
PROMPT="${2:-}"

# 检查参数
if [ -z "$IMAGE_FILE" ]; then
    echo "错误: 请提供图片文件路径"
    echo "使用方法: $0 <图片文件路径> [提示词]"
    echo "示例: $0 ./test.jpg \"请检测该图片中的违规内容\""
    exit 1
fi

# 检查文件是否存在
if [ ! -f "$IMAGE_FILE" ]; then
    echo "错误: 文件不存在: $IMAGE_FILE"
    exit 1
fi

# 构建 curl 命令
CURL_CMD="curl -X POST \"$API_URL\""

# 添加文件参数
CURL_CMD="$CURL_CMD -F \"file=@$IMAGE_FILE\""

# 如果提供了提示词，添加到请求中
if [ -n "$PROMPT" ]; then
    CURL_CMD="$CURL_CMD -F \"prompt=$PROMPT\""
fi

# 添加输出格式化和显示响应
CURL_CMD="$CURL_CMD -w \"\n\nHTTP状态码: %{http_code}\n\""

# 执行请求
echo "正在上传图片: $IMAGE_FILE"
if [ -n "$PROMPT" ]; then
    echo "提示词: $PROMPT"
fi
echo "API地址: $API_URL"
echo "----------------------------------------"
eval $CURL_CMD | python3 -m json.tool 2>/dev/null || eval $CURL_CMD
echo "----------------------------------------"

