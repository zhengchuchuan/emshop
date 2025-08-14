package mysql

import (
	"context"
	code2 "emshop/gin-micro/code"
	"emshop/internal/app/userop/srv/data/v1/interfaces"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) interfaces.MessageRepository {
	return &messageRepository{db: db}
}

// GetMessageList 获取用户留言列表
func (r *messageRepository) GetMessageList(ctx context.Context, userID int32) ([]*dto.MessageDTO, int64, error) {
	var messages []do.LeavingMessages
	var total int64

	result := r.db.WithContext(ctx).Where("user = ?", userID).Find(&messages)
	if result.Error != nil {
		log.Errorf("get message list failed: %v", result.Error)
		return nil, 0, errors.WithCode(code2.ErrDatabase, "获取留言列表失败: %v", result.Error)
	}

	total = result.RowsAffected

	// 转换为DTO
	var dtos []*dto.MessageDTO
	for _, message := range messages {
		dtos = append(dtos, &dto.MessageDTO{
			ID:          message.ID,
			UserID:      message.User,
			MessageType: message.MessageType,
			Subject:     message.Subject,
			Message:     message.Message,
			File:        message.File,
			CreatedAt:   message.CreatedAt,
		})
	}

	return dtos, total, nil
}

// CreateMessage 创建留言
func (r *messageRepository) CreateMessage(ctx context.Context, message *do.LeavingMessages) (*do.LeavingMessages, error) {
	if err := r.db.WithContext(ctx).Create(message).Error; err != nil {
		log.Errorf("create message failed: %v", err)
		return nil, errors.WithCode(code2.ErrDatabase, "创建留言失败: %v", err)
	}

	log.Infof("created message %d for user %d successfully", message.ID, message.User)
	return message, nil
}