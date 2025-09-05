package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/recovery"
)

// HTTP处理器包装器
type SafeHandler struct {
	recovery *recovery.PanicRecovery
	handler  http.HandlerFunc
}

func (h *SafeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 使用SafeGo包装HTTP处理器
	h.recovery.SafeGo("http-handler", func() {
		h.handler(w, r)
	})
}

// 用户API处理器
func userHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	// 模拟业务逻辑
	if userID == "error_user" {
		panic("用户不存在")
	}

	// 模拟数据库查询
	time.Sleep(100 * time.Millisecond)

	response := map[string]interface{}{
		"user_id": userID,
		"name":    "张三",
		"email":   "zhangsan@example.com",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 订单API处理器
func orderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("id")
	if orderID == "" {
		http.Error(w, "Missing order ID", http.StatusBadRequest)
		return
	}

	// 模拟业务逻辑
	if orderID == "invalid_order" {
		panic("订单无效")
	}

	// 模拟订单处理
	time.Sleep(200 * time.Millisecond)

	response := map[string]interface{}{
		"order_id": orderID,
		"status":   "processing",
		"amount":   100.50,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 支付API处理器（带context）
func paymentHandler(w http.ResponseWriter, r *http.Request) {
	paymentID := r.URL.Query().Get("id")
	if paymentID == "" {
		http.Error(w, "Missing payment ID", http.StatusBadRequest)
		return
	}

	// 创建带超时的context
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// 使用SafeGo启动协程，内部处理context管理
	done := make(chan bool, 1)

	recovery.GlobalPanicRecovery.SafeGo("payment-api", func() {
		// 模拟支付处理
		if paymentID == "failed_payment" {
			panic("支付失败")
		}

		// 模拟长时间处理
		select {
		case <-time.After(2 * time.Second):
			response := map[string]interface{}{
				"payment_id": paymentID,
				"status":     "success",
				"amount":     100.50,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			done <- true
		case <-ctx.Done():
			http.Error(w, "Payment timeout", http.StatusRequestTimeout)
			done <- true
		}
	})

	// 等待处理完成
	select {
	case <-done:
		// 处理完成
	case <-time.After(4 * time.Second):
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
	}
}

// 健康检查处理器
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// WrapFunction功能展示处理器
func wrapFunctionHandler(w http.ResponseWriter, r *http.Request) {
	// 使用WrapFunction包装函数
	safeFunc := recovery.GlobalPanicRecovery.WrapFunction("wrap-api", func() {
		log.Println("执行包装的API函数...")
		time.Sleep(100 * time.Millisecond)

		// 模拟偶尔的panic
		if time.Now().Second()%2 == 0 {
			panic("包装函数中的随机panic")
		}

		log.Println("包装的API函数执行完成")
	})

	// 使用WrapErrorFunction包装返回错误的函数
	safeErrorFunc := recovery.GlobalPanicRecovery.WrapErrorFunction("wrap-error-api", func() error {
		log.Println("执行包装的错误API函数...")
		time.Sleep(100 * time.Millisecond)

		// 模拟偶尔的panic
		if time.Now().Second()%3 == 0 {
			panic("包装错误函数中的随机panic")
		}

		return nil
	})

	// 执行包装的函数
	go safeFunc()

	go func() {
		if err := safeErrorFunc(); err != nil {
			log.Printf("包装错误函数返回错误: %v", err)
		}
	}()

	response := map[string]interface{}{
		"message": "WrapFunction功能演示已启动",
		"time":    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 回调功能展示处理器
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	// 使用SafeGoWithCallback展示回调功能
	recovery.GlobalPanicRecovery.SafeGoWithCallback("callback-api", func() {
		log.Println("执行带回调的API函数...")
		time.Sleep(200 * time.Millisecond)

		// 模拟偶尔的panic
		if time.Now().Second()%2 == 0 {
			panic("回调函数中的随机panic")
		}

		log.Println("带回调的API函数执行完成")
	}, func() {
		log.Println("回调函数执行 - 无论是否发生panic都会执行")
	})

	// 使用WrapFunctionWithCallback展示包装回调功能
	safeFuncWithCallback := recovery.GlobalPanicRecovery.WrapFunctionWithCallback("wrap-callback-api", func() {
		log.Println("执行包装的带回调API函数...")
		time.Sleep(200 * time.Millisecond)

		// 模拟偶尔的panic
		if time.Now().Second()%3 == 0 {
			panic("包装回调函数中的随机panic")
		}

		log.Println("包装的带回调API函数执行完成")
	}, func() {
		log.Println("包装回调函数执行 - 无论是否发生panic都会执行")
	})

	// 执行包装的函数
	go safeFuncWithCallback()

	response := map[string]interface{}{
		"message": "回调功能演示已启动",
		"time":    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 模拟后台任务
func startBackgroundTasks() {
	// 启动定时清理任务
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			recovery.GlobalPanicRecovery.SafeGo("background-cleanup", func() {
				log.Println("执行后台清理任务...")
				time.Sleep(1 * time.Second)
				log.Println("后台清理任务完成")
			})
		}
	}()

	// 启动数据同步任务
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			recovery.GlobalPanicRecovery.SafeGo("data-sync", func() {
				log.Println("执行数据同步任务...")
				time.Sleep(2 * time.Second)
				log.Println("数据同步任务完成")
			})
		}
	}()
}

// 启动Web服务器
func startWebServer() {
	// 设置路由
	http.Handle("/api/user", &SafeHandler{
		recovery: recovery.GlobalPanicRecovery,
		handler:  userHandler,
	})

	http.Handle("/api/order", &SafeHandler{
		recovery: recovery.GlobalPanicRecovery,
		handler:  orderHandler,
	})

	http.Handle("/api/payment", &SafeHandler{
		recovery: recovery.GlobalPanicRecovery,
		handler:  paymentHandler,
	})

	// 添加更多API端点展示不同功能
	http.Handle("/api/wrap", &SafeHandler{
		recovery: recovery.GlobalPanicRecovery,
		handler:  wrapFunctionHandler,
	})

	http.Handle("/api/callback", &SafeHandler{
		recovery: recovery.GlobalPanicRecovery,
		handler:  callbackHandler,
	})

	http.HandleFunc("/health", healthHandler)

	// 启动后台任务
	startBackgroundTasks()

	// 启动服务器
	log.Println("🌐 启动Web服务器在端口 8080...")
	log.Println("📋 可用的API端点:")
	log.Println("  GET /api/user?id=xxx - 用户信息")
	log.Println("  GET /api/order?id=xxx - 订单信息")
	log.Println("  GET /api/payment?id=xxx - 支付信息")
	log.Println("  GET /api/wrap - 展示WrapFunction功能")
	log.Println("  GET /api/callback - 展示回调功能")
	log.Println("  GET /health - 健康检查")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	log.Println("🚀 启动Web应用程序...")

	// 初始化全局recovery机制
	initGlobalRecovery()

	// 启动Web服务器
	startWebServer()
}

// 初始化全局recovery机制
func initGlobalRecovery() {
	log.Println("🔧 初始化全局recovery机制...")

	// 创建自定义配置
	options := &recovery.RecoveryOptions{
		EnableStackTrace: true,
		MaxHandlers:      10,
		HandlerTimeout:   5 * time.Second,
	}

	// 创建recovery实例
	customRecovery := recovery.NewPanicRecovery(options)

	// 添加自定义处理器
	customHandler := &CustomPanicHandler{name: "Web处理器"}
	if err := customRecovery.AddHandler(customHandler); err != nil {
		log.Printf("❌ 添加Web处理器失败: %v", err)
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
