package main

import (
	"log"
	"receipt/internal/handler"
	"receipt/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建Gin引擎
	r := gin.Default()

	// 配置路径 - 新系统不需要模板文件
	templatePath := "" // 不再使用模板
	outputPath := "output"

	// 创建服务
	pdfService := service.NewPDFService(templatePath, outputPath)
	receiptHandler := handler.NewReceiptHandler(pdfService)

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 注册路由
	api := r.Group("/api")
	{
		receipt := api.Group("/receipt")
		{
			receipt.POST("/generate", receiptHandler.GenerateReceipt)                  // 直接返回PDF文件
			receipt.POST("/miniprogram", receiptHandler.GenerateReceiptForMiniProgram) // 为小程序返回Base64 PDF
			receipt.POST("/generate-image", receiptHandler.GenerateReceiptImage)       // 为小程序返回Base64图片
			receipt.POST("/info", receiptHandler.GetReceiptInfo)

			// 备份管理相关接口
			backup := receipt.Group("/backup")
			{
				backup.GET("/list", receiptHandler.ListBackupReceipts)                  // 列出备份文件
				backup.GET("/download/:fileName", receiptHandler.DownloadBackupReceipt) // 下载备份文件
			}
		}
	}

	// 健康检查
	r.GET("/health", receiptHandler.HealthCheck)

	// 添加根路径处理
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "收据生成服务",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"生成收据(PDF文件)":     "POST /api/receipt/generate",
				"生成收据(小程序Base64)": "POST /api/receipt/miniprogram",
				"生成收据图片(小程序)":     "POST /api/receipt/generate-image",
				"预览信息":            "POST /api/receipt/info",
				"备份文件列表":          "GET /api/receipt/backup/list",
				"下载备份文件":          "GET /api/receipt/backup/download/{fileName}",
				"健康检查":            "GET /health",
			},
		})
	})

	// 启动服务器
	port := ":8090"
	log.Printf("收据服务启动在端口%s", port)
	log.Printf("健康检查: http://localhost%s/health", port)
	log.Printf("生成收据: POST http://localhost%s/api/receipt/generate", port)

	if err := r.Run(port); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}
