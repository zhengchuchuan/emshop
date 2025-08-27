package request

import (
	"time"
	upbv1 "emshop/api/user/v1"
)

// UpdateUserRequest 用户更新请求结构体
type UpdateUserRequest struct {
	Name     string `form:"name" json:"name" binding:"required,min=3,max=10"`
	Gender   string `form:"gender" json:"gender" binding:"required,oneof=female male"`
	Birthday string `form:"birthday" json:"birthday" binding:"required,datetime=2006-01-02"`
}

// ToProto 将请求结构体转换为 protobuf 结构体
func (r *UpdateUserRequest) ToProto(userID uint64) (*upbv1.UpdateUserInfo, error) {
	updateReq := &upbv1.UpdateUserInfo{
		Id:       int32(userID),
		NickName: &r.Name,
		Gender:   &r.Gender,
	}

	// 将前端传递过来的日期格式转换成 timestamp
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}
	
	birthDay, err := time.ParseInLocation("2006-01-02", r.Birthday, loc)
	if err != nil {
		return nil, err
	}
	
	birthDayUnix := uint64(birthDay.Unix())
	updateReq.BirthDay = &birthDayUnix

	return updateReq, nil
}