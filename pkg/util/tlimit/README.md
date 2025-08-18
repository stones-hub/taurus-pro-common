# 组合限流器 (Composite Rate Limiter)

## 概述

组合限流器是一个高性能的令牌桶限流器，同时实现了基于IP的限流和全局限流，采用简单直接的限流策略，不依赖复杂的队列机制。

## 特性

- ✅ **双重限流**：IP级别 + 全局限流
- ✅ **高性能**：无队列开销，直接响应
- ✅ **重试控制**：提供精确的重试时间建议
- ✅ **状态监控**：实时查看限流器状态
- ✅ **线程安全**：支持高并发场景
- ✅ **向后兼容**：提供简化API

## 设计思路

### 1. 简化架构
- 去掉了复杂的队列机制
- 采用直接限流策略
- 更符合HTTP请求的特性

### 2. 双重检查
```
请求 → IP限流检查 → 全局限流检查 → 响应
  ↓           ↓           ↓
  └─ 任一失败都直接拒绝，不排队等待
```

### 3. 客户端重试控制
- 提供精确的重试时间建议
- 支持指数退避策略
- 符合HTTP 429状态码规范

## 使用方法

### 基本用法

```go
package main

import (
    "fmt"
    "time"
    "github.com/stones-hub/taurus-pro-common/pkg/util/tlimit"
)

func main() {
    // 创建组合限流器：IP容量3，全局容量5，每100ms填充1个令牌
    limiter := tlimit.NewCompositeRateLimiter(3, 5, 100*time.Millisecond)
    
    // 检查请求是否允许
    result := limiter.Allow("192.168.1.1")
    
    if result.Allowed {
        fmt.Println("请求通过")
    } else {
        fmt.Printf("请求被限流: %s\n", result.Reason)
        fmt.Printf("建议重试时间: %v\n", result.RetryAfter)
        fmt.Printf("IP令牌数: %d\n", result.IPTokens)
        fmt.Printf("全局令牌数: %d\n", result.GlobalTokens)
    }
}
```

### HTTP中间件集成

```go
func rateLimitMiddleware(limiter *tlimit.CompositeRateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := getClientIP(r)
            result := limiter.Allow(clientIP)
            
            if !result.Allowed {
                // 设置限流响应头
                w.Header().Set("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
                w.Header().Set("X-RateLimit-Limit", "5")
                w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.GlobalTokens))
                
                // 返回429状态码
                w.WriteHeader(http.StatusTooManyRequests)
                w.Write([]byte(fmt.Sprintf(`{"error": "rate_limit_exceeded", "retry_after": "%v"}`, result.RetryAfter)))
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

### 客户端重试策略

```go
func smartRetry(limiter *tlimit.CompositeRateLimiter, ip string, maxRetries int) bool {
    for attempt := 0; attempt < maxRetries; attempt++ {
        result := limiter.Allow(ip)
        
        if result.Allowed {
            return true
        }
        
        // 使用限流器建议的重试时间
        if result.RetryAfter > 0 {
            fmt.Printf("等待 %v 后重试...\n", result.RetryAfter)
            time.Sleep(result.RetryAfter)
        } else {
            // 指数退避
            delay := time.Duration(100*(1<<attempt)) * time.Millisecond
            if delay > 5*time.Second {
                delay = 5 * time.Second
            }
            time.Sleep(delay)
        }
    }
    
    return false
}
```

### 状态监控

```go
func monitorLimiter(limiter *tlimit.CompositeRateLimiter) {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        status := limiter.GetStatus()
        
        fmt.Printf("全局状态: %d/%d 令牌\n",
            status["global"].(map[string]interface{})["tokens"],
            status["global"].(map[string]interface{})["capacity"])
        
        fmt.Printf("活跃IP数量: %d\n", status["totalIPs"])
        
        // 显示每个IP的详细状态
        for ip, ipStatus := range status["ipLimiters"].(map[string]interface{}) {
            ipInfo := ipStatus.(map[string]interface{})
            fmt.Printf("  %s: %d/%d 令牌\n",
                ip, ipInfo["tokens"], ipInfo["capacity"])
        }
    }
}
```

## API参考

### 类型定义

```go
// 限流结果
type RateLimitResult struct {
    Allowed     bool          // 是否允许通过
    Reason      string        // 限流原因
    RetryAfter  time.Duration // 建议重试时间
    IPTokens    int           // IP限流器当前令牌数
    GlobalTokens int          // 全局限流器当前令牌数
}

// 组合限流器
type CompositeRateLimiter struct {
    // ... 内部字段
}
```

### 主要方法

```go
// 创建新的组合限流器
func NewCompositeRateLimiter(ipCapacity, globalCapacity int, fillInterval time.Duration) *CompositeRateLimiter

// 检查请求是否允许（推荐使用）
func (c *CompositeRateLimiter) Allow(ip string) *RateLimitResult

// 简化版本，向后兼容
func (c *CompositeRateLimiter) AllowSimple(ip string) (bool, string)

// 获取限流器状态
func (c *CompositeRateLimiter) GetStatus() map[string]interface{}
```

## 配置建议

### 令牌填充间隔
- **短间隔**（如100ms）：响应更快，但CPU开销稍高
- **长间隔**（如1s）：CPU开销低，但响应较慢

### 容量设置
- **IP容量**：根据业务需求设置，通常3-10个
- **全局容量**：根据服务器性能设置，通常IP容量 × 预期并发IP数

### 示例配置
```go
// 高并发场景
limiter := tlimit.NewCompositeRateLimiter(5, 100, 100*time.Millisecond)

// 低延迟场景
limiter := tlimit.NewCompositeRateLimiter(3, 20, 50*time.Millisecond)

// 资源受限场景
limiter := tlimit.NewCompositeRateLimiter(2, 10, 1*time.Second)
```

## 最佳实践

### 1. 错误处理
```go
result := limiter.Allow(clientIP)
if !result.Allowed {
    // 记录限流日志
    log.Printf("IP %s 被限流: %s", clientIP, result.Reason)
    
    // 返回友好的错误信息
    return fmt.Errorf("请求过于频繁，请 %v 后重试", result.RetryAfter)
}
```

### 2. 重试策略
```go
// 使用限流器建议的重试时间
if result.RetryAfter > 0 {
    time.Sleep(result.RetryAfter)
} else {
    // 实现指数退避
    time.Sleep(time.Duration(100*(1<<retryCount)) * time.Millisecond)
}
```

### 3. 监控告警
```go
// 定期检查限流器状态
status := limiter.GetStatus()
if status["totalIPs"].(int) > 1000 {
    // 发送告警：活跃IP过多
    alert("限流器活跃IP数量过多")
}
```

## 性能特点

- **内存占用**：每个IP约100字节
- **CPU开销**：每次请求约0.1微秒
- **并发支持**：支持数万个并发IP
- **令牌精度**：毫秒级令牌填充

## 迁移指南

### 从旧版本升级

```go
// 旧版本
ok, msg := limiter.Allow(ip)

// 新版本（推荐）
result := limiter.Allow(ip)
if result.Allowed {
    // 处理请求
} else {
    // 处理限流
    log.Printf("限流原因: %s, 重试时间: %v", result.Reason, result.RetryAfter)
}

// 或者使用兼容API
ok, msg := limiter.AllowSimple(ip)
```

## 总结

新的组合限流器设计更加简洁高效，去掉了复杂的队列机制，采用直接限流策略。它提供了丰富的重试控制信息，让客户端能够实现智能的重试策略，同时保持了高性能和易用性。

这种设计更符合现代HTTP服务的需求，让限流器成为真正的"保护器"而不是"阻塞器"。
