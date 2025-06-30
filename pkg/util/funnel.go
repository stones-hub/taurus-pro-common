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

package util

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// Funnel model for multi-goroutine data processing

const (
	SF_PROCESS_NUM = 30 // Default number of processing goroutines
)

// SpecialFunnel represents a funnel model structure
// Used for concurrent data processing with multiple goroutines, supports graceful shutdown and heartbeat detection
type SpecialFunnel struct {
	id              string
	closeChan       chan struct{}
	dataChan        chan interface{}
	wg              *sync.WaitGroup
	handler         func(data interface{})
	tickerCloseChan chan struct{}        // 定时器关闭通道
	processedCount  int64                // 已处理的数据条数
	closed          atomic.Bool          // 标记漏斗是否关闭
	heartbeat       func(*SpecialFunnel) // 心跳函数
}

// FunnelConfig represents the funnel configuration structure
type FunnelConfig struct {
	Cap       int                    // 数据通道容量
	Interval  int                    // 心跳检测间隔（秒）
	Handler   func(data interface{}) // 数据处理函数
	Heartbeat func(*SpecialFunnel)   // 心跳处理函数
}

// NewSpecialFunnel creates a new funnel instance
// Parameters:
//   - config: Funnel configuration
//
// Returns:
//   - *SpecialFunnel: Funnel instance
//   - func(): Close function
//   - error: Error information
func NewSpecialFunnel(config *FunnelConfig) (*SpecialFunnel, func(), error) {
	f := &SpecialFunnel{
		// Generate unique ID
		id:              uuid.NewString(),
		closeChan:       make(chan struct{}),
		dataChan:        make(chan interface{}, config.Cap),
		wg:              &sync.WaitGroup{},
		handler:         config.Handler,
		tickerCloseChan: make(chan struct{}),
		processedCount:  0,
		heartbeat:       config.Heartbeat,
	}
	f.run()
	f.checkHeartbeat(config.Interval)
	return f, f.Close, nil
}

// run starts the processing goroutines
func (f *SpecialFunnel) run() {
	for i := 0; i < SF_PROCESS_NUM; i++ {
		f.wg.Add(1)
		go f.worker()
	}
}

// worker is the working goroutine
// Responsible for retrieving and processing data from the data channel
func (f *SpecialFunnel) worker() {
	defer f.wg.Done()
	for {
		select {
		case data, ok := <-f.dataChan:
			if !ok {
				log.Printf("SpecialFunnel[%s] worker channel closed, goroutine will exit.\n", f.id)
				return
			}
			f.do(data)
		case <-f.closeChan:
			// 检查下dataChan是否还有数据, 如果有数据，就继续处理，知道处理完
			for {
				// 当你调用 close(f.dataChan) 关闭通道时，即便通道里还有数据，这些数据也不会丢失，协程仍然能够把通道里剩余的数据遍历完。
				// 1. 通道一旦关闭，就不能再向其发送数据，不过可以继续从通道接收数据，直到通道里的数据被全部接收完。
				// 2. 从已关闭的通道接收数据时，若通道里还有数据，接收操作会正常返回数据和 ok 为 true；若通道里的数据已全部接收完，接收操作会返回通道元素类型的零值和 ok 为 false。
				data, ok := <-f.dataChan
				if !ok {
					return
				}
				f.do(data)
			}
		}
	}
}

// do processes a single piece of data
// Parameters:
//   - data: Data to be processed
func (f *SpecialFunnel) do(data interface{}) {
	if f.handler == nil {
		log.Printf("SpeialFunnel[%s] handler is empty, data not processed: %v", f.id, data)
		return
	}
	// The sync/atomic package provides functions like AddInt64, LoadInt64, etc.
	// Atomic counting, goroutine-safe
	atomic.AddInt64(&f.processedCount, 1)
	// 有可能handler是阻塞的，但是不可以用协程，避免无休止的开协程
	f.handler(data)
}

// checkHeartbeat starts the heartbeat detection
// Parameters:
//   - interval: Check interval (seconds)
func (f *SpecialFunnel) checkHeartbeat(interval int) {
	go func() {
		// Create a timer to check the number of processed data items every interval seconds
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if f.heartbeat != nil {
					f.heartbeat(f)
				}
			case <-f.tickerCloseChan:
				log.Printf("SpecialFunnel[%s] timer closed, goroutine exiting.\n", f.id)
				return
			}
		}
	}()
}

// Close shuts down the funnel
// Ensures all data is processed before closing
func (f *SpecialFunnel) Close() {
	if !f.closed.CompareAndSwap(false, true) {
		log.Printf("SpecialFunnel[%s] funnel already closed, duplicate call.\n", f.id)
		return
	}

	// close 是发送通知，并不会阻塞
	close(f.closeChan)          // 通知所有协程开始处理剩余数据
	close(f.dataChan)           // 关闭数据通道阻止新数据
	f.wg.Wait()                 // 等待所有协程完成, 阻塞
	close(f.tickerCloseChan)    // 关闭定时器, 不阻塞
	time.Sleep(time.Second * 1) // 等待1秒，确保所有协程都退出
	log.Printf("所有协程已退出。\n")
}

// AddData adds data to the funnel
// Parameters:
//   - data: Data to be processed
func (f *SpecialFunnel) AddData(data interface{}) {
	if f.closed.Load() {
		log.Printf("SpecialFunnel[%s] funnel already closed, cannot add data.\n", f.id)
		return
	}

	select {
	case f.dataChan <- data:
	case <-time.After(time.Second * 60): // If channel is full, wait 60 seconds before retrying
		log.Printf("SpecialFunnel[%s] data channel full, discarding data: %v\n", f.id, data)
	}
}

// GetProcessedCount gets the number of processed data items
// Returns:
//   - int64: Number of processed data items
func (f *SpecialFunnel) GetProcessedCount() int64 {
	return atomic.LoadInt64(&f.processedCount)
}
