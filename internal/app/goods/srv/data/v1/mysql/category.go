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

type categorys struct {
	db *gorm.DB
}

func newCategorys(factory *mysqlFactory) *categorys {
	return &categorys{
		db: factory.db,
	}
}

func (c *categorys) Get(ctx context.Context, ID uint64) (*do.CategoryDO, error) {
	category := &do.CategoryDO{}

	err := c.db.Preload("SubCategory").Preload("SubCategory.SubCategory").First(category, ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrCategoryNotFound, err.Error())
		}
		return nil, errors.WithCode(code2.ErrDatabase, err.Error())
	}
	return category, nil
}

func (c *categorys) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*do.CategoryDOList, error) {
	ret := &do.CategoryDOList{}

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
	query := c.db.Model(&do.CategoryDO{})
	for _, value := range orderby {
		query = query.Order(value)
	}

	d := query.Offset(offset).Limit(limit).Find(&ret.Items).Count(&ret.TotalCount)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, d.Error.Error())
	}
	return ret, nil
}

func (c *categorys) GetByLevel(ctx context.Context, level int) (*do.CategoryDOList, error) {
	ret := &do.CategoryDOList{}
	query := c.db.Where("level = ?", level)

	if level == 1 {
		// 对于一级分类，预加载子分类
		query = query.Preload("SubCategory.SubCategory")
	}

	d := query.Find(&ret.Items)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, d.Error.Error())
	}
	return ret, nil
}

func (c *categorys) GetSubCategories(ctx context.Context, parentID uint64) (*do.CategoryDOList, error) {
	ret := &do.CategoryDOList{}
	d := c.db.Where("parent_category_id = ?", parentID).Find(&ret.Items)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, d.Error.Error())
	}
	return ret, nil
}

func (c *categorys) Create(ctx context.Context, category *do.CategoryDO) error {
	tx := c.db.Create(category)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, tx.Error.Error())
	}
	return nil
}

func (c *categorys) CreateInTxn(ctx context.Context, txn *gorm.DB, category *do.CategoryDO) error {
	tx := txn.Create(category)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, tx.Error.Error())
	}
	return nil
}

func (c *categorys) Update(ctx context.Context, category *do.CategoryDO) error {
	tx := c.db.Save(category)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, tx.Error.Error())
	}
	return nil
}

func (c *categorys) UpdateInTxn(ctx context.Context, txn *gorm.DB, category *do.CategoryDO) error {
	tx := txn.Save(category)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, tx.Error.Error())
	}
	return nil
}

func (c *categorys) Delete(ctx context.Context, ID uint64) error {
	return c.db.Where("id = ?", ID).Delete(&do.CategoryDO{}).Error
}

func (c *categorys) DeleteInTxn(ctx context.Context, txn *gorm.DB, ID uint64) error {
	return txn.Where("id = ?", ID).Delete(&do.CategoryDO{}).Error
}

var _ interfaces.CategoryStore = &categorys{}