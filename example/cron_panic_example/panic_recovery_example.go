package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/cron"
)

func main() {
	// 创建cron管理器
	cm := cron.New(
		cron.WithSeconds(),
		cron.WithLogger(log.Default()),
	)

	// 创建正常任务
	normalTask := cron.NewTask("normal_task", "* * * * * *", func(ctx context.Context) error {
		fmt.Println("正常任务执行中...")
		return nil
	})

	// 创建会panic的任务
	panicTask := cron.NewTask("panic_task", "* * * * * *", func(ctx context.Context) error {
		fmt.Println("即将panic的任务...")
		panic("这是一个测试panic")
	})

	// 添加任务
	_, err1 := cm.AddTask(normalTask)
	if err1 != nil {
		log.Fatal("添加正常任务失败:", err1)
	}

	_, err2 := cm.AddTask(panicTask)
	if err2 != nil {
		log.Fatal("添加panic任务失败:", err2)
	}

	// 启动cron
	fmt.Println("启动cron管理器...")
	cm.Start()

	// 运行10秒
	time.Sleep(10 * time.Second)

	// 停止cron
	fmt.Println("停止cron管理器...")
	cm.Stop()

	// 查看任务指标
	tasks := cm.GetAllTasks()
	for id, task := range tasks {
		metrics := cm.GetTaskMetrics(id)
		fmt.Printf("任务: %s, 成功次数: %d, 失败次数: %d, 最后错误: %v\n",
			task.Name, metrics.SuccessCount, metrics.FailureCount, metrics.LastError)
	}
}
