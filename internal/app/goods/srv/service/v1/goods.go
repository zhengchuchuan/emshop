package v1

import (
	"context"
	proto "emshop/api/goods/v1"
	dataV1 "emshop/internal/app/goods/srv/data/v1"
	"emshop/internal/app/goods/srv/data/v1/interfaces"
	"emshop/internal/app/goods/srv/domain/do"
	"emshop/internal/app/goods/srv/domain/dto"
	"sync"

	"github.com/zeromicro/go-zero/core/mr"
	metav1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/log"
	"gorm.io/gorm"
)

type goodsService struct {
	// 预加载的核心DAO（日常CRUD操作）
	goodsDAO    interfaces.GoodsStore
	categoryDAO interfaces.CategoryStore
	brandDAO    interfaces.BrandsStore
	bannerDAO   interfaces.BannerStore
	db          *gorm.DB
	
	// 保留工厂管理器（复杂操作：ES同步、事务等）
	factoryManager *dataV1.FactoryManager
}

func newGoods(srv *service) *goodsService {
	dataFactory := srv.factoryManager.GetDataFactory()
	
	return &goodsService{
		// 预加载核心DAO，避免每次方法调用时重复获取
		goodsDAO:    dataFactory.Goods(),
		categoryDAO: dataFactory.Categorys(),
		brandDAO:    dataFactory.Brands(),
		bannerDAO:   dataFactory.Banners(),
		db:          dataFactory.DB(),
		
		// 保留工厂管理器用于复杂操作（ES同步、事务等）
		factoryManager: srv.factoryManager,
	}
}

func (gs *goodsService) List(ctx context.Context, opts metav1.ListMeta, req *proto.GoodsFilterRequest, orderby []string) (*dto.GoodsDTOList, error) {
	log.Debugf("Listing goods with search conditions: keywords=%v, brand=%v", req.KeyWords, req.Brand)
	
	// 检查是否有搜索条件
	hasSearchConditions := false
	var categoryIDs []interface{}
	
	// 检查各种搜索条件
	if req.KeyWords != nil && *req.KeyWords != "" {
		hasSearchConditions = true
	}
	if req.Brand != nil && *req.Brand > 0 {
		hasSearchConditions = true
	}
	if (req.PriceMin != nil && *req.PriceMin > 0) || (req.PriceMax != nil && *req.PriceMax > 0) {
		hasSearchConditions = true
	}
	if (req.IsHot != nil && *req.IsHot) || (req.IsNew != nil && *req.IsNew) || (req.IsTab != nil && *req.IsTab) {
		hasSearchConditions = true
	}
	if req.TopCategory != nil && *req.TopCategory > 0 {
		hasSearchConditions = true
		// 使用预加载的categoryDAO
		category, err := gs.categoryDAO.Get(ctx, gs.db, uint64(*req.TopCategory))
		if err != nil {
			log.Errorf("categoryData.Get err: %v", err)
			return nil, err
		}
		// 树形结构遍历
		for _, value := range retrieveIDs(category) {
			categoryIDs = append(categoryIDs, value)
		}
	}

	// 如果没有搜索条件，直接查询MySQL - 使用预加载的DAO
	if !hasSearchConditions {
		log.Debugf("No search conditions, querying MySQL directly")
		goods, err := gs.goodsDAO.List(ctx, gs.db, orderby, opts)
		if err != nil {
			log.Errorf("Failed to list goods from MySQL: %v", err)
			return nil, err
		}
		
		// 业务逻辑层：数据转换
		ret := gs.convertToGoodsDTOList(goods)
		log.Debugf("Successfully listed %d goods from MySQL, total: %d", len(ret.Items), ret.TotalCount)
		return ret, nil
	}

	// 有搜索条件时，构建ES搜索请求
	searchReq := &interfaces.GoodsFilterRequest{
		CategoryIDs: categoryIDs,
	}
	
	// 安全地解引用指针字段
	if req.KeyWords != nil {
		searchReq.KeyWords = *req.KeyWords
	}
	if req.Brand != nil {
		searchReq.BrandID = *req.Brand
	}
	if req.PriceMin != nil {
		searchReq.PriceMin = float32(*req.PriceMin)
	}
	if req.PriceMax != nil {
		searchReq.PriceMax = float32(*req.PriceMax)
	}
	if req.IsHot != nil {
		searchReq.IsHot = *req.IsHot
	}
	if req.IsNew != nil {
		searchReq.IsNew = *req.IsNew
	}
	if req.IsTab != nil {
		searchReq.OnSale = *req.IsTab
	}
	if req.Pages != nil {
		searchReq.Pages = *req.Pages
	}
	if req.PagePerNums != nil {
		searchReq.PagePerNums = *req.PagePerNums
	}

	// 确保分页参数有效
	if searchReq.Pages <= 0 {
		searchReq.Pages = int32(opts.Page)
	}
	if searchReq.PagePerNums <= 0 {
		searchReq.PagePerNums = int32(opts.PageSize)
	}

	// 调试输出
	log.Debugf("ES Search Request: %+v", searchReq)
	log.Debugf("CategoryIDs: %v", searchReq.CategoryIDs)

	// 通过搜索引擎查询（使用工厂管理器获取搜索功能）
	dataFactory := gs.factoryManager.GetDataFactory()
	goodsList, err := dataFactory.Search().Goods().Search(ctx, searchReq)
	if err != nil {
		log.Errorf("ES search failed: %v", err)
		return nil, err
	}

	log.Debugf("ES search results: total=%d, items=%d", goodsList.TotalCount, len(goodsList.Items))

	// 提取商品ID列表
	goodsIDs := make([]uint64, 0, len(goodsList.Items))
	for _, value := range goodsList.Items {
		goodsIDs = append(goodsIDs, uint64(value.ID))
	}

	// 通过ID批量查询MySQL数据 - 使用预加载的DAO
	goods, err := gs.goodsDAO.ListByIDs(ctx, gs.db, goodsIDs, orderby)
	if err != nil {
		log.Errorf("Failed to list goods by IDs from MySQL: %v", err)
		return nil, err
	}
	
	// 业务逻辑层：数据转换
	ret := &dto.GoodsDTOList{
		TotalCount: int64(goodsList.TotalCount),
		Items:      make([]*dto.GoodsDTO, 0, len(goods.Items)),
	}
	for _, value := range goods.Items {
		ret.Items = append(ret.Items, &dto.GoodsDTO{
			GoodsDO: *value,
		})
	}
	
	log.Debugf("Successfully searched and listed %d goods, total: %d", len(ret.Items), ret.TotalCount)
	return ret, nil
}

// convertToGoodsDTOList 将DO列表转换为DTO列表 - 分离业务逻辑
func (gs *goodsService) convertToGoodsDTOList(goods *do.GoodsDOList) *dto.GoodsDTOList {
	ret := &dto.GoodsDTOList{
		TotalCount: goods.TotalCount,
		Items:      make([]*dto.GoodsDTO, 0, len(goods.Items)),
	}
	for _, value := range goods.Items {
		ret.Items = append(ret.Items, &dto.GoodsDTO{
			GoodsDO: *value,
		})
	}
	return ret
}

func (gs *goodsService) Get(ctx context.Context, ID uint64) (*dto.GoodsDTO, error) {
	log.Debugf("Getting goods by ID: %d", ID)
	
	// 直接使用预加载的DAO
	goods, err := gs.goodsDAO.Get(ctx, gs.db, ID)
	if err != nil {
		log.Errorf("Failed to get goods by ID %d: %v", ID, err)
		return nil, err
	}
	
	log.Debugf("Successfully got goods by ID: %d", ID)
	return &dto.GoodsDTO{
		GoodsDO: *goods,
	}, nil
}

func (gs *goodsService) Create(ctx context.Context, goods *dto.GoodsDTO) (*dto.GoodsDTO, error) {
	log.Debugf("Creating goods: name=%s, brandID=%d, categoryID=%d", goods.Name, goods.BrandsID, goods.CategoryID)
	
	// 验证品牌和分类是否存在 - 使用预加载的DAO
	_, err := gs.brandDAO.Get(ctx, gs.db, uint64(goods.BrandsID))
	if err != nil {
		log.Errorf("Brand not found: ID=%d, error=%v", goods.BrandsID, err)
		return nil, err
	}

	_, err = gs.categoryDAO.Get(ctx, gs.db, uint64(goods.CategoryID))
	if err != nil {
		log.Errorf("Category not found: ID=%d, error=%v", goods.CategoryID, err)
		return nil, err
	}

	// 开启事务 - 使用保留的工厂管理器
	dataFactory := gs.factoryManager.GetDataFactory()
	txn := dataFactory.Begin()
	defer func() {
		if r := recover(); r != nil {
			txn.Rollback()
			log.Errorf("goodsService.Create panic: %v", r)
			return
		}
	}()

	err = dataFactory.Goods().Create(ctx, txn, &goods.GoodsDO)
	if err != nil {
		log.Errorf("data.Create err: %v", err)
		txn.Rollback()
		return nil, err
	}
	
	// 检查是否启用服务层ES同步
	esOptions := gs.factoryManager.GetEsOptions()
	if esOptions.EnableServiceSync {
		log.Debugf("Service-level ES sync enabled, syncing to Elasticsearch")
		// 构建搜索对象
		searchDO := do.GoodsSearchDO{
			ID:          goods.ID,
			CategoryID:  goods.CategoryID,
			BrandsID:    goods.BrandsID,
			OnSale:      goods.OnSale,
			ShipFree:    goods.ShipFree,
			IsNew:       goods.IsNew,
			IsHot:       goods.IsHot,
			Name:        goods.Name,
			ClickNum:    goods.ClickNum,
			SoldNum:     goods.SoldNum,
			FavNum:      goods.FavNum,
			MarketPrice: goods.MarketPrice,
			GoodsBrief:  goods.GoodsBrief,
			ShopPrice:   goods.ShopPrice,
		}

		err = dataFactory.Search().Goods().Create(ctx, &searchDO)
		if err != nil {
			txn.Rollback()
			return nil, err
		}
	} else {
		log.Debugf("Service-level ES sync disabled, relying on Canal for synchronization")
	}
	
	txn.Commit()
	
	// 获取完整的商品信息（包含关联数据）
	createdGoods, err := gs.Get(ctx, uint64(goods.ID))
	if err != nil {
		return nil, err
	}
	
	return createdGoods, nil
}

func (gs *goodsService) Update(ctx context.Context, goods *dto.GoodsDTO) error {
	log.Debugf("Updating goods: ID=%d, name=%s, brandID=%d, categoryID=%d", goods.ID, goods.Name, goods.BrandsID, goods.CategoryID)
	
	// 验证品牌和分类是否存在 - 使用预加载的DAO
	_, err := gs.brandDAO.Get(ctx, gs.db, uint64(goods.BrandsID))
	if err != nil {
		log.Errorf("Brand not found: ID=%d, error=%v", goods.BrandsID, err)
		return err
	}

	_, err = gs.categoryDAO.Get(ctx, gs.db, uint64(goods.CategoryID))
	if err != nil {
		log.Errorf("Category not found: ID=%d, error=%v", goods.CategoryID, err)
		return err
	}

	// 开启事务 - 使用保留的工厂管理器
	dataFactory := gs.factoryManager.GetDataFactory()
	txn := dataFactory.Begin()
	defer func() {
		if r := recover(); r != nil {
			txn.Rollback()
			log.Errorf("goodsService.Update panic: %v", r)
			return
		}
	}()

	// 更新MySQL数据
	err = dataFactory.Goods().Update(ctx, txn, &goods.GoodsDO)
	if err != nil {
		log.Errorf("data.Update err: %v", err)
		txn.Rollback()
		return err
	}

	// 检查是否启用服务层ES同步
	esOptions := gs.factoryManager.GetEsOptions()
	if esOptions.EnableServiceSync {
		log.Debugf("Service-level ES sync enabled, updating Elasticsearch")
		// 更新ES数据
		searchDO := do.GoodsSearchDO{
			ID:          goods.ID,
			CategoryID:  goods.CategoryID,
			BrandsID:    goods.BrandsID,
			OnSale:      goods.OnSale,
			ShipFree:    goods.ShipFree,
			IsNew:       goods.IsNew,
			IsHot:       goods.IsHot,
			Name:        goods.Name,
			ClickNum:    goods.ClickNum,
			SoldNum:     goods.SoldNum,
			FavNum:      goods.FavNum,
			MarketPrice: goods.MarketPrice,
			GoodsBrief:  goods.GoodsBrief,
			ShopPrice:   goods.ShopPrice,
		}

		err = dataFactory.Search().Goods().Update(ctx, &searchDO)
		if err != nil {
			txn.Rollback()
			return err
		}
	} else {
		log.Debugf("Service-level ES sync disabled, relying on Canal for synchronization")
	}

	txn.Commit()
	log.Infof("Successfully updated goods: ID=%d", goods.ID)
	return nil
}

func (gs *goodsService) Delete(ctx context.Context, ID uint64) error {
	log.Debugf("Deleting goods: ID=%d", ID)
	
	// 开启事务 - 使用保留的工厂管理器
	dataFactory := gs.factoryManager.GetDataFactory()
	txn := dataFactory.Begin()
	defer func() {
		if err := recover(); err != nil {
			txn.Rollback()
			log.Errorf("goodsService.Delete panic: %v", err)
			return
		}
	}()

	// 删除MySQL数据
	err := dataFactory.Goods().Delete(ctx, txn, ID)
	if err != nil {
		log.Errorf("data.DeleteInTxn err: %v", err)
		txn.Rollback()
		return err
	}

	// 检查是否启用服务层ES同步
	esOptions := gs.factoryManager.GetEsOptions()
	if esOptions.EnableServiceSync {
		log.Debugf("Service-level ES sync enabled, deleting from Elasticsearch")
		// 删除ES数据
		err = dataFactory.Search().Goods().Delete(ctx, ID)
		if err != nil {
			txn.Rollback()
			return err
		}
	} else {
		log.Debugf("Service-level ES sync disabled, relying on Canal for synchronization")
	}

	txn.Commit()
	log.Infof("Successfully deleted goods: ID=%d", ID)
	return nil
}

func (gs *goodsService) BatchGet(ctx context.Context, ids []uint64) ([]*dto.GoodsDTO, error) {
	var ret []*dto.GoodsDTO
	var callFuncs []func() error
	var mu sync.Mutex
	for _, value := range ids {
		tmp := value
		callFuncs = append(callFuncs, func() error {
			goodsDTO, err := gs.Get(ctx, tmp)
			mu.Lock()
			ret = append(ret, goodsDTO)
			mu.Unlock()
			return err
		})
	}
	err := mr.Finish(callFuncs...)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

var _ GoodsSrv = &goodsService{}

func retrieveIDs(category *do.CategoryDO) []uint64 {
	ids := []uint64{}
	if category == nil || category.ID == 0 {
		return ids
	}
	ids = append(ids, uint64(category.ID))
	for _, child := range category.SubCategory {
		subids := retrieveIDs(child)
		ids = append(ids, subids...)
	}
	return ids
}