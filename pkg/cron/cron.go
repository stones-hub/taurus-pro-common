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

package cron

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// CronManager 管理所有的定时任务
type CronManager struct {
	cron    *cron.Cron                    // cron 实例
	tasks   map[cron.EntryID]*Task        // 存储任务信息的映射
	metrics map[cron.EntryID]*TaskMetrics // 存储任务指标
	groups  map[string]*TaskGroup         // 存储任务分组
	logger  *log.Logger                   // 日志记录器
	mu      sync.RWMutex                  // 保护共享资源的互斥锁
}

// New 创建一个新的 CronManager 实例
func New(opts ...Option) *CronManager {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	cronOptions := []cron.Option{}

	// 根据并发控制模式设置相应的 JobWrapper
	var wrappers []cron.JobWrapper
	switch options.ConcurrencyMode {
	case SkipIfRunning:
		wrappers = []cron.JobWrapper{
			cron.SkipIfStillRunning(cron.DefaultLogger),
		}
	case DelayIfRunning:
		wrappers = []cron.JobWrapper{
			cron.DelayIfStillRunning(cron.DefaultLogger),
		}
	default:
		wrappers = []cron.JobWrapper{
			cron.Recover(cron.DefaultLogger), // 默认只添加错误恢复
		}
	}
	cronOptions = append(cronOptions, cron.WithChain(wrappers...))

	if options.EnableSeconds {
		cronOptions = append(cronOptions, cron.WithSeconds())
	}
	if options.Location != nil {
		cronOptions = append(cronOptions, cron.WithLocation(options.Location))
	}

	return &CronManager{
		cron:    cron.New(cronOptions...),
		tasks:   make(map[cron.EntryID]*Task),
		metrics: make(map[cron.EntryID]*TaskMetrics),
		groups:  make(map[string]*TaskGroup),
		logger:  options.Logger,
		mu:      sync.RWMutex{},
	}
}

// Start 启动 cron 调度器
func (cm *CronManager) Start() {
	cm.cron.Start()
	log.Println("\033[1;32m🔗 -> Cron initialized successfully\033[0m")
}

// Stop 停止 cron 调度器
func (cm *CronManager) Stop() {
	cm.cron.Stop()
	log.Println("\033[1;31m🔗 -> Cron stopped\033[0m")
}

// GracefulStop 优雅停止所有任务
func (cm *CronManager) GracefulStop(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 停止 cron 调度器并获取等待上下文
	cronStop := cm.cron.Stop()

	// 等待所有运行中的任务完成或超时
	select {
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %v", ctx.Err())
	case <-cronStop.Done():
		return nil
	}
}

// AddTask 添加一个新的定时任务
func (cm *CronManager) AddTask(task *Task) (cron.EntryID, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 创建任务包装函数
	wrapper := func() {
		startTime := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
		defer cancel()

		var err error
		done := make(chan error, 1)

		// 在新的 goroutine 中执行任务
		go func() {
			for i := 0; i <= task.RetryCount; i++ {
				if err = task.Func(ctx); err == nil {
					done <- nil
					return
				}
				if i < task.RetryCount {
					time.Sleep(task.RetryDelay)
				}
			}
			done <- err
		}()

		// 等待任务完成或超时
		select {
		case err = <-done:
			// 任务正常完成或重试完成
		case <-ctx.Done():
			err = fmt.Errorf("task timeout after %v", task.Timeout)
		}

		duration := time.Since(startTime)
		cm.updateMetrics(task.Name, duration, err)
	}

	id, err := cm.cron.AddFunc(task.Spec, wrapper)
	if err != nil {
		return 0, err
	}

	cm.tasks[id] = task
	cm.metrics[id] = &TaskMetrics{
		CreatedAt: time.Now(),
	}

	if task.Group != nil {
		cm.groups[task.Group.Name] = task.Group
	}

	return id, nil
}

// RemoveTask 移除一个定时任务
func (cm *CronManager) RemoveTask(id cron.EntryID) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 获取要删除的任务
	task := cm.tasks[id]
	if task != nil && task.Group != nil {
		// 检查这个分组是否还有其他任务
		hasOtherTasks := false
		for tid, t := range cm.tasks {
			if tid != id && t.Group != nil && t.Group.Name == task.Group.Name {
				hasOtherTasks = true
				break
			}
		}
		// 如果分组没有其他任务了，删除分组
		if !hasOtherTasks {
			delete(cm.groups, task.Group.Name)
		}
	}

	cm.cron.Remove(id)
	delete(cm.tasks, id)
	delete(cm.metrics, id)
}

// GetTasksByGroup 获取指定分组的所有任务
func (cm *CronManager) GetTasksByGroup(groupName string) []*Task {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var groupTasks []*Task
	for _, task := range cm.tasks {
		if task.Group != nil && task.Group.Name == groupName {
			groupTasks = append(groupTasks, task)
		}
	}
	return groupTasks
}

// GetTasksByTag 获取具有指定标签的所有任务
func (cm *CronManager) GetTasksByTag(tag string) []*Task {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var taggedTasks []*Task
	for _, task := range cm.tasks {
		if task.HasTag(tag) {
			taggedTasks = append(taggedTasks, task)
		}
	}
	return taggedTasks
}

// GetTaskMetrics 获取任务的指标信息
func (cm *CronManager) GetTaskMetrics(id cron.EntryID) *TaskMetrics {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if metrics, exists := cm.metrics[id]; exists {
		return metrics
	}
	return nil
}

// updateMetrics 更新任务指标
func (cm *CronManager) updateMetrics(taskName string, duration time.Duration, err error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for id, task := range cm.tasks {
		if task.Name == taskName {
			metrics := cm.metrics[id]
			metrics.LastRunTime = time.Now()
			metrics.LastDuration = duration

			if err != nil {
				metrics.FailureCount++
				metrics.LastError = err
			} else {
				metrics.SuccessCount++
			}

			// 更新平均执行时间
			totalCount := int64(metrics.SuccessCount + metrics.FailureCount)
			if totalCount > 1 {
				metrics.AverageDuration = time.Duration(
					(metrics.AverageDuration.Nanoseconds()*(totalCount-1) + duration.Nanoseconds()) / totalCount,
				)
			} else {
				metrics.AverageDuration = duration
			}
			break
		}
	}
}

// ListTasks 列出所有的定时任务
func (cm *CronManager) ListTasks() []*Task {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	tasks := make([]*Task, 0, len(cm.tasks))
	for _, task := range cm.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// ModifyTask 修改一个定时任务
func (cm *CronManager) ModifyTask(id cron.EntryID, newTask *Task) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.tasks[id]; !exists {
		return fmt.Errorf("task with ID %d does not exist", id)
	}

	cm.RemoveTask(id)
	newID, err := cm.AddTask(newTask)
	if err != nil {
		return err
	}

	cm.logger.Printf("Task %s modified successfully, New ID: %d", newTask.Name, newID)
	return nil
}
