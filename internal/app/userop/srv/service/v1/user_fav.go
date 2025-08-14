package v1

import (
	"context"
	datav1 "emshop/internal/app/userop/srv/data/v1"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
)

// UserFavService 用户收藏服务接口
type UserFavService interface {
	GetUserFavList(ctx context.Context, userID int32, goodsID int32) ([]*dto.UserFavDTO, int64, error)
	CreateUserFav(ctx context.Context, userID int32, goodsID int32) (*do.UserFav, error)
	DeleteUserFav(ctx context.Context, userID int32, goodsID int32) error
	GetUserFavDetail(ctx context.Context, userID int32, goodsID int32) (*do.UserFav, error)
}

type userFavService struct {
	dataFactory datav1.DataFactory
}

// NewUserFavService 创建用户收藏服务
func NewUserFavService(dataFactory datav1.DataFactory) UserFavService {
	return &userFavService{
		dataFactory: dataFactory,
	}
}

func (s *userFavService) GetUserFavList(ctx context.Context, userID int32, goodsID int32) ([]*dto.UserFavDTO, int64, error) {
	return s.dataFactory.UserFav().GetUserFavList(ctx, userID, goodsID)
}

func (s *userFavService) CreateUserFav(ctx context.Context, userID int32, goodsID int32) (*do.UserFav, error) {
	userFav := &do.UserFav{
		User:  userID,
		Goods: goodsID,
	}
	return s.dataFactory.UserFav().CreateUserFav(ctx, userFav)
}

func (s *userFavService) DeleteUserFav(ctx context.Context, userID int32, goodsID int32) error {
	return s.dataFactory.UserFav().DeleteUserFav(ctx, userID, goodsID)
}

func (s *userFavService) GetUserFavDetail(ctx context.Context, userID int32, goodsID int32) (*do.UserFav, error) {
	return s.dataFactory.UserFav().GetUserFavDetail(ctx, userID, goodsID)
}