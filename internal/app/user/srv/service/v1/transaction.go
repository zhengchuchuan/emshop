package v1

import (
    "context"

    "emshop/internal/app/pkg/code"
    "emshop/internal/app/user/srv/domain/dto"
    "emshop/internal/app/user/srv/pkg/password"
    "emshop/pkg/errors"
    "emshop/pkg/log"
)

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

