# 图片检测 API - curl 使用示例

## 基本用法

### 1. 上传图片（不带提示词）
```bash
curl -X POST http://localhost:8081/api/upload \
  -F "file=@/path/to/your/image.jpg"
```

### 2. 上传图片（带自定义提示词）
```bash
curl -X POST http://localhost:8081/api/upload \
  -F "file=@/path/to/your/image.jpg" \
  -F "prompt=请检测该图片中的违规内容"
```

### 3. 格式化输出（使用 jq）
```bash
curl -X POST http://localhost:8081/api/upload \
  -F "file=@/path/to/your/image.jpg" \
  -F "prompt=请检测该图片中的违规内容" | jq .
```

### 4. 保存响应到文件
```bash
curl -X POST http://localhost:8081/api/upload \
  -F "file=@/path/to/your/image.jpg" \
  -F "prompt=请检测该图片中的违规内容" \
  -o response.json
```

## 使用测试脚本

### 基本用法
```bash
./test_upload.sh ./test.jpg
```

### 带提示词
```bash
./test_upload.sh ./test.jpg "请检测该图片中的违规内容"
```

### 自定义 API 地址
```bash
API_URL=http://your-server:8081/api/upload ./test_upload.sh ./test.jpg
```

## 响应格式

成功响应示例：
```json
{
  "success": true,
  "message": "文件上传成功并完成推理",
  "image_url": "http://localhost:8080/files/filename.jpg",
  "inference": {
    "content": "检测结果...",
    "usage": {
      "prompt_tokens": 100,
      "completion_tokens": 50,
      "total_tokens": 150
    }
  }
}
```

错误响应示例：
```json
{
  "success": false,
  "error": "错误信息"
}
```

## 支持的图片格式

- .jpg / .jpeg
- .png
- .gif
- .webp

