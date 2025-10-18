package service

import (
    "context"
    "testing"

    "emshop/internal/app/order/srv/domain/do"
    metav1 "emshop/pkg/common/meta/v1"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gorm.io/gorm"
)

// mockOrderStore implements interfaces.OrderStore
type mockOrderStore struct{ mock.Mock }

func (m *mockOrderStore) Get(ctx context.Context, db *gorm.DB, orderSn string) (*do.OrderInfoDO, error) {
    args := m.Called(ctx, db, orderSn)
    if v := args.Get(0); v != nil { return v.(*do.OrderInfoDO), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockOrderStore) List(ctx context.Context, db *gorm.DB, userID uint64, meta metav1.ListMeta, orderby []string) (*do.OrderInfoDOList, error) {
    args := m.Called(ctx, db, userID, meta, orderby)
    if v := args.Get(0); v != nil { return v.(*do.OrderInfoDOList), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockOrderStore) Create(ctx context.Context, db *gorm.DB, order *do.OrderInfoDO) error {
    return m.Called(ctx, db, order).Error(0)
}
func (m *mockOrderStore) Update(ctx context.Context, db *gorm.DB, order *do.OrderInfoDO) error {
    return m.Called(ctx, db, order).Error(0)
}

// mockShopCartStore implements interfaces.ShopCartStore
type mockShopCartStore struct{ mock.Mock }

func (m *mockShopCartStore) List(ctx context.Context, db *gorm.DB, userID uint64, checked bool, meta metav1.ListMeta, orderby []string) (*do.ShoppingCartDOList, error) {
    args := m.Called(ctx, db, userID, checked, meta, orderby)
    if v := args.Get(0); v != nil { return v.(*do.ShoppingCartDOList), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockShopCartStore) Create(ctx context.Context, db *gorm.DB, cartItem *do.ShoppingCartDO) error {
    return m.Called(ctx, db, cartItem).Error(0)
}
func (m *mockShopCartStore) Get(ctx context.Context, db *gorm.DB, userID, goodsID uint64) (*do.ShoppingCartDO, error) {
    args := m.Called(ctx, db, userID, goodsID)
    if v := args.Get(0); v != nil { return v.(*do.ShoppingCartDO), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockShopCartStore) UpdateNum(ctx context.Context, db *gorm.DB, cartItem *do.ShoppingCartDO) error {
    return m.Called(ctx, db, cartItem).Error(0)
}
func (m *mockShopCartStore) Delete(ctx context.Context, db *gorm.DB, ID uint64) error {
    return m.Called(ctx, db, ID).Error(0)
}
func (m *mockShopCartStore) ClearCheck(ctx context.Context, db *gorm.DB, userID uint64) error {
    return m.Called(ctx, db, userID).Error(0)
}
func (m *mockShopCartStore) DeleteByGoodsIDs(ctx context.Context, db *gorm.DB, userID uint64, goodsIDs []int32) error {
    return m.Called(ctx, db, userID, goodsIDs).Error(0)
}

func newOrderSvcForUnit() (*orderService, *mockOrderStore, *mockShopCartStore, *gorm.DB) {
    ordersDAO := &mockOrderStore{}
    cartsDAO := &mockShopCartStore{}
    gdb := &gorm.DB{}
    svc := &orderService{ordersDAO: ordersDAO, shoppingCartsDAO: cartsDAO, db: gdb}
    return svc, ordersDAO, cartsDAO, gdb
}

func TestOrderService_Get_ReturnsDTO(t *testing.T) {
    svc, ordersDAO, _, gdb := newOrderSvcForUnit()
    ordersDAO.On("Get", mock.Anything, gdb, "OSN-1").Return(&do.OrderInfoDO{OrderSn: "OSN-1"}, nil)

    got, err := svc.Get(context.Background(), "OSN-1")
    assert.NoError(t, err)
    assert.Equal(t, "OSN-1", got.OrderSn)
    ordersDAO.AssertExpectations(t)
}

func TestOrderService_CartItemList_ReturnsList(t *testing.T) {
    svc, _, cartsDAO, gdb := newOrderSvcForUnit()
    items := &do.ShoppingCartDOList{TotalCount: 2, Items: []*do.ShoppingCartDO{{User: 1, Goods: 2}, {User: 1, Goods: 3}}}
    meta := metav1.ListMeta{Page: 1, PageSize: 10}
    cartsDAO.On("List", mock.Anything, gdb, uint64(1001), false, meta, []string{}).Return(items, nil)

    got, err := svc.CartItemList(context.Background(), 1001, meta)
    assert.NoError(t, err)
    assert.Equal(t, int64(2), got.TotalCount)
    assert.Len(t, got.Items, 2)
    cartsDAO.AssertExpectations(t)
}
