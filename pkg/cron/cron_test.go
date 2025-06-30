package cron

import (
	"context"
	"errors"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCronManager(t *testing.T) {
	t.Run("基本功能测试", func(t *testing.T) {
		// 创建 CronManager 实例
		cm := New(
			WithSeconds(),
			WithLogger(log.Default()),
		)
		assert.NotNil(t, cm)

		// 启动和停止
		cm.Start()
		time.Sleep(time.Second)
		cm.Stop()
	})

	t.Run("任务执行测试", func(t *testing.T) {
		cm := New(WithSeconds())

		// 创建一个计数器，用于验证任务执行
		counter := 0
		var mu sync.Mutex

		// 创建一个每秒执行的任务
		task := NewTask("test_task", "* * * * * *", func(ctx context.Context) error {
			mu.Lock()
			counter++
			mu.Unlock()
			return nil
		})

		// 添加任务
		id, err := cm.AddTask(task)
		assert.NoError(t, err)
		assert.NotEqual(t, 0, id)

		// 启动并等待任务执行
		cm.Start()
		time.Sleep(3 * time.Second)
		cm.Stop()

		// 验证任务至少执行了2次
		assert.GreaterOrEqual(t, counter, 2)
	})

	t.Run("任务超时测试", func(t *testing.T) {
		cm := New(WithSeconds())

		// 创建一个会超时的任务
		task := NewTask("timeout_task", "* * * * * *", func(ctx context.Context) error {
			time.Sleep(2 * time.Second)
			return nil
		}, WithTimeout(time.Second))

		id, err := cm.AddTask(task)
		assert.NoError(t, err)

		cm.Start()
		time.Sleep(3 * time.Second)

		// 检查指标
		metrics := cm.GetTaskMetrics(id)
		assert.NotNil(t, metrics)
		assert.NotNil(t, metrics.LastError)

		cm.Stop()
	})

	t.Run("任务重试测试", func(t *testing.T) {
		cm := New(WithSeconds())

		retryCount := 0
		task := NewTask("retry_task", "* * * * * *", func(ctx context.Context) error {
			retryCount++
			return errors.New("故意失败")
		}, WithRetry(2, time.Millisecond*100))

		id, err := cm.AddTask(task)
		assert.NoError(t, err)

		cm.Start()
		time.Sleep(2 * time.Second)
		cm.Stop()

		metrics := cm.GetTaskMetrics(id)
		assert.NotNil(t, metrics)
		assert.NotNil(t, metrics.LastError)
		assert.GreaterOrEqual(t, retryCount, 3) // 初始执行 + 2次重试
	})

	t.Run("分组功能测试", func(t *testing.T) {
		cm := New(WithSeconds())

		// 创建分组
		group := NewTaskGroup("test_group")
		group.AddTag("test_tag")

		// 创建两个任务并加入分组
		task1 := NewTask("task1", "* * * * * *", func(ctx context.Context) error {
			return nil
		}, WithGroup(group))

		task2 := NewTask("task2", "* * * * * *", func(ctx context.Context) error {
			return nil
		}, WithGroup(group))

		// 添加任务
		id1, _ := cm.AddTask(task1)
		id2, _ := cm.AddTask(task2)

		// 验证分组任务数量
		groupTasks := cm.GetTasksByGroup("test_group")
		assert.Equal(t, 2, len(groupTasks))

		// 删除一个任务
		cm.RemoveTask(id1)
		groupTasks = cm.GetTasksByGroup("test_group")
		assert.Equal(t, 1, len(groupTasks))

		// 删除另一个任务
		cm.RemoveTask(id2)
		groupTasks = cm.GetTasksByGroup("test_group")
		assert.Empty(t, groupTasks)
	})

	t.Run("标签功能测试", func(t *testing.T) {
		cm := New(WithSeconds())

		// 创建带标签的任务
		task := NewTask("tagged_task", "* * * * * *", func(ctx context.Context) error {
			return nil
		}, WithTag("important"))

		// 添加任务
		_, err := cm.AddTask(task)
		assert.NoError(t, err)

		// 按标签查找任务
		taggedTasks := cm.GetTasksByTag("important")
		assert.Equal(t, 1, len(taggedTasks))
		assert.True(t, taggedTasks[0].HasTag("important"))
	})

	t.Run("并发控制测试_Skip", func(t *testing.T) {
		cm := New(WithSeconds(), WithConcurrencyMode(SkipIfRunning))

		execCount := 0
		var mu sync.Mutex

		task := NewTask("long_task", "* * * * * *", func(ctx context.Context) error {
			mu.Lock()
			execCount++
			mu.Unlock()
			log.Println("execCount start", execCount)
			// 增加执行时间，确保下一次调度时任务还在运行
			time.Sleep(5 * time.Second)
			log.Println("execCount end", execCount)
			return nil
		})

		_, err := cm.AddTask(task)
		assert.NoError(t, err)

		cm.Start()
		// 等待足够长的时间，让任务有机会被调度多次
		time.Sleep(4 * time.Second)
		cm.Stop()

		// 由于任务执行时间为3秒，而调度间隔为1秒，
		// 在SkipIfRunning模式下，第二次调度应该被跳过
		assert.Equal(t, 1, execCount)
	})

	t.Run("优雅关闭测试", func(t *testing.T) {
		cm := New(WithSeconds())

		task := NewTask("graceful_task", "* * * * * *", func(ctx context.Context) error {
			time.Sleep(2 * time.Second)
			return nil
		})

		_, err := cm.AddTask(task)
		assert.NoError(t, err)

		cm.Start()
		time.Sleep(time.Second)

		// 测试超时情况
		err = cm.GracefulStop(time.Second)
		assert.Error(t, err)

		// 测试正常关闭
		err = cm.GracefulStop(3 * time.Second)
		assert.NoError(t, err)
	})
}
