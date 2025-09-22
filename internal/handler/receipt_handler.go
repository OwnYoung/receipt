package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	pdfBytes, err := os.ReadFile(outputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "读取PDF文件失败: " + err.Error(),
		})
		return
	}

	// 转换为Base64编码
	base64PDF := base64.StdEncoding.EncodeToString(pdfBytes)

	// 创建备份目录
	backupDir := "backup"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "创建备份目录失败: " + err.Error(),
		})
		return
	}

	// 生成备份文件名（包含时间戳和收据ID）
	backupFileName := fmt.Sprintf("receipt_%s_%s.pdf", data.ID, time.Now().Format("20060102_150405"))
	backupPath := filepath.Join(backupDir, backupFileName)

	// 创建备份文件
	err = os.WriteFile(backupPath, pdfBytes, 0644)
	if err != nil {
		// 备份失败不影响主要功能，记录错误但继续执行
		fmt.Printf("警告：备份文件失败: %v\n", err)
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
			"backupPath":   backupPath, // 返回备份路径信息
		},
	})

	// 清理临时文件
	go func() {
		time.Sleep(1 * time.Minute) // 1分钟后清理
		os.Remove(outputPath)
	}()
}

// GenerateReceiptImage 生成收据图片
// @Summary 生成收据图片
// @Description 专门为小程序设计的接口，生成收据图片并返回Base64编码数据
// @Tags 收据
// @Accept json
// @Produce json
// @Param request body model.ReceiptRequest true "收据信息"
// @Success 200 {object} map[string]interface{} "生成成功"
// @Failure 400 {object} model.ReceiptResponse "请求参数错误"
// @Failure 500 {object} model.ReceiptResponse "服务器内部错误"
// @Router /api/receipt/generate-image [post]
func (h *ReceiptHandler) GenerateReceiptImage(c *gin.Context) {
	var req model.ReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ReceiptResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 转换为收据数据
	data := service.ConvertReceiptToData(&req)

	// 生成收据图片并直接返回Base64编码
	base64Image, err := h.pdfService.GenerateReceiptImageBase64(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "生成收据图片失败: " + err.Error(),
		})
		return
	}

	// 为备份功能，将Base64转换回字节数据
	imageBytes, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "图片数据处理失败: " + err.Error(),
		})
		return
	}

	// 创建备份目录
	backupDir := "backup"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "创建备份目录失败: " + err.Error(),
		})
		return
	}

	// 生成备份文件名
	backupFileName := fmt.Sprintf("receipt_%s_%s.png", data.ID, time.Now().Format("20060102_150405"))
	backupPath := filepath.Join(backupDir, backupFileName)

	// 创建备份文件
	err = os.WriteFile(backupPath, imageBytes, 0644)
	if err != nil {
		// 备份失败不影响主要功能，记录错误但继续执行
		fmt.Printf("警告：备份图片文件失败: %v\n", err)
	}

	fileName := fmt.Sprintf("receipt_%s_%s.png", data.RoomNumber, time.Now().Format("20060102_150405"))

	// 返回JSON响应给小程序
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "收据图片生成成功",
		"data": gin.H{
			"receiptId":    data.ID,
			"fileName":     fileName,
			"fileSize":     len(imageBytes),
			"imageBase64":  base64Image,
			"contentType":  "image/png",
			"generateTime": time.Now().Format("2006-01-02 15:04:05"),
			"backupPath":   backupPath,
		},
	})
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

// ListBackupReceipts 列出备份的收据文件
// @Summary 列出备份的收据文件
// @Description 获取服务器上备份的收据文件列表
// @Tags 收据
// @Produce json
// @Success 200 {object} map[string]interface{} "获取成功"
// @Failure 500 {object} model.ReceiptResponse "服务器内部错误"
// @Router /api/receipt/backup/list [get]
func (h *ReceiptHandler) ListBackupReceipts(c *gin.Context) {
	backupDir := "backup"

	// 检查备份目录是否存在
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "暂无备份文件",
			"data": gin.H{
				"files": []interface{}{},
				"count": 0,
			},
		})
		return
	}

	// 读取备份目录
	files, err := os.ReadDir(backupDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ReceiptResponse{
			Success: false,
			Message: "读取备份目录失败: " + err.Error(),
		})
		return
	}

	var backupFiles []map[string]interface{}
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".pdf" {
			info, err := file.Info()
			if err != nil {
				continue
			}

			backupFiles = append(backupFiles, map[string]interface{}{
				"fileName":    file.Name(),
				"fileSize":    info.Size(),
				"modTime":     info.ModTime().Format("2006-01-02 15:04:05"),
				"downloadUrl": fmt.Sprintf("/api/receipt/backup/download/%s", file.Name()),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取备份文件列表成功",
		"data": gin.H{
			"files": backupFiles,
			"count": len(backupFiles),
		},
	})
}

// DownloadBackupReceipt 下载备份的收据文件
// @Summary 下载备份的收据文件
// @Description 根据文件名下载指定的备份收据文件
// @Tags 收据
// @Param fileName path string true "文件名"
// @Produce application/pdf
// @Success 200 {file} binary "PDF文件"
// @Failure 404 {object} model.ReceiptResponse "文件不存在"
// @Failure 500 {object} model.ReceiptResponse "服务器内部错误"
// @Router /api/receipt/backup/download/{fileName} [get]
func (h *ReceiptHandler) DownloadBackupReceipt(c *gin.Context) {
	fileName := c.Param("fileName")
	backupPath := filepath.Join("backup", fileName)

	// 检查文件是否存在
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, model.ReceiptResponse{
			Success: false,
			Message: "备份文件不存在",
		})
		return
	}

	// 检查文件扩展名安全性
	if filepath.Ext(fileName) != ".pdf" {
		c.JSON(http.StatusBadRequest, model.ReceiptResponse{
			Success: false,
			Message: "不支持的文件类型",
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileName))

	// 返回文件
	c.File(backupPath)
}
