package interfaces

import (
	"context"
	"emshop/internal/app/user/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"
	"gorm.io/gorm"
)

// UserStore 用户存储接口
type UserStore interface {
	Get(ctx context.Context, db *gorm.DB, ID uint64) (*do.UserDO, error)
	GetByMobile(ctx context.Context, db *gorm.DB, mobile string) (*do.UserDO, error)
	List(ctx context.Context, db *gorm.DB, orderby []string, opts metav1.ListMeta) (*do.UserDOList, error)
	Create(ctx context.Context, db *gorm.DB, user *do.UserDO) error
	Update(ctx context.Context, db *gorm.DB, user *do.UserDO) error
}