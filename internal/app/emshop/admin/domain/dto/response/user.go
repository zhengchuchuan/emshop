package response

// AdminLoginResponse 管理员登录响应
type AdminLoginResponse struct {
	ID        int32  `json:"id"`
	NickName  string `json:"nickName"`
	Mobile    string `json:"mobile"`
	Role      int32  `json:"role"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
	Message   string `json:"message"`
}