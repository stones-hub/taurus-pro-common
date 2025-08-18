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
	"fmt"
	"log"
	"sync"
)

// Task 定义协程池中的任务结构
// 任务完成不代表协程停止，因为可能存在任务数量大于协程数的情况
//
// 字段说明：
//   - Id: 任务的唯一标识符
//   - Handler: 任务的执行函数，返回error表示执行结果
//
// 使用示例：
//
//	task := &tgo.Task{
//	    Id: "task-001",
//	    Handler: func() error {
//	        // 执行具体任务...
//	        return nil
//	    },
//	}
//
// 注意事项：
//   - Id应该是唯一的
//   - Handler不应该是nil
//   - Handler应该处理自己的panic
//   - Handler执行时间不应过长
//   - 适用于需要异步执行的任务
type Task struct {
	Id      string       `json:"id"`      // 协程任务ID
	Handler func() error `json:"handler"` // 协程任务执行代码
}

// TaskResult 定义任务执行的结果结构
// 用于返回任务的执行状态和可能的错误信息
//
// 字段说明：
//   - Id: 对应任务的唯一标识符
//   - Error: 任务执行过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	results := pool.ResultPool()
//	for _, result := range results {
//	    if result.Error != nil {
//	        log.Printf("任务 %s 执行失败：%v", result.Id, result.Error)
//	    } else {
//	        log.Printf("任务 %s 执行成功", result.Id)
//	    }
//	}
//
// 注意事项：
//   - Id与原任务ID一致
//   - Error为nil表示成功
//   - 通过ResultPool获取
//   - 按任务完成顺序返回
//   - 适用于需要获取任务执行结果的场景
type TaskResult struct {
	Id    string `json:"id"`    // 协程任务ID
	Error error  `json:"error"` // 协程任务执行错误
}

// WorkerPool 提供协程池功能，用于管理和复用goroutine
// 支持任务的异步执行、结果收集和优雅关闭
//
// 字段说明：
//   - tasks: 当前已注册的任务数量
//   - mutex: 用于保护并发访问
//   - taskChan: 任务队列通道
//   - taskResultChan: 任务结果队列通道
//   - capacity: 协程池的容量（最大协程数）
//   - closeChan: 用于发送关闭信号
//   - isClosed: 协程池是否已关闭
//   - closeOnce: 确保只关闭一次
//   - sg: 用于等待所有协程完成
//
// 使用示例：
//
//	// 创建容量为5的协程池
//	pool := tgo.NewWorkerPool(5)
//	pool.Run()
//	defer pool.Close()
//
//	// 注册任务
//	pool.Register(&tgo.Task{
//	    Id: "task-001",
//	    Handler: func() error {
//	        // 执行任务...
//	        return nil
//	    },
//	})
//
//	// 获取结果
//	results := pool.ResultPool()
//
// 注意事项：
//   - 使用前必须调用Run
//   - 使用完需要调用Close
//   - 任务通道容量等于协程数
//   - 支持优雅关闭
//   - 适用于需要控制并发的场景
//   - 建议使用defer Close
type WorkerPool struct {
	tasks          int
	mutex          sync.Mutex
	taskChan       chan *Task       // 协程任务队列
	taskResultChan chan *TaskResult // 协程任务返回队列
	capacity       int              // 协程数量
	closeChan      chan byte        // 协程池关闭信号
	isClosed       bool             // 协程池是否关闭
	closeOnce      sync.Once
	sg             *sync.WaitGroup
}

// NewWorkerPool 创建一个新的协程池实例
// 参数：
//   - capacity: 协程池的容量（最大协程数）
//
// 返回值：
//   - *WorkerPool: 初始化好的协程池实例
//
// 使用示例：
//
//	pool := tgo.NewWorkerPool(10) // 创建容量为10的协程池
//	pool.Run()
//	defer pool.Close()
//
// 注意事项：
//   - capacity应该根据实际需求设置
//   - 创建后需要调用Run才会启动
//   - 任务队列容量等于协程数
//   - 结果队列容量等于协程数
//   - 默认未启动和未关闭
//   - 适用于需要限制并发数的场景
func NewWorkerPool(capacity int) *WorkerPool {
	return &WorkerPool{
		taskChan:       make(chan *Task, capacity), //  任务队列的最大值最好跟任务数一致
		taskResultChan: make(chan *TaskResult, capacity),
		capacity:       capacity,
		isClosed:       false,
		closeChan:      make(chan byte),
		sg:             &sync.WaitGroup{},
		tasks:          0,
	}
}

// ResultPool 获取所有已完成任务的结果
// 返回值：
//   - []*TaskResult: 任务结果列表
//
// 使用示例：
//
//	results := pool.ResultPool()
//	for _, result := range results {
//	    if result.Error != nil {
//	        log.Printf("任务 %s 失败：%v", result.Id, result.Error)
//	    } else {
//	        log.Printf("任务 %s 成功", result.Id)
//	    }
//	}
//
// 注意事项：
//   - 会阻塞等待所有任务完成
//   - 按任务完成顺序返回结果
//   - 返回结果数量等于注册任务数
//   - 在Close之前调用
//   - 适用于需要等待所有任务完成的场景
//   - 建议在所有任务注册完成后调用
func (w *WorkerPool) ResultPool() []*TaskResult {
	resps := make([]*TaskResult, 0)
	for i := 0; i < w.tasks; i++ {
		resps = append(resps, <-w.taskResultChan)
	}
	return resps
}

// Run 启动协程池中的所有工作协程
// 使用示例：
//
//	pool := tgo.NewWorkerPool(5)
//	pool.Run() // 启动协程池
//	defer pool.Close()
//
// 注意事项：
//   - 在注册任务前调用
//   - 会立即创建所有协程
//   - 协程数量等于capacity
//   - 协程会等待任务或关闭信号
//   - 只能调用一次
//   - 必须在Close之前调用
func (w *WorkerPool) Run() {
	// 上来先创建协程
	for i := 0; i < w.capacity; i++ { // 创建capacity个协程
		w.sg.Add(1)
		go w.worker(i)
	}
}

// Register 向协程池注册一个新任务
// 参数：
//   - task: 要执行的任务
//
// 使用示例：
//
//	pool.Register(&tgo.Task{
//	    Id: "task-001",
//	    Handler: func() error {
//	        // 执行任务...
//	        return nil
//	    },
//	})
//
// 注意事项：
//   - 任务队列满时会被忽略
//   - 协程池关闭时会被忽略
//   - 是线程安全的
//   - 不会阻塞调用者
//   - 任务会被异步执行
//   - 适用于需要异步执行任务的场景
func (w *WorkerPool) Register(task *Task) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.isClosed {
		return
	}

	select {
	case w.taskChan <- task:
		w.tasks++
	default:
		log.Println("Task channel is full, cannot register task:", task.Id)
	}
}

// Close 优雅地关闭协程池
// 会等待所有正在执行的任务完成，并清理资源
//
// 使用示例：
//
//	pool := tgo.NewWorkerPool(5)
//	pool.Run()
//	defer pool.Close() // 确保协程池被关闭
//
// 注意事项：
//   - 是线程安全的
//   - 只会执行一次
//   - 会等待所有任务完成
//   - 会关闭所有通道
//   - 会阻塞直到所有协程退出
//   - 建议使用defer调用
//   - 关闭后的任务会被忽略
//   - 关闭是不可逆的
func (w *WorkerPool) Close() {
	w.closeOnce.Do(func() { // 只调用一次
		w.mutex.Lock()
		defer w.mutex.Unlock()

		if !w.isClosed {
			w.isClosed = true
			close(w.closeChan)
			close(w.taskChan)
			close(w.taskResultChan)
		}
		// 等待所有协程结束， 用了select还要用sg的原因是，尽管已经发送了关闭信号，但是协程可能还没来得及正常执行完退出，因此还是在等待比较好
		w.sg.Wait()
	})
}

// worker 工作协程的主循环函数
// 参数：
//   - id: 协程的唯一标识符
//
// 内部处理流程：
//  1. 等待任务或关闭信号
//  2. 收到任务时执行任务
//  3. 将结果写入结果通道
//  4. 收到关闭信号时退出
//
// 注意事项：
//   - 这是一个内部方法
//   - 在Run中被调用
//   - 会一直运行直到收到关闭信号
//   - 使用select监听多个通道
//   - 处理任务队列关闭的情况
//   - 会打印执行日志
func (w *WorkerPool) worker(id int) {
	defer w.sg.Done()

	for !w.isClosed {

		select {

		case task, ok := <-w.taskChan:
			if !ok {
				goto CLOSED
			}

			// --------> 开始执行任务 <---------

			log.Printf("协程 %d 执行任务 %s", id, task.Id)
			err := task.Handler()

			// 将执行的结果写入到结果队列中
			w.taskResultChan <- &TaskResult{
				Id:    task.Id,
				Error: err,
			}

		case <-w.closeChan: // 收到关闭信号
			goto CLOSED
		}
	}

CLOSED:
	fmt.Printf("协程id : %d 结束\n", id)
}
