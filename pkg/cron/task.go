package cron

import (
	"context"
	"time"
)

// TaskFunc 定义任务函数类型
type TaskFunc func(ctx context.Context) error

// Task 定义任务结构
type Task struct {
	Name       string        // 任务名称
	Spec       string        // 任务调度表达式
	Func       TaskFunc      // 任务函数
	Timeout    time.Duration // 任务超时时间
	RetryCount int           // 重试次数
	RetryDelay time.Duration // 重试间隔
	Group      *TaskGroup    // 任务分组
	Tags       []string      // 任务标签
}

type TaskOption func(*Task)

// NewTask 创建新任务
func NewTask(name, spec string, fn TaskFunc, opts ...TaskOption) *Task {
	t := &Task{
		Name:       name,
		Spec:       spec,
		Func:       fn,
		Timeout:    time.Hour,   // 默认超时时间
		RetryCount: 0,           // 默认不重试
		RetryDelay: time.Minute, // 默认重试间隔
		Tags:       make([]string, 0),
	}

	for _, opt := range opts {
		opt(t)
	}

	return t
}

// WithTimeout 设置任务超时时间
func WithTimeout(timeout time.Duration) TaskOption {
	return func(t *Task) {
		t.Timeout = timeout
	}
}

// WithRetry 设置重试策略
func WithRetry(count int, delay time.Duration) TaskOption {
	return func(t *Task) {
		t.RetryCount = count
		t.RetryDelay = delay
	}
}

// WithGroup 设置任务分组
func WithGroup(group *TaskGroup) TaskOption {
	return func(t *Task) {
		t.Group = group
	}
}

// WithTag 添加任务标签
func WithTag(tag string) TaskOption {
	return func(t *Task) {
		t.Tags = append(t.Tags, tag)
	}
}

// TaskGroup 定义任务分组
type TaskGroup struct {
	Name string   // 分组名称
	Tags []string // 分组标签
}

// NewTaskGroup 创建新的任务分组
func NewTaskGroup(name string) *TaskGroup {
	return &TaskGroup{
		Name: name,
		Tags: make([]string, 0),
	}
}

// AddTag 添加标签到分组
func (g *TaskGroup) AddTag(tag string) {
	g.Tags = append(g.Tags, tag)
}

// HasTag 检查分组是否有指定标签
func (g *TaskGroup) HasTag(tag string) bool {
	for _, t := range g.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// RemoveTag 从分组中移除标签
func (g *TaskGroup) RemoveTag(tag string) {
	for i, t := range g.Tags {
		if t == tag {
			g.Tags = append(g.Tags[:i], g.Tags[i+1:]...)
			return
		}
	}
}

// AddTag 添加标签到任务
func (t *Task) AddTag(tag string) {
	t.Tags = append(t.Tags, tag)
}

// HasTag 检查任务是否有指定标签
func (t *Task) HasTag(tag string) bool {
	for _, tt := range t.Tags {
		if tt == tag {
			return true
		}
	}
	return false
}

// RemoveTag 从任务中移除标签
func (t *Task) RemoveTag(tag string) {
	for i, tt := range t.Tags {
		if tt == tag {
			t.Tags = append(t.Tags[:i], t.Tags[i+1:]...)
			return
		}
	}
}
