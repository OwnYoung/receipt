# 收据生成服务

基于 Golang 的收据 PDF 生成服务，接收小程序请求，直接生成标准收据 PDF 文件。

## 功能特性

- ✅ 接收小程序请求（租金、房间号、收款人等信息）
- ✅ 直接生成标准收据 PDF（无需模板）
- ✅ 返回 PDF 文件或 Base64 编码（适配小程序）
- ✅ 支持自定义日期和收费目的
- ✅ RESTful API 设计
- ✅ CORS 跨域支持
- ✅ 自动清理临时文件

## 项目结构

```
receipt/
├── cmd/                    # 应用程序入口
│   └── main.go
├── internal/              # 内部包
│   ├── handler/          # HTTP 处理器
│   │   └── receipt_handler.go
│   ├── service/          # 业务逻辑
│   │   └── pdf_service.go
│   └── model/            # 数据模型
│       └── receipt.go
├── fonts/                # 中文字体文件（如 FangZhengFangSong-GBK-1.ttf）
├── output/               # 生成的PDF输出目录
├── go.mod                # Go模块文件
└── README.md             # 说明文档
```

## API 接口

### 1. 生成收据 PDF（直接下载）

**POST** `/api/receipt/generate`

请求体：
```json
{
  "rent": 1500.00,
  "room_number": "101",
  "recipient": "张三",
  "payer": "李四",
  "date": "2025-09-21",
  "month": "2025年9月",
  "purpose": "房租"
}
```

响应：返回 PDF 文件（Content-Type: application/pdf）

### 2. 生成收据 PDF（小程序Base64接口）

**POST** `/api/receipt/miniprogram`

请求体：同上

响应：
```json
{
  "success": true,
  "message": "收据生成成功",
  "data": {
    "receiptId": "NO101202509",
    "fileName": "receipt_101_20250921_143022.pdf",
    "fileSize": 15234,
    "pdfBase64": "JVBERi0xLjQKJcOkw7zDtsOkdwoXZnNlcmdsZXJ0...",
    "contentType": "application/pdf",
    "generateTime": "2025-09-21 14:30:22"
  }
}
```

### 3. 预览收据信息

**POST** `/api/receipt/info`

请求体：同上

响应：
```json
{
  "success": true,
  "message": "获取收据信息成功",
  "data": {
    "id": "NO101202509",
    "rent": "1500.00",
    "rent_zh": "壹仟伍佰元整",
    "room_number": "101",
    "recipient": "张三",
    "payer": "李四",
    "date": "2025-09-21",
    "month": "2025年9月",
    "purpose": "房租",
    "created_at": "2025-09-21T10:30:00Z"
  }
}
```

### 4. 健康检查

**GET** `/health`

响应：
```json
{
  "status": "ok",
  "message": "收据服务运行正常",
  "service": "receipt-service"
}
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 准备字体文件

将中文字体文件（如 FangZhengFangSong-GBK-1.ttf）放置在 `fonts/` 目录下。

### 3. 运行服务

```bash
go run cmd/main.go
```

服务将在 `http://localhost:8090` 启动

### 4. 测试服务

健康检查：
```bash
curl http://localhost:8090/health
```

生成收据（下载PDF）：
```bash
curl -X POST http://localhost:8090/api/receipt/generate \
  -H "Content-Type: application/json" \
  -d '{
    "rent": 1500.00,
    "room_number": "101",
    "recipient": "张三",
    "payer": "李四"
  }' \
  --output receipt.pdf
```

生成收据（小程序Base64接口）：
```bash
curl -X POST http://localhost:8090/api/receipt/miniprogram \
  -H "Content-Type: application/json" \
  -d '{
    "rent": 1500.00,
    "room_number": "101",
    "recipient": "张三",
    "payer": "李四"
  }'
```

## 环境变量

可以通过环境变量配置：

- `PORT` - 服务端口（默认：8090）
- `OUTPUT_PATH` - 输出目录（默认：output）

## 技术栈

- **框架**: Gin (HTTP Web Framework)
- **PDF处理**: gopdf (PDF生成库)
- **语言**: Go 1.21+

## 部署建议

### Docker 部署

创建 `Dockerfile`：
```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o receipt-service cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/receipt-service .
COPY --from=builder /app/fonts ./fonts

EXPOSE 8090
CMD ["./receipt-service"]
```

构建和运行：
```bash
docker build -t receipt-service .
docker run -p 8090:8090 -v $(pwd)/output:/root/output receipt-service
```

## 注意事项

1. **字体文件**: 确保 `fonts/` 目录下有可用的中文字体文件
2. **文件权限**: 确保输出目录有写入权限
3. **内存使用**: 大量并发请求时注意内存使用情况
4. **文件清理**: 服务自动清理临时文件，建议定期检查输出目录

## 许可证

MIT License