package response

// LoginResponse 用户登录响应
type LoginResponse struct {
	ID        int32  `json:"id"`
	NickName  string `json:"nickName"`
	Token     string `json:"token"`
	ExpiredAt int64  `json:"expiredAt"`
}

// RegisterResponse 用户注册响应
type RegisterResponse struct {
	ID        int32  `json:"id"`
	NickName  string `json:"nickName"`
	Token     string `json:"token"`
	ExpiredAt int64  `json:"expiredAt"`
}
