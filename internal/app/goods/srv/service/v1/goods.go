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
)

type goodsService struct {
	factoryManager *dataV1.FactoryManager
}

func newGoods(srv *service) *goodsService {
	return &goodsService{
		factoryManager: srv.factoryManager,
	}
}

func (gs *goodsService) List(ctx context.Context, opts metav1.ListMeta, req *proto.GoodsFilterRequest, orderby []string) (*dto.GoodsDTOList, error) {
	dataFactory := gs.factoryManager.GetDataFactory()
	
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
		category, err := dataFactory.Categorys().Get(ctx, uint64(*req.TopCategory))
		if err != nil {
			log.Errorf("categoryData.Get err: %v", err)
			return nil, err
		}
		// 树形结构遍历
		for _, value := range retrieveIDs(category) {
			categoryIDs = append(categoryIDs, value)
		}
	}

	// 如果没有搜索条件，直接查询MySQL
	if !hasSearchConditions {
		log.Debugf("No search conditions, querying MySQL directly")
		goods, err := dataFactory.Goods().List(ctx, orderby, opts)
		if err != nil {
			log.Errorf("data.List err: %v", err)
			return nil, err
		}
		
		var ret dto.GoodsDTOList
		ret.TotalCount = goods.TotalCount
		for _, value := range goods.Items {
			ret.Items = append(ret.Items, &dto.GoodsDTO{
				GoodsDO: *value,
			})
		}
		return &ret, nil
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

	// 通过搜索引擎查询
	goodsList, err := dataFactory.Search().Goods().Search(ctx, searchReq)
	if err != nil {
		log.Errorf("searchData.Search err: %v", err)
		return nil, err
	}

	log.Debugf("Search es data: %v", goodsList)

	goodsIDs := []uint64{}
	for _, value := range goodsList.Items {
		goodsIDs = append(goodsIDs, uint64(value.ID))
	}

	// 通过id批量查询mysql数据
	goods, err := dataFactory.Goods().ListByIDs(ctx, goodsIDs, orderby)
	if err != nil {
		log.Errorf("data.ListByIDs err: %v", err)
		return nil, err
	}
	
	var ret dto.GoodsDTOList
	ret.TotalCount = int64(goodsList.TotalCount)
	for _, value := range goods.Items {
		ret.Items = append(ret.Items, &dto.GoodsDTO{
			GoodsDO: *value,
		})
	}
	return &ret, nil
}

func (gs *goodsService) Get(ctx context.Context, ID uint64) (*dto.GoodsDTO, error) {
	dataFactory := gs.factoryManager.GetDataFactory()
	goods, err := dataFactory.Goods().Get(ctx, ID)
	if err != nil {
		log.Errorf("data.Get err: %v", err)
		return nil, err
	}
	return &dto.GoodsDTO{
		GoodsDO: *goods,
	}, nil
}

func (gs *goodsService) Create(ctx context.Context, goods *dto.GoodsDTO) (*dto.GoodsDTO, error) {
	dataFactory := gs.factoryManager.GetDataFactory()
	
	// 验证品牌和分类是否存在
	_, err := dataFactory.Brands().Get(ctx, uint64(goods.BrandsID))
	if err != nil {
		return nil, err
	}

	_, err = dataFactory.Categorys().Get(ctx, uint64(goods.CategoryID))
	if err != nil {
		return nil, err
	}

	// 开启事务
	txn := dataFactory.Begin()
	defer func() {
		if err := recover(); err != nil {
			txn.Rollback()
			log.Errorf("goodsService.Create panic: %v", err)
			return
		}
	}()

	err = dataFactory.Goods().CreateInTxn(ctx, txn, &goods.GoodsDO)
	if err != nil {
		log.Errorf("data.CreateInTxn err: %v", err)
		txn.Rollback()
		return nil, err
	}
	
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
	
	txn.Commit()
	
	// 获取完整的商品信息（包含关联数据）
	createdGoods, err := gs.Get(ctx, uint64(goods.ID))
	if err != nil {
		return nil, err
	}
	
	return createdGoods, nil
}

func (gs *goodsService) Update(ctx context.Context, goods *dto.GoodsDTO) error {
	dataFactory := gs.factoryManager.GetDataFactory()
	
	// 验证品牌和分类是否存在
	_, err := dataFactory.Brands().Get(ctx, uint64(goods.BrandsID))
	if err != nil {
		return err
	}

	_, err = dataFactory.Categorys().Get(ctx, uint64(goods.CategoryID))
	if err != nil {
		return err
	}

	// 开启事务
	txn := dataFactory.Begin()
	defer func() {
		if err := recover(); err != nil {
			txn.Rollback()
			log.Errorf("goodsService.Update panic: %v", err)
			return
		}
	}()

	// 更新MySQL数据
	err = dataFactory.Goods().UpdateInTxn(ctx, txn, &goods.GoodsDO)
	if err != nil {
		log.Errorf("data.UpdateInTxn err: %v", err)
		txn.Rollback()
		return err
	}

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

	txn.Commit()
	return nil
}

func (gs *goodsService) Delete(ctx context.Context, ID uint64) error {
	dataFactory := gs.factoryManager.GetDataFactory()
	
	// 开启事务
	txn := dataFactory.Begin()
	defer func() {
		if err := recover(); err != nil {
			txn.Rollback()
			log.Errorf("goodsService.Delete panic: %v", err)
			return
		}
	}()

	// 删除MySQL数据
	err := dataFactory.Goods().DeleteInTxn(ctx, txn, ID)
	if err != nil {
		log.Errorf("data.DeleteInTxn err: %v", err)
		txn.Rollback()
		return err
	}

	// 删除ES数据
	err = dataFactory.Search().Goods().Delete(ctx, ID)
	if err != nil {
		txn.Rollback()
		return err
	}

	txn.Commit()
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