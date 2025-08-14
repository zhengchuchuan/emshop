package mysql

import (
	"context"
	code2 "emshop/gin-micro/code"
	"emshop/internal/app/userop/srv/data/v1/interfaces"
	"emshop/internal/app/userop/srv/domain/do"
	"emshop/internal/app/userop/srv/domain/dto"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type userFavRepository struct {
	db *gorm.DB
}

func NewUserFavRepository(db *gorm.DB) interfaces.UserFavRepository {
	return &userFavRepository{db: db}
}

// GetUserFavList 获取用户收藏列表
func (r *userFavRepository) GetUserFavList(ctx context.Context, userID int32, goodsID int32) ([]*dto.UserFavDTO, int64, error) {
	var userFavs []do.UserFav
	var total int64

	query := r.db.WithContext(ctx).Model(&do.UserFav{})
	
	// 根据条件查询
	if userID > 0 {
		query = query.Where("user = ?", userID)
	}
	if goodsID > 0 {
		query = query.Where("goods = ?", goodsID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		log.Errorf("count user favorites failed: %v", err)
		return nil, 0, errors.WithCode(code2.ErrDatabase, "获取收藏数量失败: %v", err)
	}

	// 获取列表
	if err := query.Find(&userFavs).Error; err != nil {
		log.Errorf("get user favorites failed: %v", err)
		return nil, 0, errors.WithCode(code2.ErrDatabase, "获取收藏列表失败: %v", err)
	}

	// 转换为DTO
	var dtos []*dto.UserFavDTO
	for _, userFav := range userFavs {
		dtos = append(dtos, &dto.UserFavDTO{
			ID:        userFav.ID,
			UserID:    userFav.User,
			GoodsID:   userFav.Goods,
			CreatedAt: userFav.CreatedAt,
		})
	}

	return dtos, total, nil
}

// CreateUserFav 创建用户收藏
func (r *userFavRepository) CreateUserFav(ctx context.Context, userFav *do.UserFav) (*do.UserFav, error) {
	// 检查是否已收藏
	var existingFav do.UserFav
	result := r.db.WithContext(ctx).Where("user = ? AND goods = ?", userFav.User, userFav.Goods).First(&existingFav)
	if result.Error == nil {
		log.Warnf("user %d already favorited goods %d", userFav.User, userFav.Goods)
		return &existingFav, nil // 已存在，返回现有记录
	}

	if err := r.db.WithContext(ctx).Create(userFav).Error; err != nil {
		log.Errorf("create user favorite failed: %v", err)
		return nil, errors.WithCode(code2.ErrDatabase, "添加收藏失败: %v", err)
	}

	log.Infof("user %d favorited goods %d successfully", userFav.User, userFav.Goods)
	return userFav, nil
}

// DeleteUserFav 删除用户收藏
func (r *userFavRepository) DeleteUserFav(ctx context.Context, userID int32, goodsID int32) error {
	result := r.db.WithContext(ctx).Unscoped().Where("goods = ? AND user = ?", goodsID, userID).Delete(&do.UserFav{})
	if result.Error != nil {
		log.Errorf("delete user favorite failed: %v", result.Error)
		return errors.WithCode(code2.ErrDatabase, "删除收藏失败: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		log.Warnf("favorite not found for user %d and goods %d", userID, goodsID)
		return status.Errorf(codes.NotFound, "收藏记录不存在")
	}

	log.Infof("user %d removed favorite for goods %d successfully", userID, goodsID)
	return nil
}

// GetUserFavDetail 获取用户收藏详情（检查是否收藏）
func (r *userFavRepository) GetUserFavDetail(ctx context.Context, userID int32, goodsID int32) (*do.UserFav, error) {
	var userFav do.UserFav
	result := r.db.WithContext(ctx).Where("goods = ? AND user = ?", goodsID, userID).First(&userFav)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "收藏记录不存在")
		}
		log.Errorf("get user favorite detail failed: %v", result.Error)
		return nil, errors.WithCode(code2.ErrDatabase, "获取收藏详情失败: %v", result.Error)
	}

	return &userFav, nil
}