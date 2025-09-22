package model

import "time"

// ReceiptRequest 收据请求模型
type ReceiptRequest struct {
	Rent       float64 `json:"rent" binding:"required" example:"1500.00"`    // 租金
	RoomNumber string  `json:"room_number" binding:"required" example:"101"` // 房间号
	Recipient  string  `json:"recipient" binding:"required" example:"张三"`    // 收款人
	Payer      string  `json:"payer" binding:"required" example:"李四"`        // 付款人
	Date       string  `json:"date" example:"2025-09-21"`                    // 收据日期，如果为空则使用当前日期
	Month      string  `json:"month" example:"2025年9月"`                      // 租金月份
	Purpose    string  `json:"purpose" example:"房租"`                         // 收费目的
}

// ReceiptResponse 收据响应模型
type ReceiptResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	FileName string `json:"file_name,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
}

// ReceiptData PDF 填充数据
type ReceiptData struct {
	ID         string    `json:"id"`          // 收据编号，格式：NO+房间号+月份
	Rent       string    `json:"rent"`        // 租金金额
	RentZh     string    `json:"rent_zh"`     // 租金中文大写金额
	RoomNumber string    `json:"room_number"` // 房间号
	Recipient  string    `json:"recipient"`   // 收款人
	Payer      string    `json:"payer"`       // 付款人
	Date       string    `json:"date"`        // 收据日期
	Month      string    `json:"month"`       // 租金月份
	Purpose    string    `json:"purpose"`     // 收费目的
	CreatedAt  time.Time `json:"created_at"`  // 创建时间
}
