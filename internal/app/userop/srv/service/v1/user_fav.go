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

// UserFavService 用户收藏服务接口
type UserFavService interface {
	GetUserFavList(ctx context.Context, userID int32, goodsID int32) ([]*dto.UserFavDTO, int64, error)
	CreateUserFav(ctx context.Context, userID int32, goodsID int32) (*do.UserFav, error)
	DeleteUserFav(ctx context.Context, userID int32, goodsID int32) error
	GetUserFavDetail(ctx context.Context, userID int32, goodsID int32) (*do.UserFav, error)
}

type userFavService struct {
	// 预加载的核心组件（日常CRUD操作）
	userFavDAO  interfaces.UserFavStore
	db          *gorm.DB
	
	// 保留工厂引用（复杂操作和扩展）
	dataFactory mysql.DataFactory
}

// NewUserFavService 创建用户收藏服务
func NewUserFavService(dataFactory datav1.DataFactory) UserFavService {
	// 适配器模式：将datav1.DataFactory转换为mysql.DataFactory
	mysqlFactory, ok := dataFactory.(mysql.DataFactory)
	if !ok {
		log.Errorf("dataFactory is not mysql.DataFactory type")
		// 如果类型断言失败，使用原有方式
		return &userFavService{
			dataFactory: dataFactory.(mysql.DataFactory),
		}
	}
	
	return &userFavService{
		// 预加载核心组件，避免每次方法调用时重复获取
		userFavDAO:  mysqlFactory.UserFav(),
		db:          mysqlFactory.DB(),
		
		// 保留工厂引用用于复杂操作
		dataFactory: mysqlFactory,
	}
}

func (s *userFavService) GetUserFavList(ctx context.Context, userID int32, goodsID int32) ([]*dto.UserFavDTO, int64, error) {
	log.Debugf("Getting user fav list: userID=%d, goodsID=%d", userID, goodsID)
	
	// 直接使用预加载的DAO
	favList, total, err := s.userFavDAO.GetUserFavList(ctx, s.db, userID, goodsID)
	if err != nil {
		log.Errorf("Failed to get user fav list: userID=%d, goodsID=%d, error=%v", userID, goodsID, err)
		return nil, 0, err
	}
	
	log.Debugf("Successfully got user fav list: userID=%d, total=%d", userID, total)
	return favList, total, nil
}

func (s *userFavService) CreateUserFav(ctx context.Context, userID int32, goodsID int32) (*do.UserFav, error) {
	log.Debugf("Creating user fav: userID=%d, goodsID=%d", userID, goodsID)
	
	userFav := &do.UserFav{
		User:  userID,
		Goods: goodsID,
	}
	
	// 直接使用预加载的DAO
	createdFav, err := s.userFavDAO.CreateUserFav(ctx, s.db, userFav)
	if err != nil {
		log.Errorf("Failed to create user fav: userID=%d, goodsID=%d, error=%v", userID, goodsID, err)
		return nil, err
	}
	
	log.Infof("Successfully created user fav: userID=%d, goodsID=%d", userID, goodsID)
	return createdFav, nil
}

func (s *userFavService) DeleteUserFav(ctx context.Context, userID int32, goodsID int32) error {
	log.Debugf("Deleting user fav: userID=%d, goodsID=%d", userID, goodsID)
	
	// 直接使用预加载的DAO
	err := s.userFavDAO.DeleteUserFav(ctx, s.db, userID, goodsID)
	if err != nil {
		log.Errorf("Failed to delete user fav: userID=%d, goodsID=%d, error=%v", userID, goodsID, err)
		return err
	}
	
	log.Infof("Successfully deleted user fav: userID=%d, goodsID=%d", userID, goodsID)
	return nil
}

func (s *userFavService) GetUserFavDetail(ctx context.Context, userID int32, goodsID int32) (*do.UserFav, error) {
	log.Debugf("Getting user fav detail: userID=%d, goodsID=%d", userID, goodsID)
	
	// 直接使用预加载的DAO
	favDetail, err := s.userFavDAO.GetUserFavDetail(ctx, s.db, userID, goodsID)
	if err != nil {
		log.Errorf("Failed to get user fav detail: userID=%d, goodsID=%d, error=%v", userID, goodsID, err)
		return nil, err
	}
	
	log.Debugf("Successfully got user fav detail: userID=%d, goodsID=%d", userID, goodsID)
	return favDetail, nil
}