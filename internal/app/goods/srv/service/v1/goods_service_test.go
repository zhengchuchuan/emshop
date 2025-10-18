package v1

import (
    "context"
    "testing"

    proto "emshop/api/goods/v1"
    "emshop/internal/app/goods/srv/domain/do"
    interfaces "emshop/internal/app/goods/srv/data/v1/interfaces"
    metav1 "emshop/pkg/common/meta/v1"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gorm.io/gorm"
)

// mockGoodsStore implements interfaces.GoodsStore
type mockGoodsStore struct{ mock.Mock }

func (m *mockGoodsStore) Get(ctx context.Context, db *gorm.DB, ID uint64) (*do.GoodsDO, error) {
    args := m.Called(ctx, db, ID)
    if v := args.Get(0); v != nil { return v.(*do.GoodsDO), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockGoodsStore) ListByIDs(ctx context.Context, db *gorm.DB, ids []uint64, orderby []string) (*do.GoodsDOList, error) {
    args := m.Called(ctx, db, ids, orderby)
    if v := args.Get(0); v != nil { return v.(*do.GoodsDOList), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockGoodsStore) List(ctx context.Context, db *gorm.DB, orderby []string, opts metav1.ListMeta) (*do.GoodsDOList, error) {
    args := m.Called(ctx, db, orderby, opts)
    if v := args.Get(0); v != nil { return v.(*do.GoodsDOList), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockGoodsStore) GetAllGoodsIDs(ctx context.Context, db *gorm.DB) ([]uint64, error) {
    args := m.Called(ctx, db)
    if v := args.Get(0); v != nil { return v.([]uint64), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockGoodsStore) Create(ctx context.Context, db *gorm.DB, goods *do.GoodsDO) error {
    return m.Called(ctx, db, goods).Error(0)
}
func (m *mockGoodsStore) Update(ctx context.Context, db *gorm.DB, goods *do.GoodsDO) error {
    return m.Called(ctx, db, goods).Error(0)
}
func (m *mockGoodsStore) Delete(ctx context.Context, db *gorm.DB, ID uint64) error {
    return m.Called(ctx, db, ID).Error(0)
}

// minimal mocks to satisfy struct fields (not used in these tests)
type noopCategoryStore struct{ interfaces.CategoryStore }
type noopBrandStore struct{ interfaces.BrandsStore }
type noopBannerStore struct{ interfaces.BannerStore }

func TestGoodsService_List_NoSearchConditions(t *testing.T) {
    dao := &mockGoodsStore{}
    gdb := &gorm.DB{}
    svc := &goodsService{goodsDAO: dao, db: gdb}

    opts := metav1.ListMeta{Page: 1, PageSize: 10}
    order := []string{"id desc"}
    goods := &do.GoodsDOList{TotalCount: 2, Items: []*do.GoodsDO{{Name: "A"}, {Name: "B"}}}
    dao.On("List", mock.Anything, gdb, order, opts).Return(goods, nil)

    got, err := svc.List(context.Background(), opts, &proto.GoodsFilterRequest{}, order)
    assert.NoError(t, err)
    assert.Equal(t, int64(2), got.TotalCount)
    assert.Len(t, got.Items, 2)
    dao.AssertExpectations(t)
}

func TestGoodsService_Get_ReturnsDTO(t *testing.T) {
    dao := &mockGoodsStore{}
    gdb := &gorm.DB{}
    svc := &goodsService{goodsDAO: dao, db: gdb}

    dao.On("Get", mock.Anything, gdb, uint64(101)).Return(&do.GoodsDO{Name: "Phone"}, nil)

    got, err := svc.Get(context.Background(), 101)
    assert.NoError(t, err)
    assert.Equal(t, "Phone", got.Name)
    dao.AssertExpectations(t)
}
