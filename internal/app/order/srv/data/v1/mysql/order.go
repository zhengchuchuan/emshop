package mysql

import (
	"context"
	code2 "emshop/gin-micro/code"
	"emshop/pkg/errors"

	"emshop/internal/app/order/srv/data/v1/interfaces"
	"emshop/internal/app/order/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"

	"gorm.io/gorm"
)

type orders struct {
	// 无状态结构体，不需要factory字段
}

func newOrders() *orders {
	return &orders{}
}

func (o *orders) Get(ctx context.Context, db *gorm.DB, orderSn string) (*do.OrderInfoDO, error) {
	var order do.OrderInfoDO
	err := db.WithContext(ctx).Preload("OrderGoods").Where("order_sn = ? AND deleted_at IS NULL", orderSn).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (o *orders) List(ctx context.Context, db *gorm.DB, userID uint64, meta metav1.ListMeta, orderby []string) (*do.OrderInfoDOList, error) {
	ret := &do.OrderInfoDOList{}
	// 分页
	var limit, offset int
	if meta.PageSize == 0 {
		limit = 10
	} else {
		limit = meta.PageSize
	}

	if meta.Page > 0 {
		offset = (meta.Page - 1) * limit
	}

	// 排序和过滤
	query := db.WithContext(ctx).Preload("OrderGoods").Where("deleted_at IS NULL")
	if userID > 0 {
		query = query.Where("user = ?", userID)
	}
	for _, value := range orderby {
		query = query.Order(value)
	}

	d := query.Offset(offset).Limit(limit).Find(&ret.Items).Count(&ret.TotalCount)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", d.Error.Error())
	}
	return ret, nil
}

// Create 创建订单之后要删除对应的购物车记录
func (o *orders) Create(ctx context.Context, db *gorm.DB, order *do.OrderInfoDO) error {
	return db.Create(order).Error
}

func (o *orders) Update(ctx context.Context, db *gorm.DB, order *do.OrderInfoDO) error {
	return db.Model(order).Save(order).Error
}

var _ interfaces.OrderStore = &orders{}