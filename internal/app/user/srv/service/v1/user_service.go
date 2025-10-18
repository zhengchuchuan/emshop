package v1

import (
    "context"

    datav1 "emshop/internal/app/user/srv/data/v1"
    "emshop/internal/app/user/srv/data/v1/interfaces"
    "emshop/internal/app/user/srv/data/v1/mysql"
    "emshop/internal/app/user/srv/domain/dto"
    metav1 "emshop/pkg/common/meta/v1"
    "gorm.io/gorm"
)

type userService struct {
    // 预加载的核心组件（日常CRUD操作）
    userDAO interfaces.UserStore
    db      *gorm.DB

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

func NewUserService(fm *datav1.FactoryManager) UserSrv {
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

