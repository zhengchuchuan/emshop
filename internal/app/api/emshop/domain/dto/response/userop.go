package response

// UserFavItemResponse 用户收藏项响应
type UserFavItemResponse struct {
	UserId  int32 `json:"userId"`
	GoodsId int32 `json:"goodsId"`
}

// UserFavListResponse 用户收藏列表响应
type UserFavListResponse struct {
	Total int64                 `json:"total"`
	Items []UserFavItemResponse `json:"data"`
}

// AddressItemResponse 地址项响应
type AddressItemResponse struct {
	ID           int32  `json:"id"`
	Province     string `json:"province"`
	City         string `json:"city"`
	District     string `json:"district"`
	Address      string `json:"address"`
	SignerName   string `json:"signerName"`
	SignerMobile string `json:"signerMobile"`
}

// AddressListResponse 地址列表响应
type AddressListResponse struct {
	Total int64                 `json:"total"`
	Items []AddressItemResponse `json:"data"`
}

// MessageItemResponse 留言项响应
type MessageItemResponse struct {
	ID          int32  `json:"id"`
	MessageType int32  `json:"messageType"`
	Subject     string `json:"subject"`
	Message     string `json:"message"`
	File        string `json:"file"`
}

// MessageListResponse 留言列表响应
type MessageListResponse struct {
	Total int64                 `json:"total"`
	Items []MessageItemResponse `json:"data"`
}
