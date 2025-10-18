package v1

import (
    "context"

    "emshop/internal/app/user/srv/domain/dto"
    "emshop/pkg/log"
)

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

