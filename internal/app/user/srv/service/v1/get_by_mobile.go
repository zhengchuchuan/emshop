package v1

import (
    "context"

    "emshop/internal/app/user/srv/domain/dto"
    "emshop/pkg/log"
)

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

