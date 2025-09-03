package v1

import (
	"context"
	datav1 "emshop/internal/app/userop/srv/data/v1"
	"emshop/internal/app/userop/srv/data/v1/interfaces"
	"emshop/internal/app/userop/srv/data/v1/mysql"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
	"emshop/pkg/log"
	"gorm.io/gorm"
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
	// 预加载的核心组件（日常CRUD操作）
	messageDAO  interfaces.MessageStore
	db          *gorm.DB
	
	// 保留工厂引用（复杂操作和扩展）
	dataFactory mysql.DataFactory
}

// NewMessageService 创建留言服务
func NewMessageService(dataFactory datav1.DataFactory) MessageService {
	// 适配器模式：将datav1.DataFactory转换为mysql.DataFactory
	mysqlFactory, ok := dataFactory.(mysql.DataFactory)
	if !ok {
		log.Errorf("dataFactory is not mysql.DataFactory type")
		return &messageService{
			dataFactory: dataFactory.(mysql.DataFactory),
		}
	}
	
	return &messageService{
		// 预加载核心组件，避免每次方法调用时重复获取
		messageDAO:  mysqlFactory.Message(),
		db:          mysqlFactory.DB(),
		
		// 保留工厂引用用于复杂操作
		dataFactory: mysqlFactory,
	}
}

func (s *messageService) GetMessageList(ctx context.Context, userID int32) ([]*dto.MessageDTO, int64, error) {
	log.Debugf("Getting message list for user: %d", userID)
	
	// 直接使用预加载的DAO
	messageList, total, err := s.messageDAO.GetMessageList(ctx, s.db, userID)
	if err != nil {
		log.Errorf("Failed to get message list for user %d: %v", userID, err)
		return nil, 0, err
	}
	
	log.Debugf("Successfully got message list for user %d, total: %d", userID, total)
	return messageList, total, nil
}

func (s *messageService) CreateMessage(ctx context.Context, req *MessageCreateRequest) (*do.LeavingMessages, error) {
	log.Debugf("Creating message for user: %d, subject: %s", req.UserID, req.Subject)
	
	message := &do.LeavingMessages{
		User:        req.UserID,
		MessageType: req.MessageType,
		Subject:     req.Subject,
		Message:     req.Message,
		File:        req.File,
	}
	
	// 直接使用预加载的DAO
	createdMessage, err := s.messageDAO.CreateMessage(ctx, s.db, message)
	if err != nil {
		log.Errorf("Failed to create message for user %d: %v", req.UserID, err)
		return nil, err
	}
	
	log.Infof("Successfully created message for user %d, messageID: %d", req.UserID, createdMessage.ID)
	return createdMessage, nil
}