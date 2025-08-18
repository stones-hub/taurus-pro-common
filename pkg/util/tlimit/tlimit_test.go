package tlimit

import (
	"sync"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	// 创建一个限流器，容量为5，每100ms填充一个令牌
	limiter := NewRateLimiter(5, 100*time.Millisecond)

	// 测试初始状态
	if !limiter.Allow() {
		t.Error("First request should be allowed")
	}

	// 测试突发请求
	allowed := 0
	for i := 0; i < 10; i++ {
		if limiter.Allow() {
			allowed++
		}
	}

	// 初始容量为5，所以应该只允许5个请求
	if allowed != 4 {
		t.Errorf("Expected 4 requests to be allowed, got %d", allowed)
	}

	// 等待令牌填充
	time.Sleep(300 * time.Millisecond)

	// 应该允许3个新请求（300ms应该填充3个令牌）
	allowed = 0
	for i := 0; i < 3; i++ {
		if limiter.Allow() {
			allowed++
		}
	}

	if allowed != 3 {
		t.Errorf("Expected 3 requests to be allowed after waiting, got %d", allowed)
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	limiter := NewRateLimiter(5, 100*time.Millisecond)
	var wg sync.WaitGroup
	var allowed int32
	var mutex sync.Mutex

	// 并发测试
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Allow() {
				mutex.Lock()
				allowed++
				mutex.Unlock()
			}
		}()
	}

	wg.Wait()

	// 初始容量为5，所以应该只允许5个请求
	if allowed != 5 {
		t.Errorf("Expected 5 concurrent requests to be allowed, got %d", allowed)
	}
}

func TestCompositeRateLimiter(t *testing.T) {
	// 创建组合限流器
	limiter := NewCompositeRateLimiter(3, 5, 100*time.Millisecond)

	// 测试单个IP的限流
	ip := "192.168.1.1"
	allowed := 0
	for i := 0; i < 5; i++ {
		result := limiter.Allow(ip)
		if result.Allowed {
			allowed++
		}
	}

	// IP限流器容量为3，应该只允许3个请求
	if allowed != 3 {
		t.Errorf("Expected 3 requests to be allowed for single IP, got %d", allowed)
	}

	// 测试多个IP的限流
	ips := []string{"192.168.1.2", "192.168.1.3", "192.168.1.4"}
	totalAllowed := 0
	for _, ip := range ips {
		result := limiter.Allow(ip)
		if result.Allowed {
			totalAllowed++
		}
	}

	// 全局限流器还剩2个令牌（5-3），应该允许2个请求
	if totalAllowed != 2 {
		t.Errorf("Expected 2 requests to be allowed globally, got %d", totalAllowed)
	}
}

func TestCompositeRateLimiterQueue(t *testing.T) {
	// 创建组合限流器
	limiter := NewCompositeRateLimiter(2, 3, 100*time.Millisecond)

	// 测试请求排队
	ip := "192.168.1.1"
	results := make(chan bool, 5)

	// 并发发送5个请求
	for i := 0; i < 5; i++ {
		go func() {
			result := limiter.Allow(ip)
			results <- result.Allowed
		}()
	}

	// 收集结果
	allowed := 0
	for i := 0; i < 5; i++ {
		if <-results {
			allowed++
		}
	}

	// IP限流器容量为2，应该只允许2个请求，其他请求应该被拒绝
	if allowed > 2 {
		t.Errorf("Expected at most 2 requests to be allowed, got %d", allowed)
	}
}

func TestCompositeRateLimiterTimeout(t *testing.T) {
	// 创建组合限流器
	limiter := NewCompositeRateLimiter(1, 1, 1*time.Second)

	// 第一个请求应该立即通过
	ip := "192.168.1.1"
	result := limiter.Allow(ip)
	if !result.Allowed {
		t.Errorf("First request should be allowed, got reason: %s", result.Reason)
	}

	// 第二个请求应该被拒绝
	result = limiter.Allow(ip)
	if result.Allowed {
		t.Error("Second request should be denied")
	}
	if result.Reason != "IP限流" {
		t.Errorf("Expected IP限流 reason, got: %s", result.Reason)
	}

	// 检查重试时间
	if result.RetryAfter <= 0 {
		t.Error("RetryAfter should be positive")
	}
}

func TestCompositeRateLimiterConcurrent(t *testing.T) {
	// 创建组合限流器
	limiter := NewCompositeRateLimiter(2, 4, 100*time.Millisecond)

	var wg sync.WaitGroup
	results := make(map[string]int)
	var mutex sync.Mutex

	// 测试多个IP并发请求
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
	for _, ip := range ips {
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(ip string) {
				defer wg.Done()
				result := limiter.Allow(ip)
				if result.Allowed {
					mutex.Lock()
					results[ip]++
					mutex.Unlock()
				}
			}(ip)
		}
	}

	wg.Wait()

	// 验证每个IP的限制
	for ip, count := range results {
		if count > 2 {
			t.Errorf("IP %s exceeded limit: got %d requests allowed", ip, count)
		}
	}

	// 验证全局限制
	total := 0
	for _, count := range results {
		total += count
	}
	if total > 4 {
		t.Errorf("Global limit exceeded: got %d total requests allowed", total)
	}
}

func TestRateLimiterRefill(t *testing.T) {
	// 创建限流器，容量为2，每100ms填充一个令牌
	limiter := NewRateLimiter(2, 100*time.Millisecond)

	// 消耗所有令牌
	if !limiter.Allow() {
		t.Error("First request should be allowed")
	}
	if !limiter.Allow() {
		t.Error("Second request should be allowed")
	}
	if limiter.Allow() {
		t.Error("Third request should be denied")
	}

	// 等待令牌重新填充
	time.Sleep(250 * time.Millisecond)

	// 应该有2个新令牌
	if !limiter.Allow() {
		t.Error("Request should be allowed after refill")
	}
	if !limiter.Allow() {
		t.Error("Second request should be allowed after refill")
	}
}

func TestCompositeRateLimiterCleanup(t *testing.T) {
	// 创建一个短时间间隔的限流器
	limiter := NewCompositeRateLimiter(2, 4, 50*time.Millisecond)

	// 使用一些IP
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3"}
	for _, ip := range ips {
		limiter.Allow(ip)
	}

	// 验证IP限流器是否被创建
	limiter.mutex.Lock()
	initialCount := len(limiter.ipLimiters)
	limiter.mutex.Unlock()

	if initialCount != 3 {
		t.Errorf("Expected 3 IP limiters, got %d", initialCount)
	}

	// 等待一段时间后再次检查
	time.Sleep(200 * time.Millisecond)

	// 再次使用部分IP
	limiter.Allow(ips[0])
	limiter.Allow(ips[1])

	// 验证限流器状态
	limiter.mutex.Lock()
	for ip, ipLimiter := range limiter.ipLimiters {
		if ipLimiter == nil {
			t.Errorf("IP limiter for %s should not be nil", ip)
		}
	}
	limiter.mutex.Unlock()
}
