package interfaces

import (
	"context"
	"emshop/internal/app/order/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"

	"gorm.io/gorm"
)

// OrderStore 订单存储接口
type OrderStore interface {
	Get(ctx context.Context, db *gorm.DB, orderSn string) (*do.OrderInfoDO, error)

	List(ctx context.Context, db *gorm.DB, userID uint64, meta metav1.ListMeta, orderby []string) (*do.OrderInfoDOList, error)

	Create(ctx context.Context, db *gorm.DB, order *do.OrderInfoDO) error

	Update(ctx context.Context, db *gorm.DB, order *do.OrderInfoDO) error
}