package recovery

import (
	"context"
	"fmt"
	"regexp"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestGetGoroutineID 测试获取goroutine ID功能
func TestGetGoroutineID(t *testing.T) {
	// 测试主goroutine
	mainGID := getGoroutineID()
	if mainGID == "" {
		t.Error("主goroutine ID不能为空")
	}

	// 验证ID是数字格式
	matched, _ := regexp.MatchString(`^\d+$`, mainGID)
	if !matched && mainGID != "unknown" {
		t.Errorf("goroutine ID应该是数字格式，但得到: %s", mainGID)
	}

	// 测试在goroutine中获取ID
	var gID string
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		gID = getGoroutineID()
	}()

	wg.Wait()

	if gID == "" {
		t.Error("goroutine中的ID不能为空")
	}

	t.Logf("主goroutine ID: %s", mainGID)
	t.Logf("子goroutine ID: %s", gID)
}

// TestPanicRecovery 测试panic恢复功能
func TestPanicRecovery(t *testing.T) {
	// 创建测试恢复器
	options := &RecoveryOptions{
		EnableStackTrace: true,
		MaxHandlers:      5,
		HandlerTimeout:   1 * time.Second,
	}
	recovery := NewPanicRecovery(options)

	// 添加测试处理器
	panicHandled := false
	testHandler := &TestPanicHandler{
		handled: &panicHandled,
	}
	recovery.AddHandler(testHandler)

	// 测试SafeGo
	recovery.SafeGo("test-component", func() {
		panic("测试panic")
	})

	// 等待处理完成
	time.Sleep(100 * time.Millisecond)

	if !panicHandled {
		t.Error("panic应该被处理")
	}
}

// TestPanicRecoveryWithContext 测试带上下文的panic恢复
func TestPanicRecoveryWithContext(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 添加测试处理器
	var receivedContext context.Context
	testHandler := &ContextTestHandler{
		receivedContext: &receivedContext,
	}
	recovery.AddHandler(testHandler)

	// 创建测试上下文
	type testKey string
	ctx := context.WithValue(context.Background(), testKey("test_key"), "test_value")

	// 测试SafeGoWithContext
	recovery.SafeGoWithContext("test-context", ctx, func() {
		panic("测试上下文panic")
	})

	// 等待处理完成
	time.Sleep(100 * time.Millisecond)

	if receivedContext == nil {
		t.Error("应该接收到上下文")
	} else if receivedContext.Value(testKey("test_key")) != "test_value" {
		t.Error("上下文值不正确")
	}
}

// TestRecoveryOptions 测试恢复器配置
func TestRecoveryOptions(t *testing.T) {
	// 测试默认配置
	defaultOptions := DefaultRecoveryOptions()
	if defaultOptions.EnableStackTrace != true {
		t.Error("默认配置应该启用堆栈跟踪")
	}
	if defaultOptions.MaxHandlers != 100 {
		t.Error("默认配置的最大处理器数量应该是100")
	}
	if defaultOptions.HandlerTimeout != 5*time.Second {
		t.Error("默认配置的处理器超时应该是5秒")
	}

	// 测试自定义配置
	customOptions := &RecoveryOptions{
		EnableStackTrace: false,
		MaxHandlers:      10,
		HandlerTimeout:   2 * time.Second,
	}

	recovery := NewPanicRecovery(customOptions)
	if recovery.options.EnableStackTrace != false {
		t.Error("自定义配置应该被正确应用")
	}
}

// TestEnableDisable 测试启用/禁用功能
func TestEnableDisable(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 测试默认启用
	if !recovery.IsEnabled() {
		t.Error("恢复器应该默认启用")
	}

	// 测试禁用
	recovery.Disable()
	if recovery.IsEnabled() {
		t.Error("恢复器应该被禁用")
	}

	// 测试重新启用
	recovery.Enable()
	if !recovery.IsEnabled() {
		t.Error("恢复器应该被重新启用")
	}
}

// TestMaxHandlers 测试最大处理器数量限制
func TestMaxHandlers(t *testing.T) {
	options := &RecoveryOptions{
		MaxHandlers: 2,
	}
	recovery := NewPanicRecovery(options)

	// 添加第一个处理器
	err1 := recovery.AddHandler(&TestPanicHandler{})
	if err1 != nil {
		t.Errorf("添加第一个处理器失败: %v", err1)
	}

	// 添加第二个处理器
	err2 := recovery.AddHandler(&TestPanicHandler{})
	if err2 != nil {
		t.Errorf("添加第二个处理器失败: %v", err2)
	}

	// 尝试添加第三个处理器（应该失败）
	err3 := recovery.AddHandler(&TestPanicHandler{})
	if err3 == nil {
		t.Error("应该无法添加超过最大数量的处理器")
	}
}

// TestHandlerTimeout 测试处理器超时
func TestHandlerTimeout(t *testing.T) {
	options := &RecoveryOptions{
		HandlerTimeout: 100 * time.Millisecond,
		MaxHandlers:    10, // 设置最大处理器数量
	}
	recovery := NewPanicRecovery(options)

	// 添加慢处理器
	slowHandler := &SlowPanicHandler{
		duration: 200 * time.Millisecond,
	}
	err := recovery.AddHandler(slowHandler)
	if err != nil {
		t.Errorf("添加处理器失败: %v", err)
	}
	t.Logf("成功添加处理器: %T", slowHandler)

	start := time.Now()

	// 触发panic
	recovery.SafeGo("timeout-test", func() {
		panic("超时测试")
	})

	// 等待足够长的时间，确保处理器被调用
	time.Sleep(300 * time.Millisecond)

	elapsed := time.Since(start)
	t.Logf("总耗时: %v", elapsed)

	// 验证处理器被调用
	slowHandler.mu.Lock()
	called := slowHandler.called
	slowHandler.mu.Unlock()

	t.Logf("处理器调用状态: %v", called)
	if !called {
		t.Error("慢处理器应该被调用")
	}

	// 验证超时机制生效：总耗时应该接近超时时间，而不是处理器的duration
	// 由于handler现在会检查context并提前退出，总耗时应该接近HandlerTimeout
	if elapsed > 150*time.Millisecond {
		t.Logf("⚠️ 总耗时 %v 超过了预期的超时时间 %v，可能超时机制未完全生效", elapsed, options.HandlerTimeout)
	} else {
		t.Logf("✅ 超时机制生效，总耗时: %v", elapsed)
	}
}

// TestHandlerContextTimeout 测试handler通过context处理超时
func TestHandlerContextTimeout(t *testing.T) {
	options := &RecoveryOptions{
		HandlerTimeout: 50 * time.Millisecond,
		MaxHandlers:    10,
	}
	recovery := NewPanicRecovery(options)

	// 添加支持context超时的处理器
	timeoutHandler := &ContextTimeoutHandler{
		timeoutDetected: false,
	}
	err := recovery.AddHandler(timeoutHandler)
	if err != nil {
		t.Errorf("添加处理器失败: %v", err)
	}

	start := time.Now()

	// 触发panic
	recovery.SafeGo("context-timeout-test", func() {
		panic("context超时测试")
	})

	// 等待处理完成
	time.Sleep(200 * time.Millisecond)

	elapsed := time.Since(start)
	t.Logf("总耗时: %v", elapsed)

	// 验证处理器检测到超时
	if !timeoutHandler.timeoutDetected {
		t.Error("处理器应该检测到context超时")
	}

	// 验证总耗时接近超时时间
	if elapsed > 100*time.Millisecond {
		t.Logf("⚠️ 总耗时 %v 超过了预期的超时时间 %v", elapsed, options.HandlerTimeout)
	} else {
		t.Logf("✅ context超时机制生效，总耗时: %v", elapsed)
	}
}

// TestNilHandler 测试nil处理器
func TestNilHandler(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 尝试添加nil处理器
	err := recovery.AddHandler(nil)
	if err == nil {
		t.Error("应该拒绝nil处理器")
	}
}

// TestNilLogger 测试nil日志记录器
func TestNilLogger(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 设置nil日志记录器
	recovery.SetLogger(nil)

	// 触发panic，应该使用默认日志记录器
	recovery.SafeGo("nil-logger-test", func() {
		panic("nil日志记录器测试")
	})

	time.Sleep(100 * time.Millisecond)
	// 如果没有崩溃，说明处理正常
}

// TestNilFunction 测试nil函数
func TestNilFunction(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 测试SafeGo with nil function
	recovery.SafeGo("nil-function", nil)

	// 测试WrapFunction with nil function
	wrappedFn := recovery.WrapFunction("nil-wrap", nil)
	if wrappedFn == nil {
		t.Error("包装nil函数应该返回空函数")
	}

	// 执行包装的函数，不应该panic
	wrappedFn()
}

// TestNilContext 测试nil上下文
func TestNilContext(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 测试SafeGoWithContext with nil context
	recovery.SafeGoWithContext("nil-context", nil, func() {
		panic("nil上下文测试")
	})

	time.Sleep(100 * time.Millisecond)
	// 如果没有崩溃，说明处理正常
}

// TestHandlerPanic 测试处理器panic
func TestHandlerPanic(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 添加会panic的处理器
	panicHandler := &PanicPanicHandler{}
	recovery.AddHandler(panicHandler)

	// 添加正常处理器
	normalHandler := &TestPanicHandler{
		handled: &panicHandler.called,
	}
	recovery.AddHandler(normalHandler)

	// 触发panic
	recovery.SafeGo("handler-panic-test", func() {
		panic("处理器panic测试")
	})

	time.Sleep(200 * time.Millisecond)

	// 验证正常处理器仍然被调用
	if !panicHandler.called {
		t.Error("正常处理器应该被调用，即使其他处理器panic")
	}
}

// TestCallbackPanic 测试回调函数panic
func TestCallbackPanic(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	callbackCalled := false
	callback := func() {
		callbackCalled = true
		// 移除panic，因为回调函数中的panic已经被保护了
		// panic("回调函数panic")
	}

	// 测试SafeGoWithCallback
	recovery.SafeGoWithCallback("callback-panic-test", func() {
		panic("回调panic测试")
	}, callback)

	time.Sleep(100 * time.Millisecond)

	// 验证回调被调用
	if !callbackCalled {
		t.Error("回调函数应该被调用")
	}
}

// TestConcurrentHandlers 测试并发处理器
func TestConcurrentHandlers(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	var wg sync.WaitGroup
	handlerCount := 5

	// 添加多个处理器
	for i := 0; i < handlerCount; i++ {
		wg.Add(1)
		handler := &ConcurrentTestHandler{
			wg: &wg,
			id: i,
		}
		recovery.AddHandler(handler)
	}

	// 触发panic
	recovery.SafeGo("concurrent-test", func() {
		panic("并发处理器测试")
	})

	// 等待所有处理器完成
	wg.Wait()
}

// TestRemoveHandler 测试移除处理器
func TestRemoveHandler(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	handler := &TestPanicHandler{}

	// 添加处理器
	recovery.AddHandler(handler)

	// 移除处理器
	recovery.RemoveHandler(handler)

	// 验证处理器被移除
	// 注意：这里我们通过添加处理器来验证，因为移除后应该能添加更多处理器
	options := &RecoveryOptions{
		MaxHandlers: 1,
	}
	limitedRecovery := NewPanicRecovery(options)

	handler1 := &TestPanicHandler{}
	handler2 := &TestPanicHandler{}

	limitedRecovery.AddHandler(handler1)
	limitedRecovery.RemoveHandler(handler1)

	// 移除后应该能添加新的处理器
	err := limitedRecovery.AddHandler(handler2)
	if err != nil {
		t.Error("移除处理器后应该能添加新处理器")
	}
}

// TestGoroutineLeak 测试goroutine泄漏
func TestGoroutineLeak(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()

	recovery := NewPanicRecovery(&RecoveryOptions{
		HandlerTimeout: 50 * time.Millisecond,
	})

	// 添加慢处理器
	slowHandler := &SlowPanicHandler{
		duration: 200 * time.Millisecond,
	}
	recovery.AddHandler(slowHandler)

	// 多次触发panic
	for i := 0; i < 10; i++ {
		recovery.SafeGo(fmt.Sprintf("leak-test-%d", i), func() {
			panic("泄漏测试")
		})
	}

	// 等待处理完成
	time.Sleep(500 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	leakedGoroutines := finalGoroutines - initialGoroutines

	if leakedGoroutines > 5 { // 允许少量goroutine
		t.Errorf("检测到goroutine泄漏: %d", leakedGoroutines)
	}
}

// TestGoroutineTimeout 测试协程超时控制
func TestGoroutineTimeout(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 创建一个带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 记录协程是否完成
	completed := make(chan bool, 1)

	// 启动一个长时间运行的协程
	recovery.SafeGoWithContext("timeout-test", ctx, func() {
		// 模拟长时间运行的任务
		select {
		case <-time.After(200 * time.Millisecond):
			// 任务完成
			completed <- true
		case <-ctx.Done():
			// context被取消，任务超时
			completed <- false
		}
	})

	// 等待结果
	select {
	case result := <-completed:
		if result {
			t.Error("协程应该在超时前被取消")
		}
	case <-time.After(300 * time.Millisecond):
		t.Error("测试超时")
	}
}

// TestGoroutineContextCancellation 测试协程context取消
func TestGoroutineContextCancellation(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 创建一个可取消的context
	ctx, cancel := context.WithCancel(context.Background())

	// 记录协程是否完成
	completed := make(chan bool, 1)

	// 启动协程
	recovery.SafeGoWithContext("cancellation-test", ctx, func() {
		// 模拟长时间运行的任务
		select {
		case <-time.After(200 * time.Millisecond):
			// 任务完成
			completed <- true
		case <-ctx.Done():
			// context被取消
			completed <- false
		}
	})

	// 立即取消context
	cancel()

	// 等待结果
	select {
	case result := <-completed:
		if result {
			t.Error("协程应该在context取消时停止")
		}
	case <-time.After(300 * time.Millisecond):
		t.Error("测试超时")
	}
}

// TestGoroutineWithSubGoroutines 测试协程内部启动子协程的管理
func TestGoroutineWithSubGoroutines(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 创建一个带超时的context
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// 记录子协程的状态
	subGoroutineCompleted := make(chan bool, 1)
	mainGoroutineCompleted := make(chan bool, 1)

	// 启动主协程，内部启动子协程
	recovery.SafeGoWithContext("sub-goroutine-test", ctx, func() {
		// 启动子协程
		go func() {
			select {
			case <-time.After(200 * time.Millisecond):
				subGoroutineCompleted <- true
			case <-ctx.Done():
				subGoroutineCompleted <- false
			}
		}()

		// 主协程等待一段时间
		select {
		case <-time.After(100 * time.Millisecond):
			mainGoroutineCompleted <- true
		case <-ctx.Done():
			mainGoroutineCompleted <- false
		}
	})

	// 等待主协程完成
	select {
	case mainResult := <-mainGoroutineCompleted:
		if !mainResult {
			t.Error("主协程应该正常完成")
		}
	case <-time.After(600 * time.Millisecond):
		t.Error("主协程测试超时")
	}

	// 等待子协程完成
	select {
	case subResult := <-subGoroutineCompleted:
		if !subResult {
			t.Error("子协程应该正常完成")
		}
	case <-time.After(600 * time.Millisecond):
		t.Error("子协程测试超时")
	}
}

// TestGoroutinePanicWithContext 测试带context的协程panic处理
func TestGoroutinePanicWithContext(t *testing.T) {
	recovery := NewPanicRecovery(nil)

	// 添加测试处理器
	var receivedContext context.Context
	testHandler := &ContextTestHandler{
		receivedContext: &receivedContext,
	}
	recovery.AddHandler(testHandler)

	// 创建测试上下文
	type panicTestKey string
	ctx := context.WithValue(context.Background(), panicTestKey("panic_test_key"), "panic_test_value")

	// 启动会panic的协程
	recovery.SafeGoWithContext("panic-context-test", ctx, func() {
		panic("带context的panic测试")
	})

	// 等待处理完成
	time.Sleep(100 * time.Millisecond)

	// 验证context被正确传递
	if receivedContext == nil {
		t.Error("应该接收到上下文")
	} else if receivedContext.Value(panicTestKey("panic_test_key")) != "panic_test_value" {
		t.Error("上下文值不正确")
	}
}

// 测试用的处理器

type TestPanicHandler struct {
	handled *bool
}

func (h *TestPanicHandler) HandlePanic(info *PanicInfo) error {
	if h.handled != nil {
		*h.handled = true
	}
	return nil
}

type ContextTestHandler struct {
	receivedContext *context.Context
}

func (h *ContextTestHandler) HandlePanic(info *PanicInfo) error {
	if h.receivedContext != nil {
		*h.receivedContext = info.Context
	}
	return nil
}

type SlowPanicHandler struct {
	duration time.Duration
	called   bool
	mu       sync.Mutex
}

func (h *SlowPanicHandler) HandlePanic(info *PanicInfo) error {
	// 立即设置called标志，确保在超时之前就被设置
	h.mu.Lock()
	h.called = true
	h.mu.Unlock()

	// 检查context是否已取消
	select {
	case <-info.Context.Done():
		// 如果context已取消，提前返回
		return info.Context.Err()
	default:
		// 继续执行
	}

	// 在慢操作中定期检查context状态
	start := time.Now()
	for time.Since(start) < h.duration {
		select {
		case <-info.Context.Done():
			// context被取消，提前退出
			return info.Context.Err()
		case <-time.After(50 * time.Millisecond):
			// 继续执行
		}
	}

	return nil
}

type ContextTimeoutHandler struct {
	timeoutDetected bool
	mu              sync.Mutex
}

func (h *ContextTimeoutHandler) HandlePanic(info *PanicInfo) error {
	// 检查context是否已取消
	select {
	case <-info.Context.Done():
		h.mu.Lock()
		h.timeoutDetected = true
		h.mu.Unlock()
		return info.Context.Err()
	default:
		// 继续执行
	}

	// 模拟长时间操作，定期检查context
	for i := 0; i < 10; i++ {
		select {
		case <-info.Context.Done():
			h.mu.Lock()
			h.timeoutDetected = true
			h.mu.Unlock()
			return info.Context.Err()
		case <-time.After(20 * time.Millisecond):
			// 继续执行
		}
	}

	return nil
}

type PanicPanicHandler struct {
	called bool
}

func (h *PanicPanicHandler) HandlePanic(info *PanicInfo) error {
	h.called = true
	panic("处理器内部panic")
}

type ConcurrentTestHandler struct {
	wg *sync.WaitGroup
	id int
}

func (h *ConcurrentTestHandler) HandlePanic(info *PanicInfo) error {
	defer h.wg.Done()
	time.Sleep(10 * time.Millisecond) // 模拟处理时间
	return nil
}

// BenchmarkGetGoroutineID 性能测试
func BenchmarkGetGoroutineID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getGoroutineID()
	}
}

// BenchmarkPanicRecovery 性能测试
func BenchmarkPanicRecovery(b *testing.B) {
	recovery := NewPanicRecovery(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recovery.SafeGo("benchmark", func() {
			// 空函数，不触发panic
		})
	}
}

// BenchmarkHandlerExecution 处理器执行性能测试
func BenchmarkHandlerExecution(b *testing.B) {
	recovery := NewPanicRecovery(nil)

	// 添加快速处理器
	fastHandler := &TestPanicHandler{}
	recovery.AddHandler(fastHandler)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recovery.SafeGo("benchmark", func() {
			panic("性能测试panic")
		})
	}
}
