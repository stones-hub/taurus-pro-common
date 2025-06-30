package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/cron"
	"github.com/stones-hub/taurus-pro-common/pkg/ctx"
	"github.com/stones-hub/taurus-pro-common/pkg/util"
)

func main() {
	fmt.Println("=== 表格测试 ===")
	// 定义表头
	headers := []string{"编号", "项目名称", "状态", "进度", "负责人"}

	// 准备表格数据
	data := [][]interface{}{
		{1, "项目A", "进行中", "75%", "张三"},
		{2, "项目B", "已完成", "100%", "李四"},
		{3, "项目C", "待开始", "0%", "王五"},
		{4, "项目D", "已暂停", "45%", "赵六"},
		{5, "项目E", "规划中", "10%", "孙七"},
	}

	// 使用工具包渲染表格
	util.RenderTable(headers, data)

	fmt.Println("\n=== Context 测试 ===")
	contextExample()

	cronExample()
}

func contextExample() {
	// 创建基础 context
	baseCtx := context.Background()

	// 创建一个请求ID
	requestID := "req-123456"

	// 使用 WithTaurusContext 创建新的 context
	taurusCtx := ctx.WithTaurusContext(baseCtx, requestID)

	// 获取 TaurusContext
	tc := ctx.GetTaurusContext(taurusCtx)

	// 测试设置和获取数据
	tc.Set("user_id", "user-001")
	tc.Set("role", "admin")
	tc.Set("login_time", time.Now())

	// 打印基本信息
	fmt.Printf("请求ID: %s\n", tc.GetRequestID())
	fmt.Printf("创建时间: %v\n", tc.AtTime)

	// 获取并打印存储的数据
	fmt.Printf("\n存储的数据:\n")
	fmt.Printf("用户ID: %v\n", tc.Get("user_id"))
	fmt.Printf("角色: %v\n", tc.Get("role"))
	fmt.Printf("登录时间: %v\n", tc.Get("login_time"))

	// 测试不存在的键
	fmt.Printf("\n获取不存在的键:\n")
	fmt.Printf("不存在的键: %v\n", tc.Get("not_exist"))

	// 测试存储不同类型的数据
	tc.Set("int_value", 42)
	tc.Set("float_value", 3.14)
	tc.Set("bool_value", true)
	tc.Set("slice_value", []string{"a", "b", "c"})
	tc.Set("map_value", map[string]int{"one": 1, "two": 2})

	fmt.Printf("\n不同类型的数据:\n")
	fmt.Printf("整数: %v\n", tc.Get("int_value"))
	fmt.Printf("浮点数: %v\n", tc.Get("float_value"))
	fmt.Printf("布尔值: %v\n", tc.Get("bool_value"))
	fmt.Printf("切片: %v\n", tc.Get("slice_value"))
	fmt.Printf("映射: %v\n", tc.Get("map_value"))
}

func cronExample() {
	// 创建 cron 管理器，启用秒级调度
	cm := cron.New(
		cron.WithSeconds(),                           // 启用秒级调度
		cron.WithConcurrencyMode(cron.SkipIfRunning), // 设置并发控制模式
	)

	// 创建任务分组
	orderGroup := cron.NewTaskGroup("order")
	orderGroup.AddTag("business")
	orderGroup.AddTag("core")

	userGroup := cron.NewTaskGroup("user")
	userGroup.AddTag("business")

	// 创建一个普通任务：每5秒执行一次的订单检查任务
	orderCheckTask := cron.NewTask(
		"order_check",   // 任务名称
		"*/5 * * * * *", // cron 表达式（每5秒执行一次）
		func(ctx context.Context) error {
			log.Println("正在检查订单状态...")
			// 模拟任务执行
			time.Sleep(2 * time.Second)
			return nil
		},
		cron.WithTimeout(10*time.Second), // 设置超时时间
		cron.WithGroup(orderGroup),       // 设置任务分组
		cron.WithTag("check"),            // 添加标签
		cron.WithTag("periodic"),
	)

	// 创建一个可能超时的任务：每10秒执行一次的数据同步任务
	dataSyncTask := cron.NewTask(
		"data_sync",
		"*/10 * * * * *",
		func(ctx context.Context) error {
			log.Println("开始同步数据...")
			select {
			case <-ctx.Done():
				return fmt.Errorf("数据同步超时")
			case <-time.After(8 * time.Second):
				log.Println("数据同步完成")
				return nil
			}
		},
		cron.WithTimeout(5*time.Second), // 设置一个较短的超时时间来演示超时
		cron.WithGroup(userGroup),
		cron.WithTag("sync"),
	)

	// 创建一个会失败并重试的任务：每15秒执行一次的通知任务
	notifyTask := cron.NewTask(
		"notify",
		"*/15 * * * * *",
		func(ctx context.Context) error {
			log.Println("尝试发送通知...")
			// 模拟任务失败
			return errors.New("通知发送失败")
		},
		cron.WithRetry(3, time.Second), // 设置重试策略：最多重试3次，间隔1秒
		cron.WithGroup(userGroup),
		cron.WithTag("notification"),
	)

	// 添加所有任务
	orderCheckId, _ := cm.AddTask(orderCheckTask)
	dataSyncId, _ := cm.AddTask(dataSyncTask)
	notifyId, _ := cm.AddTask(notifyTask)

	// 启动 cron 管理器
	cm.Start()
	log.Println("Cron 管理器已启动")

	// 演示查询功能
	go func() {
		time.Sleep(30 * time.Second)

		// 获取任务指标
		if metrics := cm.GetTaskMetrics(orderCheckId); metrics != nil {
			log.Printf("订单检查任务指标: 最后执行时间=%v, 最后执行耗时=%v\n",
				metrics.LastRunTime, metrics.LastDuration)
		}
		if metrics := cm.GetTaskMetrics(dataSyncId); metrics != nil {
			log.Printf("数据同步任务指标: 最后执行时间=%v, 最后执行耗时=%v, 最后错误=%v\n",
				metrics.LastRunTime, metrics.LastDuration, metrics.LastError)
		}
		if metrics := cm.GetTaskMetrics(notifyId); metrics != nil {
			log.Printf("通知任务指标: 最后执行时间=%v, 最后执行耗时=%v, 最后错误=%v\n",
				metrics.LastRunTime, metrics.LastDuration, metrics.LastError)
		}

		// 按分组获取任务
		orderTasks := cm.GetTasksByGroup("order")
		log.Printf("订单分组任务数量: %d\n", len(orderTasks))

		// 按标签获取任务
		syncTasks := cm.GetTasksByTag("sync")
		log.Printf("同步相关任务数量: %d\n", len(syncTasks))

		// 列出所有任务
		allTasks := cm.ListTasks()
		log.Printf("总任务数量: %d\n", len(allTasks))
	}()

	// 运行60秒后优雅关闭
	time.Sleep(60 * time.Second)
	log.Println("开始优雅关闭...")

	if err := cm.GracefulStop(10 * time.Second); err != nil {
		log.Printf("优雅关闭出错: %v\n", err)
	} else {
		log.Println("Cron 管理器已优雅关闭")
	}
}
