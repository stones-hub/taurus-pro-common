package cron

import "time"

// TaskMetrics 定义任务指标
type TaskMetrics struct {
	SuccessCount    uint          // 成功次数
	FailureCount    uint          // 失败次数
	LastDuration    time.Duration // 上次执行时间
	AverageDuration time.Duration // 平均执行时间
	LastError       error         // 最后一次错误
	LastRunTime     time.Time     // 最后一次执行时间
	CreatedAt       time.Time     // 创建时间
}
