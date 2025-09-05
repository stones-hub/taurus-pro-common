package recovery

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// PanicInfo panic信息结构
type PanicInfo struct {
	Component   string          // 组件名称
	Error       interface{}     // 错误信息
	Stack       string          // 堆栈信息
	Timestamp   time.Time       // 时间戳
	GoroutineID string          // goroutine ID
	Context     context.Context // 上下文
}

// ---------------------------- PanicHandler实现 --------------------------------
// PanicHandler panic处理器接口
type PanicHandler interface {
	// HandlePanic 处理panic， 如果函数开了协程，则需要自己处理协程的
	HandlePanic(info *PanicInfo) error
}

// 统一的panic处理器
type UnifiedPanicHandler struct{}

// HandlePanic 统一的panic处理逻辑
func (h *UnifiedPanicHandler) HandlePanic(info *PanicInfo) error {
	// 统一的panic处理逻辑
	// 可以根据需要添加告警、监控、重试等逻辑
	log.Printf("🔧 [UNIFIED_PANIC] Component: %s, Error: %v, Time: %s",
		info.Component, info.Error, info.Timestamp.Format("2006-01-02 15:04:05"))

	// 这里可以添加：
	// 1. 发送告警通知
	// 2. 记录到监控系统
	// 3. 触发自动重试
	// 4. 记录到专门的错误日志文件
	// 5. 发送到错误追踪系统
	// 6. 根据上下文信息进行特定处理

	return nil
}

// ---------------------------- PanicLogger实现 --------------------------------
// PanicLogger panic日志记录器接口
type PanicLogger interface {
	LogPanic(info *PanicInfo) error
}

// DefaultPanicLogger 默认panic日志记录器
type DefaultPanicLogger struct{}

// LogPanic 记录panic日志
func (l *DefaultPanicLogger) LogPanic(info *PanicInfo) error {
	timestamp := info.Timestamp.Format("2006-01-02 15:04:05")
	log.Printf("🚨 [PANIC] [%s] Component: %s, Error: %v\nStack: %s",
		timestamp, info.Component, info.Error, info.Stack)
	return nil
}

// ---------------------------- RecoveryOptions实现 --------------------------------
// RecoveryOptions 恢复器配置选项
type RecoveryOptions struct {
	EnableStackTrace bool          // 是否启用堆栈跟踪
	MaxHandlers      int           // 最大处理器数量
	HandlerTimeout   time.Duration // 处理器超时时间
}

// DefaultRecoveryOptions 默认配置
func DefaultRecoveryOptions() *RecoveryOptions {
	return &RecoveryOptions{
		EnableStackTrace: true,
		MaxHandlers:      100,
		HandlerTimeout:   5 * time.Second,
	}
}

// PanicRecovery 全局panic恢复器
type PanicRecovery struct {
	enabled  bool             // 是否启用
	handlers []PanicHandler   // 多个处理器
	logger   PanicLogger      // 日志记录器
	mu       sync.RWMutex     // 保护 enabled、handlers、logger、options 的并发访问
	options  *RecoveryOptions // 恢复器配置选项
}

// NewPanicRecovery 创建新的panic恢复器
func NewPanicRecovery(options *RecoveryOptions) *PanicRecovery {
	if options == nil {
		options = DefaultRecoveryOptions()
	}

	return &PanicRecovery{
		enabled:  true,
		handlers: make([]PanicHandler, 0, options.MaxHandlers), // 使用MaxHandlers预分配容量
		logger:   &DefaultPanicLogger{},
		options:  options,
	}
}

// Enable 启用panic恢复
func (pr *PanicRecovery) Enable() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.enabled = true
	log.Printf("✅ 全局panic恢复机制已启用")
}

// Disable 禁用panic恢复
func (pr *PanicRecovery) Disable() {
	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.enabled = false
	log.Printf("⚠️ 全局panic恢复机制已禁用")
}

// IsEnabled 检查是否启用
func (pr *PanicRecovery) IsEnabled() bool {
	pr.mu.RLock()
	defer pr.mu.RUnlock()
	return pr.enabled
}

// AddHandler 添加panic处理器
func (pr *PanicRecovery) AddHandler(handler PanicHandler) error {
	if handler == nil {
		return fmt.Errorf("处理器不能为空")
	}

	pr.mu.Lock()
	defer pr.mu.Unlock()

	if len(pr.handlers) >= pr.options.MaxHandlers {
		return fmt.Errorf("处理器数量已达上限: %d", pr.options.MaxHandlers)
	}

	pr.handlers = append(pr.handlers, handler)
	return nil
}

// RemoveHandler 移除panic处理器
func (pr *PanicRecovery) RemoveHandler(handler PanicHandler) {
	if handler == nil {
		return
	}

	pr.mu.Lock()
	defer pr.mu.Unlock()

	for i, h := range pr.handlers {
		if h == handler {
			pr.handlers = append(pr.handlers[:i], pr.handlers[i+1:]...)
			break
		}
	}
}

// SetLogger 设置日志记录器
func (pr *PanicRecovery) SetLogger(logger PanicLogger) {
	if logger == nil {
		logger = &DefaultPanicLogger{}
	}

	pr.mu.Lock()
	defer pr.mu.Unlock()
	pr.logger = logger
}

// 获取goroutine ID的辅助函数
func getGoroutineID() string {
	// 使用runtime.Stack()来获取goroutine ID
	buf := make([]byte, 64)
	n := runtime.Stack(buf, false)

	// 解析堆栈信息来提取goroutine ID
	// 格式类似: "goroutine 123 [running]:"
	stack := string(buf[:n])

	// 使用正则表达式提取goroutine ID
	re := regexp.MustCompile(`goroutine (\d+)`)
	matches := re.FindStringSubmatch(stack)
	if len(matches) >= 2 {
		return matches[1]
	}

	// 如果无法解析，返回默认值
	return "unknown"
}

// handlePanic 统一的panic处理逻辑
func (pr *PanicRecovery) handlePanic(component string, err interface{}, ctx context.Context) {

	hctx, cancel := context.WithTimeout(ctx, pr.options.HandlerTimeout)
	defer cancel()

	// 1. 构建panic信息
	info := &PanicInfo{
		Component:   component,
		Error:       err,
		Timestamp:   time.Now(),
		GoroutineID: getGoroutineID(),
		Context:     hctx,
	}

	if pr.options.EnableStackTrace {
		info.Stack = string(debug.Stack())
	}

	// 2. 记录panic日志
	if pr.logger != nil {
		if logErr := pr.logger.LogPanic(info); logErr != nil {
			log.Printf("❌ 记录panic日志失败: %v", logErr)
		}
	}

	// 3. 协程调用所有处理器
	pr.mu.RLock()
	handlers := make([]PanicHandler, len(pr.handlers))
	copy(handlers, pr.handlers)
	pr.mu.RUnlock()

	if len(handlers) == 0 {
		return
	}

	// 并发执行处理器，避免一个处理器阻塞其他处理器
	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h PanicHandler) {
			defer wg.Done()

			// 保护handler执行，防止handler中的panic影响其他handler
			defer func() {
				if r := recover(); r != nil {
					log.Printf("❌ 处理器panic: %v", r)
				}
			}()

			// 直接调用处理器, 建议考虑HandlePanic的实现，需要有超时控制
			if err := h.HandlePanic(info); err != nil {
				log.Printf("❌ 处理器执行失败: %v", err)
			}
		}(handler)
	}

	// 等待所有处理器完成
	wg.Wait()
}

// Recover 恢复panic
func (pr *PanicRecovery) Recover(component string) {
	if !pr.IsEnabled() {
		return
	}

	if r := recover(); r != nil {
		pr.handlePanic(component, r, context.Background())
	}
}

// RecoverWithContext 恢复panic并传递上下文
func (pr *PanicRecovery) RecoverWithContext(component string, ctx context.Context) {
	if !pr.IsEnabled() {
		return
	}

	if r := recover(); r != nil {
		pr.handlePanic(component, r, ctx)
	}
}

// RecoverWithCallback 恢复panic并执行回调
func (pr *PanicRecovery) RecoverWithCallback(component string, callback func()) {
	if !pr.IsEnabled() {
		return
	}

	if r := recover(); r != nil {
		pr.handlePanic(component, r, context.Background())

		// 执行回调
		if callback != nil {
			// 保护回调函数，避免回调中的panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("❌ 回调函数panic: %v", r)
					}
				}()
				callback()
			}()
		}
	}
}

// SafeGo 安全地启动goroutine
func (pr *PanicRecovery) SafeGo(component string, fn func()) {
	if fn == nil {
		log.Printf("⚠️ 尝试启动空的goroutine函数")
		return
	}

	if !pr.IsEnabled() {
		go fn()
		return
	}

	go func() {
		defer pr.Recover(component)
		fn()
	}()
}

// SafeGoWithContext 安全地启动goroutine并传递上下文
func (pr *PanicRecovery) SafeGoWithContext(component string, ctx context.Context, fn func()) {
	if fn == nil {
		log.Printf("⚠️ 尝试启动空的goroutine函数")
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if !pr.IsEnabled() {
		go fn()
		return
	}

	go func() {
		// 使用传入的上下文进行panic恢复，业务逻辑由fn内部自己管理上下文
		defer pr.RecoverWithContext(component, ctx)
		fn()
	}()
}

// SafeGoWithCallback 安全地启动goroutine并执行回调
func (pr *PanicRecovery) SafeGoWithCallback(component string, fn func(), callback func()) {
	if fn == nil {
		log.Printf("⚠️ 尝试启动空的goroutine函数")
		return
	}

	if !pr.IsEnabled() {
		go fn()
		return
	}

	go func() {
		defer pr.RecoverWithCallback(component, callback)
		fn()
	}()
}

// WrapFunction 包装函数以添加panic恢复
func (pr *PanicRecovery) WrapFunction(component string, fn func()) func() {
	if fn == nil {
		return func() {}
	}

	if !pr.IsEnabled() {
		return fn
	}

	return func() {
		defer pr.Recover(component)
		fn()
	}
}

// WrapFunctionWithContext 包装函数以添加panic恢复和上下文
func (pr *PanicRecovery) WrapFunctionWithContext(component string, ctx context.Context, fn func()) func() {
	if fn == nil {
		return func() {}
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if !pr.IsEnabled() {
		return fn
	}

	return func() {
		defer pr.RecoverWithContext(component, ctx)
		fn()
	}
}

// WrapFunctionWithCallback 包装函数以添加panic恢复和回调
func (pr *PanicRecovery) WrapFunctionWithCallback(component string, fn func(), callback func()) func() {
	if fn == nil {
		return func() {}
	}

	if !pr.IsEnabled() {
		return fn
	}

	return func() {
		defer pr.RecoverWithCallback(component, callback)
		fn()
	}
}

// WrapErrorFunction 包装返回错误的函数
func (pr *PanicRecovery) WrapErrorFunction(component string, fn func() error) func() error {
	if fn == nil {
		return func() error { return nil }
	}

	if !pr.IsEnabled() {
		return fn
	}

	return func() error {
		defer pr.Recover(component)
		return fn()
	}
}

// WrapErrorFunctionWithContext 包装返回错误的函数并传递上下文
func (pr *PanicRecovery) WrapErrorFunctionWithContext(component string, ctx context.Context, fn func() error) func() error {
	if fn == nil {
		return func() error { return nil }
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if !pr.IsEnabled() {
		return fn
	}

	return func() error {
		defer pr.RecoverWithContext(component, ctx)
		return fn()
	}
}

// 全局panic恢复器实例
var GlobalPanicRecovery = NewPanicRecovery(nil)

// 框架初始化函数
func InitFrameworkPanicRecovery() {
	// 添加统一的panic处理器
	if err := GlobalPanicRecovery.AddHandler(&UnifiedPanicHandler{}); err != nil {
		log.Printf("❌ 添加统一panic处理器失败: %v", err)
	}

	// 确保启用
	GlobalPanicRecovery.Enable()

	log.Printf("✅ 框架panic恢复机制初始化完成")
}

// 初始化全局panic处理器
func init() {
	// 默认启用框架自动panic恢复
	InitFrameworkPanicRecovery()
}
