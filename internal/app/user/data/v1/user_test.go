package v1

import (
	"context"
	"testing"

	"emshop-admin/internal/app/user/data/v1/mock"
	servicev1 "emshop-admin/internal/app/user/service/v1" // 给 service/v1 包起别名
	metav1 "emshop-admin/pkg/common/meta/v1"
)

func TestUserList(t *testing.T) {
	userSrv := servicev1.NewUserService(mock.NewUsers())
	userSrv.List(context.Background(), metav1.ListMeta{})
}