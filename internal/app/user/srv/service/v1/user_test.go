package v1

import (
	"context"
	"testing"

	"emshop-admin/internal/app/user/srv/data/v1/mock"
	metav1 "emshop-admin/pkg/common/meta/v1"
)

func TestUserList(t *testing.T) {
	userSrv := NewUserService(mock.NewUsers())
	userSrv.List(context.Background(), metav1.ListMeta{})
}