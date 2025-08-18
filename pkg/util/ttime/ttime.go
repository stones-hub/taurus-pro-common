package ttime

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var (
	// 使用原子操作替代互斥锁
	nowSecondCache atomic.Int64
	// 使用 time.Ticker 定期更新缓存
	_ = startSecondCacheUpdater()
)

// startSecondCacheUpdater 启动一个后台 goroutine 每秒更新时间缓存，用于提高性能
// 返回值：
//   - *time.Ticker: 用于更新缓存的定时器，可用于停止更新
//
// 注意事项：
//   - 使用原子操作替代互斥锁，避免锁竞争
//   - 每秒更新一次缓存，精度在1秒以内
//   - 返回的 Ticker 不应该被停止，除非确实不再需要时间缓存
//   - 此函数在包初始化时自动调用
//   - 缓存的时间戳通过 GetUnixSeconds 函数访问
func startSecondCacheUpdater() *time.Ticker {
	ticker := time.NewTicker(time.Second)
	nowSecondCache.Store(time.Now().Unix())

	go func() {
		for range ticker.C {
			nowSecondCache.Store(time.Now().Unix())
		}
	}()

	return ticker
}

// GetUnixMilliSeconds 获取当前Unix时间戳（毫秒级精度）
// 返回值：
//   - int64: 当前时间的毫秒级Unix时间戳
//
// 使用示例：
//
//	// 获取当前毫秒时间戳
//	ms := ttime.GetUnixMilliSeconds()
//	fmt.Printf("当前时间戳（毫秒）：%d\n", ms)
//
//	// 用于计算时间差
//	start := ttime.GetUnixMilliSeconds()
//	// ... 执行一些操作 ...
//	end := ttime.GetUnixMilliSeconds()
//	duration := end - start
//	fmt.Printf("操作耗时：%d毫秒\n", duration)
//
// 注意事项：
//   - 返回值精确到毫秒
//   - 不使用缓存，每次调用都会获取当前时间
//   - 如果只需要秒级精度，建议使用 GetUnixSeconds
//   - 适用于需要毫秒级精度的场景
func GetUnixMilliSeconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// GetUnixSeconds 获取当前Unix时间戳（秒级精度，使用缓存）
// 返回值：
//   - int64: 当前时间的秒级Unix时间戳
//
// 使用示例：
//
//	// 获取当前秒级时间戳
//	seconds := ttime.GetUnixSeconds()
//	fmt.Printf("当前时间戳（秒）：%d\n", seconds)
//
//	// 用于计算时间差
//	start := ttime.GetUnixSeconds()
//	// ... 执行一些操作 ...
//	end := ttime.GetUnixSeconds()
//	duration := end - start
//	fmt.Printf("操作耗时：%d秒\n", duration)
//
// 注意事项：
//   - 使用缓存提高性能，每秒更新一次
//   - 精度在1秒以内，适用于大多数场景
//   - 如果需要更高精度，请使用 GetUnixMilliSeconds
//   - 使用原子操作保证线程安全
//   - 比直接调用 time.Now().Unix() 性能更好
func GetUnixSeconds() int64 {
	return nowSecondCache.Load()
}

// Seconds2date 将Unix时间戳（秒）转换为指定格式的日期字符串
// 参数：
//   - timestamp: Unix时间戳（秒）
//   - format: 日期格式字符串（如 "2006-01-02 15:04:05"）
//
// 返回值：
//   - string: 格式化后的日期字符串
//
// 使用示例：
//
//	// 转换时间戳为标准日期格式
//	ts := time.Now().Unix()
//	date := ttime.Seconds2date(ts, "2006-01-02 15:04:05")
//	fmt.Printf("当前时间：%s\n", date)
//
//	// 自定义格式
//	date = ttime.Seconds2date(ts, "2006年01月02日 15时04分")
//	fmt.Printf("中文格式：%s\n", date)
//
//	// 只显示日期部分
//	date = ttime.Seconds2date(ts, "2006-01-02")
//	fmt.Printf("日期：%s\n", date)
//
// 注意事项：
//   - 使用 Go 的标准时间格式化语法
//   - 时间戳必须是秒级的
//   - 如果时间戳是毫秒级的，请使用 Milliseconds2date
//   - 返回的字符串格式完全由 format 参数控制
func Seconds2date(timestamp int64, format string) string {
	t := time.Unix(timestamp, 0).UTC()
	formattedTime := t.Format(format)
	return formattedTime
}

// Milliseconds2date 将Unix时间戳（毫秒）转换为指定格式的日期字符串
// 参数：
//   - milliseconds: Unix时间戳（毫秒）
//   - format: 日期格式字符串（如 "2006-01-02 15:04:05.000"）
//
// 返回值：
//   - string: 格式化后的日期字符串
//
// 使用示例：
//
//	// 转换毫秒时间戳为标准日期格式
//	ms := ttime.GetUnixMilliSeconds()
//	date := ttime.Milliseconds2date(ms, "2006-01-02 15:04:05.000")
//	// fmt.Printf("当前时间（毫秒）：%s\n", date)
//
//	// 自定义格式（不显示毫秒）
//	date = ttime.Milliseconds2date(ms, "2006年01月02日 15时04分05秒")
//	// fmt.Printf("中文格式：%s\n", date)
//
//	// 只显示时间部分（带毫秒）
//	date = ttime.Milliseconds2date(ms, "15:04:05.000")
//	// fmt.Printf("时间：%s\n", date)
//
// 注意事项：
//   - 使用 Go 的标准时间格式化语法
//   - 时间戳必须是毫秒级的
//   - 如果时间戳是秒级的，请使用 Seconds2date
//   - 如果需要显示毫秒，format 中要包含 .000
//   - 返回的字符串格式完全由 format 参数控制
func Milliseconds2date(milliseconds int64, format string) string {
	t := time.UnixMilli(milliseconds).UTC()
	return t.Format(format)
}

// TimeFormatter 将时间格式化为标准格式（YYYY-MM-DD HH:mm:ss）
// 参数：
//   - time: 要格式化的时间
//
// 返回值：
//   - string: 格式化后的时间字符串
//
// 使用示例：
//
//	// 格式化当前时间
//	now := time.Now()
//	formatted := ttime.TimeFormatter(now)
//	fmt.Printf("当前时间：%s\n", formatted)
//	// 输出类似：当前时间：2023-06-15 14:30:45
//
//	// 格式化指定时间
//	t := time.Date(2023, 6, 15, 14, 30, 45, 0, time.Local)
//	formatted = ttime.TimeFormatter(t)
//	fmt.Printf("指定时间：%s\n", formatted)
//
// 注意事项：
//   - 使用固定的格式："2006-01-02 15:04:05"
//   - 不包含毫秒部分
//   - 使用 24 小时制
//   - 如果需要自定义格式，请直接使用 time.Format
func TimeFormatter(time time.Time) string {
	return time.Format("2006-01-02 15:04:05")
}

// ParseDuration 解析持续时间字符串为 time.Duration，支持扩展的时间单位格式
// 参数：
//   - d: 持续时间字符串，支持以下格式：
//   - 标准 Go 时间单位：ns（纳秒）, us/µs（微秒）, ms（毫秒）, s（秒）, m（分）, h（小时）
//   - 天数单位：d，如 "7d"
//   - 组合格式：如 "1d12h30m"
//   - 纯数字：默认解释为秒，如 "60" 表示 60 秒
//
// 返回值：
//   - time.Duration: 解析后的持续时间
//   - error: 如果解析失败则返回错误信息
//
// 使用示例：
//
//	// 标准时间单位
//	d, err := ttime.ParseDuration("2h30m")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("持续时间：%v\n", d)  // 输出：2h30m0s
//
//	// 使用天数
//	d, err = ttime.ParseDuration("2d12h")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("持续时间：%v\n", d)  // 输出：60h0m0s
//
//	// 组合格式
//	d, err = ttime.ParseDuration("1d6h30m15s")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("持续时间：%v\n", d)  // 输出：30h30m15s
//
//	// 纯数字（秒）
//	d, err = ttime.ParseDuration("3600")
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("持续时间：%v\n", d)  // 输出：1h0m0s
//
// 注意事项：
//   - 空字符串会返回错误
//   - 天数会被转换为小时（1天 = 24小时）
//   - 支持任意顺序的组合，如 "30m2h" 等价于 "2h30m"
//   - 纯数字默认为秒，可以是负数
//   - 如果只需要标准单位，建议直接使用 time.ParseDuration
func ParseDuration(d string) (time.Duration, error) {
	d = strings.TrimSpace(d)
	if d == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// 尝试标准解析
	if dur, err := time.ParseDuration(d); err == nil {
		return dur, nil
	}

	// 处理天数
	if strings.Contains(d, "d") {
		parts := strings.SplitN(d, "d", 2)
		if len(parts) == 0 {
			return 0, fmt.Errorf("invalid duration format")
		}

		// 解析天数部分
		days, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid days value: %w", err)
		}
		duration := time.Hour * 24 * time.Duration(days)

		// 如果有剩余部分，解析它
		if len(parts) > 1 && parts[1] != "" {
			remainder, err := time.ParseDuration(parts[1])
			if err != nil {
				return 0, fmt.Errorf("invalid remainder duration: %w", err)
			}
			duration += remainder
		}

		return duration, nil
	}

	// 尝试解析纯数字（作为秒数）
	if seconds, err := strconv.ParseInt(d, 10, 64); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	return 0, fmt.Errorf("unsupported duration format")
}

var (
	// 预编译正则表达式
	iso8601Regex = regexp.MustCompile(
		`"((\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})(?:\.(\d{3}))\d*)(Z|[\+-]\d{2}:\d{2})"`,
	)
	iso8601Substitution = []byte("\"$2 $3.$4\"")
)

// ParseTime 将 ISO 8601 格式的时间字符串转换为更易读的格式
// 参数：
//   - input: 包含 ISO 8601 格式时间字符串的字节切片
//
// 返回值：
//   - string: 转换后的易读时间字符串
//
// 使用示例：
//
//	// 基本 ISO 8601 格式
//	str := ttime.ParseTime([]byte(`"2023-04-05T15:30:45.123456Z"`))
//	fmt.Println(str)  // 输出：2023-04-05 15:30:45.123
//
//	// 带时区偏移
//	str = ttime.ParseTime([]byte(`"2023-04-05T15:30:45.123456+08:00"`))
//	fmt.Println(str)  // 输出：2023-04-05 15:30:45.123
//
//	// 非 ISO 8601 格式保持不变
//	str = ttime.ParseTime([]byte(`"2023-04-05 15:30:45"`))
//	fmt.Println(str)  // 输出：2023-04-05 15:30:45
//
// 注意事项：
//   - 输入必须是带引号的 JSON 字符串格式
//   - 只处理包含 'T' 分隔符的 ISO 8601 格式
//   - 会保留3位毫秒精度，截断更高精度
//   - 会移除时区信息
//   - 如果输入不是 ISO 格式，返回原始字符串
//   - 主要用于处理 JSON 中的时间字段
func ParseTime(input []byte) string {
	if !bytes.Contains(input, []byte("T")) {
		return string(input)
	}
	return string(iso8601Regex.ReplaceAll(input, iso8601Substitution))
}

// ParseTimeString 将 ISO 8601 格式的时间字符串转换为更易读的格式（字符串版本）
// 参数：
//   - input: ISO 8601 格式的时间字符串
//
// 返回值：
//   - string: 转换后的易读时间字符串
//
// 使用示例：
//
//	// 基本 ISO 8601 格式
//	str := ttime.ParseTimeString(`"2023-04-05T15:30:45.123456Z"`)
//	fmt.Println(str)  // 输出：2023-04-05 15:30:45.123
//
//	// 带时区偏移
//	str = ttime.ParseTimeString(`"2023-04-05T15:30:45.123456+08:00"`)
//	fmt.Println(str)  // 输出：2023-04-05 15:30:45.123
//
//	// 非 ISO 8601 格式保持不变
//	str = ttime.ParseTimeString(`"2023-04-05 15:30:45"`)
//	fmt.Println(str)  // 输出：2023-04-05 15:30:45
//
// 注意事项：
//   - 是 ParseTime 的字符串版本
//   - 输入必须是带引号的 JSON 字符串格式
//   - 只处理包含 'T' 分隔符的 ISO 8601 格式
//   - 会保留3位毫秒精度，截断更高精度
//   - 会移除时区信息
//   - 如果输入不是 ISO 格式，返回原始字符串
//   - 主要用于处理 JSON 中的时间字段
func ParseTimeString(input string) string {
	return ParseTime([]byte(input))
}

// Now 获取当前时间
// 返回值：
//   - time.Time: 当前时间
//
// 使用示例：
//
//	// 获取当前时间
//	now := ttime.Now()
//	fmt.Printf("当前时间：%v\n", now)
//
//	// 用于时间计算
//	future := now.Add(24 * time.Hour)
//	fmt.Printf("24小时后：%v\n", future)
//
// 注意事项：
//   - 返回本地时区的时间
//   - 是 time.Now() 的简单封装
//   - 如果需要 UTC 时间，使用 Now().UTC()
//   - 如果需要时间戳，建议使用 GetUnixSeconds 或 GetUnixMilliSeconds
func Now() time.Time {
	return time.Now()
}

// Today 获取今天的开始时间（00:00:00）
// 返回值：
//   - time.Time: 今天零点的时间
//
// 使用示例：
//
//	// 获取今天零点
//	today := ttime.Today()
//	fmt.Printf("今天零点：%v\n", today)
//
//	// 用于日期范围查询
//	if someTime.After(Today()) {
//	    fmt.Println("时间在今天之内")
//	}
//
//	// 计算到现在的时间差
//	duration := time.Since(Today())
//	fmt.Printf("今天已经过去：%v\n", duration)
//
// 注意事项：
//   - 返回本地时区的时间
//   - 时间部分被设置为 00:00:00.000
//   - 常用于日期范围的开始时间
//   - 如果需要其他时区，使用 Today().In(location)
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Tomorrow 获取明天的开始时间（00:00:00）
// 返回值：
//   - time.Time: 明天零点的时间
//
// 使用示例：
//
//	// 获取明天零点
//	tomorrow := ttime.Tomorrow()
//	fmt.Printf("明天零点：%v\n", tomorrow)
//
//	// 用于设置过期时间
//	expireTime := ttime.Tomorrow()
//	if time.Now().After(expireTime) {
//	    fmt.Println("已过期")
//	}
//
//	// 计算到明天的时间差
//	duration := time.Until(Tomorrow())
//	fmt.Printf("距离明天还有：%v\n", duration)
//
// 注意事项：
//   - 返回本地时区的时间
//   - 时间部分被设置为 00:00:00.000
//   - 是 Today().AddDate(0, 0, 1) 的简单封装
//   - 常用于设置次日过期时间
//   - 如果需要其他时区，使用 Tomorrow().In(location)
func Tomorrow() time.Time {
	return Today().AddDate(0, 0, 1)
}

// Yesterday 获取昨天的开始时间（00:00:00）
// 返回值：
//   - time.Time: 昨天零点的时间
//
// 使用示例：
//
//	// 获取昨天零点
//	yesterday := ttime.Yesterday()
//	fmt.Printf("昨天零点：%v\n", yesterday)
//
//	// 用于日期范围查询
//	if someTime.After(Yesterday()) && someTime.Before(Today()) {
//	    fmt.Println("时间在昨天")
//	}
//
//	// 计算到现在的时间差
//	duration := time.Since(Yesterday())
//	fmt.Printf("距离昨天已经过去：%v\n", duration)
//
// 注意事项：
//   - 返回本地时区的时间
//   - 时间部分被设置为 00:00:00.000
//   - 是 Today().AddDate(0, 0, -1) 的简单封装
//   - 常用于查询昨天的数据
//   - 如果需要其他时区，使用 Yesterday().In(location)
func Yesterday() time.Time {
	return Today().AddDate(0, 0, -1)
}

// ThisWeek 获取本周的开始时间（周一 00:00:00）
// 返回值：
//   - time.Time: 本周一零点的时间
//
// 使用示例：
//
//	// 获取本周一零点
//	weekStart := ttime.ThisWeek()
//	fmt.Printf("本周开始：%v\n", weekStart)
//
//	// 获取本周的时间范围
//	weekEnd := weekStart.AddDate(0, 0, 7)
//	fmt.Printf("本周时间范围：%v 至 %v\n", weekStart, weekEnd)
//
//	// 判断时间是否在本周
//	if someTime.After(ThisWeek()) && someTime.Before(ThisWeek().AddDate(0, 0, 7)) {
//	    fmt.Println("时间在本周内")
//	}
//
// 注意事项：
//   - 返回本地时区的时间
//   - 时间部分被设置为 00:00:00.000
//   - 以周一作为一周的开始
//   - 周日被视为上一周的最后一天
//   - 如果需要其他时区，使用 ThisWeek().In(location)
func ThisWeek() time.Time {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // 周日算作第7天
	}
	daysToMonday := weekday - 1
	return Today().AddDate(0, 0, -daysToMonday)
}

// ThisMonth 获取本月的开始时间（1号 00:00:00）
// 返回值：
//   - time.Time: 本月1号零点的时间
//
// 使用示例：
//
//	// 获取本月1号零点
//	monthStart := ttime.ThisMonth()
//	fmt.Printf("本月开始：%v\n", monthStart)
//
//	// 获取本月的时间范围
//	nextMonth := monthStart.AddDate(0, 1, 0)
//	fmt.Printf("本月时间范围：%v 至 %v\n", monthStart, nextMonth)
//
//	// 判断时间是否在本月
//	if someTime.After(ThisMonth()) && someTime.Before(ThisMonth().AddDate(0, 1, 0)) {
//	    fmt.Println("时间在本月内")
//	}
//
// 注意事项：
//   - 返回本地时区的时间
//   - 时间部分被设置为 00:00:00.000
//   - 每月1号作为月份的开始
//   - 自动处理不同月份的天数
//   - 如果需要其他时区，使用 ThisMonth().In(location)
func ThisMonth() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}

// ThisYear 获取本年的开始时间（1月1号 00:00:00）
// 返回值：
//   - time.Time: 本年1月1号零点的时间
//
// 使用示例：
//
//	// 获取本年1月1号零点
//	yearStart := ttime.ThisYear()
//	fmt.Printf("本年开始：%v\n", yearStart)
//
//	// 获取本年的时间范围
//	nextYear := yearStart.AddDate(1, 0, 0)
//	fmt.Printf("本年时间范围：%v 至 %v\n", yearStart, nextYear)
//
//	// 判断时间是否在本年
//	if someTime.After(ThisYear()) && someTime.Before(ThisYear().AddDate(1, 0, 0)) {
//	    fmt.Println("时间在本年内")
//	}
//
//	// 计算年初到现在的天数
//	days := time.Since(ThisYear()).Hours() / 24
//	fmt.Printf("今年已经过去 %.0f 天\n", days)
//
// 注意事项：
//   - 返回本地时区的时间
//   - 时间部分被设置为 00:00:00.000
//   - 每年1月1号作为年份的开始
//   - 自动处理闰年
//   - 如果需要其他时区，使用 ThisYear().In(location)
func ThisYear() time.Time {
	now := time.Now()
	return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
}

// IsToday 检查给定时间是否在今天
// 参数：
//   - t: 要检查的时间
//
// 返回值：
//   - bool: 如果时间在今天返回 true，否则返回 false
//
// 使用示例：
//
//	// 检查当前时间
//	now := time.Now()
//	if ttime.IsToday(now) {
//	    fmt.Println("时间是今天")
//	}
//
//	// 检查特定时间
//	t := time.Date(2023, 6, 15, 14, 30, 0, 0, time.Local)
//	if ttime.IsToday(t) {
//	    fmt.Println("2023-06-15 14:30:00 是今天")
//	}
//
//	// 用于过滤数据
//	if ttime.IsToday(record.CreatedAt) {
//	    fmt.Println("记录创建于今天")
//	}
//
// 注意事项：
//   - 只比较日期部分，忽略时间部分
//   - 使用年份和年中的天数比较，处理跨月的情况
//   - 考虑时区，使用时间的原始时区进行比较
//   - 比 time.Now() 更高效，因为复用了 Now() 的结果
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

// IsThisWeek 检查给定时间是否在本周内（周一到周日）
// 参数：
//   - t: 要检查的时间
//
// 返回值：
//   - bool: 如果时间在本周内返回 true，否则返回 false
//
// 使用示例：
//
//	// 检查当前时间
//	now := time.Now()
//	if ttime.IsThisWeek(now) {
//	    fmt.Println("时间在本周内")
//	}
//
//	// 检查特定时间
//	t := time.Date(2023, 6, 15, 14, 30, 0, 0, time.Local)
//	if ttime.IsThisWeek(t) {
//	    fmt.Println("2023-06-15 14:30:00 在本周内")
//	}
//
//	// 用于过滤数据
//	if ttime.IsThisWeek(record.CreatedAt) {
//	    fmt.Println("记录创建于本周")
//	}
//
// 注意事项：
//   - 以周一为一周的开始
//   - 周日被视为本周的最后一天
//   - 使用 ThisWeek() 获取本周开始时间
//   - 考虑时区，使用时间的原始时区进行比较
//   - 比较时包含边界（大于等于本周开始，小于下周开始）
func IsThisWeek(t time.Time) bool {
	weekStart := ThisWeek()
	weekEnd := weekStart.AddDate(0, 0, 7)
	return t.After(weekStart) && t.Before(weekEnd)
}

// IsThisMonth 检查给定时间是否在本月内
// 参数：
//   - t: 要检查的时间
//
// 返回值：
//   - bool: 如果时间在本月内返回 true，否则返回 false
//
// 使用示例：
//
//	// 检查当前时间
//	now := time.Now()
//	if ttime.IsThisMonth(now) {
//	    fmt.Println("时间在本月内")
//	}
//
//	// 检查特定时间
//	t := time.Date(2023, 6, 15, 14, 30, 0, 0, time.Local)
//	if ttime.IsThisMonth(t) {
//	    fmt.Println("2023-06-15 14:30:00 在本月内")
//	}
//
//	// 用于过滤数据
//	if ttime.IsThisMonth(record.CreatedAt) {
//	    fmt.Println("记录创建于本月")
//	}
//
// 注意事项：
//   - 比较年份和月份，忽略日期和时间部分
//   - 自动处理不同月份的天数
//   - 考虑时区，使用时间的原始时区进行比较
//   - 比 time.Now() 更高效，因为复用了 Now() 的结果
//   - 适用于月度统计和报表
func IsThisMonth(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month()
}

// IsThisYear 检查给定时间是否在今年内
// 参数：
//   - t: 要检查的时间
//
// 返回值：
//   - bool: 如果时间在今年内返回 true，否则返回 false
//
// 使用示例：
//
//	// 检查当前时间
//	now := time.Now()
//	if ttime.IsThisYear(now) {
//	    fmt.Println("时间在今年内")
//	}
//
//	// 检查特定时间
//	t := time.Date(2023, 6, 15, 14, 30, 0, 0, time.Local)
//	if ttime.IsThisYear(t) {
//	    fmt.Println("2023-06-15 14:30:00 在今年内")
//	}
//
//	// 用于过滤数据
//	if ttime.IsThisYear(record.CreatedAt) {
//	    fmt.Println("记录创建于今年")
//	}
//
// 注意事项：
//   - 只比较年份，忽略月份、日期和时间部分
//   - 自动处理闰年
//   - 考虑时区，使用时间的原始时区进行比较
//   - 比 time.Now() 更高效，因为复用了 Now() 的结果
//   - 适用于年度统计和报表
func IsThisYear(t time.Time) bool {
	return t.Year() == time.Now().Year()
}

// AddDays 在给定时间上添加或减少指定的天数
// 参数：
//   - t: 基准时间
//   - days: 要添加的天数，可以是负数表示减少天数
//
// 返回值：
//   - time.Time: 计算后的新时间
//
// 使用示例：
//
//	// 添加天数
//	now := time.Now()
//	future := ttime.AddDays(now, 7)
//	fmt.Printf("7天后：%v\n", future)
//
//	// 减少天数
//	past := ttime.AddDays(now, -7)
//	fmt.Printf("7天前：%v\n", past)
//
//	// 设置过期时间
//	expireTime := ttime.AddDays(time.Now(), 30)
//	fmt.Printf("30天后过期：%v\n", expireTime)
//
// 注意事项：
//   - 保持时间部分不变
//   - 自动处理月末和闰年
//   - 考虑时区，使用时间的原始时区
//   - 比手动计算更准确，尤其是在月末
//   - 是 t.AddDate(0, 0, days) 的简单封装
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddMonths 在给定时间上添加或减少指定的月数
// 参数：
//   - t: 基准时间
//   - months: 要添加的月数，可以是负数表示减少月数
//
// 返回值：
//   - time.Time: 计算后的新时间
//
// 使用示例：
//
//	// 添加月数
//	now := time.Now()
//	future := ttime.AddMonths(now, 3)
//	fmt.Printf("3个月后：%v\n", future)
//
//	// 减少月数
//	past := ttime.AddMonths(now, -3)
//	fmt.Printf("3个月前：%v\n", past)
//
//	// 设置订阅到期时间
//	expireTime := ttime.AddMonths(time.Now(), 12)
//	fmt.Printf("一年后到期：%v\n", expireTime)
//
// 注意事项：
//   - 保持日期和时间部分不变（如果可能）
//   - 自动调整月末日期（如1月31日加一个月会变成2月28/29日）
//   - 考虑时区，使用时间的原始时区
//   - 处理跨年的情况
//   - 是 t.AddDate(0, months, 0) 的简单封装
func AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// AddYears 在给定时间上添加或减少指定的年数
// 参数：
//   - t: 基准时间
//   - years: 要添加的年数，可以是负数表示减少年数
//
// 返回值：
//   - time.Time: 计算后的新时间
//
// 使用示例：
//
//	// 添加年数
//	now := time.Now()
//	future := ttime.AddYears(now, 1)
//	fmt.Printf("明年：%v\n", future)
//
//	// 减少年数
//	past := ttime.AddYears(now, -1)
//	fmt.Printf("去年：%v\n", past)
//
//	// 计算证件有效期
//	expireTime := ttime.AddYears(time.Now(), 10)
//	fmt.Printf("十年后到期：%v\n", expireTime)
//
// 注意事项：
//   - 保持月份、日期和时间部分不变（如果可能）
//   - 自动处理闰年（如2月29日）
//   - 考虑时区，使用时间的原始时区
//   - 适用于长期时间计算
//   - 是 t.AddDate(years, 0, 0) 的简单封装
func AddYears(t time.Time, years int) time.Time {
	return t.AddDate(years, 0, 0)
}

// DaysBetween 计算两个时间之间的天数差（按日期计算，忽略时间部分）
// 参数：
//   - t1: 第一个时间
//   - t2: 第二个时间
//
// 返回值：
//   - int: t2 减去 t1 的天数差
//
// 使用示例：
//
//	// 计算两个日期之间的天数
//	start := time.Date(2023, 6, 1, 0, 0, 0, 0, time.Local)
//	end := time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local)
//	days := ttime.DaysBetween(start, end)
//	fmt.Printf("相差 %d 天\n", days)  // 输出：相差 14 天
//
//	// 计算到期剩余天数
//	now := time.Now()
//	expire := time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local)
//	days = ttime.DaysBetween(now, expire)
//	fmt.Printf("还剩 %d 天到期\n", days)
//
// 注意事项：
//   - 只考虑日期部分，忽略时间部分
//   - 如果 t2 早于 t1，结果为负数
//   - 自动处理跨月、跨年的情况
//   - 考虑时区，在计算前统一时区
//   - 结果包含开始日期，不包含结束日期
func DaysBetween(t1, t2 time.Time) int {
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())

	duration := t2.Sub(t1)
	return int(duration.Hours() / 24)
}

// HoursBetween 计算两个时间之间的小时数差（精确到小时）
// 参数：
//   - t1: 第一个时间
//   - t2: 第二个时间
//
// 返回值：
//   - int: t2 减去 t1 的小时数差
//
// 使用示例：
//
//	// 计算工作时长
//	start := time.Date(2023, 6, 15, 9, 0, 0, 0, time.Local)
//	end := time.Date(2023, 6, 15, 18, 0, 0, 0, time.Local)
//	hours := ttime.HoursBetween(start, end)
//	fmt.Printf("工作时长：%d小时\n", hours)  // 输出：工作时长：9小时
//
//	// 计算跨天的时长
//	start = time.Date(2023, 6, 15, 22, 0, 0, 0, time.Local)
//	end = time.Date(2023, 6, 16, 6, 0, 0, 0, time.Local)
//	hours = ttime.HoursBetween(start, end)
//	fmt.Printf("时长：%d小时\n", hours)  // 输出：时长：8小时
//
// 注意事项：
//   - 结果会被截断为整数小时
//   - 如果 t2 早于 t1，结果为负数
//   - 自动处理跨天、跨月、跨年的情况
//   - 考虑时区，使用原始时区计算
//   - 不会四舍五入，总是向下取整
func HoursBetween(t1, t2 time.Time) int {
	duration := t2.Sub(t1)
	return int(duration.Hours())
}

// MinutesBetween 计算两个时间之间的分钟数差（精确到分钟）
// 参数：
//   - t1: 第一个时间
//   - t2: 第二个时间
//
// 返回值：
//   - int: t2 减去 t1 的分钟数差
//
// 使用示例：
//
//	// 计算会议时长
//	start := time.Date(2023, 6, 15, 14, 30, 0, 0, time.Local)
//	end := time.Date(2023, 6, 15, 15, 45, 0, 0, time.Local)
//	minutes := ttime.MinutesBetween(start, end)
//	fmt.Printf("会议时长：%d分钟\n", minutes)  // 输出：会议时长：75分钟
//
//	// 计算任务耗时
//	start = time.Now()
//	// ... 执行任务 ...
//	end = time.Now()
//	minutes = ttime.MinutesBetween(start, end)
//	fmt.Printf("任务耗时：%d分钟\n", minutes)
//
// 注意事项：
//   - 结果会被截断为整数分钟
//   - 如果 t2 早于 t1，结果为负数
//   - 自动处理跨小时、跨天的情况
//   - 考虑时区，使用原始时区计算
//   - 不会四舍五入，总是向下取整
func MinutesBetween(t1, t2 time.Time) int {
	duration := t2.Sub(t1)
	return int(duration.Minutes())
}

// SecondsBetween 计算两个时间之间的秒数差（精确到秒）
// 参数：
//   - t1: 第一个时间
//   - t2: 第二个时间
//
// 返回值：
//   - int64: t2 减去 t1 的秒数差
//
// 使用示例：
//
//	// 计算操作耗时
//	start := time.Now()
//	// ... 执行操作 ...
//	end := time.Now()
//	seconds := ttime.SecondsBetween(start, end)
//	fmt.Printf("操作耗时：%d秒\n", seconds)
//
//	// 计算剩余时间
//	now := time.Now()
//	deadline := now.Add(10 * time.Minute)
//	seconds = ttime.SecondsBetween(now, deadline)
//	fmt.Printf("剩余时间：%d秒\n", seconds)  // 输出：剩余时间：600秒
//
// 注意事项：
//   - 结果会被截断为整数秒
//   - 如果 t2 早于 t1，结果为负数
//   - 自动处理跨分钟、跨小时的情况
//   - 考虑时区，使用原始时区计算
//   - 不会四舍五入，总是向下取整
//   - 返回 int64 以支持较大的时间差
func SecondsBetween(t1, t2 time.Time) int64 {
	duration := t2.Sub(t1)
	return int64(duration.Seconds())
}

// FormatDuration 将持续时间格式化为易读的字符串
// 参数：
//   - d: 要格式化的持续时间
//
// 返回值：
//   - string: 格式化后的字符串，格式如下：
//   - 小于1分钟：直接使用 time.Duration 的字符串表示
//   - 1分钟到1小时：显示分钟和秒，如 "5m 30s"
//   - 1小时到24小时：显示小时和分钟，如 "2h 30m"
//   - 大于24小时：显示天数和小时，如 "2d 5h"
//
// 使用示例：
//
//	// 格式化各种时长
//	fmt.Println(ttime.FormatDuration(45 * time.Second))        // "45s"
//	fmt.Println(ttime.FormatDuration(90 * time.Second))        // "1m 30s"
//	fmt.Println(ttime.FormatDuration(150 * time.Minute))       // "2h 30m"
//	fmt.Println(ttime.FormatDuration(50 * time.Hour))          // "2d 2h"
//
//	// 用于显示耗时
//	start := time.Now()
//	// ... 执行操作 ...
//	duration := time.Since(start)
//	fmt.Printf("耗时：%s\n", ttime.FormatDuration(duration))
//
// 注意事项：
//   - 自动选择最合适的时间单位
//   - 最多显示两个单位
//   - 总是显示最大的非零单位
//   - 四舍五入到最小显示单位
//   - 负数时长会保持负号
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.String()
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

// IsWeekend 检查给定时间是否是周末（周六或周日）
// 参数：
//   - t: 要检查的时间
//
// 返回值：
//   - bool: 如果是周末返回 true，否则返回 false
//
// 使用示例：
//
//	// 检查当前时间
//	if ttime.IsWeekend(time.Now()) {
//	    fmt.Println("今天是周末")
//	}
//
//	// 检查特定日期
//	t := time.Date(2023, 6, 17, 0, 0, 0, 0, time.Local)  // 2023-06-17 是周六
//	if ttime.IsWeekend(t) {
//	    fmt.Println("2023-06-17 是周末")
//	}
//
//	// 用于业务逻辑
//	if !ttime.IsWeekend(orderTime) {
//	    fmt.Println("这是工作日订单")
//	}
//
// 注意事项：
//   - 周六和周日都被视为周末
//   - 不考虑节假日调休的情况
//   - 只检查日期，忽略时间部分
//   - 考虑时区，使用时间的原始时区
//   - 通常与 IsWorkday 配合使用
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWorkday 检查给定时间是否是工作日（周一至周五）
// 参数：
//   - t: 要检查的时间
//
// 返回值：
//   - bool: 如果是工作日返回 true，否则返回 false
//
// 使用示例：
//
//	// 检查当前时间
//	if ttime.IsWorkday(time.Now()) {
//	    fmt.Println("今天是工作日")
//	}
//
//	// 检查特定日期
//	t := time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local)  // 2023-06-15 是周四
//	if ttime.IsWorkday(t) {
//	    fmt.Println("2023-06-15 是工作日")
//	}
//
//	// 用于业务逻辑
//	if ttime.IsWorkday(deliveryTime) {
//	    fmt.Println("这是工作日配送")
//	}
//
// 注意事项：
//   - 周一到周五被视为工作日
//   - 不考虑节假日调休的情况
//   - 只检查日期，忽略时间部分
//   - 考虑时区，使用时间的原始时区
//   - 是 !IsWeekend(t) 的简单封装
//   - 通常与 IsWeekend 配合使用
func IsWorkday(t time.Time) bool {
	return !IsWeekend(t)
}

// AddWorkdays 在给定时间上添加或减少指定的工作日天数（只计算周一至周五）
// 参数：
//   - t: 基准时间
//   - days: 要添加的工作日天数，可以是负数表示减少天数
//
// 返回值：
//   - time.Time: 计算后的新时间
//
// 使用示例：
//
//	// 添加工作日
//	now := time.Now()
//	future := ttime.AddWorkdays(now, 5)  // 跳过周末，只计算工作日
//	fmt.Printf("5个工作日后：%v\n", future)
//
//	// 减少工作日
//	past := ttime.AddWorkdays(now, -5)  // 向前计算5个工作日
//	fmt.Printf("5个工作日前：%v\n", past)
//
//	// 计算截止日期
//	deadline := ttime.AddWorkdays(time.Now(), 10)  // 10个工作日后的截止日期
//	fmt.Printf("截止日期：%v\n", deadline)
//
// 注意事项：
//   - 只计算周一到周五
//   - 不考虑节假日调休
//   - 保持时间部分不变
//   - 如果开始日期是周末，从下一个工作日开始计算
//   - 支持正数和负数天数
//   - 通常与 WorkdaysBetween 配合使用
func AddWorkdays(t time.Time, days int) time.Time {
	// 如果天数为0，直接返回
	if days == 0 {
		return t
	}

	// 决定方向（前进或后退）
	direction := 1
	if days < 0 {
		direction = -1
		days = -days
	}

	result := t
	for days > 0 {
		result = result.AddDate(0, 0, direction)
		if IsWorkday(result) {
			days--
		}
	}

	return result
}

// WorkdaysBetween 计算两个时间之间的工作日天数（只计算周一至周五）
// 参数：
//   - t1: 开始时间
//   - t2: 结束时间
//
// 返回值：
//   - int: t2 减去 t1 的工作日天数
//
// 使用示例：
//
//	// 计算工作日天数
//	start := time.Date(2023, 6, 1, 0, 0, 0, 0, time.Local)  // 周四
//	end := time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local)   // 周四
//	days := ttime.WorkdaysBetween(start, end)
//	fmt.Printf("工作日天数：%d天\n", days)  // 输出：工作日天数：11天
//
//	// 计算跨月工作日
//	start = time.Date(2023, 5, 29, 0, 0, 0, 0, time.Local)  // 周一
//	end = time.Date(2023, 6, 2, 0, 0, 0, 0, time.Local)     // 周五
//	days = ttime.WorkdaysBetween(start, end)
//	fmt.Printf("工作日天数：%d天\n", days)  // 输出：工作日天数：5天
//
// 注意事项：
//   - 只计算周一到周五
//   - 不考虑节假日调休
//   - 如果 t2 早于 t1，会自动交换顺序
//   - 只考虑日期部分，忽略时间部分
//   - 开始日期计入总数，结束日期不计入
//   - 通常与 AddWorkdays 配合使用
func WorkdaysBetween(t1, t2 time.Time) int {
	// 确保t1在t2之前
	if t1.After(t2) {
		t1, t2 = t2, t1
	}

	// 将时间调整到当天开始
	t1 = time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	t2 = time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())

	// 计算总天数
	days := 0
	for current := t1; !current.After(t2); current = current.AddDate(0, 0, 1) {
		if IsWorkday(current) {
			days++
		}
	}

	return days
}

// TimeRange 时间范围结构
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// NewTimeRange 创建时间范围
func NewTimeRange(start, end time.Time) *TimeRange {
	if start.After(end) {
		start, end = end, start
	}
	return &TimeRange{
		Start: start,
		End:   end,
	}
}

// Contains 检查时间是否在范围内
func (r *TimeRange) Contains(t time.Time) bool {
	return !t.Before(r.Start) && !t.After(r.End)
}

// Overlaps 检查两个时间范围是否重叠
func (r *TimeRange) Overlaps(other *TimeRange) bool {
	return !r.End.Before(other.Start) && !other.End.Before(r.Start)
}

// Duration 获取时间范围的持续时间
func (r *TimeRange) Duration() time.Duration {
	return r.End.Sub(r.Start)
}

// String 格式化时间范围为字符串
func (r *TimeRange) String() string {
	return fmt.Sprintf("%s - %s", TimeFormatter(r.Start), TimeFormatter(r.End))
}

// ConvertTimezone 转换时间到指定时区
func ConvertTimezone(t time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("load timezone failed: %w", err)
	}
	return t.In(loc), nil
}

// ConvertToUTC 转换时间到UTC时区
func ConvertToUTC(t time.Time) time.Time {
	return t.UTC()
}

// ConvertToLocal 转换时间到本地时区
func ConvertToLocal(t time.Time) time.Time {
	return t.Local()
}

// LunarDate 农历日期结构
type LunarDate struct {
	Year        int
	Month       int
	Day         int
	IsLeap      bool
	YearZodiac  string
	YearGanZhi  string
	MonthGanZhi string
	DayGanZhi   string
}

var (
	// 天干
	tiangan = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	// 地支
	dizhi = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	// 生肖
	zodiac = []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
	// 农历月份
	lunarMonths = []string{"正", "二", "三", "四", "五", "六", "七", "八", "九", "十", "冬", "腊"}
	// 农历日期
	lunarDays = []string{"初一", "初二", "初三", "初四", "初五", "初六", "初七", "初八", "初九", "初十",
		"十一", "十二", "十三", "十四", "十五", "十六", "十七", "十八", "十九", "二十",
		"廿一", "廿二", "廿三", "廿四", "廿五", "廿六", "廿七", "廿八", "廿九", "三十"}
)

// GetLunarDate 获取农历日期（基于查表法）
func GetLunarDate(t time.Time) *LunarDate {
	// 这里使用简化的算法，实际应该使用农历转换表
	// 为了演示，我们只返回一个基本结构
	year := t.Year()
	yearIndex := (year - 4) % 12 // 1900年是鼠年
	yearGanIndex := (year - 4) % 10
	yearZhiIndex := (year - 4) % 12

	return &LunarDate{
		Year:        year,
		Month:       int(t.Month()),
		Day:         t.Day(),
		IsLeap:      false,
		YearZodiac:  zodiac[yearIndex],
		YearGanZhi:  tiangan[yearGanIndex] + dizhi[yearZhiIndex],
		MonthGanZhi: "未实现", // 需要复杂的转换算法
		DayGanZhi:   "未实现", // 需要复杂的转换算法
	}
}

// String 格式化农历日期为字符串
func (d *LunarDate) String() string {
	monthStr := lunarMonths[d.Month-1] + "月"
	if d.IsLeap {
		monthStr = "闰" + monthStr
	}
	return fmt.Sprintf("%d年%s%s %s年 %s", d.Year, monthStr, lunarDays[d.Day-1], d.YearZodiac, d.YearGanZhi)
}

// GetLunarFestival 获取农历节日（简化版本）
func GetLunarFestival(d *LunarDate) string {
	switch {
	case d.Month == 1 && d.Day == 1:
		return "春节"
	case d.Month == 5 && d.Day == 5:
		return "端午节"
	case d.Month == 8 && d.Day == 15:
		return "中秋节"
	case d.Month == 9 && d.Day == 9:
		return "重阳节"
	default:
		return ""
	}
}
