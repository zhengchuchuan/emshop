package mysql

import (
	"context"
	"emshop/internal/app/pkg/code"
	code2 "emshop/gin-micro/code"
	metav1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"

	"gorm.io/gorm"
	"emshop/internal/app/goods/srv/data/v1/interfaces"
	"emshop/internal/app/goods/srv/domain/do"
)

type goods struct {
	db *gorm.DB
}

func newGoods(factory *mysqlFactory) *goods {
	return &goods{
		db: factory.db,
	}
}

func (g *goods) CreateInTxn(ctx context.Context, txn *gorm.DB, goods *do.GoodsDO) error {
	tx := txn.Create(goods)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (g *goods) UpdateInTxn(ctx context.Context, txn *gorm.DB, goods *do.GoodsDO) error {
	tx := txn.Model(goods).Omit("add_time", "created_at").Updates(goods)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (g *goods) DeleteInTxn(ctx context.Context, txn *gorm.DB, ID uint64) error {
	return txn.Where("id = ?", ID).Delete(&do.GoodsDO{}).Error
}

func (g *goods) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*do.GoodsDOList, error) {
	ret := &do.GoodsDOList{}

	// 分页
	var limit, offset int
	if opts.PageSize == 0 {
		limit = 10
	} else {
		limit = opts.PageSize
	}

	if opts.Page > 0 {
		offset = (opts.Page - 1) * limit
	}

	// 排序和过滤
	query := g.db.WithContext(ctx).Preload("Category").Preload("Brands").Where("deleted_at IS NULL")
	for _, value := range orderby {
		query = query.Order(value)
	}

	d := query.Offset(offset).Limit(limit).Find(&ret.Items).Count(&ret.TotalCount)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", d.Error.Error())
	}
	return ret, nil
}

func (g *goods) Get(ctx context.Context, ID uint64) (*do.GoodsDO, error) {
	good := &do.GoodsDO{}
	err := g.db.WithContext(ctx).Preload("Category").Preload("Brands").Where("id = ? AND deleted_at IS NULL", ID).First(good).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrGoodsNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	return good, nil
}

func (g *goods) ListByIDs(ctx context.Context, ids []uint64, orderby []string) (*do.GoodsDOList, error) {
	ret := &do.GoodsDOList{}

	// 排序和过滤
	query := g.db.WithContext(ctx).Preload("Category").Preload("Brands").Where("deleted_at IS NULL")
	for _, value := range orderby {
		query = query.Order(value)
	}

	d := query.Where("id in ?", ids).Find(&ret.Items).Count(&ret.TotalCount)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", d.Error.Error())
	}
	return ret, nil
}

func (g *goods) Create(ctx context.Context, goods *do.GoodsDO) error {
	tx := g.db.Create(goods)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (g *goods) Update(ctx context.Context, goods *do.GoodsDO) error {
	tx := g.db.Model(goods).Omit("add_time", "created_at").Updates(goods)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (g *goods) Delete(ctx context.Context, ID uint64) error {
	return g.db.Where("id = ?", ID).Delete(&do.GoodsDO{}).Error
}

var _ interfaces.GoodsStore = &goods{}