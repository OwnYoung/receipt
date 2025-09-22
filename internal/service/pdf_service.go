package service

import (
	"fmt"
	"os"
	"path/filepath"
	"receipt/internal/model"
	"regexp"
	"strings"
	"time"

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
