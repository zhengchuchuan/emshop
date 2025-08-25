package v1

import (
	"context"
	"fmt"
	"emshop/internal/app/inventory/srv/data/v1/mysql"
	"emshop/internal/app/pkg/code"
	"emshop/internal/app/pkg/options"
	"emshop/pkg/errors"
	"sort"

	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"

	"emshop/internal/app/inventory/srv/domain/do"
	"emshop/internal/app/inventory/srv/domain/dto"

	"emshop/pkg/log"
)

const (
	inventoryLockPrefix = "inventory_"
	orderLockPrefix     = "order_"
)

type InventorySrv interface {
	//设置库存
	Create(ctx context.Context, inv *dto.InventoryDTO) error

	//根据商品的id查询库存
	Get(ctx context.Context, goodsID uint64) (*dto.InventoryDTO, error)

	//扣减库存
	Sell(ctx context.Context, ordersn string, detail []do.GoodsDetail) error

	//归还库存
	Reback(ctx context.Context, ordersn string, detail []do.GoodsDetail) error

	// TCC分布式事务方法
	TrySell(ctx context.Context, ordersn string, detail []do.GoodsDetail) error   // Try: 冻结库存
	ConfirmSell(ctx context.Context, ordersn string, detail []do.GoodsDetail) error // Confirm: 确认扣减
	CancelSell(ctx context.Context, ordersn string, detail []do.GoodsDetail) error  // Cancel: 取消冻结，释放库存
	
	// Saga分布式事务方法 (用于支付服务集成)
	ReserveStock(ctx context.Context, ordersn string, detail []do.GoodsDetail) error  // 预留库存
	ReleaseReserved(ctx context.Context, ordersn string, detail []do.GoodsDetail) error // 释放预留库存
}

type inventoryService struct {
	data mysql.DataFactory

	redisOptions *options.RedisOptions

	pool redsyncredis.Pool
}

func (is *inventoryService) Create(ctx context.Context, inv *dto.InventoryDTO) error {
	return is.data.Inventorys().Create(ctx, is.data.DB(), &inv.InventoryDO)
}

func (is *inventoryService) Get(ctx context.Context, goodsID uint64) (*dto.InventoryDTO, error) {
	inv, err := is.data.Inventorys().Get(ctx, is.data.DB(), goodsID)
	if err != nil {
		return nil, err
	}
	return &dto.InventoryDTO{InventoryDO: *inv}, nil
}

func (is *inventoryService) Sell(ctx context.Context, ordersn string, details []do.GoodsDetail) error {
	log.Infof("订单%s扣减库存", ordersn)
	//解决了空悬挂的问题
	//先查询刚才插入的记录是否存在，如果存在则说明已经cancel就不能执行了

	rs := redsync.New(is.pool)
	//实际上批量扣减库存的时候， 我们经常会先按照商品的id排序，然后从小大小逐个扣减库存，这样可以减少锁的竞争
	//如果无序的话 那么就有可能订单a 扣减 1,3,4 订单B 扣减 3,2,1
	var detail = do.GoodsDetailList(details)
	sort.Sort(detail)

	txn := is.data.Begin()
	defer func() {
		if err := recover(); err != nil {
			_ = txn.Rollback()
			log.Error("事务进行中出现异常，回滚")
			return
		}
	}()

	sellDetail := do.StockSellDetailDO{
		OrderSn: ordersn,
		Status:  1,
		Detail:  detail,
	}

	for _, goodsInfo := range detail {
		// 使用商品ID作为锁的key，而不是订单号，避免不同商品之间的锁竞争
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodsInfo.Goods))
		if err := mutex.Lock(); err != nil {
			txn.Rollback()
			log.Errorf("订单%s获取商品%d锁失败: %v", ordersn, goodsInfo.Goods, err)
			return errors.WithCode(code.ErrConnectDB, "获取分布式锁失败")
		}

		inv, err := is.data.Inventorys().Get(ctx, is.data.DB(), uint64(goodsInfo.Goods))
		if err != nil {
			mutex.Unlock()
			txn.Rollback()
			log.Errorf("订单%s获取库存失败", ordersn)
			return err
		}

		//判断库存是否充足
		if inv.Stocks < goodsInfo.Num {
			mutex.Unlock()
			txn.Rollback() //回滚
			log.Errorf("商品%d库存%d不足, 现有库存: %d", goodsInfo.Goods, goodsInfo.Num, inv.Stocks)
			return errors.WithCode(code.ErrInvNotEnough, "库存不足")
		}

		err = is.data.Inventorys().Reduce(ctx, txn, uint64(goodsInfo.Goods), int(goodsInfo.Num))
		if err != nil {
			mutex.Unlock()
			txn.Rollback() //回滚
			log.Errorf("订单%s扣减库存失败", ordersn)
			return err
		}

		// 释放锁
		if ok, err := mutex.Unlock(); !ok || err != nil {
			log.Errorf("订单%s释放商品%d锁失败: %v", ordersn, goodsInfo.Goods, err)
		}
	}

	err := is.data.Inventorys().CreateStockSellDetail(ctx, txn, &sellDetail)
	if err != nil {
		txn.Rollback() //回滚
		log.Errorf("订单%s创建扣减库存记录失败", ordersn)
		return err
	}

	txn.Commit()
	return nil
}

func (is *inventoryService) Reback(ctx context.Context, ordersn string, details []do.GoodsDetail) error {
	log.Infof("订单%s归还库存", ordersn)

	rs := redsync.New(is.pool)

	txn := is.data.Begin()
	defer func() {
		if err := recover(); err != nil {
			_ = txn.Rollback()
			log.Error("事务进行中出现异常，回滚")
			return
		}
	}()

	//库存归还的时候有不少细节
	//1. 主动取消 2. 网络问题引起的重试 3. 超时取消 4. 退款取消
	// redis分布式锁
	mutex := rs.NewMutex(orderLockPrefix + ordersn)
	if err := mutex.Lock(); err != nil {
		txn.Rollback() //回滚
		log.Errorf("订单%s获取锁失败", ordersn)
		return err
	}
	sellDetail, err := is.data.Inventorys().GetSellDetail(ctx, txn, ordersn)
	if err != nil {
		txn.Rollback()
		_, err := mutex.Unlock()
		if err != nil {
			log.Errorf("订单%s释放锁出现异常", ordersn)
			return err
		}
		if errors.IsCode(err, code.ErrInvSellDetailNotFound) {
			//空回滚
			log.Errorf("订单%s扣减库存记录不存在, 忽略", ordersn)
			//我应该记录一条数据去记录，说 ordersn已经被cancel了
			return nil
		}
		log.Errorf("订单%s获取扣减库存记录失败", ordersn)
		return err
	}

	if sellDetail.Status == 2 {
		log.Infof("订单%s扣减库存记录已经归还, 忽略", ordersn)
		return nil
	}

	var detail = do.GoodsDetailList(details)
	sort.Sort(detail)

	for _, goodsInfo := range detail {
		inv, err := is.data.Inventorys().Get(ctx, is.data.DB(), uint64(goodsInfo.Goods))
		if err != nil {
			txn.Rollback() //回滚
			log.Errorf("订单%s获取库存失败", ordersn)
			return err
		}
		inv.Stocks += goodsInfo.Num

		err = is.data.Inventorys().Increase(ctx, txn, uint64(goodsInfo.Goods), int(goodsInfo.Num))
		if err != nil {
			txn.Rollback() //回滚
			log.Errorf("订单%s归还库存失败", ordersn)
			return err
		}
	}

	err = is.data.Inventorys().UpdateStockSellDetailStatus(ctx, txn, ordersn, 2)
	if err != nil {
		txn.Rollback() //回滚
		log.Errorf("订单%s更新扣减库存记录失败", ordersn)
		return err
	}

	txn.Commit()
	return nil
}

// TrySell 冻结库存 - TCC分布式事务Try阶段
func (is *inventoryService) TrySell(ctx context.Context, ordersn string, details []do.GoodsDetail) error {
	log.Infof("订单%s冻结库存", ordersn)

	rs := redsync.New(is.pool)
	var detail = do.GoodsDetailList(details)
	sort.Sort(detail)

	txn := is.data.Begin()
	defer func() {
		if err := recover(); err != nil {
			_ = txn.Rollback()
			log.Error("事务进行中出现异常，回滚")
			return
		}
	}()

	// 创建出库单记录
	deliveryDetail := do.DeliveryDO{
		OrderSn: ordersn,
		Status:  "1", // 1. 表示等待支付
	}

	for _, goodsInfo := range detail {
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodsInfo.Goods))
		if err := mutex.Lock(); err != nil {
			txn.Rollback()
			log.Errorf("订单%s获取商品%d锁失败: %v", ordersn, goodsInfo.Goods, err)
			return errors.WithCode(code.ErrInventoryNotFound, "获取分布式锁失败")
		}

		// 查询库存
		inv, err := is.data.Inventorys().Get(ctx, is.data.DB(), uint64(goodsInfo.Goods))
		if err != nil {
			mutex.Unlock()
			txn.Rollback()
			return err
		}

		// 判断可用库存是否充足（实际库存 - 已冻结库存）
		// 这里需要扩展库存模型支持冻结字段，或者通过业务逻辑计算
		if inv.Stocks < goodsInfo.Num {
			mutex.Unlock()
			txn.Rollback()
			return errors.WithCode(code.ErrInvNotEnough, "库存不足")
		}

		// 冻结库存而不是直接扣减
		// TODO: 这里需要使用InventoryNewDO来支持冻结字段
		// 或者通过创建冻结记录来实现

		if ok, err := mutex.Unlock(); !ok || err != nil {
			log.Errorf("订单%s释放商品%d锁失败: %v", ordersn, goodsInfo.Goods, err)
		}
	}

	// 保存出库单
	if err := txn.Create(&deliveryDetail).Error; err != nil {
		txn.Rollback()
		return errors.WithCode(code.ErrInventoryNotFound, "创建出库单失败")
	}

	txn.Commit()
	return nil
}

// ConfirmSell 确认扣减 - TCC分布式事务Confirm阶段
func (is *inventoryService) ConfirmSell(ctx context.Context, ordersn string, details []do.GoodsDetail) error {
	log.Infof("订单%s确认扣减库存", ordersn)

	txn := is.data.Begin()
	defer func() {
		if err := recover(); err != nil {
			_ = txn.Rollback()
			log.Error("事务进行中出现异常，回滚")
			return
		}
	}()

	// 更新出库单状态为支付成功
	if err := txn.Model(&do.DeliveryDO{}).Where("order_sn = ?", ordersn).Update("status", "2").Error; err != nil {
		txn.Rollback()
		return errors.WithCode(code.ErrInventoryNotFound, "更新出库单状态失败")
	}

	// 执行实际的库存扣减
	for _, goodsInfo := range details {
		err := is.data.Inventorys().Reduce(ctx, txn, uint64(goodsInfo.Goods), int(goodsInfo.Num))
		if err != nil {
			txn.Rollback()
			return err
		}
	}

	txn.Commit()
	return nil
}

// CancelSell 取消冻结 - TCC分布式事务Cancel阶段
func (is *inventoryService) CancelSell(ctx context.Context, ordersn string, details []do.GoodsDetail) error {
	log.Infof("订单%s取消库存冻结", ordersn)

	txn := is.data.Begin()
	defer func() {
		if err := recover(); err != nil {
			_ = txn.Rollback()
			log.Error("事务进行中出现异常，回滚")
			return
		}
	}()

	// 更新出库单状态为失败
	if err := txn.Model(&do.DeliveryDO{}).Where("order_sn = ?", ordersn).Update("status", "3").Error; err != nil {
		txn.Rollback()
		return errors.WithCode(code.ErrInventoryNotFound, "更新出库单状态失败")
	}

	// 释放冻结的库存
	// TODO: 这里需要根据具体的冻结实现来释放库存

	txn.Commit()
	return nil
}

// ReserveStock 预留库存 - Saga分布式事务（用于订单提交阶段）
func (is *inventoryService) ReserveStock(ctx context.Context, ordersn string, details []do.GoodsDetail) error {
	log.Infof("订单%s预留库存", ordersn)

	if len(details) == 0 {
		log.Infof("订单%s没有商品，跳过库存预留", ordersn)
		return nil
	}

	rs := redsync.New(is.pool)
	var detail = do.GoodsDetailList(details)
	sort.Sort(detail)

	txn := is.data.Begin()
	defer func() {
		if err := recover(); err != nil {
			_ = txn.Rollback()
			log.Error("事务进行中出现异常，回滚")
			return
		}
	}()

	// 检查是否已经预留过（幂等性处理）
	existingReservation, err := is.data.Inventorys().GetSellDetail(ctx, txn, ordersn)
	if err == nil && existingReservation != nil {
		if existingReservation.Status == 1 {
			log.Infof("订单%s库存已预留，跳过", ordersn)
			txn.Rollback() // 不需要修改，回滚事务
			return nil
		}
	} else if !errors.IsCode(err, code.ErrInvSellDetailNotFound) {
		txn.Rollback()
		return err
	}

	// 创建库存预留记录
	reservationDetail := do.StockSellDetailDO{
		OrderSn: ordersn,
		Status:  1, // 1: 已预留
		Detail:  detail,
	}

	// 批量检查和预留库存
	for _, goodsInfo := range detail {
		mutex := rs.NewMutex(fmt.Sprintf("goods_%d", goodsInfo.Goods))
		if err := mutex.Lock(); err != nil {
			txn.Rollback()
			log.Errorf("订单%s获取商品%d锁失败: %v", ordersn, goodsInfo.Goods, err)
			return errors.WithCode(code.ErrConnectDB, "获取分布式锁失败")
		}

		// 查询库存
		inv, err := is.data.Inventorys().Get(ctx, is.data.DB(), uint64(goodsInfo.Goods))
		if err != nil {
			mutex.Unlock()
			txn.Rollback()
			log.Errorf("订单%s获取商品%d库存失败: %v", ordersn, goodsInfo.Goods, err)
			return err
		}

		// 判断库存是否充足
		if inv.Stocks < goodsInfo.Num {
			mutex.Unlock()
			txn.Rollback()
			log.Errorf("订单%s商品%d库存不足：需要%d，现有%d", ordersn, goodsInfo.Goods, goodsInfo.Num, inv.Stocks)
			return errors.WithCode(code.ErrInvNotEnough, "库存不足")
		}

		// 注意：这里只是创建预留记录，不实际扣减库存
		// 实际的库存扣减将在支付成功后的ConfirmSell中执行

		if ok, err := mutex.Unlock(); !ok || err != nil {
			log.Errorf("订单%s释放商品%d锁失败: %v", ordersn, goodsInfo.Goods, err)
		}
	}

	// 保存预留记录
	if err := is.data.Inventorys().CreateStockSellDetail(ctx, txn, &reservationDetail); err != nil {
		txn.Rollback()
		log.Errorf("订单%s创建库存预留记录失败: %v", ordersn, err)
		return errors.WithCode(code.ErrConnectDB, "创建库存预留记录失败")
	}

	txn.Commit()
	log.Infof("订单%s库存预留成功", ordersn)
	return nil
}

// ReleaseReserved 释放预留库存 - Saga分布式事务补偿（用于订单提交失败时）
func (is *inventoryService) ReleaseReserved(ctx context.Context, ordersn string, details []do.GoodsDetail) error {
	log.Infof("订单%s释放预留库存", ordersn)

	txn := is.data.Begin()
	defer func() {
		if err := recover(); err != nil {
			_ = txn.Rollback()
			log.Error("事务进行中出现异常，回滚")
			return
		}
	}()

	// 查询预留记录
	reservationDetail, err := is.data.Inventorys().GetSellDetail(ctx, txn, ordersn)
	if err != nil {
		txn.Rollback()
		if errors.IsCode(err, code.ErrInvSellDetailNotFound) {
			// 没有预留记录，说明预留操作未成功，直接返回成功（幂等性）
			log.Infof("订单%s没有库存预留记录，忽略释放操作", ordersn)
			return nil
		}
		log.Errorf("订单%s查询预留记录失败: %v", ordersn, err)
		return err
	}

	// 检查状态，避免重复释放
	if reservationDetail.Status == 0 {
		// 已释放
		log.Infof("订单%s库存预留已释放，忽略", ordersn)
		txn.Rollback()
		return nil
	}

	// 更新预留记录状态为已释放
	if err := is.data.Inventorys().UpdateStockSellDetailStatus(ctx, txn, ordersn, 0); err != nil {
		txn.Rollback()
		log.Errorf("订单%s更新预留记录状态失败: %v", ordersn, err)
		return errors.WithCode(code.ErrConnectDB, "更新预留记录状态失败")
	}

	txn.Commit()
	log.Infof("订单%s释放预留库存成功", ordersn)
	return nil
}

func newInventoryService(s *service) *inventoryService {
	return &inventoryService{data: s.data, redisOptions: s.redisOptions, pool: s.pool}
}

var _ InventorySrv = &inventoryService{}
