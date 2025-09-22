package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

func main() {
	templatePath := filepath.Join("..", "templates", "receipt_template.pdf")

	// 检查文件是否存在
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		log.Fatalf("PDF模板文件不存在: %s", templatePath)
	}

	// 创建测试JSON数据 - 使用更简单的格式
	jsonData := `{"id":"NO101202509","rent":"150.00","rent_zh":"壹佰伍拾元整","room_number":"101","recipient":"张三","payer":"李四","date":"2025-09-21","month":"2025年9月","purpose":"房租"}`

	// 创建临时JSON文件
	jsonFile := "test_form_data.json"
	log.Printf("创建JSON文件: %s", jsonFile)
	log.Printf("JSON内容: %s", jsonData)

	if err := os.WriteFile(jsonFile, []byte(jsonData), 0644); err != nil {
		log.Fatalf("创建JSON文件失败: %v", err)
	}
	defer os.Remove(jsonFile)

	// 验证文件是否创建成功
	if content, err := os.ReadFile(jsonFile); err != nil {
		log.Fatalf("读取JSON文件失败: %v", err)
	} else {
		log.Printf("JSON文件内容确认: %s", string(content))
	}

	// 输出文件路径
	outputPath := "test_filled_receipt.pdf"

	log.Println("尝试填充PDF表单...")
	log.Printf("模板文件: %s", templatePath)
	log.Printf("JSON数据文件: %s", jsonFile)
	log.Printf("输出文件: %s", outputPath)

	// 尝试填充表单
	if err := api.FillFormFile(templatePath, jsonFile, outputPath, nil); err != nil {
		log.Printf("填充失败: %v", err)

		// 如果填充失败，尝试导出表单字段信息
		log.Println("尝试导出表单字段信息...")
		if err := api.ExportFormFile(templatePath, "form_fields.json", nil); err != nil {
			log.Printf("导出表单字段失败: %v", err)
		} else {
			log.Println("表单字段信息已导出到 form_fields.json")
			if data, err := os.ReadFile("form_fields.json"); err == nil {
				fmt.Println("表单字段详情:")
				fmt.Println(string(data))
			}
		}
	} else {
		log.Printf("填充成功！输出文件: %s", outputPath)
	}
}
