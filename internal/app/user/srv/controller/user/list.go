package user

import (
	"context"
	"fmt"

	upbv1 "emshop/api/user/v1"
	"emshop/internal/app/user/srv/domain/dto"
	metav1 "emshop/pkg/common/meta/v1"
)

func DTOToResponse(userDTO dto.UserDTO) *upbv1.UserInfoResponse {
	//在grpc的message中字段有默认值，不能随便赋值nil进去，容易出错
	//这里要搞清， 哪些字段是有默认值
	userInfoRsp := upbv1.UserInfoResponse{
		Id:       userDTO.ID,
		PassWord: userDTO.Password,
		NickName: userDTO.NickName,
		Gender:   userDTO.Gender,
		Role:     int32(userDTO.Role),
		Mobile:   userDTO.Mobile,
	}
	// Birthday是time.Time类型，不能直接赋值nil
	if userDTO.Birthday != nil {
		userInfoRsp.BirthDay = uint64(userDTO.Birthday.Unix())
	}
	// 内部有mutex, 不能拷贝
	return &userInfoRsp
}
/*
controller 层依赖了service， service层依赖了data层：
controller层能否直接依赖data层： 可以的
controller依赖service并不是直接依赖了具体的struct而是依赖了interface
*/

func (us *userServer)GetUserList(ctx context.Context, info *upbv1.PageInfo) (*upbv1.UserListResponse, error) {
	srvOpts := metav1.ListMeta{}
	
	// Handle optional pagination parameters
	if info.Pn == nil && info.PSize == nil {
		// Return all data when no pagination parameters are specified
		srvOpts.Page = 1
		srvOpts.PageSize = 10000 // Large page size to get all data
	} else {
		// Use provided pagination parameters with default values for nil fields
		if info.Pn != nil {
			srvOpts.Page = int(*info.Pn)
		} else {
			srvOpts.Page = 1 // Default page
		}
		if info.PSize != nil {
			srvOpts.PageSize = int(*info.PSize)
		} else {
			srvOpts.PageSize = 10 // Default page size
		}
	}
	
	dtoList, err := us.srv.List(ctx,[]string{}, srvOpts)
	if err != nil {
		return nil, err
	}
	var rsp upbv1.UserListResponse
	for _, value := range dtoList.Items {
		// 这里可以对dtoList.Items进行处理
		userRsp := DTOToResponse(*value)
		rsp.Data = append(rsp.Data, userRsp)
	}
	fmt.Printf("dtoList.TotalCount: %d\n", dtoList.TotalCount)
	rsp.Total = int32(dtoList.TotalCount)  // 设置总数
	fmt.Printf("rsp.Total: %d\n", rsp.Total)
	return &rsp, nil
}
