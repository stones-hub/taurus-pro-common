package tgo

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/util/tjson"
)

// Funnel model for multi-goroutine data processing
type SpecialFunnel struct {
	id             string
	processNum     int                          // 工作协程数量
	dataChan       chan interface{}             // 数据通道
	handler        func(data interface{}) error // 数据处理函数
	handlerTimeout time.Duration                // 每条数据处理超时时间

	errorChan    chan error  // 错误通道
	errorHandler func(error) // 错误处理函数

	heartbeatCloseChan chan struct{}        // 定时器关闭通道
	heartbeat          func(*SpecialFunnel) // 心跳函数

	timeout time.Duration // 超时时间， adddata， close等非数据worker的协程超时时间

	closed atomic.Bool        // 标记漏斗是否关闭
	ctx    context.Context    // 协程上下文
	cancel context.CancelFunc // 取消函数
	wg     *sync.WaitGroup    // 等待组

	// 统计信息
	processedCount int64 // 已处理的数据条数
	goroutineCount int64 // 当前运行的协程数量
}

type FunnelConfig struct {
	Id             string                       // 漏斗实例的唯一标识符
	ProcessNum     int                          // 处理协程数量， 默认30
	Cap            int                          // 数据通道容量
	Handler        func(data interface{}) error // 数据处理函数
	HandlerTimeout time.Duration                // 每条数据处理超时时间， 默认60秒

	ErrorHandler func(error) // 错误处理函数

	Interval  int                  // 心跳检测间隔（秒）
	Heartbeat func(*SpecialFunnel) // 心跳处理函数

	Timeout time.Duration // 超时时间， adddata， close等非数据worker的协程超时时间
}

func NewSpecialFunnel(config *FunnelConfig) (*SpecialFunnel, func(), error) {
	if config.Id == "" {
		config.Id = time.Now().Format("20060102") + strconv.Itoa(rand.Intn(1000))
	}
	if config.ProcessNum <= 0 {
		config.ProcessNum = 30
	}
	if config.Cap <= 0 {
		config.Cap = 1000 // 默认通道容量
	}
	if config.HandlerTimeout <= 0 {
		config.HandlerTimeout = 60 * time.Second
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultErrorHandler
	}
	if config.Timeout <= 0 {
		config.Timeout = 3 * time.Second
	}
	if config.Interval <= 0 {
		config.Interval = 30 // 默认心跳间隔30秒
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	f := &SpecialFunnel{
		id: config.Id, // 漏斗实例的唯一标识符

		processNum:     config.ProcessNum,                  // 工作协程数量
		dataChan:       make(chan interface{}, config.Cap), // 数据通道
		handler:        config.Handler,                     // 数据处理函数
		handlerTimeout: config.HandlerTimeout,              // 每条数据处理超时时间

		errorChan:    make(chan error, config.Cap), // 错误通道
		errorHandler: config.ErrorHandler,          // 错误处理函数

		heartbeatCloseChan: make(chan struct{}), // 定时器关闭通道
		heartbeat:          config.Heartbeat,    // 心跳处理函数

		timeout: config.Timeout, // 超时时间， adddata， close等非数据worker的协程超时时间

		processedCount: 0,                 // 已处理的数据条数
		goroutineCount: 0,                 // 当前运行的协程数量
		ctx:            ctx,               // 协程上下文
		cancel:         cancel,            // 取消函数
		wg:             &sync.WaitGroup{}, // 等待组
	}

	// 启动工作协程
	f.run()
	// 启动心跳检测协程
	f.wg.Add(1)
	go f.checkHeartbeat(config.Interval)

	// 启动错误处理协程
	f.wg.Add(1)
	go f.errorWorker()

	return f, f.Close, nil
}

func (f *SpecialFunnel) run() {
	for i := 0; i < f.processNum; i++ {
		f.wg.Add(1)
		go f.worker(i)
	}
}

func (f *SpecialFunnel) worker(id int) {
	atomic.AddInt64(&f.goroutineCount, 1)
	defer func() {
		atomic.AddInt64(&f.goroutineCount, -1)
		if r := recover(); r != nil {
			log.Printf("漏斗模型[%s]: 工作协程%d 发生panic: %v", f.id, id, r)

			// 检查错误通道是否已关闭，避免向已关闭的通道写入
			if !f.closed.Load() {
				select {
				case f.errorChan <- fmt.Errorf("漏斗模型[%s]: 工作协程%d 发生panic: %v", f.id, id, r):
				default:
					// 如果错误通道已满或已关闭，则记录到日志
					log.Printf("漏斗模型[%s]: 错误通道已满或已关闭, 错误信息无法写入错误通道: %v", f.id, r)
				}
			} else {
				// 漏斗已关闭，直接记录到日志
				log.Printf("漏斗模型[%s]: 工作协程%d panic信息无法写入错误通道(漏斗已关闭): %v", f.id, id, r)
			}
		}
		// 确保wg.Done()总是被调用，即使在panic的情况下
		f.wg.Done()
	}()

	for {
		select {
		case <-f.ctx.Done():
			log.Printf("漏斗模型[%s]: 工作协程%d 检测到关闭信号, 立即退出.\n", f.id, id)
			return
		case data, ok := <-f.dataChan:
			if !ok {
				log.Printf("漏斗模型[%s]: 工作协程%d 检测到数据通道已关闭, 立即退出.\n", f.id, id)
				return
			}
			if err := f.do(data); err != nil {
				// 检查错误通道是否已关闭，避免向已关闭的通道写入
				if !f.closed.Load() {
					select {
					case f.errorChan <- err:
					default:
						// 如果错误通道已满或已关闭，则记录到日志
						log.Printf("漏斗模型[%s]: 错误通道已满或已关闭, 错误信息无法写入错误通道: %v", f.id, err)
					}
				} else {
					// 漏斗已关闭，直接记录到日志
					log.Printf("漏斗模型[%s]: 工作协程%d 错误信息无法写入错误通道(漏斗已关闭): %v", f.id, id, err)
				}
			}
		}
	}
}

func (f *SpecialFunnel) do(data interface{}) error {

	if f.handler == nil {
		return fmt.Errorf("漏斗模型[%s]: 数据处理funnel.handler为空", f.id)
	}

	// 使用漏斗的上下文创建超时上下文，使用正确的handlerTimeout
	ctx, cancel := context.WithTimeout(f.ctx, f.handlerTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		done <- f.handler(data)
	}()

	select {
	// handler处理超时
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		atomic.AddInt64(&f.processedCount, 1)
		return err
	}
}

func (f *SpecialFunnel) checkHeartbeat(interval int) {
	atomic.AddInt64(&f.goroutineCount, 1)
	defer func() {
		atomic.AddInt64(&f.goroutineCount, -1)
		if r := recover(); r != nil {
			// 检查错误通道是否已关闭，避免向已关闭的通道写入
			if !f.closed.Load() {
				select {
				case f.errorChan <- fmt.Errorf("漏斗模型[%s]: 心跳检测协程发生panic: %v", f.id, r):
				default:
					// 如果错误通道已满或已关闭，则记录到日志
					log.Printf("漏斗模型[%s]: 错误通道已满或已关闭, 错误信息无法写入错误通道: %v", f.id, r)
				}
			} else {
				// 漏斗已关闭，直接记录到日志
				log.Printf("漏斗模型[%s]: 心跳检测协程panic信息无法写入错误通道(漏斗已关闭): %v", f.id, r)
			}
		}
		// 确保wg.Done()总是被调用，即使在panic的情况下
		f.wg.Done()
	}()

	// Create a timer to check the number of processed data items every interval seconds
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if f.heartbeat != nil {
				f.heartbeat(f)
			}
		case <-f.ctx.Done():
			log.Printf("漏斗模型[%s]: 心跳检测协程检测到关闭信号, 立即退出.\n", f.id)
			return
		case <-f.heartbeatCloseChan:
			log.Printf("漏斗模型[%s]: 心跳检测定时器已关闭, 协程退出.\n", f.id)
			return
		}
	}
}

func (f *SpecialFunnel) errorWorker() {
	atomic.AddInt64(&f.goroutineCount, 1)
	defer func() {
		atomic.AddInt64(&f.goroutineCount, -1)
		if r := recover(); r != nil {
			log.Printf("漏斗模型[%s]: 错误处理协程发生panic: %v", f.id, r)
		}
		// 确保wg.Done()总是被调用，即使在panic的情况下
		f.wg.Done()
	}()

	for {
		select {
		case <-f.ctx.Done():
			log.Printf("漏斗模型[%s]: 错误处理协程检测到关闭信号, 立即退出.\n", f.id)
			return
		case err, ok := <-f.errorChan:
			if !ok {
				log.Printf("漏斗模型[%s]: 错误处理协程检测到错误通道已关闭, 立即退出.\n", f.id)
				return
			}

			if err := f.errorDo(err); err != nil {
				log.Printf("漏斗模型[%s]: 错误处理协程处理错误失败: %v", f.id, err)
			}
		}
	}
}

func (f *SpecialFunnel) errorDo(err error) error {
	// 使用漏斗的context，确保漏斗关闭时错误处理协程也能响应
	ctx, cancel := context.WithTimeout(f.ctx, f.timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		defer close(done)
		f.errorHandler(err)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (f *SpecialFunnel) AddData(data interface{}) {
	if f.closed.Load() {
		log.Printf("漏斗模型[%s]: 漏斗已关闭, 无法添加数据.\n", f.id)
		return
	}

	// 使用漏斗的context，确保漏斗关闭时AddData也能响应
	ctx, cancel := context.WithTimeout(f.ctx, f.timeout)
	defer cancel()

	select {
	case f.dataChan <- data:
	case <-ctx.Done():
		select {
		case f.errorChan <- fmt.Errorf("漏斗模型[%s]: 数据通道已满, 丢弃数据: %s", f.id, tjson.ToJSONString(data)):
		default:
			// 如果错误通道已满或已关闭，则记录到日志
			log.Printf("漏斗模型[%s]: 错误通道已满或已关闭, 错误信息无法写入错误通道: %v", f.id, data)
		}
	}
}

func (f *SpecialFunnel) GetProcessedCount() int64 {
	return atomic.LoadInt64(&f.processedCount)
}

func (f *SpecialFunnel) GetGoroutineCount() int64 {
	return atomic.LoadInt64(&f.goroutineCount)
}

func (f *SpecialFunnel) Close() {
	if !f.closed.CompareAndSwap(false, true) {
		log.Printf("漏斗模型[%s]: 漏斗已关闭, 请勿重复调用.\n", f.id)
		return
	}

	// 1. 先关闭数据通道，阻止新数据进入
	// 当数据通道关闭后，所有 worker 会自然处理完剩余数据并退出
	close(f.dataChan)
	// 2. 关闭错误通道，防止协程在写入错误时阻塞
	// 注意：错误通道要在心跳通道之前关闭，因为心跳协程会写入错误通道
	close(f.errorChan)
	// 3. 关闭定时器通道
	close(f.heartbeatCloseChan)

	// 4. 等待所有 worker 处理完剩余数据并退出
	done := make(chan struct{})
	go func() {
		defer close(done)
		f.wg.Wait()
	}()

	// 使用context控制超时，避免time.After泄漏
	waitCtx, waitCancel := context.WithTimeout(context.Background(), f.timeout)
	defer waitCancel()

	// 超时等待
	select {
	case <-done:
		log.Printf("漏斗模型[%s]: 漏斗已关闭, 所有协程已退出.\n", f.id)
	case <-waitCtx.Done():
		log.Printf("漏斗模型[%s]: 漏斗关闭超时, 开始强制关闭.\n", f.id)
		// 超时后直接取消上下文，所有协程会立即收到信号并退出
		f.cancel()

		// 再等待一段时间让协程响应取消信号并完成清理
		time.Sleep(f.timeout)

		// 检查是否还有协程在运行
		runningCount := atomic.LoadInt64(&f.goroutineCount)
		if runningCount > 0 {
			log.Printf("漏斗模型[%s]: 漏斗关闭超时, 仍有 %d 个协程未退出, 程序将继续, 请检查是否存在死循环或卡死.\n", f.id, runningCount)
		}
	}
}

func defaultErrorHandler(err error) {
	// 默认错误处理函数， 什么都不做
	log.Printf("%s", err.Error())
}

// IsClosed 检查漏斗是否已关闭
func (f *SpecialFunnel) IsClosed() bool {
	return f.closed.Load()
}

// GetID 获取漏斗实例ID
func (f *SpecialFunnel) GetID() string {
	return f.id
}

// GetProcessNum 获取工作协程数量
func (f *SpecialFunnel) GetProcessNum() int {
	return f.processNum
}

// GetCap 获取数据通道容量
func (f *SpecialFunnel) GetCap() int {
	return cap(f.dataChan)
}

// GetPendingDataCount 获取待处理数据数量（近似值）
func (f *SpecialFunnel) GetPendingDataCount() int {
	return len(f.dataChan)
}

// GetPendingErrorCount 获取待处理错误数量（近似值）
func (f *SpecialFunnel) GetPendingErrorCount() int {
	return len(f.errorChan)
}

// GetHandlerTimeout 获取数据处理超时时间
func (f *SpecialFunnel) GetHandlerTimeout() time.Duration {
	return f.handlerTimeout
}

// GetTimeout 获取通用超时时间
func (f *SpecialFunnel) GetTimeout() time.Duration {
	return f.timeout
}
