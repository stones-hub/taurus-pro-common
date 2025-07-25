package hook

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// HookFunc 定义钩子函数类型
type HookFunc func(ctx context.Context) error

// HookInfo 钩子信息
type HookInfo struct {
	Name     string
	Priority int // 优先级，数字越小优先级越高
	Hook     HookFunc
}

// HookManager 钩子管理器
type HookManager struct {
	startHooks []HookInfo
	stopHooks  []HookInfo
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewHookManager 创建新的钩子管理器
func NewHookManager() *HookManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &HookManager{
		startHooks: make([]HookInfo, 0),
		stopHooks:  make([]HookInfo, 0),
		ctx:        ctx,
		cancel:     cancel,
	}
}

// RegisterStartHook 注册启动钩子
func (hm *HookManager) RegisterStartHook(name string, hook HookFunc, priority int) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.startHooks = append(hm.startHooks, HookInfo{
		Name:     name,
		Priority: priority,
		Hook:     hook,
	})

	log.Printf("[HookManager] 注册启动钩子: %s (优先级: %d)", name, priority)
}

// RegisterStopHook 注册停止钩子
func (hm *HookManager) RegisterStopHook(name string, hook HookFunc, priority int) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.stopHooks = append(hm.stopHooks, HookInfo{
		Name:     name,
		Priority: priority,
		Hook:     hook,
	})

	log.Printf("[HookManager] 注册停止钩子: %s (优先级: %d)", name, priority)
}

// RegisterStartHookDefault 注册启动钩子（默认优先级）
func (hm *HookManager) RegisterStartHookDefault(name string, hook HookFunc) {
	hm.RegisterStartHook(name, hook, 100)
}

// RegisterStopHookDefault 注册停止钩子（默认优先级）
func (hm *HookManager) RegisterStopHookDefault(name string, hook HookFunc) {
	hm.RegisterStopHook(name, hook, 100)
}

// Start 执行所有启动钩子
func (hm *HookManager) Start(ctx context.Context) error {
	hm.mu.RLock()
	hooks := make([]HookInfo, len(hm.startHooks))
	copy(hooks, hm.startHooks)
	hm.mu.RUnlock()

	// 按优先级排序
	sortHooksByPriority(hooks)

	log.Printf("[HookManager] 开始执行 %d 个启动钩子", len(hooks))

	for _, hook := range hooks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.Printf("[HookManager] 执行启动钩子: %s", hook.Name)
			if err := hook.Hook(ctx); err != nil {
				log.Printf("[HookManager] 启动钩子执行失败: %s, 错误: %v", hook.Name, err)
				return fmt.Errorf("启动钩子 %s 执行失败: %w", hook.Name, err)
			}
			log.Printf("[HookManager] 启动钩子执行成功: %s", hook.Name)
		}
	}

	log.Printf("[HookManager] 所有启动钩子执行完成")
	return nil
}

// Stop 执行所有停止钩子
func (hm *HookManager) Stop(ctx context.Context) error {
	hm.mu.RLock()
	hooks := make([]HookInfo, len(hm.stopHooks))
	copy(hooks, hm.stopHooks)
	hm.mu.RUnlock()

	// 按优先级排序（停止时优先级高的先执行）
	sortHooksByPriority(hooks)

	log.Printf("[HookManager] 开始执行 %d 个停止钩子", len(hooks))

	var lastErr error
	for _, hook := range hooks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.Printf("[HookManager] 执行停止钩子: %s", hook.Name)
			if err := hook.Hook(ctx); err != nil {
				log.Printf("[HookManager] 停止钩子执行失败: %s, 错误: %v", hook.Name, err)
				lastErr = fmt.Errorf("停止钩子 %s 执行失败: %w", hook.Name, err)
				// 继续执行其他钩子，不中断
			} else {
				log.Printf("[HookManager] 停止钩子执行成功: %s", hook.Name)
			}
		}
	}

	log.Printf("[HookManager] 所有停止钩子执行完成")
	return lastErr
}

// WaitForShutdown 等待关闭信号并执行停止钩子
func (hm *HookManager) WaitForShutdown(timeout time.Duration) error {
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("[HookManager] 等待关闭信号...")

	// 等待信号
	select {
	case sig := <-sigChan:
		log.Printf("[HookManager] 收到信号: %v", sig)
	case <-hm.ctx.Done():
		log.Printf("[HookManager] 上下文已取消")
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 执行停止钩子
	return hm.Stop(ctx)
}

// GetHookCount 获取钩子数量
func (hm *HookManager) GetHookCount() (startCount, stopCount int) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return len(hm.startHooks), len(hm.stopHooks)
}

// ClearHooks 清空所有钩子
func (hm *HookManager) ClearHooks() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.startHooks = make([]HookInfo, 0)
	hm.stopHooks = make([]HookInfo, 0)

	log.Printf("[HookManager] 已清空所有钩子")
}

// Cancel 取消管理器上下文, 会停止所有钩子
func (hm *HookManager) Cancel() {
	hm.cancel()
}

// sortHooksByPriority 按优先级排序钩子
func sortHooksByPriority(hooks []HookInfo) {
	// 简单的冒泡排序，按优先级升序排列
	for i := 0; i < len(hooks)-1; i++ {
		for j := 0; j < len(hooks)-1-i; j++ {
			if hooks[j].Priority > hooks[j+1].Priority {
				hooks[j], hooks[j+1] = hooks[j+1], hooks[j]
			}
		}
	}
}
