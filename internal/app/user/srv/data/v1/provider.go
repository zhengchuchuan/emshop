package v1

import (
	"github.com/google/wire"
)

// ProviderSet 数据层提供器集合
var ProviderSet = wire.NewSet(
	NewFactoryManager,
)