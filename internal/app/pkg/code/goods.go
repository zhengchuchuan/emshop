package code

const (
	// ErrGoodsNotFound - 404: Goods not found.
	ErrGoodsNotFound int = iota + 100501

	// ErrCategoryNotFound - 404: Category not found.
	ErrCategoryNotFound

	// ErrBrandNotFound - 404: Brand not found.
	ErrBrandNotFound

	// ErrBannerNotFound - 404: Banner not found.
	ErrBannerNotFound

	// ErrCategoryBrandNotFound - 404: CategoryBrand not found.
	ErrCategoryBrandNotFound

	// ErrEsQuery - 500: Es query error.
	ErrEsQuery

	// ErrEsUnmarshal - 500: Es unmarshal error.
	ErrEsUnmarshal

	// ErrCategoryHasChildren - 400: Category has children.
	ErrCategoryHasChildren
)
