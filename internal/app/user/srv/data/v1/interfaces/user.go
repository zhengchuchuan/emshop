package interfaces

import (
	"context"
	"emshop/internal/app/user/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
)

// UserStore 用户存储接口
type UserStore interface {
	Get(ctx context.Context, ID uint64) (*do.UserDO, error)
	GetByMobile(ctx context.Context, mobile string) (*do.UserDO, error)
	List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*do.UserDOList, error)
	Create(ctx context.Context, user *do.UserDO) error
	Update(ctx context.Context, user *do.UserDO) error
}