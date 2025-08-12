package v1

import (
	"context"
	"encoding/json"
	
	"emshop/internal/app/goods/srv/domain/do"
	"emshop/internal/app/goods/srv/domain/dto"
	"emshop/internal/app/pkg/code"
	metav1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"
	"emshop/pkg/log"
)

type category struct {
	srv *service
}

func newCategory(srv *service) CategorySrv {
	return &category{srv: srv}
}

var _ CategorySrv = &category{}

func (c *category) List(ctx context.Context, opts metav1.ListMeta, orderby []string) (*dto.CategoryDTOList, error) {
	dataFactory := c.srv.factoryManager.GetDataFactory()
	categories, err := dataFactory.Categorys().List(ctx, orderby, opts)
	if err != nil {
		log.Errorf("get categories list error: %v", err)
		return nil, err
	}

	ret := &dto.CategoryDTOList{
		TotalCount: categories.TotalCount,
		Items: make([]*dto.CategoryDTO, 0),
	}

	for _, item := range categories.Items {
		categoryDTO := &dto.CategoryDTO{
			CategoryDO: *item,
		}
		ret.Items = append(ret.Items, categoryDTO)
	}

	return ret, nil
}

func (c *category) ListAll(ctx context.Context, orderby []string) (*dto.CategoryDTOList, error) {
	dataFactory := c.srv.factoryManager.GetDataFactory()
	// 获取一级分类（包括子分类）
	categories, err := dataFactory.Categorys().GetByLevel(ctx, 1)
	if err != nil {
		log.Errorf("get all categories error: %v", err)
		return nil, err
	}

	ret := &dto.CategoryDTOList{
		TotalCount: int64(len(categories.Items)),
		Items: make([]*dto.CategoryDTO, 0),
	}

	for _, item := range categories.Items {
		categoryDTO := &dto.CategoryDTO{
			CategoryDO:    *item,
			SubCategories: make([]*dto.CategoryDTO, 0),
		}

		// 获取二级分类
		if subCategories, err := dataFactory.Categorys().GetSubCategories(ctx, uint64(item.ID)); err == nil {
			for _, sub := range subCategories.Items {
				subCategoryDTO := &dto.CategoryDTO{
					CategoryDO:    *sub,
					SubCategories: make([]*dto.CategoryDTO, 0),
				}

				// 获取三级分类
				if subSubCategories, err := dataFactory.Categorys().GetSubCategories(ctx, uint64(sub.ID)); err == nil {
					for _, subSub := range subSubCategories.Items {
						subSubCategoryDTO := &dto.CategoryDTO{
							CategoryDO: *subSub,
						}
						subCategoryDTO.SubCategories = append(subCategoryDTO.SubCategories, subSubCategoryDTO)
					}
				}
				categoryDTO.SubCategories = append(categoryDTO.SubCategories, subCategoryDTO)
			}
		}
		ret.Items = append(ret.Items, categoryDTO)
	}

	return ret, nil
}

func (c *category) GetSubCategories(ctx context.Context, parentID int32) (*dto.CategoryDTOList, error) {
	dataFactory := c.srv.factoryManager.GetDataFactory()
	
	// 先检查父分类是否存在
	parent, err := dataFactory.Categorys().Get(ctx, uint64(parentID))
	if err != nil {
		log.Errorf("parent category not found: %v", err)
		return nil, errors.WithCode(code.ErrCategoryNotFound, "父分类不存在")
	}

	categories, err := dataFactory.Categorys().GetSubCategories(ctx, uint64(parentID))
	if err != nil {
		log.Errorf("get sub categories error: %v", err)
		return nil, err
	}

	ret := &dto.CategoryDTOList{
		TotalCount: int64(len(categories.Items)),
		Items: make([]*dto.CategoryDTO, 0),
		ParentInfo: &dto.CategoryDTO{
			CategoryDO: *parent,
		},
	}

	for _, item := range categories.Items {
		categoryDTO := &dto.CategoryDTO{
			CategoryDO: *item,
		}
		ret.Items = append(ret.Items, categoryDTO)
	}

	return ret, nil
}

func (c *category) Get(ctx context.Context, ID int32) (*dto.CategoryDTO, error) {
	dataFactory := c.srv.factoryManager.GetDataFactory()
	category, err := dataFactory.Categorys().Get(ctx, uint64(ID))
	if err != nil {
		log.Errorf("get category error: %v", err)
		return nil, err
	}

	return &dto.CategoryDTO{
		CategoryDO: *category,
	}, nil
}

func (c *category) Create(ctx context.Context, category *dto.CategoryDTO) error {
	dataFactory := c.srv.factoryManager.GetDataFactory()

	// 验证父分类是否存在（如果不是一级分类）
	if category.Level != 1 && category.ParentCategoryID != 0 {
		_, err := dataFactory.Categorys().Get(ctx, uint64(category.ParentCategoryID))
		if err != nil {
			log.Errorf("parent category not found: %v", err)
			return errors.WithCode(code.ErrCategoryNotFound, "父分类不存在")
		}
	}

	categoryDO := &do.CategoryDO{
		Name:             category.Name,
		ParentCategoryID: category.ParentCategoryID,
		Level:            category.Level,
		IsTab:            category.IsTab,
	}

	err := dataFactory.Categorys().Create(ctx, categoryDO)
	if err != nil {
		log.Errorf("create category error: %v", err)
		return err
	}

	category.ID = categoryDO.ID
	return nil
}

func (c *category) Update(ctx context.Context, category *dto.CategoryDTO) error {
	dataFactory := c.srv.factoryManager.GetDataFactory()

	// 检查分类是否存在
	existing, err := dataFactory.Categorys().Get(ctx, uint64(category.ID))
	if err != nil {
		log.Errorf("category not found: %v", err)
		return errors.WithCode(code.ErrCategoryNotFound, "分类不存在")
	}

	// 验证父分类是否存在（如果不是一级分类）
	if category.Level != 1 && category.ParentCategoryID != 0 {
		_, err := dataFactory.Categorys().Get(ctx, uint64(category.ParentCategoryID))
		if err != nil {
			log.Errorf("parent category not found: %v", err)
			return errors.WithCode(code.ErrCategoryNotFound, "父分类不存在")
		}
	}

	// 更新字段
	existing.Name = category.Name
	existing.ParentCategoryID = category.ParentCategoryID
	existing.Level = category.Level
	existing.IsTab = category.IsTab

	err = dataFactory.Categorys().Update(ctx, existing)
	if err != nil {
		log.Errorf("update category error: %v", err)
		return err
	}

	return nil
}

func (c *category) Delete(ctx context.Context, ID int32) error {
	dataFactory := c.srv.factoryManager.GetDataFactory()

	// 检查分类是否存在
	_, err := dataFactory.Categorys().Get(ctx, uint64(ID))
	if err != nil {
		log.Errorf("category not found: %v", err)
		return errors.WithCode(code.ErrCategoryNotFound, "分类不存在")
	}

	// 检查是否有子分类
	subCategories, err := dataFactory.Categorys().GetSubCategories(ctx, uint64(ID))
	if err == nil && len(subCategories.Items) > 0 {
		return errors.WithCode(code.ErrCategoryHasChildren, "分类下存在子分类，无法删除")
	}

	err = dataFactory.Categorys().Delete(ctx, uint64(ID))
	if err != nil {
		log.Errorf("delete category error: %v", err)
		return err
	}

	return nil
}

// GetAllCategoriesJSON 获取所有分类的JSON格式数据（与原有接口兼容）
func (c *category) GetAllCategoriesJSON(ctx context.Context) (string, error) {
	categories, err := c.ListAll(ctx, []string{})
	if err != nil {
		return "", err
	}

	// 构建树形结构
	type CategoryNode struct {
		ID            int32           `json:"id"`
		Name          string          `json:"name"`
		Level         int32           `json:"level"`
		IsTab         bool            `json:"is_tab"`
		ParentID      int32           `json:"parent,omitempty"`
		SubCategories []*CategoryNode `json:"sub_category,omitempty"`
	}

	var buildTree func([]*dto.CategoryDTO) []*CategoryNode
	buildTree = func(categories []*dto.CategoryDTO) []*CategoryNode {
		var nodes []*CategoryNode
		for _, cat := range categories {
			node := &CategoryNode{
				ID:       cat.ID,
				Name:     cat.Name,
				Level:    cat.Level,
				IsTab:    cat.IsTab,
				ParentID: cat.ParentCategoryID,
			}
			if len(cat.SubCategories) > 0 {
				node.SubCategories = buildTree(cat.SubCategories)
			}
			nodes = append(nodes, node)
		}
		return nodes
	}

	tree := buildTree(categories.Items)
	jsonData, err := json.Marshal(tree)
	if err != nil {
		log.Errorf("marshal categories to json error: %v", err)
		return "", err
	}

	return string(jsonData), nil
}