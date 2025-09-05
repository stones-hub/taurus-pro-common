package co

import (
	"context"
	"sync"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/recovery"
)

// 便捷函数 - 直接使用全局实例 （安全函数）

// Recover panic恢复
func Recover(component string) {
	recovery.GlobalPanicRecovery.Recover(component)
}

// RecoverWithCallback panic恢复后，执行回调(只有在函数中自己处理了panic，才会执行回调)
func RecoverWithCallback(component string, callback func()) {
	recovery.GlobalPanicRecovery.RecoverWithCallback(component, callback)
}

// SafeGo 安全地启动goroutine, 并保证goroutine中panic会被恢复
func Go(component string, fn func()) {
	recovery.GlobalPanicRecovery.SafeGo(component, fn)
}

// SafeGoWithCallback 安全地启动goroutine, 并保证goroutine中panic会被恢复, panic恢复后，执行回调
func GoWithCallback(component string, fn func(), callback func()) {
	recovery.GlobalPanicRecovery.SafeGoWithCallback(component, fn, callback)
}

// WrapFunction 包装函数以添加panic恢复, 并保证函数中panic会被恢复(不是goroutine中)
func WrapFunction(component string, fn func()) func() {
	return recovery.GlobalPanicRecovery.WrapFunction(component, fn)
}

// WrapFunctionWithCallback 包装函数以添加panic恢复和回调, 并保证函数中panic会被恢复(不是goroutine中)
func WrapFunctionWithCallback(component string, fn func(), callback func()) func() {
	return recovery.GlobalPanicRecovery.WrapFunctionWithCallback(component, fn, callback)
}

// WrapErrorFunction 包装返回错误的函数, 并保证函数中panic会被恢复(不是goroutine中)
func WrapErrorFunction(component string, fn func() error) func() error {
	return recovery.GlobalPanicRecovery.WrapErrorFunction(component, fn)
}

// WrapErrorFunctionWithCallback 包装返回错误的函数并执行回调, 并保证函数中panic会被恢复(不是goroutine中)
func WrapErrorFunctionWithCallback(component string, fn func() error, callback func()) func() error {
	return recovery.GlobalPanicRecovery.WrapErrorFunctionWithCallback(component, fn, callback)
}

// GoWithTimeout 同步版本，带超时控制函数
func GoWithTimeout(component string, timeout time.Duration, fn func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	done := make(chan struct{})
	var once sync.Once
	closeDone := func() { once.Do(func() { close(done) }) }
	defer closeDone() // 保险：确保done channel总是被关闭

	recovery.GlobalPanicRecovery.SafeGo(component, func() {
		defer closeDone() // 确保无论是否panic都会关闭channel
		fn(ctx)
	})

	select {
	case <-ctx.Done():
		// 超时，立即返回
		return
	case <-done:
		// 正常完成，立即返回
		return
	}
}

// AsyncGoWithTimeout 异步版本，带超时控制函数
func AsyncGoWithTimeout(component string, timeout time.Duration, fn func(ctx context.Context)) {
	recovery.GlobalPanicRecovery.SafeGo(component, func() {
		GoWithTimeout(component, timeout, fn)
	})
}

// GoWithTimeoutCallback 同步版本，带回调的超时控制函数
func GoWithTimeoutCallback(component string, timeout time.Duration, fn func(ctx context.Context), callback func()) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	done := make(chan struct{})
	var once sync.Once
	closeDone := func() { once.Do(func() { close(done) }) }
	defer closeDone() // 保险：确保done channel总是被关闭

	recovery.GlobalPanicRecovery.SafeGoWithCallback(component, func() {
		defer closeDone() // 确保无论是否panic都会关闭channel
		fn(ctx)
	}, callback)

	select {
	case <-ctx.Done():
		// 超时，立即返回
		return
	case <-done:
		// 正常完成，立即返回
		return
	}
}

// AsyncGoWithTimeoutCallback 异步版本，带回调的超时控制函数
func AsyncGoWithTimeoutCallback(component string, timeout time.Duration, fn func(ctx context.Context), callback func()) {
	recovery.GlobalPanicRecovery.SafeGo(component, func() {
		GoWithTimeoutCallback(component, timeout, fn, callback)
	})
}
