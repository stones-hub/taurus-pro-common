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

package goroutine

import (
	"fmt"
	"log"
	"sync"
)

// 协程需要执行的任务信息, 任务完成不代表协程停止， 毕竟有可能任务多，协程数少
type Task struct {
	Id      string       `json:"id"`      // 协程任务ID
	Handler func() error `json:"handler"` // 协程任务执行代码
}

// 协程任务返回的信息
type TaskResult struct {
	Id    string `json:"id"`    // 协程任务ID
	Error error  `json:"error"` // 协程任务执行错误
}

// 协程池
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

// 创建协程池
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

func (w *WorkerPool) ResultPool() []*TaskResult {
	resps := make([]*TaskResult, 0)
	for i := 0; i < w.tasks; i++ {
		resps = append(resps, <-w.taskResultChan)
	}
	return resps
}

// 启动协程
func (w *WorkerPool) Run() {
	// 上来先创建协程
	for i := 0; i < w.capacity; i++ { // 创建capacity个协程
		w.sg.Add(1)
		go w.worker(i)
	}
}

// 注册任务到协程池子里面
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

// 强制关闭整个协程池, 避免重入调用， 加锁
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

// 协程中的运行Worker
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
