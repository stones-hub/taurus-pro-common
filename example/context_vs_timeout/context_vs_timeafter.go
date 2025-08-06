package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// 使用time.After()的问题版本
func timeAfterProblem() {
	fmt.Println("=== time.After() 的问题 ===")

	initialGoroutines := runtime.NumGoroutine()
	fmt.Printf("开始前goroutine数量: %d\n", initialGoroutines)

	for i := 0; i < 10; i++ {
		go func(id int) {
			done := make(chan bool, 1)

			// 模拟一个可能很慢的任务
			go func() {
				time.Sleep(200 * time.Millisecond)
				select {
				case done <- true:
				default:
					// 如果已经超时，这里会阻塞
				}
			}()

			select {
			case <-done:
				fmt.Printf("任务 %d 完成\n", id)
			case <-time.After(100 * time.Millisecond):
				fmt.Printf("任务 %d 超时\n", id)
				// 问题：time.After()创建的goroutine可能还在运行
			}
		}(i)
	}

	time.Sleep(500 * time.Millisecond)
	currentGoroutines := runtime.NumGoroutine()
	fmt.Printf("结束后goroutine数量: %d\n", currentGoroutines)
	fmt.Printf("泄漏的goroutine: %d\n", currentGoroutines-initialGoroutines)
}

// 使用context的修复版本
func contextSolution() {
	fmt.Println("\n=== Context 解决方案 ===")

	initialGoroutines := runtime.NumGoroutine()
	fmt.Printf("开始前goroutine数量: %d\n", initialGoroutines)

	for i := 0; i < 10; i++ {
		go func(id int) {
			// 使用context控制超时
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel() // 确保context被取消

			done := make(chan bool, 1)

			go func() {
				time.Sleep(200 * time.Millisecond)
				select {
				case done <- true:
				case <-ctx.Done():
					// context已取消，不再发送结果
					fmt.Printf("任务 %d 的goroutine检测到context取消\n", id)
				}
			}()

			select {
			case <-done:
				fmt.Printf("任务 %d 完成\n", id)
			case <-ctx.Done():
				fmt.Printf("任务 %d 超时\n", id)
				// 优势：context取消后，相关goroutine会自动退出
			}
		}(i)
	}

	time.Sleep(500 * time.Millisecond)
	currentGoroutines := runtime.NumGoroutine()
	fmt.Printf("结束后goroutine数量: %d\n", currentGoroutines)
	fmt.Printf("泄漏的goroutine: %d\n", currentGoroutines-initialGoroutines)
}

// 演示context的链式取消
func contextChainCancellation() {
	fmt.Println("\n=== Context 链式取消 ===")

	// 创建父context
	parentCtx, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	// 创建子context
	childCtx, childCancel := context.WithTimeout(parentCtx, 200*time.Millisecond)
	defer childCancel()

	go func() {
		select {
		case <-childCtx.Done():
			fmt.Println("子context被取消")
		case <-time.After(300 * time.Millisecond):
			fmt.Println("子context超时")
		}
	}()

	// 模拟父context被取消
	time.Sleep(100 * time.Millisecond)
	parentCancel() // 取消父context，子context也会被取消

	time.Sleep(100 * time.Millisecond)
}

// 演示time.After()在高频调用下的问题
func timeAfterHighFrequency() {
	fmt.Println("\n=== time.After() 高频调用问题 ===")

	initialGoroutines := runtime.NumGoroutine()
	fmt.Printf("开始前goroutine数量: %d\n", initialGoroutines)

	// 模拟高频调用
	for i := 0; i < 100; i++ {
		go func(id int) {
			select {
			case <-time.After(50 * time.Millisecond):
				// 快速超时
			}
		}(i)
	}

	time.Sleep(200 * time.Millisecond)
	currentGoroutines := runtime.NumGoroutine()
	fmt.Printf("结束后goroutine数量: %d\n", currentGoroutines)
	fmt.Printf("泄漏的goroutine: %d\n", currentGoroutines-initialGoroutines)
}

// 演示context在高频调用下的优势
func contextHighFrequency() {
	fmt.Println("\n=== Context 高频调用优势 ===")

	initialGoroutines := runtime.NumGoroutine()
	fmt.Printf("开始前goroutine数量: %d\n", initialGoroutines)

	// 模拟高频调用
	for i := 0; i < 100; i++ {
		go func(id int) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			select {
			case <-ctx.Done():
				// 快速超时
			}
		}(i)
	}

	time.Sleep(200 * time.Millisecond)
	currentGoroutines := runtime.NumGoroutine()
	fmt.Printf("结束后goroutine数量: %d\n", currentGoroutines)
	fmt.Printf("泄漏的goroutine: %d\n", currentGoroutines-initialGoroutines)
}

func main() {
	fmt.Println("Context vs time.After() 详细对比")
	fmt.Println("==================================")

	// 基础对比
	timeAfterProblem()
	contextSolution()

	// 高级特性对比
	contextChainCancellation()

	// 高频调用对比
	timeAfterHighFrequency()
	contextHighFrequency()

	fmt.Println("\n=== 总结 ===")
	fmt.Println("Context的优势:")
	fmt.Println("1. 更好的资源管理 - 自动清理相关goroutine")
	fmt.Println("2. 链式取消 - 父context取消会传播到子context")
	fmt.Println("3. 更精确的控制 - 可以传递值、设置截止时间等")
	fmt.Println("4. 标准库支持 - 与Go生态更好地集成")
	fmt.Println("5. 避免goroutine泄漏 - 特别是在高频调用场景")
}
