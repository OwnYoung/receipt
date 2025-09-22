package main

import (
	"fmt"
	"image/png"
	"log"
	"os"

	"github.com/gen2brain/go-fitz"
)

func main() {
	// 测试 go-fitz 库
	pdfPath := "output/receipt_101_20250921_214202.pdf"

	// 检查文件是否存在
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		log.Printf("PDF文件不存在: %s", pdfPath)
		return
	}

	// 打开PDF文档
	doc, err := fitz.New(pdfPath)
	if err != nil {
		log.Fatal(err)
	}
	defer doc.Close()

	// 获取第一页
	img, err := doc.Image(0) // 0 是第一页
	if err != nil {
		log.Fatal(err)
	}

	// 保存为PNG
	f, err := os.Create("output/receipt_converted.png")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = png.Encode(f, img)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("PDF转图片成功: output/receipt_converted.png")
}
