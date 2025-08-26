package shutdown

import (
	"context"
	"emshop/pkg/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Manager 优雅关闭管理器
type Manager struct {
	mu        sync.Mutex
	callbacks []func() error
	done      chan struct{}
}

// NewManager 创建新的关闭管理器
func NewManager() *Manager {
	return &Manager{
		callbacks: make([]func() error, 0),
		done:      make(chan struct{}),
	}
}

// AddCallback 添加关闭回调函数
func (m *Manager) AddCallback(callback func() error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callbacks = append(m.callbacks, callback)
}

// WaitForShutdown 等待关闭信号并执行清理
func (m *Manager) WaitForShutdown() {
	// 监听关闭信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	sig := <-sigCh
	log.Infof("收到关闭信号: %v", sig)

	// 执行关闭流程
	m.shutdown()
}

// WaitForShutdownWithContext 带上下文的等待关闭
func (m *Manager) WaitForShutdownWithContext(ctx context.Context) {
	// 监听关闭信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Infof("收到关闭信号: %v", sig)
		m.shutdown()
	case <-ctx.Done():
		log.Info("上下文被取消，开始关闭")
		m.shutdown()
	}
}

// Shutdown 立即执行关闭流程
func (m *Manager) Shutdown() {
	m.shutdown()
}

// shutdown 执行实际的关闭流程
func (m *Manager) shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Info("开始执行优雅关闭...")

	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 执行所有回调
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := len(m.callbacks) - 1; i >= 0; i-- {
			if err := m.callbacks[i](); err != nil {
				log.Errorf("关闭回调执行失败: %v", err)
			}
		}
	}()

	// 等待完成或超时
	select {
	case <-done:
		log.Info("优雅关闭完成")
	case <-ctx.Done():
		log.Warn("关闭超时，强制退出")
	}

	close(m.done)
}

// Done 返回关闭完成信号通道
func (m *Manager) Done() <-chan struct{} {
	return m.done
}