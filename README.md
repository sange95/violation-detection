# VLLM Show 服务

一个基于 Gin 框架的图片上传和 VLLM 推理服务，提供文件下载和图片识别功能。

## 功能特性

- ✅ 文件服务器：提供静态文件下载服务
- ✅ 图片上传API：支持上传图片并存储到本地
- ✅ VLLM推理集成：上传后自动调用VLLM进行图片识别
- ✅ 双端口服务：文件服务器和API服务器独立运行
- ✅ 支持自定义提示词

## 服务架构

```
┌─────────────────┐
│  文件服务器      │ 端口: 8080 (默认)
│  (File Server)  │ 提供文件下载
└─────────────────┘
         ↑
         │ 文件URL
         │
┌─────────────────┐
│  API服务器       │ 端口: 8081 (默认)
│  (API Server)   │ 处理上传和推理
└─────────────────┘
         │
         ↓
┌─────────────────┐
│  VLLM服务        │ 端口: 8001 (默认)
│  (推理服务)      │ 图片内容识别
└─────────────────┘
```

## 环境变量配置

### 必需配置

- `VLLM_BASE_URL`: VLLM服务地址（默认: http://localhost:8001）

### 可选配置

- `FILE_SERVER_PORT`: 文件服务器端口（默认: 8080）
- `FILE_SERVER_PATH`: 文件存储路径（默认: ./uploads）
- `API_SERVER_PORT`: API服务器端口（默认: 8081）
- `FILE_SERVER_HOST`: 文件服务器主机地址，用于生成文件URL（默认: localhost:8080）
- `VLLM_MODEL`: VLLM模型名称（可选）
- `VLLM_PROMPT`: 默认提示词（默认: "请检测该图片中内容。"）

## 快速开始

### 方式一：直接运行

#### 1. 安装依赖

```bash
go mod download
```

#### 2. 运行服务

```bash
# 设置环境变量（可选）
export VLLM_BASE_URL=http://localhost:8001
export FILE_SERVER_PORT=8080
export API_SERVER_PORT=8081

# 运行服务
go run .
```

### 方式二：使用Docker

#### 1. 构建镜像

```bash
docker build -t vllm-show .
```

#### 2. 运行容器

```bash
docker run -d \
  --name vllm-show \
  -p 8080:8080 \
  -p 8081:8081 \
  -e VLLM_BASE_URL=http://host.docker.internal:8001 \
  -e FILE_SERVER_HOST=localhost:8080 \
  -v $(pwd)/uploads:/app/uploads \
  vllm-show
```

#### 3. 使用Docker Compose（推荐）

```bash
# 编辑 docker-compose.yml 中的环境变量
# 然后运行
docker-compose up -d
```

**注意**：
- 如果VLLM服务在Docker容器外运行，使用 `http://host.docker.internal:8001`
- 如果VLLM服务在同一Docker网络中，使用服务名和端口，如 `http://vllm-service:8001`

### 3. 测试上传接口

```bash
# 上传图片
curl -X POST http://localhost:8081/api/upload \
  -F "file=@/path/to/your/image.jpg" \
  -F "prompt=请检测该图片中内容。"

# 或者使用自定义提示词
curl -X POST http://localhost:8081/api/upload \
  -F "file=@/path/to/your/image.jpg" \
  -F "prompt=这是什么？"
```

### 4. 访问上传的文件

上传成功后，可以通过文件服务器访问：

```
http://localhost:8080/files/<filename>
```

## API接口说明

### 上传图片

**接口**: `POST /api/upload`

**请求参数**:
- `file` (multipart/form-data): 图片文件（必需）
- `prompt` (form-data): 自定义提示词（可选）

**支持的文件格式**: jpg, jpeg, png, gif, webp

**响应示例**:
```json
{
  "success": true,
  "message": "文件上传成功并完成推理",
  "image_url": "http://localhost:8080/files/1234567890_image.jpg",
  "inference": {
    "content": "这是VLLM返回的识别结果...",
    "usage": {
      "prompt_tokens": 100,
      "completion_tokens": 50,
      "total_tokens": 150
    }
  }
}
```

### 健康检查

**文件服务器**: `GET http://localhost:8080/health`

**API服务器**: `GET http://localhost:8081/health`

## VLLM服务要求

VLLM服务需要支持OpenAI兼容的API格式：

```json
POST /v1/chat/completions
{
  "model": "",
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "image_url",
          "image_url": {
            "url": "http://..."
          }
        },
        {
          "type": "text",
          "text": "请检测该图片中内容。"
        }
      ]
    }
  ]
}
```

## 使用示例

### Python示例

```python
import requests

url = "http://localhost:8081/api/upload"
files = {'file': open('image.jpg', 'rb')}
data = {'prompt': '请检测该图片中内容。'}

response = requests.post(url, files=files, data=data)
result = response.json()

print(f"上传成功: {result['success']}")
print(f"图片URL: {result['image_url']}")
print(f"识别结果: {result['inference']['content']}")
```

### JavaScript示例

```javascript
const formData = new FormData();
formData.append('file', fileInput.files[0]);
formData.append('prompt', '请检测该图片中内容。');

fetch('http://localhost:8081/api/upload', {
  method: 'POST',
  body: formData
})
.then(response => response.json())
.then(data => {
  console.log('上传成功:', data.success);
  console.log('图片URL:', data.image_url);
  console.log('识别结果:', data.inference.content);
});
```

## 项目结构

```
vllm-show/
├── main.go              # 主程序入口
├── config.go            # 配置管理
├── file_server.go       # 文件服务器
├── api_server.go        # API服务器
├── vllm_client.go       # VLLM客户端
├── go.mod               # 依赖管理
└── README.md            # 文档
```

## 注意事项

1. 确保VLLM服务已启动并可访问
2. 文件服务器需要能够访问存储的文件路径
3. 上传的文件会保存在 `FILE_SERVER_PATH` 指定的目录
4. 文件服务器主机地址需要正确配置，以便生成可访问的文件URL
5. 如果VLLM服务在远程，确保文件URL可以被VLLM服务访问

## 开发

### 运行测试

```bash
go test ./...
```

### 构建

```bash
go build -o vllm-show .
```

### 运行

```bash
./vllm-show
```

