package user

import (
	"context"

	upbv1 "emshop-admin/api/user/v1"
	srvv1 "emshop-admin/internal/app/user/service/v1"
	metav1 "emshop-admin/pkg/common/meta/v1"
)

func DTOToResponse(userdto srvv1.UserDTO) *upbv1.UserInfoResponse {
	return &upbv1.UserInfoResponse{}
}


func GetUserList(ctx context.Context, info *upbv1.PageInfo) (*upbv1.UserListResponse, error) {
	srvOpts := metav1.ListMeta{
		Page:    int(info.Pn),
		PageSize: int(info.PSize),
	}
	dtoList, err := srvv1.List(ctx, srvOpts)
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
