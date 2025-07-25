package hook

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestHookManager_RegisterAndExecuteHooks(t *testing.T) {
	hm := NewHookManager()

	// 记录执行顺序
	executionOrder := make([]string, 0)

	// 注册启动钩子
	hm.RegisterStartHook("service1", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "start1")
		return nil
	}, 10)

	hm.RegisterStartHook("service2", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "start2")
		return nil
	}, 5)

	// 注册停止钩子
	hm.RegisterStopHook("service1", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "stop1")
		return nil
	}, 10)

	hm.RegisterStopHook("service2", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "stop2")
		return nil
	}, 5)

	// 执行启动钩子
	ctx := context.Background()
	if err := hm.Start(ctx); err != nil {
		t.Fatalf("启动钩子执行失败: %v", err)
	}

	// 执行停止钩子
	if err := hm.Stop(ctx); err != nil {
		t.Fatalf("停止钩子执行失败: %v", err)
	}

	// 验证执行顺序
	expectedOrder := []string{"start2", "start1", "stop2", "stop1"}
	if len(executionOrder) != len(expectedOrder) {
		t.Fatalf("执行顺序长度不匹配，期望: %d, 实际: %d", len(expectedOrder), len(executionOrder))
	}

	for i, expected := range expectedOrder {
		if executionOrder[i] != expected {
			t.Fatalf("执行顺序错误，位置 %d: 期望 %s, 实际 %s", i, expected, executionOrder[i])
		}
	}
}

func TestHookManager_ErrorHandling(t *testing.T) {
	hm := NewHookManager()

	// 注册一个会失败的启动钩子
	hm.RegisterStartHookDefault("failing_service", func(ctx context.Context) error {
		return fmt.Errorf("启动失败")
	})

	// 注册一个正常的启动钩子
	hm.RegisterStartHookDefault("normal_service", func(ctx context.Context) error {
		return nil
	})

	// 执行启动钩子，应该因第一个钩子失败而停止
	ctx := context.Background()
	err := hm.Start(ctx)
	if err == nil {
		t.Fatal("期望启动失败，但实际成功了")
	}

	if err.Error() != "启动钩子 failing_service 执行失败: 启动失败" {
		t.Fatalf("错误信息不匹配: %v", err)
	}
}

func TestHookManager_StopHookErrorHandling(t *testing.T) {
	hm := NewHookManager()

	// 注册一个会失败的停止钩子
	hm.RegisterStopHookDefault("failing_stop", func(ctx context.Context) error {
		return fmt.Errorf("停止失败")
	})

	// 注册一个正常的停止钩子
	hm.RegisterStopHookDefault("normal_stop", func(ctx context.Context) error {
		return nil
	})

	// 执行停止钩子，应该继续执行所有钩子，但返回最后一个错误
	ctx := context.Background()
	err := hm.Stop(ctx)
	if err == nil {
		t.Fatal("期望停止时返回错误，但实际没有错误")
	}

	if err.Error() != "停止钩子 failing_stop 执行失败: 停止失败" {
		t.Fatalf("错误信息不匹配: %v", err)
	}
}

func TestHookManager_ContextCancellation(t *testing.T) {
	hm := NewHookManager()

	// 注册一个长时间运行的钩子
	hm.RegisterStartHookDefault("long_running", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return nil
		}
	})

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 执行启动钩子，应该因上下文超时而失败
	err := hm.Start(ctx)
	if err == nil {
		t.Fatal("期望因上下文超时而失败，但实际成功了")
	}

	// 检查错误是否包含超时信息
	if err.Error() != "启动钩子 long_running 执行失败: context deadline exceeded" {
		t.Fatalf("错误信息不匹配，期望包含超时信息，实际: %v", err)
	}
}

func TestHookManager_GetHookCount(t *testing.T) {
	hm := NewHookManager()

	// 初始状态
	startCount, stopCount := hm.GetHookCount()
	if startCount != 0 || stopCount != 0 {
		t.Fatalf("初始状态钩子数量错误，期望: 0,0, 实际: %d,%d", startCount, stopCount)
	}

	// 注册钩子
	hm.RegisterStartHookDefault("start1", func(ctx context.Context) error { return nil })
	hm.RegisterStartHookDefault("start2", func(ctx context.Context) error { return nil })
	hm.RegisterStopHookDefault("stop1", func(ctx context.Context) error { return nil })

	// 检查钩子数量
	startCount, stopCount = hm.GetHookCount()
	if startCount != 2 || stopCount != 1 {
		t.Fatalf("钩子数量错误，期望: 2,1, 实际: %d,%d", startCount, stopCount)
	}
}

func TestHookManager_ClearHooks(t *testing.T) {
	hm := NewHookManager()

	// 注册钩子
	hm.RegisterStartHookDefault("start1", func(ctx context.Context) error { return nil })
	hm.RegisterStopHookDefault("stop1", func(ctx context.Context) error { return nil })

	// 清空钩子
	hm.ClearHooks()

	// 检查钩子数量
	startCount, stopCount := hm.GetHookCount()
	if startCount != 0 || stopCount != 0 {
		t.Fatalf("清空后钩子数量错误，期望: 0,0, 实际: %d,%d", startCount, stopCount)
	}
}

func TestHookManager_DefaultPriority(t *testing.T) {
	hm := NewHookManager()

	executionOrder := make([]string, 0)

	// 使用默认优先级注册钩子
	hm.RegisterStartHookDefault("service1", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "service1")
		return nil
	})

	hm.RegisterStartHookDefault("service2", func(ctx context.Context) error {
		executionOrder = append(executionOrder, "service2")
		return nil
	})

	// 执行启动钩子
	ctx := context.Background()
	if err := hm.Start(ctx); err != nil {
		t.Fatalf("启动钩子执行失败: %v", err)
	}

	// 验证执行顺序（默认优先级相同，按注册顺序执行）
	if len(executionOrder) != 2 {
		t.Fatalf("执行数量错误，期望: 2, 实际: %d", len(executionOrder))
	}

	if executionOrder[0] != "service1" || executionOrder[1] != "service2" {
		t.Fatalf("执行顺序错误: %v", executionOrder)
	}
}
