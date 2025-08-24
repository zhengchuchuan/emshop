package consumer

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	
	"emshop/internal/app/goods/srv/data/v1/sync"
)

// MockSyncManager 模拟同步管理器
type MockSyncManager struct {
	mock.Mock
}

func (m *MockSyncManager) SyncToSearch(ctx context.Context, entityType string, entityID uint64) error {
	args := m.Called(ctx, entityType, entityID)
	return args.Error(0)
}

func (m *MockSyncManager) SyncToCache(ctx context.Context, entityType string, entityID uint64) error {
	args := m.Called(ctx, entityType, entityID)
	return args.Error(0)
}

func (m *MockSyncManager) RemoveFromSearch(ctx context.Context, entityType string, entityID uint64) error {
	args := m.Called(ctx, entityType, entityID)
	return args.Error(0)
}

func (m *MockSyncManager) SyncAllGoodsToSearch(ctx context.Context, forceSync bool, goodsIds []uint64) (*sync.SyncResult, error) {
	args := m.Called(ctx, forceSync, goodsIds)
	return args.Get(0).(*sync.SyncResult), args.Error(1)
}

func TestCanalConsumer_processMessage(t *testing.T) {
	// 创建模拟同步管理器
	mockSyncManager := new(MockSyncManager)
	
	// 创建消费者配置
	config := &CanalConsumerConfig{
		NameServers:   []string{"localhost:9876"},
		ConsumerGroup: "test-group",
		Topic:         "test-topic",
		MaxReconsume:  3,
	}
	
	// 创建消费者
	consumer := NewCanalConsumer(config, mockSyncManager)
	
	tests := []struct {
		name        string
		message     *CanalMessage
		setupMocks  func()
		wantErr     bool
		description string
	}{
		{
			name: "处理商品插入消息",
			message: &CanalMessage{
				Database: "emshop",
				Table:    "goods",
				Type:     "INSERT",
				Data: []map[string]interface{}{
					{"id": float64(123), "name": "测试商品"},
				},
				IsDdl: false,
			},
			setupMocks: func() {
				mockSyncManager.On("SyncToSearch", mock.Anything, "goods", uint64(123)).Return(nil)
			},
			wantErr:     false,
			description: "商品插入操作应该调用SyncToSearch",
		},
		{
			name: "处理商品更新消息",
			message: &CanalMessage{
				Database: "emshop",
				Table:    "goods",
				Type:     "UPDATE",
				Data: []map[string]interface{}{
					{"id": float64(456), "name": "更新商品"},
				},
				IsDdl: false,
			},
			setupMocks: func() {
				mockSyncManager.On("SyncToSearch", mock.Anything, "goods", uint64(456)).Return(nil)
			},
			wantErr:     false,
			description: "商品更新操作应该调用SyncToSearch",
		},
		{
			name: "处理商品删除消息",
			message: &CanalMessage{
				Database: "emshop",
				Table:    "goods",
				Type:     "DELETE",
				Data: []map[string]interface{}{
					{"id": float64(789)},
				},
				IsDdl: false,
			},
			setupMocks: func() {
				mockSyncManager.On("RemoveFromSearch", mock.Anything, "goods", uint64(789)).Return(nil)
			},
			wantErr:     false,
			description: "商品删除操作应该调用RemoveFromSearch",
		},
		{
			name: "忽略非emshop数据库消息",
			message: &CanalMessage{
				Database: "other_db",
				Table:    "goods",
				Type:     "INSERT",
				Data:     []map[string]interface{}{{"id": float64(123)}},
				IsDdl:    false,
			},
			setupMocks: func() {
				// 不应该调用任何同步方法
			},
			wantErr:     false,
			description: "非目标数据库的消息应该被忽略",
		},
		{
			name: "忽略DDL操作",
			message: &CanalMessage{
				Database: "emshop",
				Table:    "goods",
				Type:     "CREATE",
				Data:     []map[string]interface{}{},
				IsDdl:    true,
			},
			setupMocks: func() {
				// 不应该调用任何同步方法
			},
			wantErr:     false,
			description: "DDL操作应该被忽略",
		},
		{
			name: "忽略不支持的表",
			message: &CanalMessage{
				Database: "emshop",
				Table:    "unsupported_table",
				Type:     "INSERT",
				Data:     []map[string]interface{}{{"id": float64(123)}},
				IsDdl:    false,
			},
			setupMocks: func() {
				// 不应该调用任何同步方法
			},
			wantErr:     false,
			description: "不支持的表消息应该被忽略",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置模拟对象
			mockSyncManager.ExpectedCalls = nil
			mockSyncManager.Calls = nil
			
			// 设置模拟预期
			tt.setupMocks()
			
			// 执行测试
			ctx := context.Background()
			err := consumer.processMessage(ctx, tt.message)
			
			// 验证结果
			if tt.wantErr {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
			
			// 验证模拟调用
			mockSyncManager.AssertExpectations(t)
		})
	}
}

func TestCanalConsumer_handleGoodsChange(t *testing.T) {
	mockSyncManager := new(MockSyncManager)
	
	config := &CanalConsumerConfig{
		NameServers:   []string{"localhost:9876"},
		ConsumerGroup: "test-group",
		Topic:         "test-topic",
		MaxReconsume:  3,
	}
	
	consumer := NewCanalConsumer(config, mockSyncManager)

	tests := []struct {
		name        string
		message     *CanalMessage
		setupMocks  func()
		wantErr     bool
		description string
	}{
		{
			name: "处理多条商品插入",
			message: &CanalMessage{
				Type: "INSERT",
				Data: []map[string]interface{}{
					{"id": float64(100), "name": "商品1"},
					{"id": float64(101), "name": "商品2"},
				},
			},
			setupMocks: func() {
				mockSyncManager.On("SyncToSearch", mock.Anything, "goods", uint64(100)).Return(nil)
				mockSyncManager.On("SyncToSearch", mock.Anything, "goods", uint64(101)).Return(nil)
			},
			wantErr:     false,
			description: "应该处理多条商品插入",
		},
		{
			name: "处理ID类型为int64的情况",
			message: &CanalMessage{
				Type: "INSERT",
				Data: []map[string]interface{}{
					{"id": int64(200), "name": "商品"},
				},
			},
			setupMocks: func() {
				mockSyncManager.On("SyncToSearch", mock.Anything, "goods", uint64(200)).Return(nil)
			},
			wantErr:     false,
			description: "应该正确处理int64类型的ID",
		},
		{
			name: "处理ID类型为string的情况",
			message: &CanalMessage{
				Type: "INSERT",
				Data: []map[string]interface{}{
					{"id": "300", "name": "商品"},
				},
			},
			setupMocks: func() {
				mockSyncManager.On("SyncToSearch", mock.Anything, "goods", uint64(300)).Return(nil)
			},
			wantErr:     false,
			description: "应该正确解析字符串类型的ID",
		},
		{
			name: "处理不支持的操作类型",
			message: &CanalMessage{
				Type: "UNKNOWN",
				Data: []map[string]interface{}{
					{"id": float64(400)},
				},
			},
			setupMocks: func() {
				// 不应该调用任何方法
			},
			wantErr:     false,
			description: "不支持的操作类型应该被忽略",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置模拟对象
			mockSyncManager.ExpectedCalls = nil
			mockSyncManager.Calls = nil
			
			// 设置模拟预期
			tt.setupMocks()
			
			// 执行测试
			ctx := context.Background()
			err := consumer.handleGoodsChange(ctx, tt.message)
			
			// 验证结果
			if tt.wantErr {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
			
			// 验证模拟调用
			mockSyncManager.AssertExpectations(t)
		})
	}
}

func TestCanalMessage_JSON_Unmarshal(t *testing.T) {
	// 测试Canal消息的JSON反序列化
	jsonData := `{
		"database": "emshop",
		"table": "goods",
		"type": "INSERT",
		"data": [
			{
				"id": 123,
				"name": "测试商品",
				"price": 99.99
			}
		],
		"old": [],
		"es": 1234567890,
		"ts": 1234567890123,
		"isDdl": false,
		"executeTime": 1234567890000
	}`
	
	var msg CanalMessage
	err := json.Unmarshal([]byte(jsonData), &msg)
	
	assert.NoError(t, err, "JSON反序列化应该成功")
	assert.Equal(t, "emshop", msg.Database, "数据库名应该正确")
	assert.Equal(t, "goods", msg.Table, "表名应该正确")
	assert.Equal(t, "INSERT", msg.Type, "操作类型应该正确")
	assert.Len(t, msg.Data, 1, "数据数组长度应该正确")
	assert.Equal(t, float64(123), msg.Data[0]["id"], "ID应该正确")
	assert.False(t, msg.IsDdl, "IsDdl应该为false")
}

func TestParseUint64(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		wantErr  bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"999999", 999999, false},
		{"abc", 0, true},
		{"12a3", 0, true},
		{"", 0, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseUint64(tt.input)
			
			if tt.wantErr {
				assert.Error(t, err, "应该返回错误")
			} else {
				assert.NoError(t, err, "不应该返回错误")
				assert.Equal(t, tt.expected, result, "结果应该匹配")
			}
		})
	}
}