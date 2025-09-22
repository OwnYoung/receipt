#!/bin/bash

# 收据API测试脚本

BASE_URL="http://localhost:8090"

echo "🚀 开始测试收据生成API..."

# 测试数据
TEST_DATA='{
  "rent": 1500.00,
  "room_number": "101",
  "recipient": "张三",
  "payer": "李四",
  "date": "2025-09-21",
  "month": "2025年09月",
  "purpose": "房租"
}'

echo ""
echo "📋 测试数据:"
echo "$TEST_DATA" | jq '.'

# 1. 健康检查
echo ""
echo "1️⃣ 测试健康检查..."
curl -s "$BASE_URL/health" | jq '.'

# 2. 测试预览信息
echo ""
echo "2️⃣ 测试预览信息..."
curl -s -X POST "$BASE_URL/api/receipt/info" \
  -H "Content-Type: application/json" \
  -d "$TEST_DATA" | jq '.'

# 3. 测试小程序接口（Base64）
echo ""
echo "3️⃣ 测试小程序接口 (Base64)..."
RESPONSE=$(curl -s -X POST "$BASE_URL/api/receipt/miniprogram" \
  -H "Content-Type: application/json" \
  -d "$TEST_DATA")

echo "$RESPONSE" | jq '. | del(.data.pdfBase64)'  # 隐藏Base64数据以便查看
echo "📄 PDF Base64数据长度: $(echo "$RESPONSE" | jq -r '.data.pdfBase64 // ""' | wc -c) 字符"

# 4. 测试直接下载PDF
echo ""
echo "4️⃣ 测试直接下载PDF..."
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
OUTPUT_FILE="test_receipt_$TIMESTAMP.pdf"

HTTP_CODE=$(curl -s -o "$OUTPUT_FILE" -w "%{http_code}" -X POST "$BASE_URL/api/receipt/generate" \
  -H "Content-Type: application/json" \
  -d "$TEST_DATA")

if [ "$HTTP_CODE" = "200" ]; then
    FILE_SIZE=$(stat -f%z "$OUTPUT_FILE" 2>/dev/null || stat -c%s "$OUTPUT_FILE" 2>/dev/null)
    echo "✅ PDF文件下载成功: $OUTPUT_FILE (大小: $FILE_SIZE 字节)"
    
    # 验证PDF文件
    if command -v file >/dev/null 2>&1; then
        FILE_TYPE=$(file "$OUTPUT_FILE")
        echo "📁 文件类型: $FILE_TYPE"
    fi
else
    echo "❌ PDF下载失败，HTTP状态码: $HTTP_CODE"
    cat "$OUTPUT_FILE"
    rm -f "$OUTPUT_FILE"
fi

# 5. 测试备份文件列表
echo ""
echo "5️⃣ 测试备份文件列表..."
curl -s "$BASE_URL/api/receipt/backup/list" | jq '.'

# 6. 等待一下让备份文件生成，然后再次查看备份列表
echo ""
echo "6️⃣ 等待备份文件生成后再次查看..."
sleep 2
BACKUP_RESPONSE=$(curl -s "$BASE_URL/api/receipt/backup/list")
echo "$BACKUP_RESPONSE" | jq '.'

# 如果有备份文件，测试下载第一个
FIRST_FILE=$(echo "$BACKUP_RESPONSE" | jq -r '.data.files[0].fileName // empty')
if [ ! -z "$FIRST_FILE" ] && [ "$FIRST_FILE" != "null" ]; then
    echo ""
    echo "7️⃣ 测试下载第一个备份文件: $FIRST_FILE"
    BACKUP_OUTPUT="backup_$TIMESTAMP.pdf"
    HTTP_CODE=$(curl -s -o "$BACKUP_OUTPUT" -w "%{http_code}" "$BASE_URL/api/receipt/backup/download/$FIRST_FILE")
    
    if [ "$HTTP_CODE" = "200" ]; then
        BACKUP_SIZE=$(stat -f%z "$BACKUP_OUTPUT" 2>/dev/null || stat -c%s "$BACKUP_OUTPUT" 2>/dev/null)
        echo "✅ 备份文件下载成功: $BACKUP_OUTPUT (大小: $BACKUP_SIZE 字节)"
        rm -f "$BACKUP_OUTPUT"
    else
        echo "❌ 备份文件下载失败，HTTP状态码: $HTTP_CODE"
        rm -f "$BACKUP_OUTPUT"
    fi
fi

echo ""
echo "🎉 测试完成！"

# 如果有jq命令，显示服务端点信息
echo ""
echo "📡 服务端点信息:"
curl -s "$BASE_URL/" | jq '.'