package v1

import (
	"context"
	"emshop/internal/app/pkg/code"
	v1 "emshop/internal/app/user/srv/data/v1"
	"emshop/internal/app/user/srv/data/v1/interfaces"
	"emshop/internal/app/user/srv/data/v1/mysql"
	"emshop/internal/app/user/srv/domain/do"
	"emshop/internal/app/user/srv/domain/dto"
	"emshop/internal/app/user/srv/pkg/password"
	metav1 "emshop/pkg/common/meta/v1"
	"emshop/pkg/errors"
	"emshop/pkg/log"
	"gorm.io/gorm"
)


type userService struct {
	// 预加载的核心组件（日常CRUD操作）
	userDAO     interfaces.UserStore
	db          *gorm.DB
	
	// 保留工厂引用（复杂操作和扩展）
	dataFactory mysql.DataFactory
}

type UserSrv interface {
	List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*dto.UserDTOList, error)
	Create(ctx context.Context, user *dto.UserDTO) error
	Update(ctx context.Context, user *dto.UserDTO) error
	GetByID(ctx context.Context, ID uint64) (*dto.UserDTO, error)
	GetByMobile(ctx context.Context, mobile string) (*dto.UserDTO, error)
}



func NewUserService(fm *v1.FactoryManager) UserSrv {
	dataFactory := fm.GetDataFactory()
	
	return &userService{
		// 预加载核心组件，避免每次方法调用时重复获取
		userDAO:     dataFactory.Users(),
		db:          dataFactory.DB(),
		
		// 保留工厂引用用于复杂操作
		dataFactory: dataFactory,
	}
}
var _ UserSrv = &userService{}




func (u *userService) Create(ctx context.Context, user *dto.UserDTO) error {
	log.Debugf("Creating user with mobile: %s", user.Mobile)
	
	// 检查用户是否存在 - 直接使用预加载的DAO
	_, err := u.userDAO.GetByMobile(ctx, u.db, user.Mobile)
	if err != nil && errors.IsCode(err, code.ErrUserNotFound) {
		// 密码加密逻辑在service层
		encryptedPassword, err := password.EncryptPassword(user.Password)
		if err != nil {
			log.Errorf("Password encryption failed for user %s: %v", user.Mobile, err)
			return errors.WithCode(code.ErrEncryptionFailed, "密码加密失败")
		}
		
		// 更新用户密码为加密后的密码
		user.Password = encryptedPassword
		
		// 直接创建用户 - 使用预加载的DAO和DB
		if err := u.userDAO.Create(ctx, u.db, &user.UserDO); err != nil {
			log.Errorf("Failed to create user %s: %v", user.Mobile, err)
			return err
		}
		
		log.Infof("Successfully created user: %s", user.Mobile)
		return nil
	}

	// 用户已存在或其他数据访问错误
	if err != nil {
		log.Errorf("Database error while checking user %s: %v", user.Mobile, err)
		return err
	}
	
	log.Warnf("User already exists: %s", user.Mobile)
	return errors.WithCode(code.ErrUserAlreadyExists, "用户已经存在")
}

func (u *userService) Update(ctx context.Context, user *dto.UserDTO) error {
	log.Debugf("Updating user ID: %d", user.ID)
	
	// 先查询用户是否存在 - 直接使用预加载的DAO
	_, err := u.userDAO.Get(ctx, u.db, uint64(user.ID))
	if err != nil {
		log.Errorf("User not found for update, ID: %d, error: %v", user.ID, err)
		return err
	}

	// 直接更新用户
	if err := u.userDAO.Update(ctx, u.db, &user.UserDO); err != nil {
		log.Errorf("Failed to update user ID: %d, error: %v", user.ID, err)
		return err
	}
	
	log.Infof("Successfully updated user ID: %d", user.ID)
	return nil
}

func (u *userService) GetByID(ctx context.Context, ID uint64) (*dto.UserDTO, error) {
	// 直接使用预加载的DAO - 无需每次获取工厂
	userDO, err := u.userDAO.Get(ctx, u.db, ID)
	if err != nil {
		log.Errorf("Failed to get user by ID: %d, error: %v", ID, err)
		return nil, err
	}

	log.Debugf("Successfully retrieved user ID: %d", ID)
	return &dto.UserDTO{UserDO: *userDO}, nil
}

func (u *userService) GetByMobile(ctx context.Context, mobile string) (*dto.UserDTO, error) {
	// 直接使用预加载的DAO - 无需每次获取工厂
	userDO, err := u.userDAO.GetByMobile(ctx, u.db, mobile)
	if err != nil {
		log.Errorf("Failed to get user by mobile: %s, error: %v", mobile, err)
		return nil, err
	}

	log.Debugf("Successfully retrieved user by mobile: %s", mobile)
	return &dto.UserDTO{UserDO: *userDO}, nil
}


func (u *userService) List(ctx context.Context, orderby []string, opts metav1.ListMeta) (*dto.UserDTOList, error) {
	log.Debugf("Listing users with page: %d, size: %d", opts.Page, opts.PageSize)
	
	// 数据访问层：直接使用预加载的DAO，无需每次获取工厂
	doList, err := u.userDAO.List(ctx, u.db, orderby, opts)
	if err != nil {
		log.Errorf("Failed to list users: %v", err)
		return nil, err
	}

	// 业务逻辑层：数据转换逻辑分离到专门方法
	userDTOList := u.convertToUserDTOList(doList)
	
	log.Debugf("Successfully listed %d users, total: %d", len(userDTOList.Items), userDTOList.TotalCount)
	return userDTOList, nil
}

// convertToUserDTOList 将DO列表转换为DTO列表 - 分离业务逻辑
func (u *userService) convertToUserDTOList(doList *do.UserDOList) *dto.UserDTOList {
	userDTOList := &dto.UserDTOList{
		TotalCount: doList.TotalCount,
		Items:      make([]*dto.UserDTO, 0, len(doList.Items)),
	}
	
	for _, userDO := range doList.Items {
		userDTO := &dto.UserDTO{UserDO: *userDO}
		userDTOList.Items = append(userDTOList.Items, userDTO)
	}
	
	return userDTOList
}

// CreateWithTransaction 演示复杂事务操作 - 使用保留的工厂引用
func (u *userService) CreateWithTransaction(ctx context.Context, user *dto.UserDTO) error {
	log.Debugf("Creating user with transaction: %s", user.Mobile)
	
	// 对于事务操作，使用保留的工厂获取事务DB
	txDB := u.dataFactory.Begin()
	defer func() {
		if r := recover(); r != nil {
			txDB.Rollback()
			log.Errorf("Transaction panic during user creation: %v", r)
		}
	}()
	
	// 检查用户是否存在（使用事务DB）
	_, err := u.userDAO.GetByMobile(ctx, txDB, user.Mobile)
	if err != nil && errors.IsCode(err, code.ErrUserNotFound) {
		// 密码加密
		encryptedPassword, err := password.EncryptPassword(user.Password)
		if err != nil {
			txDB.Rollback()
			return errors.WithCode(code.ErrEncryptionFailed, "密码加密失败")
		}
		user.Password = encryptedPassword
		
		// 创建用户（使用事务DB）
		if err := u.userDAO.Create(ctx, txDB, &user.UserDO); err != nil {
			txDB.Rollback()
			log.Errorf("Failed to create user in transaction: %v", err)
			return err
		}
		
		// 这里可以添加其他相关操作，如创建用户profile等
		// 所有操作都在同一事务中
		
		// 提交事务
		if err := txDB.Commit().Error; err != nil {
			log.Errorf("Failed to commit user creation transaction: %v", err)
			return err
		}
		
		log.Infof("Successfully created user with transaction: %s", user.Mobile)
		return nil
	}
	
	// 用户已存在或其他错误
	txDB.Rollback()
	if err != nil {
		return err
	}
	return errors.WithCode(code.ErrUserAlreadyExists, "用户已经存在")
}
