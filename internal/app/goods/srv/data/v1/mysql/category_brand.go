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

type categoryBrands struct {
	// 无状态结构体，不需要db字段
}

func newCategoryBrands() *categoryBrands {
	return &categoryBrands{}
}

func (cb *categoryBrands) Get(ctx context.Context, db *gorm.DB, ID uint64) (*do.GoodsCategoryBrandDO, error) {
	categoryBrand := &do.GoodsCategoryBrandDO{}
	err := db.WithContext(ctx).Preload("Category").Preload("Brand").First(categoryBrand, ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrCategoryBrandNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	return categoryBrand, nil
}

func (cb *categoryBrands) List(ctx context.Context, db *gorm.DB, orderby []string, opts metav1.ListMeta) (*do.GoodsCategoryBrandDOList, error) {
	ret := &do.GoodsCategoryBrandDOList{}

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
	query := db.WithContext(ctx).Model(&do.GoodsCategoryBrandDO{}).Preload("Category").Preload("Brand")
	for _, value := range orderby {
		query = query.Order(value)
	}

	d := query.Offset(offset).Limit(limit).Find(&ret.Items).Count(&ret.TotalCount)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", d.Error.Error())
	}
	return ret, nil
}

func (cb *categoryBrands) GetByCategory(ctx context.Context, db *gorm.DB, categoryID uint64) (*do.GoodsCategoryBrandDOList, error) {
	ret := &do.GoodsCategoryBrandDOList{}
	d := db.WithContext(ctx).Preload("Category").Preload("Brand").Where("category_id = ?", categoryID).Find(&ret.Items)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", d.Error.Error())
	}
	return ret, nil
}

func (cb *categoryBrands) GetByBrand(ctx context.Context, db *gorm.DB, brandID uint64) (*do.GoodsCategoryBrandDOList, error) {
	ret := &do.GoodsCategoryBrandDOList{}
	d := db.WithContext(ctx).Preload("Category").Preload("Brand").Where("brand_id = ?", brandID).Find(&ret.Items)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", d.Error.Error())
	}
	return ret, nil
}

func (cb *categoryBrands) Create(ctx context.Context, db *gorm.DB, categoryBrand *do.GoodsCategoryBrandDO) error {
	tx := db.WithContext(ctx).Create(categoryBrand)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (cb *categoryBrands) CreateInTxn(ctx context.Context, txn *gorm.DB, categoryBrand *do.GoodsCategoryBrandDO) error {
	tx := txn.Create(categoryBrand)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (cb *categoryBrands) Update(ctx context.Context, db *gorm.DB, categoryBrand *do.GoodsCategoryBrandDO) error {
	tx := db.WithContext(ctx).Save(categoryBrand)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (cb *categoryBrands) UpdateInTxn(ctx context.Context, txn *gorm.DB, categoryBrand *do.GoodsCategoryBrandDO) error {
	tx := txn.Save(categoryBrand)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (cb *categoryBrands) Delete(ctx context.Context, db *gorm.DB, ID uint64) error {
	return db.WithContext(ctx).Where("id = ?", ID).Delete(&do.GoodsCategoryBrandDO{}).Error
}

func (cb *categoryBrands) DeleteInTxn(ctx context.Context, txn *gorm.DB, ID uint64) error {
	return txn.Where("id = ?", ID).Delete(&do.GoodsCategoryBrandDO{}).Error
}

var _ interfaces.GoodsCategoryBrandStore = &categoryBrands{}