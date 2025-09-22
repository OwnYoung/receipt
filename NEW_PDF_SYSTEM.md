# 新的PDF生成系统

## 🎉 更新说明

现在使用 `gopdf` 库从头生成PDF，不再依赖PDF模板文件！

## ✨ 新功能特性

1. **无需PDF模板** - 程序自动生成完整的收据PDF
2. **中文支持** - 自动检测系统中文字体，支持中文显示
3. **备用方案** - 如果没有中文字体，使用英文标签生成PDF
4. **美观布局** - 专业的收据格式，包含所有必要信息

## 🚀 使用方法

### 启动服务
```bash
./receipt-service
```

服务将在端口 8090 启动

### API 请求示例
```bash
# 生成收据PDF
curl -X POST http://localhost:8090/api/receipt/generate \
  -H "Content-Type: application/json" \
  -d '{
    "rent": 1500.00,
    "room_number": "101",
    "recipient": "张三",
    "payer": "李四",
    "date": "2025-09-21",
    "month": "2025年9月",
    "purpose": "房租"
  }' \
  --output receipt.pdf

# 预览收据信息
curl -X POST http://localhost:8090/api/receipt/info \
  -H "Content-Type: application/json" \
  -d '{
    "rent": 1500.00,
    "room_number": "101",
    "recipient": "张三",
    "payer": "李四"
  }'
```

## 📄 PDF内容

生成的PDF包含以下信息：
- 收据编号（自动生成：NO+房间号+年月）
- 日期
- 收款人和房间号
- 收费项目和金额
- 金额中文大写
- 付款人
- 月份说明
- 生成时间戳

## 🔧 中文字体支持

程序会自动查找以下字体：
- `fonts/NotoSansCJK-Regular.ttc`
- `fonts/simhei.ttf`
- `fonts/simsun.ttf`
- macOS系统字体：`/System/Library/Fonts/PingFang.ttc`
- macOS系统字体：`/System/Library/Fonts/STHeiti Light.ttc`

如果需要更好的中文显示效果，可以将中文字体文件放在 `fonts/` 目录中。

## 💡 优势

1. **无依赖问题** - 不需要PDF模板文件
2. **自动化** - 完全程序化生成，格式统一
3. **灵活性** - 可以轻松修改布局和样式
4. **兼容性** - 即使没有中文字体也能生成可用的PDF
5. **性能** - 生成速度快，资源占用少

## 🛠️ 技术栈

- **PDF生成库**: gopdf
- **中文大写金额**: 自定义算法（已修复）
- **自动布局**: 响应式PDF布局
- **字体回退**: 智能字体检测和回退机制

您现在可以完全不依赖PDF模板，程序会自动生成专业格式的收据PDF！