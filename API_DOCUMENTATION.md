# 收据生成API接口文档

## 接口概览

服务地址：`http://localhost:8090`

## 1. 生成收据PDF文件 (直接下载)

**接口:** `POST /api/receipt/generate`

**用途:** 直接返回PDF文件，适合Web浏览器下载

**请求示例:**
```bash
curl -X POST http://localhost:8090/api/receipt/generate \
  -H "Content-Type: application/json" \
  -d '{
    "rent": 1500.00,
    "room_number": "101",
    "recipient": "张三",
    "payer": "李四",
    "date": "2025-09-21",
    "month": "2025年09月",
    "purpose": "房租"
  }' \
  --output receipt.pdf
```

**响应:** 直接返回PDF文件

---

## 2. 为小程序生成收据 (Base64编码) ⭐

**接口:** `POST /api/receipt/miniprogram`

**用途:** 返回Base64编码的PDF数据，适合小程序处理

**请求示例:**
```bash
curl -X POST http://localhost:8090/api/receipt/miniprogram \
  -H "Content-Type: application/json" \
  -d '{
    "rent": 1500.00,
    "room_number": "101",
    "recipient": "张三",
    "payer": "李四",
    "date": "2025-09-21",
    "month": "2025年09月",
    "purpose": "房租"
  }'
```

**响应示例:**
```json
{
  "success": true,
  "message": "收据生成成功",
  "data": {
    "receiptId": "NO10120250901",
    "fileName": "receipt_101_20250921_143022.pdf",
    "fileSize": 15234,
    "pdfBase64": "JVBERi0xLjQKJcOkw7zDtsOkdwoXZnNlcmdsZXJ0...",
    "contentType": "application/pdf",
    "generateTime": "2025-09-21 14:30:22"
  }
}
```

---

## 3. 预览收据信息

**接口:** `POST /api/receipt/info`

**用途:** 只返回处理后的收据信息，不生成PDF

**请求示例:**
```bash
curl -X POST http://localhost:8090/api/receipt/info \
  -H "Content-Type: application/json" \
  -d '{
    "rent": 1500.00,
    "room_number": "101",
    "recipient": "张三",
    "payer": "李四"
  }'
```

**响应示例:**
```json
{
  "success": true,
  "message": "获取收据信息成功",
  "data": {
    "id": "NO10120250901",
    "rent": "1500.00",
    "rent_zh": "壹仟伍佰元整",
    "room_number": "101",
    "recipient": "张三",
    "payer": "李四",
    "date": "2025-09-21",
    "month": "2025年09月",
    "purpose": "房租"
  }
}
```

---

## 4. 健康检查

**接口:** `GET /health`

**用途:** 检查服务状态

---

## 小程序集成建议

### 方案1: 使用Base64接口 (推荐)

```javascript
// 小程序代码示例
wx.request({
  url: 'http://localhost:8090/api/receipt/miniprogram',
  method: 'POST',
  header: {
    'Content-Type': 'application/json'
  },
  data: {
    rent: 1500.00,
    room_number: "101",
    recipient: "张三",
    payer: "李四",
    date: "2025-09-21",
    month: "2025年09月",
    purpose: "房租"
  },
  success: function(res) {
    if (res.data.success) {
      const pdfData = res.data.data;
      
      // 将Base64转换为临时文件
      const base64Data = pdfData.pdfBase64;
      const fileName = pdfData.fileName;
      
      // 保存到本地
      wx.getFileSystemManager().writeFile({
        filePath: wx.env.USER_DATA_PATH + '/' + fileName,
        data: base64Data,
        encoding: 'base64',
        success: function() {
          // 预览PDF
          wx.openDocument({
            filePath: wx.env.USER_DATA_PATH + '/' + fileName,
            fileType: 'pdf'
          });
        }
      });
    }
  }
});
```

### 方案2: 直接下载文件

```javascript
// 直接下载PDF文件
wx.downloadFile({
  url: 'http://localhost:8090/api/receipt/generate',
  method: 'POST',
  header: {
    'Content-Type': 'application/json'
  },
  // 注意：小程序的downloadFile不支持POST body，建议使用方案1
});
```

---

## 错误处理

所有接口在出错时返回统一格式：

```json
{
  "success": false,
  "message": "错误描述"
}
```

常见错误：
- 400: 请求参数错误
- 500: 服务器内部错误（字体加载失败、PDF生成失败等）

---

## 部署建议

1. **生产环境配置:**
   - 修改CORS设置，只允许小程序域名
   - 启用HTTPS
   - 配置适当的文件清理策略

2. **小程序域名配置:**
   - 在小程序管理后台添加服务器域名到request合法域名
   - 如果使用downloadFile，还需添加到download合法域名