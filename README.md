# 收据生成服务

基于 Golang 的收据 PDF 生成服务，接收小程序请求，使用 acroForm PDF 模板生成收据。

## 功能特性

- ✅ 接收小程序请求（租金、房间号、收款人等信息）
- ✅ 基于 acroForm PDF 模板自动填充
- ✅ 返回生成的收据 PDF 文件
- ✅ 支持自定义日期和收费目的
- ✅ RESTful API 设计
- ✅ CORS 跨域支持

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
├── templates/            # PDF 模板文件
│   └── receipt_template.pdf
├── output/              # 生成的PDF输出目录
├── go.mod              # Go模块文件
└── README.md           # 说明文档
```

## API 接口

### 1. 生成收据 PDF

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

### 2. 预览收据信息

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

### 3. 健康检查

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

### 2. 准备 PDF 模板

将您的 acroForm PDF 模板文件放置在 `templates/receipt_template.pdf`

模板中的表单字段名称应该包括：
- `id` - 收据编号（格式：NO+房间号+月份）
- `rent` - 租金金额
- `rent_zh` - 租金中文大写金额
- `room_number` - 房间号
- `recipient` - 收款人
- `payer` - 付款人
- `date` - 收据日期
- `month` - 租金月份
- `purpose` - 收费目的

### 3. 运行服务

```bash
go run cmd/main.go
```

服务将在 `http://localhost:8080` 启动

### 4. 测试服务

健康检查：
```bash
curl http://localhost:8080/health
```

生成收据：
```bash
curl -X POST http://localhost:8080/api/receipt/generate \
  -H "Content-Type: application/json" \
  -d '{
    "rent": 1500.00,
    "room_number": "101",
    "recipient": "张三",
    "payer": "李四"
  }' \
  --output receipt.pdf
```

## 环境变量

可以通过环境变量配置：

- `PORT` - 服务端口（默认：8080）
- `TEMPLATE_PATH` - PDF模板路径（默认：templates/receipt_template.pdf）
- `OUTPUT_PATH` - 输出目录（默认：output）

## 技术栈

- **框架**: Gin (HTTP Web Framework)
- **PDF处理**: pdfcpu (PDF处理库)
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
COPY --from=builder /app/templates ./templates

EXPOSE 8080
CMD ["./receipt-service"]
```

构建和运行：
```bash
docker build -t receipt-service .
docker run -p 8080:8080 -v $(pwd)/output:/root/output receipt-service
```

## 注意事项

1. **PDF 模板**: 确保 PDF 模板包含正确的表单字段名称
2. **文件权限**: 确保输出目录有写入权限
3. **内存使用**: 大量并发请求时注意内存使用情况
4. **文件清理**: 建议定期清理输出目录中的旧文件

## 许可证

MIT License