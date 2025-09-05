package co

import (
	"context"
	"runtime"
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

	t.Run("不监听context的业务函数", func(t *testing.T) {
		start := time.Now()

		// 业务函数不监听context，应该能正常超时
		GoWithTimeout("test-no-context", 100*time.Millisecond, func(ctx context.Context) {
			time.Sleep(200 * time.Millisecond) // 超过超时时间
		})

		// 应该能快速返回，不等待业务函数完成
		if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
			t.Errorf("应该快速超时返回: %v", elapsed)
		}
	})

	t.Run("空函数", func(t *testing.T) {
		start := time.Now()

		GoWithTimeout("test-empty", 100*time.Millisecond, func(ctx context.Context) {
			// 空函数
		})

		if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
			t.Errorf("空函数应该快速完成: %v", elapsed)
		}
	})

	t.Run("零超时时间", func(t *testing.T) {
		start := time.Now()

		GoWithTimeout("test-zero-timeout", 0, func(ctx context.Context) {
			time.Sleep(10 * time.Millisecond)
		})

		// 应该立即超时
		if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
			t.Errorf("零超时应该立即返回: %v", elapsed)
		}
	})

	t.Run("并发安全", func(t *testing.T) {
		const goroutines = 100
		var wg sync.WaitGroup
		wg.Add(goroutines)

		for i := 0; i < goroutines; i++ {
			go func(id int) {
				defer wg.Done()
				GoWithTimeout("test-concurrent", 50*time.Millisecond, func(ctx context.Context) {
					time.Sleep(20 * time.Millisecond)
				})
			}(i)
		}

		wg.Wait()
		t.Log("✅ 并发测试通过")
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

	t.Run("异步超时", func(t *testing.T) {
		start := time.Now()
		completed := make(chan struct{})

		AsyncGoWithTimeout("test-async-timeout", 100*time.Millisecond, func(ctx context.Context) {
			select {
			case <-ctx.Done():
				close(completed)
			case <-time.After(200 * time.Millisecond):
				t.Error("不应该执行到这里")
			}
		})

		// 应该立即返回
		if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
			t.Errorf("异步函数应该立即返回: %v", elapsed)
		}

		// 等待超时发生
		select {
		case <-completed:
			t.Log("✅ 异步函数正确超时")
		case <-time.After(300 * time.Millisecond):
			t.Error("测试超时")
		}
	})

	t.Run("异步panic", func(t *testing.T) {
		AsyncGoWithTimeout("test-async-panic", 100*time.Millisecond, func(ctx context.Context) {
			panic("测试panic")
		})

		// 应该立即返回，不阻塞
		time.Sleep(50 * time.Millisecond)
		t.Log("✅ 异步panic恢复正常")
	})
}

// TestGoWithTimeoutCallback 测试带回调的同步超时函数
func TestGoWithTimeoutCallback(t *testing.T) {
	t.Run("正常完成带回调", func(t *testing.T) {
		executed := make(chan struct{})
		callbackCalled := false

		GoWithTimeoutCallback("test-normal-callback", 200*time.Millisecond,
			func(ctx context.Context) {
				time.Sleep(100 * time.Millisecond)
				close(executed)
			},
			func() {
				callbackCalled = true
			},
		)

		select {
		case <-executed:
			if callbackCalled {
				t.Error("正常完成时不应该调用回调")
			}
			t.Log("✅ 函数正常完成")
		case <-time.After(500 * time.Millisecond):
			t.Error("函数执行超时")
		}
	})

	t.Run("超时带回调", func(t *testing.T) {
		completed := make(chan struct{})
		callbackCalled := make(chan struct{})

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
				close(callbackCalled)
			},
		)

		select {
		case <-completed:
			// 超时时回调不会被调用，因为回调只在panic时调用
			select {
			case <-callbackCalled:
				t.Error("超时时不应该调用回调")
			case <-time.After(50 * time.Millisecond):
				t.Log("✅ 函数正确超时，回调未被调用（符合预期）")
			}
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

	t.Run("不监听context的业务函数", func(t *testing.T) {
		start := time.Now()

		GoWithTimeoutCallback("test-no-context-callback", 100*time.Millisecond,
			func(ctx context.Context) {
				time.Sleep(200 * time.Millisecond) // 超过超时时间
			},
			func() {
				// 回调函数
			},
		)

		// 应该能快速返回
		if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
			t.Errorf("应该快速超时返回: %v", elapsed)
		}
	})

	t.Run("空回调函数", func(t *testing.T) {
		start := time.Now()

		GoWithTimeoutCallback("test-empty-callback", 100*time.Millisecond,
			func(ctx context.Context) {
				time.Sleep(50 * time.Millisecond)
			},
			nil, // 空回调
		)

		if elapsed := time.Since(start); elapsed > 100*time.Millisecond {
			t.Errorf("空回调应该正常工作: %v", elapsed)
		}
	})

	t.Run("回调函数panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Error("回调函数panic应该被恢复")
			}
		}()

		GoWithTimeoutCallback("test-callback-panic", 100*time.Millisecond,
			func(ctx context.Context) {
				panic("业务函数panic")
			},
			func() {
				panic("回调函数panic")
			},
		)
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
		callbackCalled := make(chan struct{})

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
				close(callbackCalled)
			},
		)

		// 等待超时发生
		select {
		case <-completed:
			// 超时时回调不会被调用，因为回调只在panic时调用
			select {
			case <-callbackCalled:
				t.Error("超时时不应该调用回调")
			case <-time.After(50 * time.Millisecond):
				t.Log("✅ 异步函数正确超时，回调未被调用（符合预期）")
			}
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

// TestGoroutineLeak 测试goroutine泄漏
func TestGoroutineLeak(t *testing.T) {
	// 记录测试前的goroutine数量
	runtime.GC()
	before := runtime.NumGoroutine()

	// 执行大量超时操作
	const iterations = 1000
	for i := 0; i < iterations; i++ {
		GoWithTimeout("test-leak", 10*time.Millisecond, func(ctx context.Context) {
			time.Sleep(50 * time.Millisecond) // 超过超时时间
		})
	}

	// 等待一段时间让goroutine完成
	time.Sleep(200 * time.Millisecond)
	runtime.GC()

	// 检查goroutine数量
	after := runtime.NumGoroutine()
	leaked := after - before

	if leaked > 10 { // 允许一些误差
		t.Errorf("可能存在goroutine泄漏: 测试前%d个, 测试后%d个, 泄漏%d个", before, after, leaked)
	} else {
		t.Logf("✅ 无goroutine泄漏: 测试前%d个, 测试后%d个", before, after)
	}
}

// TestContextCancellation 测试context取消
func TestContextCancellation(t *testing.T) {
	t.Run("context正确取消", func(t *testing.T) {
		done := make(chan struct{})
		GoWithTimeout("test-context-cancel", 100*time.Millisecond, func(fnCtx context.Context) {
			select {
			case <-fnCtx.Done():
				close(done)
			case <-time.After(300 * time.Millisecond):
				t.Error("不应该执行到这里")
			}
		})

		select {
		case <-done:
			t.Log("✅ context正确取消")
		case <-time.After(300 * time.Millisecond):
			t.Error("context应该被取消")
		}
	})

	t.Run("context在业务函数中正确传递", func(t *testing.T) {
		receivedCtx := make(chan context.Context, 1)

		GoWithTimeout("test-context-pass", 100*time.Millisecond, func(ctx context.Context) {
			receivedCtx <- ctx
		})

		select {
		case ctx := <-receivedCtx:
			if ctx == nil {
				t.Error("context不应该为nil")
			}
			// 检查context是否有超时
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Error("context应该有deadline")
			}
			if time.Until(deadline) > 200*time.Millisecond {
				t.Error("context的deadline应该正确设置")
			}
		case <-time.After(200 * time.Millisecond):
			t.Error("应该接收到context")
		}
	})
}

// TestEdgeCases 测试边界情况
func TestEdgeCases(t *testing.T) {
	t.Run("负数超时时间", func(t *testing.T) {
		start := time.Now()

		GoWithTimeout("test-negative-timeout", -100*time.Millisecond, func(ctx context.Context) {
			time.Sleep(10 * time.Millisecond)
		})

		// 负数超时应该立即超时
		if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
			t.Errorf("负数超时应该立即返回: %v", elapsed)
		}
	})

	t.Run("极大超时时间", func(t *testing.T) {
		start := time.Now()
		executed := false

		GoWithTimeout("test-large-timeout", 24*time.Hour, func(ctx context.Context) {
			time.Sleep(10 * time.Millisecond)
			executed = true
		})

		if !executed {
			t.Error("函数应该被执行")
		}
		if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
			t.Errorf("应该快速完成: %v", elapsed)
		}
	})

	t.Run("空组件名", func(t *testing.T) {
		GoWithTimeout("", 100*time.Millisecond, func(ctx context.Context) {
			time.Sleep(10 * time.Millisecond)
		})
		// 应该不panic
	})

	t.Run("nil函数", func(t *testing.T) {
		// nil函数会被recovery机制捕获，不会导致程序崩溃
		GoWithTimeout("test-nil-fn", 100*time.Millisecond, nil)
		// 应该能正常返回，不panic
	})
}

// BenchmarkGoWithTimeout 性能测试
func BenchmarkGoWithTimeout(b *testing.B) {
	b.Run("正常完成", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GoWithTimeout("bench-normal", 100*time.Millisecond, func(ctx context.Context) {
				// 空操作
			})
		}
	})

	b.Run("超时", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GoWithTimeout("bench-timeout", 1*time.Millisecond, func(ctx context.Context) {
				time.Sleep(10 * time.Millisecond)
			})
		}
	})
}
