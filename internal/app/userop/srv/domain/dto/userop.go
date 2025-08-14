package dto

import "time"

// UserFavDTO 用户收藏传输对象
type UserFavDTO struct {
	ID        int32     `json:"id"`
	UserID    int32     `json:"user_id"`
	GoodsID   int32     `json:"goods_id"`
	CreatedAt time.Time `json:"created_at"`
}

// AddressDTO 地址传输对象
type AddressDTO struct {
	ID           int32     `json:"id"`
	UserID       int32     `json:"user_id"`
	Province     string    `json:"province"`
	City         string    `json:"city"`
	District     string    `json:"district"`
	Address      string    `json:"address"`
	SignerName   string    `json:"signer_name"`
	SignerMobile string    `json:"signer_mobile"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MessageDTO 留言传输对象
type MessageDTO struct {
	ID          int32     `json:"id"`
	UserID      int32     `json:"user_id"`
	MessageType int32     `json:"message_type"`
	Subject     string    `json:"subject"`
	Message     string    `json:"message"`
	File        string    `json:"file"`
	CreatedAt   time.Time `json:"created_at"`
}