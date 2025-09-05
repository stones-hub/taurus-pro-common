package co

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestGoWithTimeout 测试同步超时函数
func TestGoWithTimeout(t *testing.T) {
	t.Run("正常完成", func(t *testing.T) {
		start := time.Now()
		executed := false

		GoWithTimeout("test-normal", 200*time.Millisecond, func(ctx context.Context) {
			time.Sleep(100 * time.Millisecond) // 模拟工作耗时
			executed = true
		})

		if !executed {
			t.Error("函数应该被执行")
		}
		if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
			t.Errorf("执行时间过长: %v", elapsed)
		}
	})

	t.Run("超时情况", func(t *testing.T) {
		start := time.Now()
		completed := make(chan struct{})

		GoWithTimeout("test-timeout", 100*time.Millisecond, func(ctx context.Context) {
			select {
			case <-time.After(200 * time.Millisecond):
				t.Error("不应该执行到这里")
			case <-ctx.Done():
				// 正确的超时处理
				completed <- struct{}{}
			}
		})

		select {
		case <-completed:
			if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
				t.Errorf("超时处理时间过长: %v", elapsed)
			}
		case <-time.After(300 * time.Millisecond):
			t.Error("测试超时")
		}
	})

	t.Run("panic恢复", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error("panic应该被恢复")
			}
		}()

		GoWithTimeout("test-panic", 100*time.Millisecond, func(ctx context.Context) {
			panic("测试panic")
		})
	})
}

// TestAsyncGoWithTimeout 测试异步超时函数
func TestAsyncGoWithTimeout(t *testing.T) {
	t.Run("异步执行", func(t *testing.T) {
		start := time.Now()
		var wg sync.WaitGroup
		wg.Add(1)

		AsyncGoWithTimeout("test-async", 200*time.Millisecond, func(ctx context.Context) {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)
		})

		// 应该立即返回
		if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
			t.Errorf("异步函数应该立即返回: %v", elapsed)
		}

		// 等待实际完成
		wg.Wait()
	})
}

// TestGoWithTimeoutCallback 测试带回调的同步超时函数
func TestGoWithTimeoutCallback(t *testing.T) {
	t.Run("正常完成带回调", func(t *testing.T) {
		executed := make(chan struct{})

		GoWithTimeoutCallback("test-normal-callback", 200*time.Millisecond,
			func(ctx context.Context) {
				time.Sleep(100 * time.Millisecond)
				close(executed)
			},
			func() {
				t.Error("正常完成时不应该调用回调")
			},
		)

		select {
		case <-executed:
			t.Log("✅ 函数正常完成")
		case <-time.After(500 * time.Millisecond):
			t.Error("函数执行超时")
		}
	})

	t.Run("超时带回调", func(t *testing.T) {
		completed := make(chan struct{})

		GoWithTimeoutCallback("test-timeout-callback", 100*time.Millisecond,
			func(ctx context.Context) {
				select {
				case <-ctx.Done():
					close(completed)
				case <-time.After(200 * time.Millisecond):
					t.Error("不应该执行到这里")
				}
			},
			func() {
				t.Error("超时时不应该调用回调")
			},
		)

		select {
		case <-completed:
			t.Log("✅ 函数正确超时")
		case <-time.After(300 * time.Millisecond):
			t.Error("测试超时")
		}
	})

	t.Run("panic带回调", func(t *testing.T) {
		callbackCalled := make(chan struct{})

		GoWithTimeoutCallback("test-panic-callback", 100*time.Millisecond,
			func(ctx context.Context) {
				panic("测试panic")
			},
			func() {
				close(callbackCalled)
			},
		)

		// 等待回调执行
		select {
		case <-callbackCalled:
			t.Log("✅ panic恢复后回调正确执行")
		case <-time.After(200 * time.Millisecond):
			t.Error("panic恢复后回调应该被执行")
		}
	})
}

// TestAsyncGoWithTimeoutCallback 测试带回调的异步超时函数
func TestAsyncGoWithTimeoutCallback(t *testing.T) {
	t.Run("异步执行带回调", func(t *testing.T) {
		start := time.Now()
		var wg sync.WaitGroup
		wg.Add(1)

		AsyncGoWithTimeoutCallback("test-async-callback", 200*time.Millisecond,
			func(ctx context.Context) {
				defer wg.Done()
				time.Sleep(100 * time.Millisecond)
			},
			func() {
				t.Error("正常完成时不应该调用回调")
			},
		)

		// 应该立即返回
		if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
			t.Errorf("异步函数应该立即返回: %v", elapsed)
		}

		// 等待实际完成
		wg.Wait()
		t.Log("✅ 异步函数正常完成")
	})

	t.Run("异步超时带回调", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		completed := make(chan struct{})

		AsyncGoWithTimeoutCallback("test-async-timeout-callback", 100*time.Millisecond,
			func(ctx context.Context) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					close(completed)
				case <-time.After(200 * time.Millisecond):
					t.Error("不应该执行到这里")
				}
			},
			func() {
				t.Error("超时时不应该调用回调")
			},
		)

		// 等待超时发生
		select {
		case <-completed:
			t.Log("✅ 异步函数正确超时")
		case <-time.After(300 * time.Millisecond):
			t.Error("测试超时")
		}

		wg.Wait()
	})

	t.Run("异步panic带回调", func(t *testing.T) {
		callbackCalled := make(chan struct{})

		AsyncGoWithTimeoutCallback("test-async-panic-callback", 100*time.Millisecond,
			func(ctx context.Context) {
				panic("测试panic")
			},
			func() {
				close(callbackCalled)
			},
		)

		// 等待回调执行
		select {
		case <-callbackCalled:
			t.Log("✅ 异步panic恢复后回调正确执行")
		case <-time.After(200 * time.Millisecond):
			t.Error("panic恢复后回调应该被执行")
		}
	})
}
