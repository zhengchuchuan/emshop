package v1

import (
    "context"
    "testing"

    datadto "emshop/internal/app/userop/srv/domain/dto"
    datado "emshop/internal/app/userop/srv/domain/do"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "gorm.io/gorm"
)

// mockAddressStore implements interfaces.AddressStore
type mockAddressStore struct{ mock.Mock }

func (m *mockAddressStore) GetAddressList(ctx context.Context, db *gorm.DB, userID int32) ([]*datadto.AddressDTO, int64, error) {
    args := m.Called(ctx, db, userID)
    if v := args.Get(0); v != nil { return v.([]*datadto.AddressDTO), args.Get(1).(int64), args.Error(2) }
    return nil, 0, args.Error(2)
}
func (m *mockAddressStore) CreateAddress(ctx context.Context, db *gorm.DB, address *datado.Address) (*datado.Address, error) {
    args := m.Called(ctx, db, address)
    if v := args.Get(0); v != nil { return v.(*datado.Address), args.Error(1) }
    return nil, args.Error(1)
}
func (m *mockAddressStore) UpdateAddress(ctx context.Context, db *gorm.DB, address *datado.Address) error {
    return m.Called(ctx, db, address).Error(0)
}
func (m *mockAddressStore) DeleteAddress(ctx context.Context, db *gorm.DB, addressID int32, userID int32) error {
    return m.Called(ctx, db, addressID, userID).Error(0)
}
func (m *mockAddressStore) GetAddressByID(ctx context.Context, db *gorm.DB, addressID int32, userID int32) (*datado.Address, error) {
    args := m.Called(ctx, db, addressID, userID)
    if v := args.Get(0); v != nil { return v.(*datado.Address), args.Error(1) }
    return nil, args.Error(1)
}

func TestAddressService_GetAddressList(t *testing.T) {
    dao := &mockAddressStore{}
    gdb := &gorm.DB{}
    svc := &addressService{addressDAO: dao, db: gdb}

    _ = datado.Address{} // silence import if not used by other methods
    expected := []*datadto.AddressDTO{{Address: "路1"}, {Address: "路2"}}
    dao.On("GetAddressList", mock.Anything, gdb, int32(1001)).Return(expected, int64(2), nil)

    list, total, err := svc.GetAddressList(context.Background(), 1001)
    assert.NoError(t, err)
    assert.Equal(t, int64(2), total)
    assert.Len(t, list, 2)
    dao.AssertExpectations(t)
}
