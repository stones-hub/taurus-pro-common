package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/recovery"
)

// 自定义的panic处理器
type CustomPanicHandler struct {
	name string
}

func (h *CustomPanicHandler) HandlePanic(info *recovery.PanicInfo) error {
	log.Printf("🔧 [%s] 捕获到panic: Component=%s, Error=%v, Time=%s",
		h.name, info.Component, info.Error, info.Timestamp.Format("2006-01-02 15:04:05"))

	// 这里可以添加自定义的处理逻辑，比如：
	// 1. 发送告警通知
	// 2. 记录到监控系统
	// 3. 触发自动重试
	// 4. 记录到专门的错误日志文件

	return nil
}

// 模拟的业务组件
type UserService struct {
	recovery *recovery.PanicRecovery
}

func NewUserService(recovery *recovery.PanicRecovery) *UserService {
	return &UserService{
		recovery: recovery,
	}
}

func (s *UserService) GetUserInfo(userID string) {
	// 使用SafeGo启动协程，如果发生panic会被框架捕获
	s.recovery.SafeGo("user-service", func() {
		log.Printf("正在获取用户信息: %s", userID)

		// 模拟业务逻辑
		if userID == "error_user" {
			panic("用户不存在")
		}

		// 模拟正常业务逻辑
		time.Sleep(100 * time.Millisecond)
		log.Printf("成功获取用户信息: %s", userID)
	})
}

func (s *UserService) GetUserInfoWithContext(userID string) {
	// 使用SafeGo启动协程，内部处理context管理
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	s.recovery.SafeGo("user-service-context", func() {
		log.Printf("正在获取用户信息(带context): %s", userID)

		// 检查context是否被取消
		select {
		case <-ctx.Done():
			log.Printf("获取用户信息被取消: %s", userID)
			return
		default:
			// 继续执行
		}

		// 模拟业务逻辑
		if userID == "error_user" {
			panic("用户不存在(带context)")
		}

		// 模拟长时间运行的任务
		select {
		case <-time.After(3 * time.Second):
			log.Printf("成功获取用户信息(带context): %s", userID)
		case <-ctx.Done():
			log.Printf("获取用户信息超时: %s", userID)
		}
	})
}

// 模拟的订单服务
type OrderService struct {
	recovery *recovery.PanicRecovery
}

func NewOrderService(recovery *recovery.PanicRecovery) *OrderService {
	return &OrderService{
		recovery: recovery,
	}
}

func (s *OrderService) ProcessOrder(orderID string) {
	// 使用SafeGoWithCallback启动协程，支持回调函数
	s.recovery.SafeGoWithCallback("order-service", func() {
		log.Printf("正在处理订单: %s", orderID)

		// 模拟业务逻辑
		if orderID == "invalid_order" {
			panic("订单无效")
		}

		// 模拟正常业务逻辑
		time.Sleep(200 * time.Millisecond)
		log.Printf("订单处理完成: %s", orderID)
	}, func() {
		// 回调函数，无论是否发生panic都会执行
		log.Printf("订单处理回调: %s", orderID)
	})
}

// 模拟的支付服务
type PaymentService struct {
	recovery *recovery.PanicRecovery
}

func NewPaymentService(recovery *recovery.PanicRecovery) *PaymentService {
	return &PaymentService{
		recovery: recovery,
	}
}

func (s *PaymentService) ProcessPayment(paymentID string) {
	// 使用SafeGo启动协程，内部处理context管理
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.recovery.SafeGo("payment-service", func() {
		log.Printf("正在处理支付: %s", paymentID)

		// 启动子协程处理支付验证
		go func() {
			select {
			case <-time.After(1 * time.Second):
				log.Printf("支付验证完成: %s", paymentID)
			case <-ctx.Done():
				log.Printf("支付验证被取消: %s", paymentID)
			}
		}()

		// 主协程处理支付逻辑
		if paymentID == "failed_payment" {
			panic("支付失败")
		}

		// 模拟支付处理
		select {
		case <-time.After(2 * time.Second):
			log.Printf("支付处理完成: %s", paymentID)
		case <-ctx.Done():
			log.Printf("支付处理超时: %s", paymentID)
		}
	})
}

// 模拟的定时任务
type ScheduledTask struct {
	recovery *recovery.PanicRecovery
}

func NewScheduledTask(recovery *recovery.PanicRecovery) *ScheduledTask {
	return &ScheduledTask{
		recovery: recovery,
	}
}

func (s *ScheduledTask) StartCleanupTask() {
	// 模拟定时清理任务
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.recovery.SafeGo("cleanup-task", func() {
			log.Printf("开始执行清理任务")

			// 模拟清理逻辑
			time.Sleep(1 * time.Second)

			// 模拟偶尔出现的错误
			if time.Now().Second()%30 == 0 {
				panic("清理任务出错")
			}

			log.Printf("清理任务完成")
		})
	}
}

func main() {
	log.Println("🚀 启动应用程序...")

	// 1. 初始化全局recovery机制
	initGlobalRecovery()

	// 2. 创建业务服务
	userService := NewUserService(recovery.GlobalPanicRecovery)
	orderService := NewOrderService(recovery.GlobalPanicRecovery)
	paymentService := NewPaymentService(recovery.GlobalPanicRecovery)
	scheduledTask := NewScheduledTask(recovery.GlobalPanicRecovery)

	// 3. 启动定时任务
	go scheduledTask.StartCleanupTask()

	// 4. 模拟业务请求
	log.Println("📝 开始模拟业务请求...")

	// 正常请求
	userService.GetUserInfo("user123")
	orderService.ProcessOrder("order123")
	paymentService.ProcessPayment("payment123")

	// 带context的请求
	userService.GetUserInfoWithContext("user456")

	// 会出错的请求
	userService.GetUserInfo("error_user")
	orderService.ProcessOrder("invalid_order")
	paymentService.ProcessPayment("failed_payment")

	// 5. 展示WrapFunction功能
	log.Println("🔧 展示WrapFunction功能...")
	demonstrateWrapFunctions()

	// 6. 保持程序运行
	log.Println("✅ 应用程序启动完成，等待业务请求...")
	log.Println("💡 观察日志输出，可以看到panic被框架捕获而不会导致程序退出")

	// 保持程序运行
	select {}
}

// 展示WrapFunction功能
func demonstrateWrapFunctions() {
	// 使用WrapFunction包装普通函数
	safeFunc := recovery.GlobalPanicRecovery.WrapFunction("wrapped-function", func() {
		log.Println("执行包装的函数...")
		time.Sleep(100 * time.Millisecond)
		log.Println("包装的函数执行完成")
	})

	// 使用WrapFunctionWithCallback包装带回调的函数
	safeFuncWithCallback := recovery.GlobalPanicRecovery.WrapFunctionWithCallback("wrapped-function-with-callback", func() {
		log.Println("执行带回调的包装函数...")
		time.Sleep(100 * time.Millisecond)
		panic("包装函数中的panic")
	}, func() {
		log.Println("回调函数执行")
	})

	// 使用WrapErrorFunction包装返回错误的函数
	safeErrorFunc := recovery.GlobalPanicRecovery.WrapErrorFunction("wrapped-error-function", func() error {
		log.Println("执行包装的错误函数...")
		time.Sleep(100 * time.Millisecond)
		panic("包装错误函数中的panic")
		// return nil // 这行代码永远不会执行，因为上面会panic
	})

	// 使用WrapErrorFunctionWithCallback包装带回调的错误函数
	safeErrorFuncWithCallback := recovery.GlobalPanicRecovery.WrapErrorFunctionWithCallback("wrapped-error-function-with-callback", func() error {
		log.Println("执行带回调的包装错误函数...")
		time.Sleep(100 * time.Millisecond)
		return fmt.Errorf("模拟错误")
	}, func() {
		log.Println("错误函数回调执行")
	})

	// 执行包装的函数
	go safeFunc()
	go safeFuncWithCallback()
	go func() {
		if err := safeErrorFunc(); err != nil {
			log.Printf("包装错误函数返回错误: %v", err)
		}
	}()
	go func() {
		if err := safeErrorFuncWithCallback(); err != nil {
			log.Printf("带回调的包装错误函数返回错误: %v", err)
		}
	}()
}

// 初始化全局recovery机制
func initGlobalRecovery() {
	log.Println("🔧 初始化全局recovery机制...")

	// 创建自定义配置
	options := &recovery.RecoveryOptions{
		EnableStackTrace: true,
		MaxHandlers:      10,
		HandlerTimeout:   3 * time.Second,
	}

	// 创建recovery实例
	customRecovery := recovery.NewPanicRecovery(options)

	// 添加自定义处理器
	customHandler := &CustomPanicHandler{name: "业务处理器"}
	if err := customRecovery.AddHandler(customHandler); err != nil {
		log.Printf("❌ 添加自定义处理器失败: %v", err)
	}

	// 添加告警处理器
	alertHandler := &CustomPanicHandler{name: "告警处理器"}
	if err := customRecovery.AddHandler(alertHandler); err != nil {
		log.Printf("❌ 添加告警处理器失败: %v", err)
	}

	// 替换全局recovery实例
	recovery.GlobalPanicRecovery = customRecovery

	log.Println("✅ 全局recovery机制初始化完成")
}
