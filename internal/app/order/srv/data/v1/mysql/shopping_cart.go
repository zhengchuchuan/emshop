package mysql

import (
	"context"
	code2 "emshop/gin-micro/code"
	"emshop/internal/app/pkg/code"
	"emshop/pkg/errors"

	"emshop/internal/app/order/srv/data/v1/interfaces"
	"emshop/internal/app/order/srv/domain/do"
	metav1 "emshop/pkg/common/meta/v1"

	"gorm.io/gorm"
)

type shoppingCarts struct {
	// 无状态结构体，不需要factory字段
}

func newShoppingCarts() *shoppingCarts {
	return &shoppingCarts{}
}

// 这个在事务中执行，建议大家使用消息队列来实现
func (sc *shoppingCarts) DeleteByGoodsIDs(ctx context.Context, db *gorm.DB, userID uint64, goodsIDs []int32) error {
	return db.Where("user = ? AND goods IN (?)", userID, goodsIDs).Delete(&do.ShoppingCartDO{}).Error
}

func (sc *shoppingCarts) List(ctx context.Context, db *gorm.DB, userID uint64, checked bool, meta metav1.ListMeta, orderby []string) (*do.ShoppingCartDOList, error) {
	ret := &do.ShoppingCartDOList{}
	query := db.WithContext(ctx)
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

	if userID > 0 {
		query = query.Where("user = ?", userID)
	}
	if checked {
		query = query.Where("checked = ?", true)
	}
	
	// 过滤已删除的记录
	query = query.Where("deleted_at IS NULL")

	// 排序
	for _, value := range orderby {
		query = query.Order(value)
	}

	d := query.Offset(offset).Limit(limit).Find(&ret.Items).Count(&ret.TotalCount)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", d.Error.Error())
	}
	return ret, nil
}

func (sc *shoppingCarts) Create(ctx context.Context, db *gorm.DB, cartItem *do.ShoppingCartDO) error {
	tx := db.Create(cartItem)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (sc *shoppingCarts) Get(ctx context.Context, db *gorm.DB, userID, goodsID uint64) (*do.ShoppingCartDO, error) {
	var shopCart do.ShoppingCartDO
	err := db.WithContext(ctx).Where("user = ? AND goods = ? AND deleted_at IS NULL", userID, goodsID).First(&shopCart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrShopCartItemNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	return &shopCart, nil
}

func (sc *shoppingCarts) UpdateNum(ctx context.Context, db *gorm.DB, cartItem *do.ShoppingCartDO) error {
	return db.Model(&do.ShoppingCartDO{}).Where("user = ? AND goods = ?", cartItem.User, cartItem.Goods).Update("nums", cartItem.Nums).Update("checked", cartItem.Checked).Error
}

func (sc *shoppingCarts) Delete(ctx context.Context, db *gorm.DB, ID uint64) error {
	return db.Where("id = ?", ID).Delete(&do.ShoppingCartDO{}).Error
}

// 清空check状态
func (sc *shoppingCarts) ClearCheck(ctx context.Context, db *gorm.DB, userID uint64) error {
	return db.WithContext(ctx).Model(&do.ShoppingCartDO{}).Where("user = ?", userID).Update("checked", false).Error
}

var _ interfaces.ShopCartStore = &shoppingCarts{}