package handler

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"receipt/internal/model"
	"receipt/internal/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ReceiptHandler struct {
	pdfService *service.PDFService
}

func NewReceiptHandler(pdfService *service.PDFService) *ReceiptHandler {
	return &ReceiptHandler{
		pdfService: pdfService,
	}
}

// GenerateReceipt 生成收据PDF
// @Summary 生成收据PDF
// @Description 接收小程序发送的租金、房间号、收款人等信息，生成收据PDF并返回
// @Tags 收据
// @Accept json
// @Produce application/pdf
// @Param request body model.ReceiptRequest true "收据信息"
// @Success 200 {file} binary "PDF文件"
// @Failure 400 {object} model.ReceiptResponse "请求参数错误"
// @Failure 500 {object} model.ReceiptResponse "服务器内部错误"
// @Router /api/receipt/generate [post]
func (h *ReceiptHandler) GenerateReceipt(c *gin.Context) {
	var req model.ReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ReceiptResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 转换为PDF填充数据
	data := service.ConvertReceiptToData(&req)

	// 生成PDF
	outputPath, err := h.pdfService.FillReceipt(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "生成收据PDF失败: " + err.Error(),
		})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "PDF文件生成失败",
		})
		return
	}

	// 获取文件信息
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "无法获取文件信息: " + err.Error(),
		})
		return
	}

	// 设置小程序友好的响应头
	fileName := fmt.Sprintf("receipt_%s_%s.pdf", data.RoomNumber, time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileName))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// 返回PDF文件给小程序
	c.File(outputPath)

	// 延迟清理临时文件（可选）
	go func() {
		time.Sleep(5 * time.Minute) // 5分钟后清理
		os.Remove(outputPath)
	}()
}

// GenerateReceiptForMiniProgram 为小程序生成收据PDF（返回Base64）
// @Summary 为小程序生成收据PDF
// @Description 专门为小程序设计的接口，返回Base64编码的PDF数据
// @Tags 收据
// @Accept json
// @Produce json
// @Param request body model.ReceiptRequest true "收据信息"
// @Success 200 {object} map[string]interface{} "生成成功"
// @Failure 400 {object} model.ReceiptResponse "请求参数错误"
// @Failure 500 {object} model.ReceiptResponse "服务器内部错误"
// @Router /api/receipt/miniprogram [post]
func (h *ReceiptHandler) GenerateReceiptForMiniProgram(c *gin.Context) {
	var req model.ReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ReceiptResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 转换为PDF填充数据
	data := service.ConvertReceiptToData(&req)

	// 生成PDF
	outputPath, err := h.pdfService.FillReceipt(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "生成收据PDF失败: " + err.Error(),
		})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "PDF文件生成失败",
		})
		return
	}

	// 读取PDF文件内容
	pdfBytes, err := ioutil.ReadFile(outputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "读取PDF文件失败: " + err.Error(),
		})
		return
	}

	// 转换为Base64编码
	base64PDF := base64.StdEncoding.EncodeToString(pdfBytes)

	// 获取文件信息
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "无法获取文件信息: " + err.Error(),
		})
		return
	}

	fileName := fmt.Sprintf("receipt_%s_%s.pdf", data.RoomNumber, time.Now().Format("20060102_150405"))

	// 返回JSON响应给小程序
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "收据生成成功",
		"data": gin.H{
			"receiptId":    data.ID,
			"fileName":     fileName,
			"fileSize":     fileInfo.Size(),
			"pdfBase64":    base64PDF,
			"contentType":  "application/pdf",
			"generateTime": time.Now().Format("2006-01-02 15:04:05"),
		},
	})

	// 清理临时文件
	go func() {
		time.Sleep(1 * time.Minute) // 1分钟后清理
		os.Remove(outputPath)
	}()
}

// GetReceiptInfo 获取收据信息（仅返回JSON，不生成PDF）
// @Summary 获取收据信息
// @Description 接收小程序发送的租金、房间号、收款人等信息，返回处理后的信息（预览功能）
// @Tags 收据
// @Accept json
// @Produce json
// @Param request body model.ReceiptRequest true "收据信息"
// @Success 200 {object} model.ReceiptResponse "处理成功"
// @Failure 400 {object} model.ReceiptResponse "请求参数错误"
// @Router /api/receipt/info [post]
func (h *ReceiptHandler) GetReceiptInfo(c *gin.Context) {
	var req model.ReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ReceiptResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 转换为PDF填充数据
	data := service.ConvertReceiptToData(&req)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取收据信息成功",
		"data":    data,
	})
}

// HealthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务状态
// @Tags 系统
// @Produce json
// @Success 200 {object} map[string]interface{} "服务正常"
// @Router /health [get]
func (h *ReceiptHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "收据服务运行正常",
		"service": "receipt-service",
	})
}
