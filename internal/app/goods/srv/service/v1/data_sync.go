package v1

import (
	"context"
	dataV1 "emshop/internal/app/goods/srv/data/v1"
	"emshop/internal/app/goods/srv/data/v1/sync"
	"emshop/pkg/log"
)

type dataSyncService struct {
	factoryManager *dataV1.FactoryManager
}

func newDataSync(srv *service) *dataSyncService {
	return &dataSyncService{
		factoryManager: srv.factoryManager,
	}
}

func (ds *dataSyncService) SyncGoodsData(ctx context.Context, forceSync bool, goodsIds []uint64) (*sync.SyncResult, error) {
	log.Infof("starting data sync: forceSync=%v, goodsIds=%v", forceSync, goodsIds)
	
	syncManager := ds.factoryManager.GetSyncManager()
	result, err := syncManager.SyncAllGoodsToSearch(ctx, forceSync, goodsIds)
	if err != nil {
		log.Errorf("data sync failed: %v", err)
		return nil, err
	}
	
	log.Infof("data sync completed: synced=%d, failed=%d", result.SyncedCount, result.FailedCount)
	return result, nil
}

var _ DataSyncSrv = &dataSyncService{}