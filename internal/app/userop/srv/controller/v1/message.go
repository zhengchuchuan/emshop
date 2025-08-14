package v1

import (
	"context"
	pb "emshop/api/userop/v1"
	servicev1 "emshop/internal/app/userop/srv/service/v1"
	"emshop/pkg/log"
)

// MessageController 留言控制器
type MessageController struct {
	pb.UnimplementedUserOpServer
	service servicev1.Service
}

// NewMessageController 创建留言控制器
func NewMessageController(service servicev1.Service) *MessageController {
	return &MessageController{
		service: service,
	}
}

// MessageList 获取留言列表
func (c *MessageController) MessageList(ctx context.Context, req *pb.MessageRequest) (*pb.MessageListResponse, error) {
	log.Infof("MessageList request: user_id=%d", req.UserId)

	messages, total, err := c.service.MessageService().GetMessageList(ctx, req.UserId)
	if err != nil {
		log.Errorf("get message list failed: %v", err)
		return nil, err
	}

	var data []*pb.MessageResponse
	for _, msg := range messages {
		data = append(data, &pb.MessageResponse{
			Id:          msg.ID,
			UserId:      msg.UserID,
			MessageType: msg.MessageType,
			Subject:     msg.Subject,
			Message:     msg.Message,
			File:        msg.File,
		})
	}

	return &pb.MessageListResponse{
		Total: int32(total),
		Data:  data,
	}, nil
}

// CreateMessage 创建留言
func (c *MessageController) CreateMessage(ctx context.Context, req *pb.MessageRequest) (*pb.MessageResponse, error) {
	log.Infof("CreateMessage request: user_id=%d, message_type=%d", req.UserId, req.MessageType)

	createReq := &servicev1.MessageCreateRequest{
		UserID:      req.UserId,
		MessageType: req.MessageType,
		Subject:     req.Subject,
		Message:     req.Message,
		File:        req.File,
	}

	message, err := c.service.MessageService().CreateMessage(ctx, createReq)
	if err != nil {
		log.Errorf("create message failed: %v", err)
		return nil, err
	}

	return &pb.MessageResponse{
		Id:          message.ID,
		UserId:      message.User,
		MessageType: message.MessageType,
		Subject:     message.Subject,
		Message:     message.Message,
		File:        message.File,
	}, nil
}