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

package tlimit

import (
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

	// 当前时间
	now := time.Now()
	// 计算上次添加令牌到现在的时间间隔
	elapsed := now.Sub(rl.lastTokenTime)

	// 根据经过的时间添加令牌，如果时间间隔大于填充间隔，则添加令牌
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
// 同时实现了基于 IP 的限流和全局限流，采用简单直接的限流策略
type CompositeRateLimiter struct {
	ipLimiters     map[string]*RateLimiter // IP限流器映射表，每个IP一个限流器
	globalLimiter  *RateLimiter            // 全局限流器，控制总体流量
	ipCapacity     int                     // 每个IP的令牌桶容量
	globalCapacity int                     // 全局令牌桶容量
	fillInterval   time.Duration           // 填充令牌的时间间隔
	mutex          sync.Mutex              // 用于保护共享状态的互斥锁
}

// RateLimitResult 限流结果，包含是否允许和重试信息
type RateLimitResult struct {
	Allowed      bool          // 是否允许通过
	Reason       string        // 限流原因
	RetryAfter   time.Duration // 建议重试时间
	IPTokens     int           // IP限流器当前令牌数
	GlobalTokens int           // 全局限流器当前令牌数
}

// NewCompositeRateLimiter 创建一个新的组合限流器
// ipCapacity: 每个IP的令牌桶容量
// globalCapacity: 全局令牌桶容量
// fillInterval: 填充令牌的时间间隔
func NewCompositeRateLimiter(ipCapacity, globalCapacity int, fillInterval time.Duration) *CompositeRateLimiter {
	return &CompositeRateLimiter{
		ipLimiters:     make(map[string]*RateLimiter),
		globalLimiter:  NewRateLimiter(globalCapacity, fillInterval),
		ipCapacity:     ipCapacity,
		globalCapacity: globalCapacity,
		fillInterval:   fillInterval,
	}
}

// Allow 检查指定IP的请求是否允许通过
// 返回值：限流结果，包含是否允许和重试信息
func (compositeRateLimiter *CompositeRateLimiter) Allow(ip string) *RateLimitResult {
	compositeRateLimiter.mutex.Lock()
	defer compositeRateLimiter.mutex.Unlock()

	// 获取或创建IP专用的限流器
	ipLimiter := compositeRateLimiter.getOrCreateIPLimiter(ip)

	// 1. 先检查IP限流
	if !ipLimiter.Allow() {
		return compositeRateLimiter.createRateLimitResult(false, "IP限流", ipLimiter, compositeRateLimiter.globalLimiter)
	}

	// 2. 再检查全局限流
	if !compositeRateLimiter.globalLimiter.Allow() {
		return compositeRateLimiter.createRateLimitResult(false, "全局限流", ipLimiter, compositeRateLimiter.globalLimiter)
	}

	// 两个限流器都通过
	return compositeRateLimiter.createRateLimitResult(true, "", ipLimiter, compositeRateLimiter.globalLimiter)
}

// AllowSimple 简化版本，只返回是否允许（向后兼容）
func (compositeRateLimiter *CompositeRateLimiter) AllowSimple(ip string) (bool, string) {
	result := compositeRateLimiter.Allow(ip)
	return result.Allowed, result.Reason
}

// getOrCreateIPLimiter 获取或创建IP专用的限流器
func (compositeRateLimiter *CompositeRateLimiter) getOrCreateIPLimiter(ip string) *RateLimiter {
	ipLimiter, exists := compositeRateLimiter.ipLimiters[ip]
	if !exists {
		ipLimiter = NewRateLimiter(compositeRateLimiter.ipCapacity, compositeRateLimiter.fillInterval)
		compositeRateLimiter.ipLimiters[ip] = ipLimiter
	}
	return ipLimiter
}

// createRateLimitResult 创建限流结果
func (compositeRateLimiter *CompositeRateLimiter) createRateLimitResult(allowed bool, reason string, ipLimiter, globalLimiter *RateLimiter) *RateLimitResult {
	result := &RateLimitResult{
		Allowed:      allowed,
		Reason:       reason,
		IPTokens:     ipLimiter.tokens,
		GlobalTokens: globalLimiter.tokens,
	}

	if !allowed {
		// 计算建议重试时间
		result.RetryAfter = compositeRateLimiter.calculateRetryAfter(ipLimiter, globalLimiter)
	}

	return result
}

// calculateRetryAfter 计算建议重试时间
func (compositeRateLimiter *CompositeRateLimiter) calculateRetryAfter(ipLimiter, globalLimiter *RateLimiter) time.Duration {
	now := time.Now()

	// 计算IP限流器的重试时间
	ipRetryAfter := compositeRateLimiter.calculateLimiterRetryAfter(ipLimiter, now)

	// 计算全局限流器的重试时间
	globalRetryAfter := compositeRateLimiter.calculateLimiterRetryAfter(globalLimiter, now)

	// 返回较大的重试时间，确保两个限流器都有令牌
	if ipRetryAfter > globalRetryAfter {
		return ipRetryAfter
	}
	return globalRetryAfter
}

// calculateLimiterRetryAfter 计算单个限流器的重试时间
func (compositeRateLimiter *CompositeRateLimiter) calculateLimiterRetryAfter(limiter *RateLimiter, now time.Time) time.Duration {
	if limiter.tokens > 0 {
		return 0 // 有令牌，可以立即重试
	}

	// 计算下次令牌填充的时间
	elapsed := now.Sub(limiter.lastTokenTime)
	nextTokenTime := limiter.fillInterval - elapsed

	if nextTokenTime <= 0 {
		return 0 // 令牌应该已经填充
	}

	return nextTokenTime
}

// GetStatus 获取限流器状态信息（用于监控和调试）
func (compositeRateLimiter *CompositeRateLimiter) GetStatus() map[string]interface{} {
	compositeRateLimiter.mutex.Lock()
	defer compositeRateLimiter.mutex.Unlock()

	status := map[string]interface{}{
		"global": map[string]interface{}{
			"capacity":     compositeRateLimiter.globalCapacity,
			"tokens":       compositeRateLimiter.globalLimiter.tokens,
			"fillInterval": compositeRateLimiter.fillInterval.String(),
		},
		"ipLimiters": map[string]interface{}{},
		"totalIPs":   len(compositeRateLimiter.ipLimiters),
	}

	for ip, limiter := range compositeRateLimiter.ipLimiters {
		status["ipLimiters"].(map[string]interface{})[ip] = map[string]interface{}{
			"capacity":     compositeRateLimiter.ipCapacity,
			"tokens":       limiter.tokens,
			"fillInterval": limiter.fillInterval.String(),
		}
	}

	return status
}
