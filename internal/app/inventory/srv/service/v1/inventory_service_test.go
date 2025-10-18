package v1

import (
    "context"
    "testing"

    "emshop/internal/app/inventory/srv/domain/do"
    "emshop/internal/app/inventory/srv/domain/dto"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gorm.io/gorm"
)

// mockInventoryStore implements interfaces.InventoryStore
type mockInventoryStore struct{ mock.Mock }

func (m *mockInventoryStore) Create(ctx context.Context, db *gorm.DB, inv *do.InventoryDO) error {
    return m.Called(ctx, db, inv).Error(0)
}
func (m *mockInventoryStore) Get(ctx context.Context, db *gorm.DB, goodsID uint64) (*do.InventoryDO, error) {
    args := m.Called(ctx, db, goodsID)
    if v := args.Get(0); v != nil { return v.(*do.InventoryDO), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockInventoryStore) GetSellDetail(ctx context.Context, db *gorm.DB, ordersn string) (*do.StockSellDetailDO, error) {
    args := m.Called(ctx, db, ordersn)
    if v := args.Get(0); v != nil { return v.(*do.StockSellDetailDO), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockInventoryStore) Reduce(ctx context.Context, db *gorm.DB, goodsID uint64, num int) error {
    return m.Called(ctx, db, goodsID, num).Error(0)
}
func (m *mockInventoryStore) Increase(ctx context.Context, db *gorm.DB, goodsID uint64, num int) error {
    return m.Called(ctx, db, goodsID, num).Error(0)
}
func (m *mockInventoryStore) CreateStockSellDetail(ctx context.Context, db *gorm.DB, detail *do.StockSellDetailDO) error {
    return m.Called(ctx, db, detail).Error(0)
}
func (m *mockInventoryStore) UpdateStockSellDetailStatus(ctx context.Context, db *gorm.DB, ordersn string, status int32) error {
    return m.Called(ctx, db, ordersn, status).Error(0)
}

func TestInventory_CreateAndGet(t *testing.T) {
    dao := &mockInventoryStore{}
    gdb := &gorm.DB{}
    svc := &inventoryService{inventoryDAO: dao, db: gdb}

    // Create
    inv := &dto.InventoryDTO{InventoryDO: do.InventoryDO{Goods: 1001, Stocks: 5}}
    dao.On("Create", mock.Anything, gdb, &inv.InventoryDO).Return(nil)
    err := svc.Create(context.Background(), inv)
    assert.NoError(t, err)

    // Get
    dao.On("Get", mock.Anything, gdb, uint64(1001)).Return(&do.InventoryDO{Goods: 1001, Stocks: 5}, nil)
    got, err := svc.Get(context.Background(), 1001)
    assert.NoError(t, err)
    assert.Equal(t, int32(1001), got.Goods)
    assert.Equal(t, int32(5), got.Stocks)

    dao.AssertExpectations(t)
}
