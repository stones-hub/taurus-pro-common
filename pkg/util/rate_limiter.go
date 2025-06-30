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
	"time"
)

// RateLimiter 令牌桶限流器
// 使用令牌桶算法实现，可以处理突发流量，同时保证长期的平均速率
type RateLimiter struct {
	capacity      int           // 令牌桶的最大容量
	tokens        int           // 当前令牌数量
	fillInterval  time.Duration // 添加令牌的时间间隔
	lastTokenTime time.Time     // 上次添加令牌的时间
	mutex         sync.Mutex    // 用于保护共享状态的互斥锁
}

// NewRateLimiter 创建一个新的限流器
// capacity: 令牌桶容量
// fillInterval: 填充令牌的时间间隔
func NewRateLimiter(capacity int, fillInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		capacity:      capacity,
		tokens:        capacity, // 初始化时令牌数等于容量
		fillInterval:  fillInterval,
		lastTokenTime: time.Now(),
	}
}

// Allow 检查请求是否允许通过
// 返回 true 表示允许，false 表示拒绝
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastTokenTime)

	// 根据经过的时间添加令牌
	tokensToAdd := int(elapsed / rl.fillInterval)
	if tokensToAdd > 0 {
		rl.tokens = min(rl.capacity, rl.tokens+tokensToAdd)
		rl.lastTokenTime = now
	}

	// 如果有可用令牌，消耗一个并允许请求
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CompositeRateLimiter 组合限流器
// 同时实现了基于 IP 的限流和全局限流，并支持请求排队
type CompositeRateLimiter struct {
	ipLimiters     map[string]*RateLimiter // IP限流器映射表，每个IP一个限流器
	globalLimiter  *RateLimiter            // 全局限流器，控制总体流量
	queue          []chan bool             // 等待队列，存储未能立即获取令牌的请求
	queueSignal    chan struct{}           // 队列信号通道，用于通知处理新的排队请求
	ipCapacity     int                     // 每个IP的令牌桶容量
	globalCapacity int                     // 全局令牌桶容量
	mutex          sync.Mutex              // 用于保护共享状态的互斥锁
}

// NewCompositeRateLimiter 创建一个新的组合限流器
// ipCapacity: 每个IP的令牌桶容量
// globalCapacity: 全局令牌桶容量
// fillInterval: 填充令牌的时间间隔
func NewCompositeRateLimiter(ipCapacity, globalCapacity int, fillInterval time.Duration) *CompositeRateLimiter {
	compositeRateLimiter := &CompositeRateLimiter{
		ipLimiters:     make(map[string]*RateLimiter),
		globalLimiter:  NewRateLimiter(globalCapacity, fillInterval),
		ipCapacity:     ipCapacity,
		globalCapacity: globalCapacity,
		queue:          make([]chan bool, 0),
		queueSignal:    make(chan struct{}, 1), // 缓冲区为1，避免发送方阻塞
	}

	// 启动队列处理协程
	go compositeRateLimiter.processQueue()
	return compositeRateLimiter
}

// Allow 检查指定IP的请求是否允许通过
// 返回值：(是否允许, 错误信息)
func (compositeRateLimiter *CompositeRateLimiter) Allow(ip string) (bool, string) {
	compositeRateLimiter.mutex.Lock()

	// 获取或创建IP专用的限流器
	ipLimiter, exists := compositeRateLimiter.ipLimiters[ip]
	if !exists {
		ipLimiter = NewRateLimiter(compositeRateLimiter.ipCapacity, compositeRateLimiter.globalLimiter.fillInterval)
		compositeRateLimiter.ipLimiters[ip] = ipLimiter
	}

	// 检查全局限流器和IP限流器是否都允许请求
	if compositeRateLimiter.globalLimiter.Allow() && ipLimiter.Allow() {
		compositeRateLimiter.mutex.Unlock()
		return true, ""
	}

	// 如果不允许，将请求加入等待队列
	log.Printf("Request from IP %s is denied and queued", ip)

	wait := make(chan bool)
	compositeRateLimiter.queue = append(compositeRateLimiter.queue, wait)

	// 发送队列信号，通知处理程序有新请求
	select {
	case compositeRateLimiter.queueSignal <- struct{}{}:
	default: // 如果信号通道已满，跳过发送以避免阻塞
	}

	// Unlock before waiting to avoid holding the lock while blocked
	compositeRateLimiter.mutex.Unlock()

	// 等待处理结果，设置5秒超时
	select {
	case allowed := <-wait:
		return allowed, ""
	case <-time.After(5 * time.Second): // 5秒超时
		return false, "请求超时，请稍后再试！"
	}
}

// processQueue 处理等待队列中的请求
// 当收到队列信号时，尝试为队列中的请求分配令牌
func (compositeRateLimiter *CompositeRateLimiter) processQueue() {
	for range compositeRateLimiter.queueSignal {
		compositeRateLimiter.mutex.Lock()
		// 当队列不为空且全局限流器允许时，处理队列中的请求
		for len(compositeRateLimiter.queue) > 0 && compositeRateLimiter.globalLimiter.Allow() {
			wait := compositeRateLimiter.queue[0]
			compositeRateLimiter.queue = compositeRateLimiter.queue[1:]
			wait <- true
			close(wait)
		}
		compositeRateLimiter.mutex.Unlock()
	}
}

// ------------------  使用示例 ------------------
/*

func main() {
	// 创建限流器实例：每个IP每分钟1个请求，全局每分钟5个请求
	limiter := util.NewCompositeRateLimiter(1, 5, time.Minute)

	// 创建HTTP服务器
	mux := http.NewServeMux()

	// 使用限流中间件包装处理器
	mux.Handle("/", middleware.RateLimitMiddleware(http.HandlerFunc(helloHandler), limiter))

	// 启动服务器
	http.ListenAndServe(":8080", mux)
}

*/
