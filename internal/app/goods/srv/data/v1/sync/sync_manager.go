package sync

import (
	"context"
	"fmt"
	"emshop/internal/app/goods/srv/domain/do"
	"emshop/internal/app/goods/srv/data/v1/mysql"
	"emshop/internal/app/goods/srv/data/v1/elasticsearch"
	"emshop/pkg/log"
)

// DataSyncManager 数据同步管理器实现
type DataSyncManager struct {
	dataFactory   mysql.DataFactory
	searchFactory elasticsearch.SearchFactory
}

// NewDataSyncManager 创建数据同步管理器
func NewDataSyncManager(dataFactory mysql.DataFactory, searchFactory elasticsearch.SearchFactory) *DataSyncManager {
	return &DataSyncManager{
		dataFactory:   dataFactory,
		searchFactory: searchFactory,
	}
}

// SyncToSearch 同步数据到搜索引擎
func (dsm *DataSyncManager) SyncToSearch(ctx context.Context, entityType string, entityID uint64) error {
	switch entityType {
	case "goods":
		return dsm.syncGoodsToSearch(ctx, entityID)
	default:
		log.Warnf("unsupported entity type for search sync: %s", entityType)
		return nil
	}
}

// SyncToCache 同步数据到缓存（预留接口）
func (dsm *DataSyncManager) SyncToCache(ctx context.Context, entityType string, entityID uint64) error {
	// TODO: 实现缓存同步逻辑
	log.Debugf("cache sync not implemented for entity: %s, id: %d", entityType, entityID)
	return nil
}

// syncGoodsToSearch 同步商品数据到搜索引擎
func (dsm *DataSyncManager) syncGoodsToSearch(ctx context.Context, goodsID uint64) error {
	// 从主数据库获取商品信息
	goods, err := dsm.dataFactory.Goods().Get(ctx, dsm.dataFactory.DB(), goodsID)
	if err != nil {
		log.Errorf("failed to get goods from database: %v", err)
		return err
	}

	// 转换为搜索对象
	searchGoods := &do.GoodsSearchDO{
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

	// 同步到搜索引擎 - 先尝试更新，如果失败则创建
	err = dsm.searchFactory.Goods().Update(ctx, searchGoods)
	if err != nil {
		log.Debugf("update failed for goods %d, trying create: %v", goodsID, err)
		// 如果更新失败（通常是文档不存在），尝试创建
		err = dsm.searchFactory.Goods().Create(ctx, searchGoods)
		if err != nil {
			log.Errorf("failed to sync goods to search engine: %v", err)
			return err
		}
		log.Debugf("created goods %d in search engine", goodsID)
	} else {
		log.Debugf("updated goods %d in search engine", goodsID)
	}

	log.Debugf("successfully synced goods %d to search engine", goodsID)
	return nil
}

// RemoveFromSearch 从搜索引擎删除数据
func (dsm *DataSyncManager) RemoveFromSearch(ctx context.Context, entityType string, entityID uint64) error {
	switch entityType {
	case "goods":
		return dsm.searchFactory.Goods().Delete(ctx, entityID)
	default:
		log.Warnf("unsupported entity type for search removal: %s", entityType)
		return nil
	}
}

// SyncResult 同步结果
type SyncResult struct {
	SyncedCount int
	FailedCount int
	Errors      []string
}

// SyncAllGoodsToSearch 同步所有商品数据到搜索引擎
func (dsm *DataSyncManager) SyncAllGoodsToSearch(ctx context.Context, forceSync bool, goodsIds []uint64) (*SyncResult, error) {
	result := &SyncResult{
		Errors: make([]string, 0),
	}

	var goodsList []uint64

	if len(goodsIds) > 0 {
		// 指定商品ID同步
		goodsList = goodsIds
	} else {
		// 获取所有商品ID - 使用专用方法
		goodsIds, err := dsm.dataFactory.Goods().GetAllGoodsIDs(ctx, dsm.dataFactory.DB())
		if err != nil {
			log.Errorf("failed to get all goods IDs: %v", err)
			return nil, err
		}
		goodsList = goodsIds
	}

	log.Infof("starting sync %d goods to search engine", len(goodsList))

	// 批量同步
	for _, goodsID := range goodsList {
		err := dsm.syncGoodsToSearch(ctx, goodsID)
		if err != nil {
			result.FailedCount++
			errorMsg := fmt.Sprintf("failed to sync goods %d: %v", goodsID, err)
			result.Errors = append(result.Errors, errorMsg)
			log.Errorf(errorMsg)
			continue
		}
		result.SyncedCount++
	}

	log.Infof("sync completed: synced=%d, failed=%d", result.SyncedCount, result.FailedCount)
	return result, nil
}

// 定义接口
type DataSyncManagerInterface interface {
	SyncToSearch(ctx context.Context, entityType string, entityID uint64) error
	SyncToCache(ctx context.Context, entityType string, entityID uint64) error
	RemoveFromSearch(ctx context.Context, entityType string, entityID uint64) error
	SyncAllGoodsToSearch(ctx context.Context, forceSync bool, goodsIds []uint64) (*SyncResult, error)
}

var _ DataSyncManagerInterface = &DataSyncManager{}