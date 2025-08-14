package interfaces

import (
	"context"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
)

// UserFavRepository 用户收藏数据访问接口
type UserFavRepository interface {
	// GetUserFavList 获取用户收藏列表
	GetUserFavList(ctx context.Context, userID int32, goodsID int32) ([]*dto.UserFavDTO, int64, error)
	
	// CreateUserFav 创建用户收藏
	CreateUserFav(ctx context.Context, userFav *do.UserFav) (*do.UserFav, error)
	
	// DeleteUserFav 删除用户收藏 
	DeleteUserFav(ctx context.Context, userID int32, goodsID int32) error
	
	// GetUserFavDetail 获取用户收藏详情（检查是否收藏）
	GetUserFavDetail(ctx context.Context, userID int32, goodsID int32) (*do.UserFav, error)
}