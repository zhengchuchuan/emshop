package dto

import "emshop/internal/app/goods/srv/domain/do"

type GoodsDTO struct {
	do.GoodsDO
}

type GoodsDTOList struct {
	TotalCount int64       `json:"total_count,omitempty"`
	Items      []*GoodsDTO `json:"data"`
}

type CategoryDTO struct {
	do.CategoryDO
	SubCategories []*CategoryDTO `json:"sub_categories,omitempty"`
}

type CategoryDTOList struct {
	TotalCount int64         `json:"total_count,omitempty"`
	Items      []*CategoryDTO `json:"data"`
	ParentInfo *CategoryDTO   `json:"parent_info,omitempty"`
}

type BrandDTO struct {
	do.BrandsDO
}

type BrandDTOList struct {
	TotalCount int64       `json:"total_count,omitempty"`
	Items      []*BrandDTO `json:"data"`
}

type BannerDTO struct {
	do.BannerDO
}

type BannerDTOList struct {
	TotalCount int64        `json:"total_count,omitempty"`
	Items      []*BannerDTO `json:"data"`
}

type CategoryBrandDTO struct {
	do.GoodsCategoryBrandDO
}

type CategoryBrandDTOList struct {
	TotalCount int64               `json:"total_count,omitempty"`
	Items      []*CategoryBrandDTO `json:"data"`
}
