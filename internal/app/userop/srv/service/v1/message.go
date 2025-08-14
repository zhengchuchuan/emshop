package v1

import (
	"context"
	datav1 "emshop/internal/app/userop/srv/data/v1"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
)

// MessageService 留言服务接口
type MessageService interface {
	GetMessageList(ctx context.Context, userID int32) ([]*dto.MessageDTO, int64, error)
	CreateMessage(ctx context.Context, req *MessageCreateRequest) (*do.LeavingMessages, error)
}

// MessageCreateRequest 创建留言请求
type MessageCreateRequest struct {
	UserID      int32
	MessageType int32
	Subject     string
	Message     string
	File        string
}

type messageService struct {
	dataFactory datav1.DataFactory
}

// NewMessageService 创建留言服务
func NewMessageService(dataFactory datav1.DataFactory) MessageService {
	return &messageService{
		dataFactory: dataFactory,
	}
}

func (s *messageService) GetMessageList(ctx context.Context, userID int32) ([]*dto.MessageDTO, int64, error) {
	return s.dataFactory.Message().GetMessageList(ctx, userID)
}

func (s *messageService) CreateMessage(ctx context.Context, req *MessageCreateRequest) (*do.LeavingMessages, error) {
	message := &do.LeavingMessages{
		User:        req.UserID,
		MessageType: req.MessageType,
		Subject:     req.Subject,
		Message:     req.Message,
		File:        req.File,
	}
	return s.dataFactory.Message().CreateMessage(ctx, message)
}