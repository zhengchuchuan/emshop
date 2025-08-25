package mysql

import (
	"context"
	"emshop/internal/app/payment/srv/data/v1/interfaces"
	"emshop/internal/app/payment/srv/domain/do"
	"emshop/internal/app/pkg/code"
	v1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"time"

	"gorm.io/gorm"
)

type paymentOrderData struct {
	db *gorm.DB
}

// NewPaymentOrderData 创建支付订单数据访问对象
func NewPaymentOrderData(db *gorm.DB) interfaces.PaymentOrderDataInterface {
	return &paymentOrderData{db: db}
}

// Create 创建支付订单
func (p *paymentOrderData) Create(ctx context.Context, db *gorm.DB, payment *do.PaymentOrderDO) error {
	if db == nil {
		db = p.db
	}
	
	if err := db.WithContext(ctx).Create(payment).Error; err != nil {
		log.Errorf("创建支付订单失败: %v", err)
		return errors.WithCode(code.ErrConnectDB, "创建支付订单失败")
	}
	
	return nil
}

// Update 更新支付订单
func (p *paymentOrderData) Update(ctx context.Context, db *gorm.DB, payment *do.PaymentOrderDO) error {
	if db == nil {
		db = p.db
	}
	
	if err := db.WithContext(ctx).Save(payment).Error; err != nil {
		log.Errorf("更新支付订单失败: %v", err)
		return errors.WithCode(code.ErrConnectDB, "更新支付订单失败")
	}
	
	return nil
}

// Delete 删除支付订单
func (p *paymentOrderData) Delete(ctx context.Context, db *gorm.DB, paymentSn string) error {
	if db == nil {
		db = p.db
	}
	
	if err := db.WithContext(ctx).Where("payment_sn = ?", paymentSn).Delete(&do.PaymentOrderDO{}).Error; err != nil {
		log.Errorf("删除支付订单失败: %v", err)
		return errors.WithCode(code.ErrConnectDB, "删除支付订单失败")
	}
	
	return nil
}

// Get 根据支付单号获取支付订单
func (p *paymentOrderData) Get(ctx context.Context, db *gorm.DB, paymentSn string) (*do.PaymentOrderDO, error) {
	if db == nil {
		db = p.db
	}
	
	var payment do.PaymentOrderDO
	if err := db.WithContext(ctx).Where("payment_sn = ?", paymentSn).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrPaymentNotFound, "支付订单不存在")
		}
		log.Errorf("查询支付订单失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查询支付订单失败")
	}
	
	return &payment, nil
}

// GetByOrderSn 根据订单号获取支付订单
func (p *paymentOrderData) GetByOrderSn(ctx context.Context, db *gorm.DB, orderSn string) (*do.PaymentOrderDO, error) {
	if db == nil {
		db = p.db
	}
	
	var payment do.PaymentOrderDO
	if err := db.WithContext(ctx).Where("order_sn = ?", orderSn).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WithCode(code.ErrPaymentNotFound, "支付订单不存在")
		}
		log.Errorf("查询支付订单失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查询支付订单失败")
	}
	
	return &payment, nil
}

// List 获取支付订单列表
func (p *paymentOrderData) List(ctx context.Context, db *gorm.DB, userID int32, meta v1.ListMeta, orderby []string) (*do.PaymentOrderDOList, error) {
	if db == nil {
		db = p.db
	}
	
	ret := &do.PaymentOrderDOList{}
	
	// 构建查询条件
	query := db.WithContext(ctx).Model(&do.PaymentOrderDO{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	
	// 获取总数
	if err := query.Count(&ret.TotalCount).Error; err != nil {
		log.Errorf("查询支付订单总数失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查询支付订单总数失败")
	}
	
	// 分页和排序
	if meta.PageSize > 0 {
		offset := (meta.Page - 1) * meta.PageSize
		query = query.Offset(int(offset)).Limit(int(meta.PageSize))
	}
	
	if len(orderby) > 0 {
		for _, order := range orderby {
			query = query.Order(order)
		}
	} else {
		query = query.Order("created_at DESC")
	}
	
	// 查询数据
	if err := query.Find(&ret.Items).Error; err != nil {
		log.Errorf("查询支付订单列表失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查询支付订单列表失败")
	}
	
	return ret, nil
}

// UpdateStatus 更新支付状态
func (p *paymentOrderData) UpdateStatus(ctx context.Context, db *gorm.DB, paymentSn string, status do.PaymentStatus) error {
	if db == nil {
		db = p.db
	}
	
	updates := map[string]interface{}{
		"payment_status": status,
	}
	
	// 如果是支付成功，更新支付时间
	if status == do.PaymentStatusPaid {
		now := time.Now()
		updates["paid_at"] = &now
	}
	
	if err := db.WithContext(ctx).Model(&do.PaymentOrderDO{}).Where("payment_sn = ?", paymentSn).Updates(updates).Error; err != nil {
		log.Errorf("更新支付状态失败: %v", err)
		return errors.WithCode(code.ErrConnectDB, "更新支付状态失败")
	}
	
	return nil
}

// UpdatePaidInfo 更新支付信息
func (p *paymentOrderData) UpdatePaidInfo(ctx context.Context, db *gorm.DB, paymentSn string, thirdPartySn *string, paidAt *time.Time) error {
	if db == nil {
		db = p.db
	}
	
	updates := map[string]interface{}{}
	if thirdPartySn != nil {
		updates["third_party_sn"] = *thirdPartySn
	}
	if paidAt != nil {
		updates["paid_at"] = *paidAt
	}
	
	if len(updates) == 0 {
		return nil
	}
	
	if err := db.WithContext(ctx).Model(&do.PaymentOrderDO{}).Where("payment_sn = ?", paymentSn).Updates(updates).Error; err != nil {
		log.Errorf("更新支付信息失败: %v", err)
		return errors.WithCode(code.ErrConnectDB, "更新支付信息失败")
	}
	
	return nil
}

// FindExpiredPayments 查找过期的支付订单
func (p *paymentOrderData) FindExpiredPayments(ctx context.Context, db *gorm.DB, beforeTime time.Time) ([]*do.PaymentOrderDO, error) {
	if db == nil {
		db = p.db
	}
	
	var payments []*do.PaymentOrderDO
	if err := db.WithContext(ctx).
		Where("payment_status = ? AND expired_at < ?", do.PaymentStatusPending, beforeTime).
		Find(&payments).Error; err != nil {
		log.Errorf("查找过期支付订单失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查找过期支付订单失败")
	}
	
	return payments, nil
}

// CountByStatus 按状态统计支付订单数量
func (p *paymentOrderData) CountByStatus(ctx context.Context, db *gorm.DB, status do.PaymentStatus) (int64, error) {
	if db == nil {
		db = p.db
	}
	
	var count int64
	if err := db.WithContext(ctx).Model(&do.PaymentOrderDO{}).Where("payment_status = ?", status).Count(&count).Error; err != nil {
		log.Errorf("统计支付订单数量失败: %v", err)
		return 0, errors.WithCode(code.ErrConnectDB, "统计支付订单数量失败")
	}
	
	return count, nil
}