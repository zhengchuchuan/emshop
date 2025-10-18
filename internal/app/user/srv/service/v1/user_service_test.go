package v1

import (
    "context"
    "testing"

    "emshop/internal/app/pkg/code"
    "emshop/internal/app/user/srv/domain/do"
    "emshop/internal/app/user/srv/domain/dto"
    "emshop/internal/app/user/srv/pkg/password"
    errorsx "emshop/pkg/errors"
    metav1 "emshop/pkg/common/meta/v1"
    dbpkg "emshop/pkg/db"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gorm.io/gorm"
)

// mockUserStore implements interfaces.UserStore for unit testing
type mockUserStore struct{ mock.Mock }

func (m *mockUserStore) Get(ctx context.Context, db *gorm.DB, id uint64) (*do.UserDO, error) {
    args := m.Called(ctx, db, id)
    if v := args.Get(0); v != nil {
        return v.(*do.UserDO), args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *mockUserStore) GetByMobile(ctx context.Context, db *gorm.DB, mobile string) (*do.UserDO, error) {
    args := m.Called(ctx, db, mobile)
    if v := args.Get(0); v != nil {
        return v.(*do.UserDO), args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *mockUserStore) List(ctx context.Context, db *gorm.DB, orderby []string, opts metav1.ListMeta) (*do.UserDOList, error) {
    args := m.Called(ctx, db, orderby, opts)
    if v := args.Get(0); v != nil {
        return v.(*do.UserDOList), args.Error(1)
    }
    return nil, args.Error(1)
}

func (m *mockUserStore) Create(ctx context.Context, db *gorm.DB, user *do.UserDO) error {
    args := m.Called(ctx, db, user)
    return args.Error(0)
}

func (m *mockUserStore) Update(ctx context.Context, db *gorm.DB, user *do.UserDO) error {
    args := m.Called(ctx, db, user)
    return args.Error(0)
}

func newUnitSvc() (*userService, *mockUserStore, *gorm.DB) {
    dao := &mockUserStore{}
    gdb := &gorm.DB{}
    return &userService{userDAO: dao, db: gdb}, dao, gdb
}

func TestUserService_Create_Success(t *testing.T) {
    svc, dao, gdb := newUnitSvc()

    dao.On("GetByMobile", mock.Anything, gdb, "13800138000").
        Return((*do.UserDO)(nil), errorsx.WithCode(code.ErrUserNotFound, "not found"))

    var created do.UserDO
    dao.On("Create", mock.Anything, gdb, mock.AnythingOfType("*do.UserDO")).
        Return(nil).
        Run(func(args mock.Arguments) { created = *args.Get(2).(*do.UserDO) })

    err := svc.Create(context.Background(), &dto.UserDTO{
        UserDO: do.UserDO{Mobile: "13800138000", Password: "plainPwd"},
    })

    assert.NoError(t, err)
    assert.NotEqual(t, "plainPwd", created.Password)
    assert.True(t, password.VerifyPassword("plainPwd", created.Password))
    dao.AssertExpectations(t)
}

func TestUserService_Create_AlreadyExists(t *testing.T) {
    svc, dao, gdb := newUnitSvc()

    dao.On("GetByMobile", mock.Anything, gdb, "13800138000").
        Return(&do.UserDO{Mobile: "13800138000"}, nil)

    err := svc.Create(context.Background(), &dto.UserDTO{
        UserDO: do.UserDO{Mobile: "13800138000", Password: "x"},
    })

    assert.Error(t, err)
    assert.True(t, errorsx.IsCode(err, code.ErrUserAlreadyExists))
    dao.AssertExpectations(t)
}

func TestUserService_Update_Success(t *testing.T) {
    svc, dao, gdb := newUnitSvc()

    dao.On("Get", mock.Anything, gdb, uint64(888)).
        Return(&do.UserDO{Mobile: "13800138000"}, nil)
    dao.On("Update", mock.Anything, gdb, mock.AnythingOfType("*do.UserDO")).
        Return(nil)

    err := svc.Update(context.Background(), &dto.UserDTO{UserDO: do.UserDO{BaseModel: dbpkg.BaseModel{ID: 888}, Mobile: "13800138000"}})
    assert.NoError(t, err)
    dao.AssertExpectations(t)
}

func TestUserService_GetByID_Success(t *testing.T) {
    svc, dao, gdb := newUnitSvc()
    dao.On("Get", mock.Anything, gdb, uint64(101)).
        Return(&do.UserDO{Mobile: "13800138000"}, nil)

    got, err := svc.GetByID(context.Background(), 101)
    assert.NoError(t, err)
    assert.Equal(t, "13800138000", got.Mobile)
    dao.AssertExpectations(t)
}

func TestUserService_GetByMobile_Success(t *testing.T) {
    svc, dao, gdb := newUnitSvc()
    dao.On("GetByMobile", mock.Anything, gdb, "13800138000").
        Return(&do.UserDO{Mobile: "13800138000"}, nil)

    got, err := svc.GetByMobile(context.Background(), "13800138000")
    assert.NoError(t, err)
    assert.Equal(t, "13800138000", got.Mobile)
    dao.AssertExpectations(t)
}

func TestUserService_List_Success(t *testing.T) {
    svc, dao, gdb := newUnitSvc()
    order := []string{"id desc"}
    opts := metav1.ListMeta{Page: 1, PageSize: 10}
    list := &do.UserDOList{TotalCount: 2, Items: []*do.UserDO{
        {Mobile: "13800138001"}, {Mobile: "13800138002"},
    }}
    dao.On("List", mock.Anything, gdb, order, opts).Return(list, nil)

    got, err := svc.List(context.Background(), order, opts)
    assert.NoError(t, err)
    assert.Equal(t, int64(2), got.TotalCount)
    assert.Len(t, got.Items, 2)
    assert.Equal(t, "13800138001", got.Items[0].Mobile)
    assert.Equal(t, "13800138002", got.Items[1].Mobile)
    dao.AssertExpectations(t)
}

