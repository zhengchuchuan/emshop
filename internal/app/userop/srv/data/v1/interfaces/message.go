package interfaces

import (
	"context"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
	"gorm.io/gorm"
)

// MessageStore 留言数据访问接口
type MessageStore interface {
	// GetMessageList 获取用户留言列表
	GetMessageList(ctx context.Context, db *gorm.DB, userID int32) ([]*dto.MessageDTO, int64, error)
	
	// CreateMessage 创建留言
	CreateMessage(ctx context.Context, db *gorm.DB, message *do.LeavingMessages) (*do.LeavingMessages, error)
}