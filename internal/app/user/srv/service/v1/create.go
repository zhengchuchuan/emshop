package v1

import (
    "context"

    "emshop/internal/app/pkg/code"
    "emshop/internal/app/user/srv/domain/dto"
    "emshop/internal/app/user/srv/pkg/password"
    "emshop/pkg/errors"
    "emshop/pkg/log"
)

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

