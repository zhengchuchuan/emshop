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
	// 无状态结构体，不需要db字段
}

func newGoods() *goods {
	return &goods{}
}


func (g *goods) List(ctx context.Context, db *gorm.DB, orderby []string, opts metav1.ListMeta) (*do.GoodsDOList, error) {
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

	// 构建基础查询
	baseQuery := db.WithContext(ctx).Where("deleted_at IS NULL")
	
	// 先获取总数
	if err := baseQuery.Model(&do.GoodsDO{}).Count(&ret.TotalCount).Error; err != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}

	// 排序和过滤
	query := db.WithContext(ctx).Preload("Category").Preload("Brands").Where("deleted_at IS NULL")
	for _, value := range orderby {
		query = query.Order(value)
	}

	// 应用分页并查询数据
	if err := query.Offset(offset).Limit(limit).Find(&ret.Items).Error; err != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	
	return ret, nil
}

func (g *goods) Get(ctx context.Context, db *gorm.DB, ID uint64) (*do.GoodsDO, error) {
	good := &do.GoodsDO{}
	err := db.WithContext(ctx).Preload("Category").Preload("Brands").Where("id = ? AND deleted_at IS NULL", ID).First(good).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrGoodsNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	return good, nil
}

func (g *goods) ListByIDs(ctx context.Context, db *gorm.DB, ids []uint64, orderby []string) (*do.GoodsDOList, error) {
	ret := &do.GoodsDOList{}

	// 先获取总数
	baseQuery := db.WithContext(ctx).Where("deleted_at IS NULL").Where("id in ?", ids)
	if err := baseQuery.Model(&do.GoodsDO{}).Count(&ret.TotalCount).Error; err != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}

	// 排序和过滤
	query := db.WithContext(ctx).Preload("Category").Preload("Brands").Where("deleted_at IS NULL")
	for _, value := range orderby {
		query = query.Order(value)
	}

	if err := query.Where("id in ?", ids).Find(&ret.Items).Error; err != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	return ret, nil
}

func (g *goods) Create(ctx context.Context, db *gorm.DB, goods *do.GoodsDO) error {
	tx := db.WithContext(ctx).Create(goods)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (g *goods) Update(ctx context.Context, db *gorm.DB, goods *do.GoodsDO) error {
	tx := db.WithContext(ctx).Model(goods).Omit("add_time", "created_at").Updates(goods)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (g *goods) Delete(ctx context.Context, db *gorm.DB, ID uint64) error {
	return db.WithContext(ctx).Where("id = ?", ID).Delete(&do.GoodsDO{}).Error
}

func (g *goods) GetAllGoodsIDs(ctx context.Context, db *gorm.DB) ([]uint64, error) {
	var ids []uint64
	
	// 只查询ID字段，避免加载完整对象
	err := db.WithContext(ctx).Model(&do.GoodsDO{}).
		Where("deleted_at IS NULL").
		Pluck("id", &ids).Error
		
	return ids, err
}

var _ interfaces.GoodsStore = &goods{}