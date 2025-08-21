package mysql

import (
	"context"
	"emshop/gin-micro/code"
	"emshop/internal/app/inventory/srv/data/v1/interfaces"
	"emshop/internal/app/inventory/srv/domain/do"
	code2 "emshop/internal/app/pkg/code"
	"emshop/pkg/errors"
	"emshop/pkg/log"

	"gorm.io/gorm"
)

type inventorys struct {
	// 无状态结构体，不需要factory字段
}

func (i *inventorys) UpdateStockSellDetailStatus(ctx context.Context, db *gorm.DB, ordersn string, status int32) error {

	// update语句如果没有更新的话那么不会报错，但是他会返回一个影响的行数，所以我们可以根据影响的行数来判断是否更新成功
	result := db.Model(do.StockSellDetailDO{}).Where("order_sn = ?", ordersn).Update("status", status)
	if result.Error != nil {
		return errors.WithCode(code.ErrDatabase, "%s", result.Error.Error())
	}

	return nil
}

func (i *inventorys) GetSellDetail(ctx context.Context, db *gorm.DB, ordersn string) (*do.StockSellDetailDO, error) {
	var ordersellDetail do.StockSellDetailDO
	err := db.Where("order_sn = ?", ordersn).First(&ordersellDetail).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code2.ErrInvSellDetailNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}
	return &ordersellDetail, err
}

func (i *inventorys) Reduce(ctx context.Context, db *gorm.DB, goodsID uint64, num int) error {
	return db.Model(&do.InventoryDO{}).Where("goods=?", goodsID).Where("stocks >= ?", num).UpdateColumn("stocks", gorm.Expr("stocks - ?", num)).Error
}

func (i *inventorys) Increase(ctx context.Context, db *gorm.DB, goodsID uint64, num int) error {
	err := db.Model(&do.InventoryDO{}).Where("goods=?", goodsID).UpdateColumn("stocks", gorm.Expr("stocks + ?", num)).Error
	return err
}

func (i *inventorys) CreateStockSellDetail(ctx context.Context, db *gorm.DB, detail *do.StockSellDetailDO) error {

	tx := db.Create(&detail)
	if tx.Error != nil {
		return errors.WithCode(code.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (i *inventorys) Create(ctx context.Context, db *gorm.DB, inv *do.InventoryDO) error {
	// 设置库存， 如果我要更新库存
	tx := db.Create(&inv)
	if tx.Error != nil {
		return errors.WithCode(code.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (i *inventorys) Get(ctx context.Context, db *gorm.DB, goodsID uint64) (*do.InventoryDO, error) {
	inv := do.InventoryDO{}
	err := db.Where("goods = ?", goodsID).First(&inv).Error
	if err != nil {
		log.Errorf("get inv err: %v", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code2.ErrInventoryNotFound, "%s", err.Error())
		}

		return nil, errors.WithCode(code.ErrDatabase, "%s", err.Error())
	}

	return &inv, nil
}

func newInventorys() *inventorys {
	return &inventorys{}
}

var _ interfaces.InventoryStore = &inventorys{}