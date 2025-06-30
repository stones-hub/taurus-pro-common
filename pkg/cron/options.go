package cron

import (
	"log"
	"time"
)

// ConcurrencyMode 定义任务并发控制模式
type ConcurrencyMode int

const (
	// AllowConcurrent 允许并发执行（默认模式）
	AllowConcurrent ConcurrencyMode = iota
	// SkipIfRunning 如果任务还在运行则跳过本次执行
	SkipIfRunning
	// DelayIfRunning 如果任务还在运行则等待执行完成后再执行
	DelayIfRunning
)

// Options 定义配置选项
type Options struct {
	Location        *time.Location  // 时区设置
	EnableSeconds   bool            // 是否启用秒级调度
	Logger          *log.Logger     // 自定义日志器
	ConcurrencyMode ConcurrencyMode // 任务并发控制模式
}

// Option 定义配置函数类型
type Option func(*Options)

// WithLocation 设置时区
func WithLocation(loc *time.Location) Option {
	return func(o *Options) {
		o.Location = loc
	}
}

// WithSeconds 启用秒级调度
func WithSeconds() Option {
	return func(o *Options) {
		o.EnableSeconds = true
	}
}

// WithLogger 设置自定义日志器
func WithLogger(logger *log.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithConcurrencyMode 设置任务并发控制模式
func WithConcurrencyMode(mode ConcurrencyMode) Option {
	return func(o *Options) {
		o.ConcurrencyMode = mode
	}
}

// defaultOptions 返回默认配置
func defaultOptions() *Options {
	return &Options{
		Location:        time.Local,
		EnableSeconds:   false,
		Logger:          log.Default(),
		ConcurrencyMode: AllowConcurrent,
	}
}
