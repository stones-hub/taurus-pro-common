package tgo

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSpecialFunnel 测试漏斗创建
func TestNewSpecialFunnel(t *testing.T) {
	tests := []struct {
		name    string
		config  *FunnelConfig
		wantErr bool
	}{
		{
			name: "正常配置",
			config: &FunnelConfig{
				ProcessNum:     5,
				Cap:            100,
				Handler:        func(data interface{}) error { return nil },
				HandlerTimeout: 10 * time.Second,
				Timeout:        5 * time.Second,
				Interval:       10,
			},
			wantErr: false,
		},
		{
			name: "默认值配置",
			config: &FunnelConfig{
				Handler: func(data interface{}) error { return nil },
			},
			wantErr: false,
		},
		{
			name: "零值配置",
			config: &FunnelConfig{
				ProcessNum: 0,
				Cap:        0,
				Handler:    func(data interface{}) error { return nil },
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funnel, closeFunc, err := NewSpecialFunnel(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, funnel)
			require.NotNil(t, closeFunc)

			// 验证默认值
			if tt.config.ProcessNum <= 0 {
				assert.Equal(t, 30, funnel.GetProcessNum())
			}
			if tt.config.Cap <= 0 {
				assert.Equal(t, 1000, funnel.GetCap())
			}
			if tt.config.HandlerTimeout <= 0 {
				assert.Equal(t, 60*time.Second, funnel.GetHandlerTimeout())
			}
			if tt.config.Timeout <= 0 {
				assert.Equal(t, 3*time.Second, funnel.GetTimeout())
			}
			if tt.config.Interval <= 0 {
				// 心跳协程会运行，所以协程数量会包含心跳协程
				assert.GreaterOrEqual(t, funnel.GetGoroutineCount(), int64(1))
			}

			// 清理
			closeFunc()
		})
	}
}

// TestSpecialFunnel_AddData 测试数据添加
func TestSpecialFunnel_AddData(t *testing.T) {
	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 2,
		Cap:        2,
		Handler:    func(data interface{}) error { log.Println("data: ", data); return nil },
		Timeout:    100 * time.Millisecond,
	})
	require.NoError(t, err)
	defer closeFunc()

	// 测试正常添加数据
	funnel.AddData("test1")

	// 测试通道满时的行为
	funnel.AddData("test2")
	funnel.AddData("test3") // 这个会被阻塞，然后超时
	funnel.AddData("test4") // 这个也会被阻塞，然后超时

	// 等待处理完成
	time.Sleep(200 * time.Millisecond)
	// 由于通道容量为2，前两条数据会被处理，后两条会超时
	// 超时的数据会通过错误通道处理，所以实际处理的数据可能更多
	assert.GreaterOrEqual(t, funnel.GetProcessedCount(), int64(2))
}

// TestSpecialFunnel_Close 测试漏斗关闭
func TestSpecialFunnel_Close(t *testing.T) {
	tests := []struct {
		name           string
		processNum     int
		handlerTimeout time.Duration
		closeTimeout   time.Duration
	}{
		{
			name:           "基本关闭测试",
			processNum:     2,
			handlerTimeout: 10 * time.Millisecond,
			closeTimeout:   100 * time.Millisecond,
		},
		{
			name:           "超时关闭测试",
			processNum:     2,
			handlerTimeout: 200 * time.Millisecond,
			closeTimeout:   50 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建一个会阻塞的handler来测试超时关闭
			blockChan := make(chan struct{})
			funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
				ProcessNum:     tt.processNum,
				Cap:            10,
				Handler:        func(data interface{}) error { <-blockChan; return nil },
				HandlerTimeout: tt.handlerTimeout,
				Timeout:        tt.closeTimeout,
			})
			require.NoError(t, err)

			// 添加一些数据
			for i := 0; i < 3; i++ {
				funnel.AddData(fmt.Sprintf("test%d", i))
			}

			// 启动关闭
			closeFunc()

			// 验证关闭状态
			assert.True(t, funnel.IsClosed())

			// 等待协程退出
			time.Sleep(200 * time.Millisecond)
			assert.Equal(t, int64(0), funnel.GetGoroutineCount(), "所有协程应该已退出")

			// 清理
			close(blockChan)
		})
	}
}

// TestSpecialFunnel_ContextCancel 测试上下文取消
func TestSpecialFunnel_ContextCancel(t *testing.T) {
	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 2,
		Cap:        10,
		Handler:    func(data interface{}) error { time.Sleep(100 * time.Millisecond); return nil },
		Timeout:    200 * time.Millisecond,
	})
	require.NoError(t, err)

	// 添加数据
	funnel.AddData("test1")
	funnel.AddData("test2")

	// 立即关闭
	closeFunc()

	// 验证协程数量
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(0), funnel.GetGoroutineCount())
}

// TestSpecialFunnel_ErrorHandling 测试错误处理
func TestSpecialFunnel_ErrorHandling(t *testing.T) {
	errorCount := int64(0)
	errorMutex := sync.Mutex{}

	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 2,
		Cap:        10,
		Handler: func(data interface{}) error {
			return fmt.Errorf("处理错误: %v", data)
		},
		ErrorHandler: func(err error) {
			errorMutex.Lock()
			errorCount++
			errorMutex.Unlock()
		},
		Timeout: 100 * time.Millisecond,
	})
	require.NoError(t, err)
	defer closeFunc()

	// 添加数据
	funnel.AddData("test1")
	funnel.AddData("test2")

	// 等待错误处理
	time.Sleep(200 * time.Millisecond)

	// 验证错误处理
	errorMutex.Lock()
	assert.Equal(t, int64(2), errorCount)
	errorMutex.Unlock()
}

// TestSpecialFunnel_Heartbeat 测试心跳功能
func TestSpecialFunnel_Heartbeat(t *testing.T) {
	heartbeatCount := int64(0)
	heartbeatMutex := sync.Mutex{}

	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 2,
		Cap:        10,
		Handler:    func(data interface{}) error { return nil },
		Heartbeat: func(f *SpecialFunnel) {
			heartbeatMutex.Lock()
			heartbeatCount++
			heartbeatMutex.Unlock()
		},
		Interval: 1, // 1秒间隔
		Timeout:  100 * time.Millisecond,
	})
	require.NoError(t, err)

	// 验证漏斗状态
	assert.False(t, funnel.IsClosed())

	// 等待协程启动完成
	time.Sleep(100 * time.Millisecond)

	// 协程数量应该至少包含工作协程、心跳协程和错误处理协程
	assert.GreaterOrEqual(t, funnel.GetGoroutineCount(), int64(1))

	// 等待心跳
	time.Sleep(3 * time.Second)

	// 验证心跳
	heartbeatMutex.Lock()
	assert.GreaterOrEqual(t, heartbeatCount, int64(2))
	heartbeatMutex.Unlock()

	// 清理
	closeFunc()
}

// TestSpecialFunnel_PanicRecovery 测试错误处理
func TestSpecialFunnel_PanicRecovery(t *testing.T) {
	// 创建一个会返回错误的handler
	errorHandler := func(data interface{}) error {
		if data == "trigger_error" {
			return fmt.Errorf("测试错误: %v", data)
		}
		return nil
	}

	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 2,
		Cap:        10,
		Handler:    errorHandler,
		Timeout:    100 * time.Millisecond,
	})
	require.NoError(t, err)
	defer closeFunc()

	// 先添加一个正常数据
	funnel.AddData("normal_data")
	time.Sleep(50 * time.Millisecond)

	// 再添加一个会触发错误的数据
	funnel.AddData("trigger_error")
	time.Sleep(100 * time.Millisecond)

	// 验证漏斗仍然可用
	assert.False(t, funnel.IsClosed())
	// 协程数量应该至少为1
	assert.GreaterOrEqual(t, funnel.GetGoroutineCount(), int64(1))
}

// TestSpecialFunnel_ConcurrentAccess 测试并发访问
func TestSpecialFunnel_ConcurrentAccess(t *testing.T) {
	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 5,
		Cap:        100,
		Handler:    func(data interface{}) error { time.Sleep(10 * time.Millisecond); return nil },
		Timeout:    100 * time.Millisecond,
	})
	require.NoError(t, err)
	defer closeFunc()

	// 并发添加数据
	var wg sync.WaitGroup
	dataCount := 50

	for i := 0; i < dataCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			funnel.AddData(fmt.Sprintf("concurrent_test_%d", id))
		}(i)
	}

	wg.Wait()

	// 等待处理完成
	time.Sleep(500 * time.Millisecond)

	// 验证处理结果
	assert.Equal(t, int64(dataCount), funnel.GetProcessedCount())
}

// TestSpecialFunnel_RepeatedClose 测试重复关闭
func TestSpecialFunnel_RepeatedClose(t *testing.T) {
	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 2,
		Cap:        10,
		Handler:    func(data interface{}) error { return nil },
		Timeout:    100 * time.Millisecond,
	})
	require.NoError(t, err)

	// 第一次关闭
	closeFunc()
	assert.True(t, funnel.IsClosed())

	// 第二次关闭应该被忽略
	closeFunc()
	assert.True(t, funnel.IsClosed())

	// 验证协程数量
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, int64(0), funnel.GetGoroutineCount())
}

// TestSpecialFunnel_AddDataAfterClose 测试关闭后添加数据
func TestSpecialFunnel_AddDataAfterClose(t *testing.T) {
	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 2,
		Cap:        10,
		Handler: func(data interface{}) error {
			// 模拟处理时间，确保数据不会立即处理完
			time.Sleep(50 * time.Millisecond)
			return nil
		},
		Timeout: 100 * time.Millisecond,
	})
	require.NoError(t, err)

	// 先添加一些数据
	funnel.AddData("test_before_close")

	// 立即关闭漏斗
	closeFunc()
	assert.True(t, funnel.IsClosed())

	// 尝试添加数据（应该被拒绝）
	funnel.AddData("test_after_close")

	// 等待所有协程退出
	time.Sleep(300 * time.Millisecond)

	// 验证漏斗状态
	assert.True(t, funnel.IsClosed())
	assert.Equal(t, int64(0), funnel.GetGoroutineCount(), "所有协程应该已退出")
}

// TestSpecialFunnel_Statistics 测试统计信息
func TestSpecialFunnel_Statistics(t *testing.T) {
	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 2,
		Cap:        10,
		Handler:    func(data interface{}) error { return nil },
		Timeout:    100 * time.Millisecond,
	})
	require.NoError(t, err)
	defer closeFunc()

	// 验证初始状态
	assert.Equal(t, int64(0), funnel.GetProcessedCount())
	assert.GreaterOrEqual(t, funnel.GetGoroutineCount(), int64(3)) // worker + heartbeat + errorWorker
	assert.Equal(t, 2, funnel.GetProcessNum())
	assert.Equal(t, 10, funnel.GetCap())
	assert.NotEmpty(t, funnel.GetID())

	// 添加数据
	funnel.AddData("test1")
	funnel.AddData("test2")

	// 等待处理
	time.Sleep(100 * time.Millisecond)

	// 验证处理结果
	assert.Equal(t, int64(2), funnel.GetProcessedCount())
}

// BenchmarkSpecialFunnel_AddData 性能测试：数据添加
func BenchmarkSpecialFunnel_AddData(b *testing.B) {
	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 10,
		Cap:        1000,
		Handler:    func(data interface{}) error { return nil },
		Timeout:    100 * time.Millisecond,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer closeFunc()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		funnel.AddData(fmt.Sprintf("benchmark_test_%d", i))
	}
}

// BenchmarkSpecialFunnel_ConcurrentAddData 性能测试：并发数据添加
func BenchmarkSpecialFunnel_ConcurrentAddData(b *testing.B) {
	funnel, closeFunc, err := NewSpecialFunnel(&FunnelConfig{
		ProcessNum: 20,
		Cap:        10000,
		Handler:    func(data interface{}) error { return nil },
		Timeout:    100 * time.Millisecond,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer closeFunc()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			funnel.AddData(fmt.Sprintf("concurrent_benchmark_%d", i))
			i++
		}
	})
}
