package v1

import (
    "context"

    "emshop/internal/app/user/srv/domain/do"
    "emshop/internal/app/user/srv/domain/dto"
    metav1 "emshop/pkg/common/meta/v1"
    "emshop/pkg/log"
)

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

