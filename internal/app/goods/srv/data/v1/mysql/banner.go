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

type banner struct {
	// 无状态结构体，不需要db字段
}

func newBanner() *banner {
	return &banner{}
}

func (b *banner) Get(ctx context.Context, db *gorm.DB, ID uint64) (*do.BannerDO, error) {
	banner := &do.BannerDO{}
	err := db.WithContext(ctx).First(banner, ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrBannerNotFound, "%s", err.Error())
		}
		return nil, errors.WithCode(code2.ErrDatabase, "%s", err.Error())
	}
	return banner, nil
}

func (b *banner) List(ctx context.Context, db *gorm.DB, orderby []string, opts metav1.ListMeta) (*do.BannerDOList, error) {
	ret := &do.BannerDOList{}

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
	query := db.WithContext(ctx).Model(&do.BannerDO{})
	for _, value := range orderby {
		query = query.Order(value)
	}

	d := query.Offset(offset).Limit(limit).Find(&ret.Items).Count(&ret.TotalCount)
	if d.Error != nil {
		return nil, errors.WithCode(code2.ErrDatabase, "%s", d.Error.Error())
	}
	return ret, nil
}

func (b *banner) Create(ctx context.Context, db *gorm.DB, banner *do.BannerDO) error {
	tx := db.WithContext(ctx).Create(banner)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (b *banner) Update(ctx context.Context, db *gorm.DB, banner *do.BannerDO) error {
	tx := db.WithContext(ctx).Save(banner)
	if tx.Error != nil {
		return errors.WithCode(code2.ErrDatabase, "%s", tx.Error.Error())
	}
	return nil
}

func (b *banner) Delete(ctx context.Context, db *gorm.DB, ID uint64) error {
	return db.WithContext(ctx).Where("id = ?", ID).Delete(&do.BannerDO{}).Error
}


var _ interfaces.BannerStore = &banner{}