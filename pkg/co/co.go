package co

import (
	"context"

	"github.com/stones-hub/taurus-pro-common/pkg/recovery"
)

// 便捷函数 - 直接使用全局实例 （安全函数）

// Recover 全局panic恢复
func Recover(component string) {
	recovery.GlobalPanicRecovery.Recover(component)
}

// RecoverWithContext 全局panic恢复并传递上下文
func RecoverWithContext(component string, ctx context.Context) {
	recovery.GlobalPanicRecovery.RecoverWithContext(component, ctx)
}

// RecoverWithCallback 全局panic恢复并执行回调
func RecoverWithCallback(component string, callback func()) {
	recovery.GlobalPanicRecovery.RecoverWithCallback(component, callback)
}

// SafeGo 安全地启动goroutine
func Go(component string, fn func()) {
	recovery.GlobalPanicRecovery.SafeGo(component, fn)
}

// SafeGoWithContext 安全地启动goroutine并传递上下文
func GoWithContext(component string, ctx context.Context, fn func()) {
	recovery.GlobalPanicRecovery.SafeGoWithContext(component, ctx, fn)
}

// SafeGoWithCallback 安全地启动goroutine并执行回调
func GoWithCallback(component string, fn func(), callback func()) {
	recovery.GlobalPanicRecovery.SafeGoWithCallback(component, fn, callback)
}

// WrapFunction 包装函数以添加panic恢复
func WrapFunction(component string, fn func()) func() {
	return recovery.GlobalPanicRecovery.WrapFunction(component, fn)
}

// WrapFunctionWithContext 包装函数以添加panic恢复和上下文
func WrapFunctionWithContext(component string, ctx context.Context, fn func()) func() {
	return recovery.GlobalPanicRecovery.WrapFunctionWithContext(component, ctx, fn)
}

// WrapFunctionWithCallback 包装函数以添加panic恢复和回调
func WrapFunctionWithCallback(component string, fn func(), callback func()) func() {
	return recovery.GlobalPanicRecovery.WrapFunctionWithCallback(component, fn, callback)
}

// WrapErrorFunction 包装返回错误的函数
func WrapErrorFunction(component string, fn func() error) func() error {
	return recovery.GlobalPanicRecovery.WrapErrorFunction(component, fn)
}

// WrapErrorFunctionWithContext 包装返回错误的函数并传递上下文
func WrapErrorFunctionWithContext(component string, ctx context.Context, fn func() error) func() error {
	return recovery.GlobalPanicRecovery.WrapErrorFunctionWithContext(component, ctx, fn)
}
