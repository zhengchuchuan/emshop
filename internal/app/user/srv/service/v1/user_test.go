package v1

import (
	"context"
	"testing"

	"emshop/internal/app/user/srv/data/v1/mock"
	metav1 "emshop/pkg/common/meta/v1"
)

func TestUserList(t *testing.T) {
	userSrv := NewUserService(mock.NewUsers())
	userSrv.List(context.Background(), nil, metav1.ListMeta{})
}