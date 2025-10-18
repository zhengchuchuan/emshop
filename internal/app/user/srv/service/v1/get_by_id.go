package v1

import (
    "context"

    "emshop/internal/app/user/srv/domain/dto"
    "emshop/pkg/log"
)

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

