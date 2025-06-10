package user

import (
	"context"

	upbv1 "emshop/api/user/v1"
	srvv1 "emshop/internal/app/user/srv/service/v1"
	metav1 "emshop/pkg/common/meta/v1"
)

func DTOToResponse(userdto srvv1.UserDTO) *upbv1.UserInfoResponse {
	return &upbv1.UserInfoResponse{}
}
/*
controller 层依赖了service， service层依赖了data层：
controller层能否直接依赖data层： 可以的
controller依赖service并不是直接依赖了具体的struct而是依赖了interface
*/

func (us *userServer)GetUserList(ctx context.Context, info *upbv1.PageInfo) (*upbv1.UserListResponse, error) {
	srvOpts := metav1.ListMeta{
		Page:    int(info.Pn),
		PageSize: int(info.PSize),
	}
	dtoList, err := us.srv.List(ctx, srvOpts)
	if err != nil {
		return nil, err
	}
	var rsp upbv1.UserListResponse
	for _, value := range dtoList.Items {
		// 这里可以对dtoList.Items进行处理
		userRsp := DTOToResponse(*value)
		rsp.Data = append(rsp.Data, userRsp)
	}
	return &rsp, nil
}
