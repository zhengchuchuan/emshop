package v1

import (
	"context"
	pb "emshop/api/userop/v1"
	servicev1 "emshop/internal/app/userop/srv/service/v1"
	"emshop/pkg/log"
)

// UserFavController 用户收藏控制器
type UserFavController struct {
	pb.UnimplementedUserOpServer
	service servicev1.Service
}

// NewUserFavController 创建用户收藏控制器
func NewUserFavController(service servicev1.Service) *UserFavController {
	return &UserFavController{
		service: service,
	}
}

// UserFavList 获取用户收藏列表
func (c *UserFavController) UserFavList(ctx context.Context, req *pb.UserFavListRequest) (*pb.UserFavListResponse, error) {
	log.Infof("UserFavList request: user_id=%d", req.UserId)

	favs, total, err := c.service.UserFavService().GetUserFavList(ctx, req.UserId, 0)
	if err != nil {
		log.Errorf("get user fav list failed: %v", err)
		return nil, err
	}

	var data []*pb.UserFavResponse
	for _, fav := range favs {
		data = append(data, &pb.UserFavResponse{
			UserId:  fav.UserID,
			GoodsId: fav.GoodsID,
		})
	}

	return &pb.UserFavListResponse{
		Total: int32(total),
		Data:  data,
	}, nil
}

// CreateUserFav 创建用户收藏
func (c *UserFavController) CreateUserFav(ctx context.Context, req *pb.UserFavRequest) (*pb.UserFavResponse, error) {
	log.Infof("CreateUserFav request: user_id=%d, goods_id=%d", req.UserId, req.GoodsId)

	userFav, err := c.service.UserFavService().CreateUserFav(ctx, req.UserId, req.GoodsId)
	if err != nil {
		log.Errorf("create user fav failed: %v", err)
		return nil, err
	}

	return &pb.UserFavResponse{
		UserId:  userFav.User,
		GoodsId: userFav.Goods,
	}, nil
}

// DeleteUserFav 删除用户收藏
func (c *UserFavController) DeleteUserFav(ctx context.Context, req *pb.UserFavRequest) (*pb.UserFavResponse, error) {
	log.Infof("DeleteUserFav request: user_id=%d, goods_id=%d", req.UserId, req.GoodsId)

	err := c.service.UserFavService().DeleteUserFav(ctx, req.UserId, req.GoodsId)
	if err != nil {
		log.Errorf("delete user fav failed: %v", err)
		return nil, err
	}

	return &pb.UserFavResponse{}, nil
}

// GetUserFavDetail 获取用户收藏详情（检查是否收藏）
func (c *UserFavController) GetUserFavDetail(ctx context.Context, req *pb.UserFavRequest) (*pb.UserFavResponse, error) {
	log.Infof("GetUserFavDetail request: user_id=%d, goods_id=%d", req.UserId, req.GoodsId)

	userFav, err := c.service.UserFavService().GetUserFavDetail(ctx, req.UserId, req.GoodsId)
	if err != nil {
		log.Errorf("get user fav detail failed: %v", err)
		return nil, err
	}

	return &pb.UserFavResponse{
		UserId:  userFav.User,
		GoodsId: userFav.Goods,
	}, nil
}