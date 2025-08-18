// Copyright (c) 2025 Taurus Team. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: yelei
// Email: 61647649@qq.com
// Date: 2025-06-13

package tgo

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestNewWorkerPool 测试协程池的创建
func TestNewWorkerPool(t *testing.T) {
	capacity := 5
	pool := NewWorkerPool(capacity)

	if pool == nil {
		t.Fatal("协程池创建失败")
	}

	if pool.capacity != capacity {
		t.Errorf("期望容量为 %d，实际为 %d", capacity, pool.capacity)
	}

	if pool.isClosed != false {
		t.Error("新创建的协程池应该是未关闭状态")
	}

	if cap(pool.taskChan) != capacity {
		t.Errorf("任务通道容量期望为 %d，实际为 %d", capacity, cap(pool.taskChan))
	}

	if cap(pool.taskResultChan) != capacity {
		t.Errorf("结果通道容量期望为 %d，实际为 %d", capacity, cap(pool.taskResultChan))
	}

	if pool.tasks != 0 {
		t.Errorf("新协程池的任务数应该为 0，实际为 %d", pool.tasks)
	}
}

// TestWorkerPool_Run 测试协程池的启动
func TestWorkerPool_Run(t *testing.T) {
	capacity := 3
	pool := NewWorkerPool(capacity)

	// 启动协程池
	pool.Run()

	// 等待一下让协程启动
	time.Sleep(100 * time.Millisecond)

	// 验证协程池状态
	if pool.isClosed {
		t.Error("启动后的协程池不应该是关闭状态")
	}
}

// TestWorkerPool_Register 测试任务注册
func TestWorkerPool_Register(t *testing.T) {
	capacity := 2
	pool := NewWorkerPool(capacity)
	pool.Run()
	defer pool.Close()

	// 创建测试任务
	task1 := &Task{
		Id: "task-001",
		Handler: func() error {
			return nil
		},
	}

	task2 := &Task{
		Id: "task-002",
		Handler: func() error {
			return errors.New("模拟错误")
		},
	}

	// 注册任务
	pool.Register(task1)
	pool.Register(task2)

	// 等待任务执行完成
	time.Sleep(200 * time.Millisecond)

	// 获取结果
	results := pool.ResultPool()

	if len(results) != 2 {
		t.Errorf("期望结果数量为 2，实际为 %d", len(results))
	}

	// 由于任务执行顺序不确定，需要通过ID来查找对应的结果
	task1Result := (*TaskResult)(nil)
	task2Result := (*TaskResult)(nil)

	for _, result := range results {
		switch result.Id {
		case "task-001":
			task1Result = result
		case "task-002":
			task2Result = result
		}
	}

	// 验证任务1成功执行
	if task1Result == nil {
		t.Error("未找到任务1的结果")
	} else if task1Result.Error != nil {
		t.Errorf("任务1应该成功执行，但返回了错误: %v", task1Result.Error)
	}

	// 验证任务2执行失败
	if task2Result == nil {
		t.Error("未找到任务2的结果")
	} else if task2Result.Error == nil {
		t.Error("任务2应该执行失败，但没有返回错误")
	}
}

// TestWorkerPool_RegisterWhenClosed 测试协程池关闭后注册任务
func TestWorkerPool_RegisterWhenClosed(t *testing.T) {
	capacity := 2
	pool := NewWorkerPool(capacity)
	pool.Run()

	// 关闭协程池
	pool.Close()

	// 尝试注册任务
	task := &Task{
		Id: "task-after-close",
		Handler: func() error {
			return nil
		},
	}

	pool.Register(task)

	// 验证任务没有被注册
	if pool.tasks != 0 {
		t.Errorf("关闭后不应该能注册任务，任务数: %d", pool.tasks)
	}
}

// TestWorkerPool_ResultPool 测试结果收集
func TestWorkerPool_ResultPool(t *testing.T) {
	capacity := 3
	pool := NewWorkerPool(capacity)
	pool.Run()
	defer pool.Close()

	// 创建多个任务
	tasks := []*Task{
		{Id: "task-1", Handler: func() error { return nil }},
		{Id: "task-2", Handler: func() error { return errors.New("错误2") }},
		{Id: "task-3", Handler: func() error { return nil }},
	}

	// 注册任务
	for _, task := range tasks {
		pool.Register(task)
	}

	// 等待任务执行完成
	time.Sleep(300 * time.Millisecond)

	// 获取结果
	results := pool.ResultPool()

	if len(results) != 3 {
		t.Errorf("期望结果数量为 3，实际为 %d", len(results))
	}

	// 验证结果
	expectedResults := map[string]bool{
		"task-1": false, // 无错误
		"task-2": true,  // 有错误
		"task-3": false, // 无错误
	}

	for _, result := range results {
		expectedError := expectedResults[result.Id]
		if expectedError && result.Error == nil {
			t.Errorf("任务 %s 应该返回错误", result.Id)
		}
		if !expectedError && result.Error != nil {
			t.Errorf("任务 %s 不应该返回错误: %v", result.Id, result.Error)
		}
	}
}

// TestWorkerPool_Close 测试协程池关闭
func TestWorkerPool_Close(t *testing.T) {
	capacity := 2
	pool := NewWorkerPool(capacity)
	pool.Run()

	// 注册一个任务
	task := &Task{
		Id: "task-before-close",
		Handler: func() error {
			time.Sleep(100 * time.Millisecond) // 模拟长时间执行
			return nil
		},
	}
	pool.Register(task)

	// 关闭协程池
	pool.Close()

	// 验证协程池已关闭
	if !pool.isClosed {
		t.Error("协程池应该已关闭")
	}

	// 尝试再次关闭（应该不会panic）
	pool.Close()
}

// TestWorkerPool_ConcurrentRegister 测试并发注册任务
func TestWorkerPool_ConcurrentRegister(t *testing.T) {
	capacity := 10
	pool := NewWorkerPool(capacity)
	pool.Run()
	defer pool.Close()

	var wg sync.WaitGroup
	taskCount := 50

	// 并发注册任务
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task := &Task{
				Id:      fmt.Sprintf("concurrent-task-%d", id),
				Handler: func() error { return nil },
			}
			pool.Register(task)
		}(i)
	}

	wg.Wait()

	// 等待任务执行完成
	time.Sleep(500 * time.Millisecond)

	// 获取结果
	results := pool.ResultPool()

	// 由于通道容量限制，实际执行的任务数可能少于注册的任务数
	// 但应该至少执行了通道容量的任务数
	if len(results) < capacity {
		t.Errorf("期望至少执行 %d 个任务，实际执行了 %d 个", capacity, len(results))
	}

	// 验证所有返回的结果都是有效的
	for _, result := range results {
		if result.Id == "" {
			t.Error("任务ID不应该为空")
		}
	}
}

// TestWorkerPool_TaskChannelFull 测试任务通道满的情况
func TestWorkerPool_TaskChannelFull(t *testing.T) {
	capacity := 2
	pool := NewWorkerPool(capacity)
	pool.Run()
	defer pool.Close()

	// 创建阻塞的任务处理器
	blockingHandler := func() error {
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	// 注册任务填满通道
	for i := 0; i < capacity; i++ {
		task := &Task{
			Id:      fmt.Sprintf("blocking-task-%d", i),
			Handler: blockingHandler,
		}
		pool.Register(task)
	}

	// 尝试注册更多任务（应该被忽略）
	extraTask := &Task{
		Id:      "extra-task",
		Handler: func() error { return nil },
	}

	pool.Register(extraTask)

	// 等待一段时间让阻塞任务开始执行
	time.Sleep(100 * time.Millisecond)

	// 验证额外任务没有被注册
	if pool.tasks != capacity {
		t.Errorf("期望任务数为 %d，实际为 %d", capacity, pool.tasks)
	}

	// 等待所有任务完成
	time.Sleep(300 * time.Millisecond)

	// 获取结果验证
	results := pool.ResultPool()
	if len(results) != capacity {
		t.Errorf("期望结果数量为 %d，实际为 %d", capacity, len(results))
	}
}

// TestWorkerPool_PanicRecovery 测试任务panic的恢复
// 注意：根据设计，任务处理器应该自己处理panic
func TestWorkerPool_PanicRecovery(t *testing.T) {
	capacity := 2
	pool := NewWorkerPool(capacity)
	pool.Run()
	defer pool.Close()

	// 创建会panic但被恢复的任务
	panicTask := &Task{
		Id: "panic-task",
		Handler: func() error {
			defer func() {
				if r := recover(); r != nil {
					// 恢复panic，返回错误而不是让程序崩溃
				}
			}()
			panic("模拟panic")
		},
	}

	// 注册任务
	pool.Register(panicTask)

	// 等待任务执行完成
	time.Sleep(200 * time.Millisecond)

	// 获取结果
	results := pool.ResultPool()

	if len(results) != 1 {
		t.Errorf("期望结果数量为 1，实际为 %d", len(results))
	}

	// 验证任务ID
	if results[0].Id != "panic-task" {
		t.Errorf("期望任务ID为 panic-task，实际为 %s", results[0].Id)
	}
}

// TestWorkerPool_EmptyResultPool 测试空结果池
func TestWorkerPool_EmptyResultPool(t *testing.T) {
	capacity := 2
	pool := NewWorkerPool(capacity)
	pool.Run()
	defer pool.Close()

	// 不注册任何任务，直接获取结果
	results := pool.ResultPool()

	if len(results) != 0 {
		t.Errorf("期望结果数量为 0，实际为 %d", len(results))
	}
}

// TestWorkerPool_MultipleClose 测试多次关闭
func TestWorkerPool_MultipleClose(t *testing.T) {
	capacity := 2
	pool := NewWorkerPool(capacity)
	pool.Run()

	// 第一次关闭
	pool.Close()

	// 验证状态
	if !pool.isClosed {
		t.Error("协程池应该已关闭")
	}

	// 再次关闭（应该不会panic）
	pool.Close()

	// 状态应该保持不变
	if !pool.isClosed {
		t.Error("协程池应该保持关闭状态")
	}
}

// TestWorkerPool_ZeroCapacity 测试零容量协程池
func TestWorkerPool_ZeroCapacity(t *testing.T) {
	pool := NewWorkerPool(0)

	if pool.capacity != 0 {
		t.Errorf("期望容量为 0，实际为 %d", pool.capacity)
	}

	if cap(pool.taskChan) != 0 {
		t.Errorf("任务通道容量期望为 0，实际为 %d", cap(pool.taskChan))
	}

	if cap(pool.taskResultChan) != 0 {
		t.Errorf("结果通道容量期望为 0，实际为 %d", cap(pool.taskResultChan))
	}
}

// TestWorkerPool_LargeCapacity 测试大容量协程池
func TestWorkerPool_LargeCapacity(t *testing.T) {
	capacity := 100
	pool := NewWorkerPool(capacity)
	pool.Run()
	defer pool.Close()

	if pool.capacity != capacity {
		t.Errorf("期望容量为 %d，实际为 %d", capacity, pool.capacity)
	}

	// 注册大量任务
	taskCount := 50
	for i := 0; i < taskCount; i++ {
		task := &Task{
			Id:      fmt.Sprintf("large-capacity-task-%d", i),
			Handler: func() error { return nil },
		}
		pool.Register(task)
	}

	// 等待任务执行完成
	time.Sleep(500 * time.Millisecond)

	// 获取结果
	results := pool.ResultPool()

	if len(results) != taskCount {
		t.Errorf("期望结果数量为 %d，实际为 %d", taskCount, len(results))
	}
}

// BenchmarkWorkerPool_Register 基准测试：任务注册性能
func BenchmarkWorkerPool_Register(b *testing.B) {
	capacity := 100
	pool := NewWorkerPool(capacity)
	pool.Run()
	defer pool.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &Task{
			Id:      fmt.Sprintf("bench-task-%d", i),
			Handler: func() error { return nil },
		}
		pool.Register(task)
	}

	// 等待所有任务完成
	time.Sleep(100 * time.Millisecond)
	pool.ResultPool()
}

// BenchmarkWorkerPool_Execution 基准测试：任务执行性能
func BenchmarkWorkerPool_Execution(b *testing.B) {
	capacity := 10
	pool := NewWorkerPool(capacity)
	pool.Run()
	defer pool.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &Task{
			Id:      fmt.Sprintf("bench-exec-task-%d", i),
			Handler: func() error { return nil },
		}
		pool.Register(task)
	}

	// 等待所有任务完成
	time.Sleep(100 * time.Millisecond)
	pool.ResultPool()
}
