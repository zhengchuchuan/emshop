package v1

import (
	"github.com/google/wire"
	"emshop/internal/app/user/srv/data/v1/interfaces"
)

// ProviderSet 数据层提供器集合
var ProviderSet = wire.NewSet(
	NewFactoryManager,
	ProvideUserStore,
)

// ProvideUserStore 提供用户存储接口
func ProvideUserStore(fm *FactoryManager) interfaces.UserStore {
	return fm.GetDataFactory().Users()
}