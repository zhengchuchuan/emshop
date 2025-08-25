package mysql

import (
	"context"
	"emshop/internal/app/payment/srv/data/v1/interfaces"
	"emshop/internal/app/payment/srv/domain/do"
	"emshop/internal/app/pkg/code"
	v1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"
	"emshop/pkg/log"

	"gorm.io/gorm"
)

type paymentLogData struct {
	db *gorm.DB
}

// NewPaymentLogData 创建支付日志数据访问对象
func NewPaymentLogData(db *gorm.DB) interfaces.PaymentLogDataInterface {
	return &paymentLogData{db: db}
}

// Create 创建支付日志
func (p *paymentLogData) Create(ctx context.Context, db *gorm.DB, paymentLog *do.PaymentLogDO) error {
	if db == nil {
		db = p.db
	}
	
	if err := db.WithContext(ctx).Create(paymentLog).Error; err != nil {
		log.Errorf("创建支付日志失败: %v", err)
		return errors.WithCode(code.ErrConnectDB, "创建支付日志失败")
	}
	
	return nil
}

// List 获取支付日志列表
func (p *paymentLogData) List(ctx context.Context, db *gorm.DB, paymentSn string, meta v1.ListMeta) (*do.PaymentLogDOList, error) {
	if db == nil {
		db = p.db
	}
	
	ret := &do.PaymentLogDOList{}
	
	// 构建查询条件
	query := db.WithContext(ctx).Model(&do.PaymentLogDO{}).Where("payment_sn = ?", paymentSn)
	
	// 获取总数
	if err := query.Count(&ret.TotalCount).Error; err != nil {
		log.Errorf("查询支付日志总数失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查询支付日志总数失败")
	}
	
	// 分页
	if meta.PageSize > 0 {
		offset := (meta.Page - 1) * meta.PageSize
		query = query.Offset(int(offset)).Limit(int(meta.PageSize))
	}
	
	// 按时间倒序排列
	query = query.Order("created_at DESC")
	
	// 查询数据
	if err := query.Find(&ret.Items).Error; err != nil {
		log.Errorf("查询支付日志列表失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查询支付日志列表失败")
	}
	
	return ret, nil
}

// GetByAction 根据操作类型获取支付日志
func (p *paymentLogData) GetByAction(ctx context.Context, db *gorm.DB, paymentSn string, action string) ([]*do.PaymentLogDO, error) {
	if db == nil {
		db = p.db
	}
	
	var logs []*do.PaymentLogDO
	if err := db.WithContext(ctx).
		Where("payment_sn = ? AND action = ?", paymentSn, action).
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		log.Errorf("查询支付日志失败: %v", err)
		return nil, errors.WithCode(code.ErrConnectDB, "查询支付日志失败")
	}
	
	return logs, nil
}