package mysql

import (
	"context"
	"emshop/internal/app/pkg/code"
	code2 "emshop/gin-micro/code"
	metav1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"

	"gorm.io/gorm"
	"emshop/internal/app/goods/srv/domain/do"
	"emshop/internal/app/goods/srv/data/v1/interfaces"
)

type brands struct {
	// 无状态结构体，不需要db字段
}

func newBrands() *brands {
	return &brands{}
}

func (b *brands) Get(ctx context.Context, db *gorm.DB, ID uint64) (*do.BrandsDO, error) {
	brand := &do.BrandsDO{}
	err := db.WithContext(ctx).First(brand, ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrBrandNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	return brand, nil
}

func (b *brands) List(ctx context.Context, db *gorm.DB, orderby []string, opts metav1.ListMeta) (*do.BrandsDOList, error) {
	ret := &do.BrandsDOList{}

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

	// 排序
	query := db.WithContext(ctx).Model(&do.BrandsDO{})
	for _, value := range orderby {
		query = query.Order(value)
	}

	d := query.Offset(offset).Limit(limit).Find(&ret.Items).Count(&ret.TotalCount)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", d.Error.Error())
	}
	return ret, nil
}

func (b *brands) Create(ctx context.Context, db *gorm.DB, brand *do.BrandsDO) error {
	tx := db.WithContext(ctx).Create(brand)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (b *brands) Update(ctx context.Context, db *gorm.DB, brand *do.BrandsDO) error {
	tx := db.WithContext(ctx).Save(brand)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (b *brands) Delete(ctx context.Context, db *gorm.DB, ID uint64) error {
	return db.WithContext(ctx).Where("id = ?", ID).Delete(&do.BrandsDO{}).Error
}

var _ interfaces.BrandsStore = &brands{}