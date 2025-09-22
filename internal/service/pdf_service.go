package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"receipt/internal/model"
	"regexp"
	"strings"
	"time"

	"github.com/gen2brain/go-fitz"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/signintech/gopdf"
)

type PDFService struct {
	templatePath string
	outputPath   string
}

func NewPDFService(templatePath, outputPath string) *PDFService {
	return &PDFService{
		templatePath: templatePath,
		outputPath:   outputPath,
	}
}

// FillReceipt 生成收据PDF
func (s *PDFService) FillReceipt(data *model.ReceiptData) (string, error) {
	// 生成输出文件名
	timestamp := time.Now().Format("20060102_150405")
	outputFileName := fmt.Sprintf("receipt_%s_%s.pdf", data.RoomNumber, timestamp)
	outputFilePath := filepath.Join(s.outputPath, outputFileName)

	// 确保输出目录存在
	if err := os.MkdirAll(s.outputPath, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 生成PDF
	if err := s.generatePDF(data, outputFilePath); err != nil {
		return "", fmt.Errorf("生成PDF失败: %v", err)
	}

	return outputFilePath, nil
}

// GenerateReceiptImage 生成收据图片 (先生成PDF再转换为图片)
func (s *PDFService) GenerateReceiptImage(data *model.ReceiptData) (string, error) {
	// 先生成PDF
	pdfPath, err := s.FillReceipt(data)
	if err != nil {
		return "", fmt.Errorf("生成PDF失败: %v", err)
	}

	// 使用go-fitz将PDF转换为图片
	imageFilePath, err := s.convertPDFToImage(pdfPath, data.RoomNumber)
	if err != nil {
		return "", fmt.Errorf("PDF转图片失败: %v", err)
	}

	return imageFilePath, nil
}

// convertPDFToImage 使用go-fitz将PDF转换为图片
func (s *PDFService) convertPDFToImage(pdfPath, roomNumber string) (string, error) {
	// 生成输出图片文件名
	timestamp := time.Now().Format("20060102_150405")
	imageFileName := fmt.Sprintf("receipt_%s_%s.png", roomNumber, timestamp)
	imageFilePath := filepath.Join(s.outputPath, imageFileName)

	// 打开PDF文档
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return "", fmt.Errorf("打开PDF文档失败: %v", err)
	}
	defer doc.Close()

	// 获取第一页作为图片
	img, err := doc.Image(0) // 0 是第一页
	if err != nil {
		return "", fmt.Errorf("获取PDF页面图片失败: %v", err)
	}

	// 保存为PNG格式
	f, err := os.Create(imageFilePath)
	if err != nil {
		return "", fmt.Errorf("创建图片文件失败: %v", err)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		return "", fmt.Errorf("保存PNG图片失败: %v", err)
	}

	return imageFilePath, nil
}

// GenerateReceiptImageBase64 生成收据图片并返回Base64编码
func (s *PDFService) GenerateReceiptImageBase64(data *model.ReceiptData) (string, error) {
	// 先生成PDF
	pdfPath, err := s.FillReceipt(data)
	if err != nil {
		return "", fmt.Errorf("生成PDF失败: %v", err)
	}

	// 使用go-fitz将PDF转换为图片
	base64String, err := s.convertPDFToImageBase64(pdfPath)
	if err != nil {
		return "", fmt.Errorf("PDF转图片失败: %v", err)
	}

	return base64String, nil
}

// convertPDFToImageBase64 将PDF转换为图片并返回Base64编码
func (s *PDFService) convertPDFToImageBase64(pdfPath string) (string, error) {
	// 打开PDF文档
	doc, err := fitz.New(pdfPath)
	if err != nil {
		return "", fmt.Errorf("打开PDF文档失败: %v", err)
	}
	defer doc.Close()

	// 获取第一页作为图片
	img, err := doc.Image(0) // 0 是第一页
	if err != nil {
		return "", fmt.Errorf("获取PDF页面图片失败: %v", err)
	}

	// 将图片编码为PNG格式并转换为Base64
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return "", fmt.Errorf("编码PNG图片失败: %v", err)
	}

	// 转换为Base64字符串
	base64String := base64.StdEncoding.EncodeToString(buf.Bytes())
	return base64String, nil
}

// generateReceiptImage 生成收据图片
func (s *PDFService) generateReceiptImage(data *model.ReceiptData, outputPath string) error {
	// 创建图片画布 (收据尺寸 176mm × 85mm，转换为像素，使用300DPI)
	// 176mm * 300DPI / 25.4 ≈ 2079 pixels
	// 85mm * 300DPI / 25.4 ≈ 1004 pixels
	width := 2079
	height := 1004

	// 创建RGBA图像
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充白色背景
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.ZP, draw.Src)

	// 绘制收据内容
	s.drawReceiptContent(img, data, width, height)

	// 保存图片
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建图片文件失败: %v", err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("编码PNG图片失败: %v", err)
	}

	return nil
}

// drawReceiptContent 在图片上绘制收据内容
func (s *PDFService) drawReceiptContent(img *image.RGBA, data *model.ReceiptData, width, height int) {
	// 加载中文字体
	fontBytes, err := os.ReadFile("fonts/FangZhengFangSong-GBK-1.ttf")
	if err != nil {
		// 如果加载中文字体失败，使用简化的英文绘制
		s.drawReceiptContentSimple(img, data, width, height)
		return
	}

	// 解析字体
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		s.drawReceiptContentSimple(img, data, width, height)
		return
	}

	// 创建FreeType context
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f)
	c.SetFontSize(24)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color.RGBA{0, 0, 0, 255}))

	// 绘制边框
	s.drawBorder(img, width, height)

	// 标题
	c.SetFontSize(36)
	pt := freetype.Pt(width/2-80, 120)
	c.DrawString("收款收据", pt)

	// 右上角信息
	c.SetFontSize(20)
	pt = freetype.Pt(width-400, 140)
	c.DrawString(fmt.Sprintf("收据号: %s", data.ID), pt)
	pt = freetype.Pt(width-400, 170)
	c.DrawString(fmt.Sprintf("日期: %s", data.Date), pt)

	// 主要内容区域
	c.SetFontSize(24)
	y := 250
	lineHeight := 70

	pt = freetype.Pt(120, y)
	c.DrawString(fmt.Sprintf("今收到 %s", data.Payer), pt)
	y += lineHeight

	pt = freetype.Pt(120, y)
	c.DrawString(fmt.Sprintf("交来: %s", data.Purpose), pt)
	y += lineHeight

	pt = freetype.Pt(120, y)
	c.DrawString(fmt.Sprintf("金额(大写) %s", data.RentZh), pt)
	y += lineHeight + 50

	// 小写金额
	c.SetFontSize(28)
	pt = freetype.Pt(120, y)
	c.DrawString(fmt.Sprintf("¥ %s", data.Rent), pt)

	// 支付方式
	c.SetFontSize(18)
	pt = freetype.Pt(width/2-300, y+50)
	c.DrawString("现金 □  转账 □  支票 □  微信支付宝 □", pt)

	// 右下角
	pt = freetype.Pt(width-250, y+50)
	c.DrawString("(盖章)", pt)
	pt = freetype.Pt(width-250, height-120)
	c.DrawString("经手人", pt)
}

// drawReceiptContentSimple 简化版本（备用方案）
func (s *PDFService) drawReceiptContentSimple(img *image.RGBA, data *model.ReceiptData, width, height int) {
	// 绘制边框
	s.drawBorder(img, width, height)

	// 使用简单的矩形来表示文本区域（当字体加载失败时）
	col := color.RGBA{128, 128, 128, 255}

	// 标题区域
	s.drawFilledRect(img, width/2-80, 80, 160, 40, col)

	// 内容区域
	y := 200
	lineHeight := 60

	for i := 0; i < 6; i++ {
		s.drawFilledRect(img, 100, y, width-200, 30, col)
		y += lineHeight
	}
}

// drawFilledRect 绘制填充矩形
func (s *PDFService) drawFilledRect(img *image.RGBA, x, y, w, h int, col color.Color) {
	for i := x; i < x+w && i < img.Bounds().Max.X; i++ {
		for j := y; j < y+h && j < img.Bounds().Max.Y; j++ {
			if i >= 0 && j >= 0 {
				img.Set(i, j, col)
			}
		}
	}
} // drawBorder 绘制边框
func (s *PDFService) drawBorder(img *image.RGBA, width, height int) {
	col := color.RGBA{0, 0, 0, 255}
	margin := 50

	// 外边框
	s.drawRectangle(img, margin, margin, width-margin*2, height-margin*2, col)

	// 内部分割线
	y := 180
	for i := 0; i < 3; i++ {
		s.drawLine(img, margin+20, y, width-margin-20, y, col)
		y += 60
	}
}

// drawLine 绘制直线
func (s *PDFService) drawLine(img *image.RGBA, x1, y1, x2, y2 int, col color.Color) {
	// 简单的水平线绘制
	if y1 == y2 {
		for x := x1; x <= x2; x++ {
			if x >= 0 && x < img.Bounds().Max.X && y1 >= 0 && y1 < img.Bounds().Max.Y {
				img.Set(x, y1, col)
			}
		}
	}
}

// drawRectangle 绘制矩形边框
func (s *PDFService) drawRectangle(img *image.RGBA, x, y, w, h int, col color.Color) {
	// 绘制四条边
	s.drawLine(img, x, y, x+w, y, col)     // 上边
	s.drawLine(img, x, y+h, x+w, y+h, col) // 下边
	s.drawLine(img, x, y, x, y+h, col)     // 左边
	s.drawLine(img, x+w, y, x+w, y+h, col) // 右边
}

// generatePDF 使用gopdf从头生成PDF收据
func (s *PDFService) generatePDF(data *model.ReceiptData, outputPath string) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	// 直接使用简单PDF生成，避免字体问题
	return s.generateSimplePDF(data, outputPath)
}

// ConvertReceiptToData 将请求数据转换为PDF填充数据

// generateSimplePDF 生成简单的纯文本PDF（使用内置字体，避免字体问题）
func (s *PDFService) generateSimplePDF(data *model.ReceiptData, outputPath string) error {
	// 自定义收据尺寸：176mm × 85mm (54开)
	// 1mm ≈ 2.83465 points
	receiptWidth := 176.0 * 2.83465 // 约498.5 points
	receiptHeight := 85.0 * 2.83465 // 约240.9 points

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: gopdf.Rect{W: receiptWidth, H: receiptHeight},
	})
	pdf.AddPage()

	// 首先尝试添加中文字体
	fontPath := "fonts/FangZhengFangSong-GBK-1.ttf"
	err := pdf.AddTTFFont("chinese", fontPath)
	if err != nil {
		// 如果中文字体加载失败，使用备用方案
		return fmt.Errorf("加载字体失败: %v", err)
	}

	// 设置字体
	err = pdf.SetFont("chinese", "", 12)
	if err != nil {
		return fmt.Errorf("设置字体失败: %v", err)
	}

	// 页面设置和边距
	margin := 8.0 // 减小边距适应小尺寸
	topMargin := 10.0

	// 绘制收据内容
	err = s.drawReceiptTemplate(&pdf, data, receiptWidth, receiptHeight, margin, topMargin)
	if err != nil {
		return err
	}

	// 保存PDF文件
	return pdf.WritePdf(outputPath)
}

// drawReceiptTemplate 按照模板样式绘制收据
func (s *PDFService) drawReceiptTemplate(pdf *gopdf.GoPdf, data *model.ReceiptData, width, height, margin, topMargin float64) error {
	var err error

	// 1. 绘制外边框
	pdf.SetLineWidth(1.5)
	pdf.RectFromUpperLeft(margin, topMargin, width-2*margin, height-2*topMargin)

	// 2. 标题区域："收款收据"
	err = pdf.SetFont("chinese", "", 14)
	if err != nil {
		return fmt.Errorf("设置标题字体失败: %v", err)
	}

	titleY := topMargin + 12
	pdf.SetXY(width/2-35, titleY)
	err = pdf.Text("收款收据")
	if err != nil {
		return fmt.Errorf("写入标题失败: %v", err)
	}

	// 3. 右上角收据编号和日期区域
	err = pdf.SetFont("chinese", "", 8)
	if err != nil {
		return fmt.Errorf("设置编号字体失败: %v", err)
	}

	// 收据号
	receiptNoY := topMargin + 20
	pdf.SetXY(width-110, receiptNoY)
	err = pdf.Text("收据号:")
	if err != nil {
		return fmt.Errorf("写入收据号标签失败: %v", err)
	}

	pdf.SetXY(width-75, receiptNoY)
	err = pdf.Text(data.ID)
	if err != nil {
		return fmt.Errorf("写入收据号失败: %v", err)
	}

	// 日期
	dateY := receiptNoY + 10
	pdf.SetXY(width-110, dateY)
	err = pdf.Text("日期:")
	if err != nil {
		return fmt.Errorf("写入日期标签失败: %v", err)
	}

	pdf.SetXY(width-75, dateY)
	err = pdf.Text(data.Date)
	if err != nil {
		return fmt.Errorf("写入日期失败: %v", err)
	}

	// 4. "今收到"区域
	todayReceivedY := topMargin + 45
	pdf.SetLineWidth(1)
	pdf.RectFromUpperLeft(margin+5, todayReceivedY, width-2*margin-10, 18)

	err = pdf.SetFont("chinese", "", 9)
	if err != nil {
		return fmt.Errorf("设置今收到字体失败: %v", err)
	}

	pdf.SetXY(margin+10, todayReceivedY+11)
	// "今收到"前缀
	err = pdf.Text("今收到 ")
	if err != nil {
		return fmt.Errorf("写入今收到前缀失败: %v", err)
	}
	// 设置颜色为蓝色
	pdf.SetTextColor(0, 0, 255)
	err = pdf.Text(data.Payer)
	if err != nil {
		return fmt.Errorf("写入今收到Payer失败: %v", err)
	}
	// 恢复为黑色
	pdf.SetTextColor(0, 0, 0)

	// 5. "交来"区域
	jiaolaiY := todayReceivedY + 22
	pdf.RectFromUpperLeft(margin+5, jiaolaiY, width-2*margin-10, 18)

	pdf.SetXY(margin+10, jiaolaiY+11)
	err = pdf.Text(fmt.Sprintf("交来: %s %s", data.Month, data.Purpose))
	if err != nil {
		return fmt.Errorf("写入交来失败: %v", err)
	}

	// 6. "金额(大写)"区域
	amountY := jiaolaiY + 22
	pdf.RectFromUpperLeft(margin+5, amountY, width-2*margin-10, 18)

	pdf.SetXY(margin+10, amountY+11)
	err = pdf.Text(fmt.Sprintf("金额(大写) 人民币 %s", data.RentZh))
	if err != nil {
		return fmt.Errorf("写入金额大写失败: %v", err)
	}

	// 7. 底部区域：小写金额和支付方式
	bottomY := amountY + 30

	// 小写金额
	err = pdf.SetFont("chinese", "", 12)
	if err != nil {
		return fmt.Errorf("设置金额字体失败: %v", err)
	}

	// 设置颜色为蓝色
	pdf.SetTextColor(0, 0, 255)
	pdf.SetXY(margin+10, bottomY)
	err = pdf.Text(fmt.Sprintf("人民币¥ %s", data.Rent))
	if err != nil {
		return fmt.Errorf("写入小写金额失败: %v", err)
	}
	// 恢复为黑色（RGB: 0, 0, 0），避免影响后续文本
	pdf.SetTextColor(0, 0, 0)

	// 支付方式选项
	err = pdf.SetFont("chinese", "", 7)
	if err != nil {
		return fmt.Errorf("设置支付方式字体失败: %v", err)
	}

	paymentY := bottomY + 3
	pdf.SetXY(width/2-50, paymentY)
	err = pdf.Text("现金 □     转账 □")
	if err != nil {
		return fmt.Errorf("写入支付方式失败: %v", err)
	}

	pdf.SetXY(width/2-50, paymentY+8)
	err = pdf.Text("支票 □  微信支付宝 □")
	if err != nil {
		return fmt.Errorf("写入支付方式2失败: %v", err)
	}

	// 右下角"(盖章)"
	pdf.SetXY(width-50, paymentY+3)
	err = pdf.Text("(盖章)")
	if err != nil {
		return fmt.Errorf("写入盖章失败: %v", err)
	}

	// 8. 底部"经手人"
	// 增大字体
	err = pdf.SetFont("chinese", "", 9)
	if err != nil {
		return fmt.Errorf("设置经手人字体失败: %v", err)
	}
	pdf.SetXY(width-90, height-40)
	err = pdf.Text(fmt.Sprintf("经手人： %s", data.Recipient))
	if err != nil {
		return fmt.Errorf("写入经手人失败: %v", err)
	}

	return nil
}

// GetFileSize 获取文件大小

// ConvertReceiptToData 将请求数据转换为PDF填充数据
func ConvertReceiptToData(req *model.ReceiptRequest) *model.ReceiptData {
	data := &model.ReceiptData{
		Rent:       fmt.Sprintf("%.2f", req.Rent),
		RentZh:     NumberToChinese(req.Rent),
		RoomNumber: req.RoomNumber,
		Recipient:  req.Recipient,
		Payer:      req.Payer,
		CreatedAt:  time.Now(),
	}

	// 处理日期
	if req.Date != "" {
		data.Date = req.Date
	} else {
		data.Date = time.Now().Format("2006-01-02")
	}

	// 处理月份
	if req.Month != "" {
		data.Month = req.Month
	} else {
		data.Month = time.Now().Format("2006年01月")
	}

	// 处理目的
	if req.Purpose != "" {
		data.Purpose = req.Purpose
	} else {
		data.Purpose = "房租"
	}

	// 生成收据ID：NO+房间号+月份
	data.ID = generateReceiptID(data.RoomNumber, data.Month)

	return data
}

// GetFileSize 获取文件大小
func GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// generateReceiptID 生成收据ID：NO+房间号+月份
func generateReceiptID(roomNumber, month string) string {
	// 提取月份中的数字部分，如"2025年9月" -> "202509"
	re := regexp.MustCompile(`(\d{4})年(\d{1,2})月`)
	matches := re.FindStringSubmatch(month)

	var monthCode string
	if len(matches) >= 3 {
		year := matches[1]
		monthNum := matches[2]
		if len(monthNum) == 1 {
			monthNum = "0" + monthNum // 补零
		}
		monthCode = year + monthNum
	} else {
		// 如果无法解析，使用当前时间
		monthCode = time.Now().Format("200601")
	}

	return fmt.Sprintf("NO%s%s", roomNumber, monthCode)
}

// NumberToChinese 将数字转换为中文大写金额
func NumberToChinese(amount float64) string {
	if amount == 0 {
		return "零元整"
	}

	// 处理负数
	negative := ""
	if amount < 0 {
		negative = "负"
		amount = -amount
	}

	// 分离整数和小数部分
	intPart := int64(amount)
	decPart := int64((amount-float64(intPart))*100 + 0.5) // 四舍五入到分

	// 中文数字
	digits := []string{"零", "壹", "贰", "叁", "肆", "伍", "陆", "柒", "捌", "玖"}

	var result strings.Builder

	// 处理整数部分
	if intPart == 0 {
		result.WriteString("零")
	} else {
		intStr := convertToChineseInt(intPart, digits)
		result.WriteString(intStr)
	}

	result.WriteString("元")

	// 处理小数部分
	if decPart == 0 {
		result.WriteString("整")
	} else {
		jiao := decPart / 10
		fen := decPart % 10

		if jiao > 0 {
			result.WriteString(digits[jiao])
			result.WriteString("角")
		}

		if fen > 0 {
			if jiao == 0 && intPart > 0 {
				result.WriteString("零")
			}
			result.WriteString(digits[fen])
			result.WriteString("分")
		}
	}

	return negative + result.String()
}

// convertToChineseInt 转换整数部分为中文
func convertToChineseInt(num int64, digits []string) string {
	if num == 0 {
		return "零"
	}

	// 处理超大数字的单位
	units := []string{"", "万", "亿", "万亿"}

	var parts []string
	unitIndex := 0

	for num > 0 && unitIndex < len(units) {
		part := num % 10000
		if part > 0 {
			partStr := convertFourDigits(part, digits)
			if unitIndex > 0 {
				partStr += units[unitIndex]
			}
			parts = append([]string{partStr}, parts...)
		} else if len(parts) > 0 {
			// 需要补零的情况
			if unitIndex > 0 && len(parts) > 0 {
				parts[0] = "零" + parts[0]
			}
		}
		num /= 10000
		unitIndex++
	}

	result := strings.Join(parts, "")

	// 清理多余的零
	result = strings.ReplaceAll(result, "零零", "零")
	result = strings.TrimSuffix(result, "零")

	return result
}

// convertFourDigits 转换四位以内的数字
func convertFourDigits(num int64, digits []string) string {
	if num == 0 {
		return ""
	}

	var result strings.Builder

	// 千位
	qian := num / 1000
	if qian > 0 {
		result.WriteString(digits[qian])
		result.WriteString("仟")
	}
	num %= 1000

	// 百位
	bai := num / 100
	if bai > 0 {
		if qian > 0 || result.Len() == 0 {
			result.WriteString(digits[bai])
		} else {
			result.WriteString("零" + digits[bai])
		}
		result.WriteString("佰")
	} else if qian > 0 && num > 0 {
		result.WriteString("零")
	}
	num %= 100

	// 十位
	shi := num / 10
	if shi > 0 {
		if shi == 1 && result.Len() == 0 {
			// 10-19 的情况，"一十" 可以简化为 "十"
			result.WriteString("拾")
		} else {
			if (bai > 0 || qian > 0) && result.String()[len(result.String())-3:] != "零" {
				if bai == 0 {
					result.WriteString("零")
				}
			}
			result.WriteString(digits[shi])
			result.WriteString("拾")
		}
	} else if (bai > 0 || qian > 0) && num > 0 {
		result.WriteString("零")
	}
	num %= 10

	// 个位
	if num > 0 {
		result.WriteString(digits[num])
	}

	return result.String()
}

// generateImageBasedOnPDFStyle 基于PDF样式生成图片
func (s *PDFService) generateImageBasedOnPDFStyle(data *model.ReceiptData, imagePath string) error {
	// PDF尺寸：176mm × 85mm (54开)
	// 转换为像素：使用300DPI
	// 176mm * 300DPI / 25.4 ≈ 2079 pixels
	// 85mm * 300DPI / 25.4 ≈ 1004 pixels
	width := 2079
	height := 1004

	// 创建RGBA图像
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充白色背景
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{255, 255, 255, 255}}, image.ZP, draw.Src)

	// 绘制收据内容 (使用与PDF相同的布局)
	if err := s.drawReceiptContentLikePDF(img, data, width, height); err != nil {
		return fmt.Errorf("绘制收据内容失败: %v", err)
	}

	// 保存图片
	file, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("创建图片文件失败: %v", err)
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("编码PNG图片失败: %v", err)
	}

	return nil
}

// drawReceiptContentLikePDF 按照PDF样式绘制图片内容
func (s *PDFService) drawReceiptContentLikePDF(img *image.RGBA, data *model.ReceiptData, width, height int) error {
	// 加载中文字体
	fontPath := "fonts/FangZhengFangSong-GBK-1.ttf"
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return fmt.Errorf("读取字体文件失败: %v", err)
	}

	font, err := truetype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("解析字体失败: %v", err)
	}

	// 创建FreeType上下文
	c := freetype.NewContext()
	c.SetDPI(300)
	c.SetFont(font)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color.RGBA{0, 0, 0, 255}))

	// PDF布局参数转换为像素
	// PDF: 176mm × 85mm, 使用points转换，但需要放大以适应图片显示
	// 使用更大的字体和间距以确保可读性
	margin := 80
	topMargin := 100

	// 1. 绘制外边框
	s.drawPDFStyleBorder(img, width, height, margin, topMargin)

	// 2. 绘制标题："收款收据" (放大字体)
	c.SetFontSize(120) // 增大标题字体
	titleY := topMargin + 100
	pt := freetype.Pt(width/2-240, titleY) // 重新计算居中位置
	_, err = c.DrawString("收款收据", pt)
	if err != nil {
		return fmt.Errorf("绘制标题失败: %v", err)
	}

	// 3. 右上角收据编号和日期 (增大字体)
	c.SetFontSize(60) // 增大字体
	receiptNoY := topMargin + 60
	pt = freetype.Pt(width-600, receiptNoY)
	_, err = c.DrawString("收据号:", pt)
	if err != nil {
		return fmt.Errorf("绘制收据号标签失败: %v", err)
	}

	pt = freetype.Pt(width-400, receiptNoY)
	_, err = c.DrawString(data.ID, pt)
	if err != nil {
		return fmt.Errorf("绘制收据号失败: %v", err)
	}

	// 日期
	dateY := receiptNoY + 80
	pt = freetype.Pt(width-600, dateY)
	_, err = c.DrawString("日期:", pt)
	if err != nil {
		return fmt.Errorf("绘制日期标签失败: %v", err)
	}

	pt = freetype.Pt(width-400, dateY)
	_, err = c.DrawString(data.Date, pt)
	if err != nil {
		return fmt.Errorf("绘制日期失败: %v", err)
	}

	// 4. "今收到"区域 (增大字体和间距)
	c.SetFontSize(80) // 增大字体
	todayReceivedY := topMargin + 250
	s.drawPDFStyleRect(img, margin+50, todayReceivedY, width-2*margin-100, 120)

	pt = freetype.Pt(margin+100, todayReceivedY+80)
	_, err = c.DrawString("今收到 ", pt)
	if err != nil {
		return fmt.Errorf("绘制今收到前缀失败: %v", err)
	}

	// 设置蓝色字体
	c.SetSrc(image.NewUniform(color.RGBA{0, 0, 255, 255}))
	pt = freetype.Pt(margin+300, todayReceivedY+80)
	_, err = c.DrawString(data.Payer, pt)
	if err != nil {
		return fmt.Errorf("绘制付款人失败: %v", err)
	}
	// 恢复黑色字体
	c.SetSrc(image.NewUniform(color.RGBA{0, 0, 0, 255}))

	// 5. "交来"区域
	jiaolaiY := todayReceivedY + 150
	s.drawPDFStyleRect(img, margin+50, jiaolaiY, width-2*margin-100, 120)

	pt = freetype.Pt(margin+100, jiaolaiY+80)
	text := fmt.Sprintf("交来: %s %s", data.Month, data.Purpose)
	_, err = c.DrawString(text, pt)
	if err != nil {
		return fmt.Errorf("绘制交来失败: %v", err)
	}

	// 6. "金额(大写)"区域
	amountY := jiaolaiY + 150
	s.drawPDFStyleRect(img, margin+50, amountY, width-2*margin-100, 120)

	pt = freetype.Pt(margin+100, amountY+80)
	text = fmt.Sprintf("金额(大写) 人民币 %s", data.RentZh)
	_, err = c.DrawString(text, pt)
	if err != nil {
		return fmt.Errorf("绘制金额大写失败: %v", err)
	}

	// 7. 底部区域：小写金额和支付方式 (增大字体)
	c.SetFontSize(100) // 增大字体
	bottomY := amountY + 200

	// 设置蓝色字体
	c.SetSrc(image.NewUniform(color.RGBA{0, 0, 255, 255}))
	pt = freetype.Pt(margin+100, bottomY)
	text = fmt.Sprintf("人民币¥ %s", data.Rent)
	_, err = c.DrawString(text, pt)
	if err != nil {
		return fmt.Errorf("绘制小写金额失败: %v", err)
	}
	// 恢复黑色字体
	c.SetSrc(image.NewUniform(color.RGBA{0, 0, 0, 255}))

	// 支付方式选项 (增大字体)
	c.SetFontSize(60) // 增大字体
	paymentY := bottomY + 50
	pt = freetype.Pt(width/2-300, paymentY)
	_, err = c.DrawString("现金 □     转账 □", pt)
	if err != nil {
		return fmt.Errorf("绘制支付方式失败: %v", err)
	}

	pt = freetype.Pt(width/2-300, paymentY+80)
	_, err = c.DrawString("支票 □  微信支付宝 □", pt)
	if err != nil {
		return fmt.Errorf("绘制支付方式2失败: %v", err)
	}

	// 右下角"(盖章)"
	pt = freetype.Pt(width-300, paymentY+40)
	_, err = c.DrawString("(盖章)", pt)
	if err != nil {
		return fmt.Errorf("绘制盖章失败: %v", err)
	}

	// 8. 底部"经手人" (增大字体)
	c.SetFontSize(80) // 增大字体
	handlerY := paymentY + 150
	pt = freetype.Pt(margin+100, handlerY)
	_, err = c.DrawString("经手人: ________________", pt)
	if err != nil {
		return fmt.Errorf("绘制经手人失败: %v", err)
	}

	return nil
}

// drawPDFStyleBorder 绘制PDF样式的边框
func (s *PDFService) drawPDFStyleBorder(img *image.RGBA, width, height, margin, topMargin int) {
	col := color.RGBA{0, 0, 0, 255}

	// 外边框 (增加线条宽度)
	borderWidth := 12 // 增大边框宽度

	for i := 0; i < borderWidth; i++ {
		// 上边
		s.drawLine(img, margin+i, topMargin+i, width-margin-i, topMargin+i, col)
		// 下边
		s.drawLine(img, margin+i, height-topMargin-i, width-margin-i, height-topMargin-i, col)
		// 左边
		s.drawLine(img, margin+i, topMargin+i, margin+i, height-topMargin-i, col)
		// 右边
		s.drawLine(img, width-margin-i, topMargin+i, width-margin-i, height-topMargin-i, col)
	}
}

// drawPDFStyleRect 绘制PDF样式的矩形框
func (s *PDFService) drawPDFStyleRect(img *image.RGBA, x, y, w, h int) {
	col := color.RGBA{0, 0, 0, 255}

	// 绘制更粗的矩形边框
	lineWidth := 3
	for i := 0; i < lineWidth; i++ {
		// 上边
		s.drawLine(img, x+i, y+i, x+w-i, y+i, col)
		// 下边
		s.drawLine(img, x+i, y+h-i, x+w-i, y+h-i, col)
		// 左边
		s.drawLine(img, x+i, y+i, x+i, y+h-i, col)
		// 右边
		s.drawLine(img, x+w-i, y+i, x+w-i, y+h-i, col)
	}
}
