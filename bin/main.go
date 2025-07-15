package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/cron"
	"github.com/stones-hub/taurus-pro-common/pkg/ctx"
	"github.com/stones-hub/taurus-pro-common/pkg/logx"
	"github.com/stones-hub/taurus-pro-common/pkg/templates"
	"github.com/stones-hub/taurus-pro-common/pkg/util"
)

func main() {
	// 运行日志示例
	loggerExample()

	// 运行模板示例
	templateExample()

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

	// 使用工具包渲染表格（终端格式）
	util.RenderTable(headers, data, "terminal")

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
	tc, err := ctx.GetTaurusContext(taurusCtx)
	if err != nil {
		log.Fatalf("获取 TaurusContext 失败: %v", err)
	}

	// 测试设置和获取数据
	tc.Set("user_id", "user-001")
	tc.Set("role", "admin")
	tc.Set("login_time", time.Now())

	// 打印基本信息
	fmt.Printf("请求ID: %s\n", tc.GetRequestID())
	fmt.Printf("创建时间: %v\n", tc.AtTime)

	// 获取并打印存储的数据
	fmt.Printf("\n存储的数据:\n")
	if val, err := tc.Get("user_id"); err == nil {
		fmt.Printf("用户ID: %v\n", val)
	}
	if val, err := tc.Get("role"); err == nil {
		fmt.Printf("角色: %v\n", val)
	}
	if val, err := tc.Get("login_time"); err == nil {
		fmt.Printf("登录时间: %v\n", val)
	}

	// 测试不存在的键
	fmt.Printf("\n获取不存在的键:\n")
	if val, err := tc.Get("not_exist"); err != nil {
		fmt.Printf("不存在的键: %v\n", err)
	} else {
		fmt.Printf("不存在的键的值: %v\n", val)
	}

	// 测试存储不同类型的数据
	tc.Set("int_value", 42)
	tc.Set("float_value", 3.14)
	tc.Set("bool_value", true)
	tc.Set("slice_value", []string{"a", "b", "c"})
	tc.Set("map_value", map[string]int{"one": 1, "two": 2})

	fmt.Printf("\n不同类型的数据:\n")
	if val, err := tc.Get("int_value"); err == nil {
		fmt.Printf("整数: %v\n", val)
	}
	if val, err := tc.Get("float_value"); err == nil {
		fmt.Printf("浮点数: %v\n", val)
	}
	if val, err := tc.Get("bool_value"); err == nil {
		fmt.Printf("布尔值: %v\n", val)
	}
	if val, err := tc.Get("slice_value"); err == nil {
		fmt.Printf("切片: %v\n", val)
	}
	if val, err := tc.Get("map_value"); err == nil {
		fmt.Printf("映射: %v\n", val)
	}
}

func cronExample() {
	// 加载上海时区
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Printf("加载时区失败: %v, 将使用系统默认时区\n", err)
		location = time.Local
	}

	// 创建 cron 管理器，启用秒级调度
	cm := cron.New(
		cron.WithSeconds(),                           // 启用秒级调度
		cron.WithLocation(location),                  // 设置时区
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

func loggerExample() {
	fmt.Println("\n=== Logger 测试 ===")

	// 1. 准备日志配置
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	// 2. 创建日志管理器
	manager, cleanup, err := logx.BuildManager(
		// 控制台日志记录器
		logx.LoggerOptions{
			Name:   "console",
			Output: logx.Console,
			Level:  logx.Debug,
		},
		// 主文件日志记录器
		logx.LoggerOptions{
			Name:       "file",
			Output:     logx.File,
			Level:      logx.Info,
			FilePath:   filepath.Join(logDir, "app.log"),
			MaxSize:    10,   // 10MB
			MaxBackups: 5,    // 保留5个备份
			MaxAge:     30,   // 保留30天
			Compress:   true, // 压缩旧日志
		},
		// 用户模块日志记录器
		logx.LoggerOptions{
			Name:       "user",
			Output:     logx.File,
			Level:      logx.Info,
			FilePath:   filepath.Join(logDir, "user.log"),
			MaxSize:    5,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
		// 订单模块日志记录器
		logx.LoggerOptions{
			Name:       "order",
			Output:     logx.File,
			Level:      logx.Info,
			FilePath:   filepath.Join(logDir, "order.log"),
			MaxSize:    5,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
		// 支付模块日志记录器
		logx.LoggerOptions{
			Name:       "payment",
			Output:     logx.File,
			Level:      logx.Info,
			FilePath:   filepath.Join(logDir, "payment.log"),
			MaxSize:    5,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   true,
		},
	)
	if err != nil {
		log.Fatalf("创建日志管理器失败: %v", err)
	}
	defer cleanup()

	// 3. 演示通过管理器使用日志记录器
	fmt.Println("通过管理器使用日志记录器:")

	// 使用控制台日志记录器
	manager.LInfo("console", "通过管理器获取的控制台日志记录器")
	manager.LDebug("console", "这是一条调试日志")

	// 使用文件日志记录器
	manager.LInfo("file", "通过管理器获取的文件日志记录器")
	manager.LWarn("file", "这是一条警告日志")

	// 4. 演示业务模块日志记录
	fmt.Println("\n业务模块日志记录:")

	// 用户模块日志
	manager.LInfo("user", "用户 %s 注册成功", "user123")
	manager.LWarn("user", "用户 %s 登录失败次数过多", "user456")

	// 订单模块日志
	manager.LInfo("order", "订单 %s 创建成功", "order789")
	manager.LError("order", "订单 %s 支付超时", "order012")

	// 支付模块日志
	manager.LInfo("payment", "支付交易 %s 完成", "pay345")
	manager.LError("payment", "支付交易 %s 失败: %v", "pay678", errors.New("余额不足"))

	// 5. 演示结构化数据记录
	fmt.Println("\n记录结构化数据:")

	// 用户注册事件
	manager.LInfo("user", "新用户注册: %+v", map[string]interface{}{
		"user_id":    "u123",
		"username":   "张三",
		"email":      "zhangsan@example.com",
		"age":        25,
		"created_at": time.Now(),
	})

	// 订单创建事件
	manager.LInfo("order", "订单创建: %+v", map[string]interface{}{
		"order_id":   "o456",
		"user_id":    "u123",
		"amount":     99.9,
		"status":     "pending",
		"created_at": time.Now(),
		"items": []map[string]interface{}{
			{
				"id":       "item1",
				"name":     "商品1",
				"price":    49.9,
				"quantity": 1,
			},
			{
				"id":       "item2",
				"name":     "商品2",
				"price":    50.0,
				"quantity": 1,
			},
		},
	})

	// 支付事件
	manager.LInfo("payment", "支付处理: %+v", map[string]interface{}{
		"payment_id": "p789",
		"order_id":   "o456",
		"user_id":    "u123",
		"amount":     99.9,
		"method":     "alipay",
		"status":     "success",
		"created_at": time.Now(),
		"metadata": map[string]string{
			"transaction_id": "ali_123456",
			"channel":        "mobile",
		},
	})

	// 6. 演示错误处理和恢复
	fmt.Println("\n错误处理和恢复:")

	// 使用不存在的日志记录器（会使用默认日志记录器）
	manager.LInfo("not_exists", "这条日志会使用默认日志记录器")

	// 使用 recover 中间件记录 panic
	defer func() {
		if r := recover(); r != nil {
			// 在 defer 函数中，我们需要减少一层调用栈，因为 defer 本身会增加一层
			manager.LWithSkip("file", logx.Error, 0, "捕获到 panic: %v", r)
		}
	}()

	// 模拟一个会导致 panic 的操作
	//var slice []string
	//fmt.Println(slice[100]) // 这会导致 panic
}

func templateExample() {
	fmt.Println("\n=== Template 测试 ===")

	// 1. 创建模板管理器
	manager, cleanup, err := templates.New(
		templates.TemplateOptions{
			Name: "main",
			Path: "bin/templates/main", // 放在 bin 目录下
		},
		templates.TemplateOptions{
			Name: "email",
			Path: "bin/templates/email", // 放在 bin 目录下
		},
	)
	if err != nil {
		log.Fatalf("创建模板管理器失败: %v", err)
	}
	defer cleanup()

	// 2. 动态添加基础布局模板
	err = manager.AddTemplate("dynamic", "layout", `
		{{define "layout"}}
		<!DOCTYPE html>
		<html>
		<head>
			<title>{{.Title}}</title>
			<style>
				.container { padding: 20px; }
				.header { background: #f0f0f0; padding: 10px; }
				.content { margin: 20px 0; }
				.footer { text-align: center; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">{{template "header" .}}</div>
				<div class="content">{{template "content" .}}</div>
				<div class="footer">{{template "footer" .}}</div>
			</div>
		</body>
		</html>
		{{end}}
	`)
	if err != nil {
		log.Fatalf("添加布局模板失败: %v", err)
	}

	// 3. 动态添加页面组件
	components := map[string]string{
		"header": `
			{{define "header"}}
			<h1>{{.Title}}</h1>
			<nav>
				{{range .NavItems}}
				<a href="{{.URL}}">{{.Text}}</a>
				{{end}}
			</nav>
			{{end}}
		`,
		"content": `
			{{define "content"}}
			<article>
				<h2>{{.Subtitle}}</h2>
				<p>{{.Content}}</p>
				{{if .ShowFeatures}}
				<ul>
					{{range .Features}}
					<li>{{.}}</li>
					{{end}}
				</ul>
				{{end}}
			</article>
			{{end}}
		`,
		"footer": `
			{{define "footer"}}
			<footer>
				<p>{{.Copyright}}</p>
				{{if .ShowContact}}
				<p>联系我们: {{.Contact}}</p>
				{{end}}
			</footer>
			{{end}}
		`,
	}

	for name, content := range components {
		if err := manager.AddTemplate("dynamic", name, content); err != nil {
			log.Fatalf("添加组件模板 %s 失败: %v", name, err)
		}
	}

	// 4. 准备渲染数据
	data := map[string]interface{}{
		"Title":    "欢迎使用 Taurus",
		"Subtitle": "功能特性",
		"Content":  "Taurus 是一个强大的 Go 工具库，提供了丰富的功能。",
		"NavItems": []struct {
			URL  string
			Text string
		}{
			{"/home", "首页"},
			{"/docs", "文档"},
			{"/about", "关于"},
		},
		"ShowFeatures": true,
		"Features": []string{
			"日志管理",
			"模板引擎",
			"定时任务",
			"工具集合",
		},
		"Copyright":   "© 2025 Taurus Team",
		"ShowContact": true,
		"Contact":     "contact@taurus.com",
	}

	// 5. 渲染完整页面
	result, err := manager.Render("dynamic", "layout", data)
	if err != nil {
		log.Fatalf("渲染页面失败: %v", err)
	}

	// 6. 保存渲染结果
	outputFile := "bin/templates/output.html"
	if err := os.WriteFile(outputFile, []byte(result), 0644); err != nil {
		log.Fatalf("保存渲染结果失败: %v", err)
	}
	fmt.Printf("页面已渲染并保存到 %s\n", outputFile)

	// 7. 动态添加邮件模板
	emailTemplate := `
		{{define "notification"}}
		<div style="max-width: 600px; margin: 0 auto;">
			<h2>{{.Subject}}</h2>
			<p>亲爱的 {{.Username}}，</p>
			<p>{{.Message}}</p>
			{{if .ShowButton}}
			<div style="text-align: center;">
				<a href="{{.ButtonURL}}" style="display: inline-block; padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 5px;">
					{{.ButtonText}}
				</a>
			</div>
			{{end}}
			<p>祝好，<br>{{.Signature}}</p>
		</div>
		{{end}}
	`

	if err := manager.AddTemplate("email", "notification", emailTemplate); err != nil {
		log.Fatalf("添加邮件模板失败: %v", err)
	}

	// 8. 渲染邮件内容
	emailData := map[string]interface{}{
		"Subject":    "欢迎加入 Taurus",
		"Username":   "张三",
		"Message":    "感谢您注册 Taurus 账号。请点击下面的按钮验证您的邮箱地址。",
		"ShowButton": true,
		"ButtonURL":  "https://taurus.com/verify?token=abc123",
		"ButtonText": "验证邮箱",
		"Signature":  "Taurus 团队",
	}

	emailContent, err := manager.Render("email", "notification", emailData)
	if err != nil {
		log.Fatalf("渲染邮件失败: %v", err)
	}

	// 保存邮件内容
	if err := os.WriteFile("bin/templates/email.html", []byte(emailContent), 0644); err != nil {
		log.Fatalf("保存邮件内容失败: %v", err)
	}
	fmt.Printf("邮件内容已保存到 email.html\n")
}
