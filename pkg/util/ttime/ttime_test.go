package ttime

import (
	"testing"
	"time"
)

func TestGetUnixMilliSeconds(t *testing.T) {
	// 获取当前时间的毫秒时间戳
	ms := GetUnixMilliSeconds()

	// 验证时间戳是否在合理范围内
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if ms < now-1000 || ms > now+1000 {
		t.Errorf("GetUnixMilliSeconds() = %v, want value between %v and %v", ms, now-1000, now+1000)
	}
}

func TestGetUnixSeconds(t *testing.T) {
	// 获取当前时间的秒时间戳
	s := GetUnixSeconds()

	// 验证时间戳是否在合理范围内
	now := time.Now().Unix()
	if s < now-1 || s > now+1 {
		t.Errorf("GetUnixSeconds() = %v, want value between %v and %v", s, now-1, now+1)
	}
}

func TestSeconds2date(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		format    string
		expected  string
	}{
		{
			name:      "标准格式",
			timestamp: 1609459200, // 2021-01-01 00:00:00
			format:    "2006-01-02 15:04:05",
			expected:  "2021-01-01 00:00:00",
		},
		{
			name:      "自定义格式",
			timestamp: 1609459200,
			format:    "2006年01月02日",
			expected:  "2021年01月01日",
		},
		{
			name:      "只显示时间",
			timestamp: 1609459200,
			format:    "15:04:05",
			expected:  "00:00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Seconds2date(tt.timestamp, tt.format)
			if result != tt.expected {
				t.Errorf("Seconds2date() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMilliseconds2date(t *testing.T) {
	tests := []struct {
		name         string
		milliseconds int64
		format       string
		expected     string
	}{
		{
			name:         "标准格式",
			milliseconds: 1609459200000, // 2021-01-01 00:00:00.000
			format:       "2006-01-02 15:04:05.000",
			expected:     "2021-01-01 00:00:00.000",
		},
		{
			name:         "不显示毫秒",
			milliseconds: 1609459200000,
			format:       "2006-01-02 15:04:05",
			expected:     "2021-01-01 00:00:00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Milliseconds2date(tt.milliseconds, tt.format)
			if result != tt.expected {
				t.Errorf("Milliseconds2date() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimeFormatter(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "零时区时间",
			time:     time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2021-01-01 00:00:00",
		},
		{
			name:     "带毫秒时间",
			time:     time.Date(2021, 1, 1, 12, 34, 56, 789000000, time.UTC),
			expected: "2021-01-01 12:34:56",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TimeFormatter(tt.time)
			if result != tt.expected {
				t.Errorf("TimeFormatter() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{
			name:     "标准时间单位",
			input:    "2h30m",
			expected: 2*time.Hour + 30*time.Minute,
			wantErr:  false,
		},
		{
			name:     "天数",
			input:    "2d12h",
			expected: 60 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "组合格式",
			input:    "1d6h30m15s",
			expected: 30*time.Hour + 30*time.Minute + 15*time.Second,
			wantErr:  false,
		},
		{
			name:     "纯数字（秒）",
			input:    "3600",
			expected: time.Hour,
			wantErr:  false,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "ISO 8601格式",
			input:    []byte(`"2023-04-05T15:30:45.123456Z"`),
			expected: `"2023-04-05 15:30:45.123"`,
		},
		{
			name:     "带时区偏移",
			input:    []byte(`"2023-04-05T15:30:45.123456+08:00"`),
			expected: `"2023-04-05 15:30:45.123"`,
		},
		{
			name:     "非ISO格式",
			input:    []byte(`"2023-04-05 15:30:45"`),
			expected: `"2023-04-05 15:30:45"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseTime(tt.input)
			if result != tt.expected {
				t.Errorf("ParseTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToday(t *testing.T) {
	today := Today()
	now := time.Now()

	// 检查日期部分是否相同
	if today.Year() != now.Year() || today.Month() != now.Month() || today.Day() != now.Day() {
		t.Errorf("Today() date = %v, want %v", today, now)
	}

	// 检查时间部分是否为零点
	if today.Hour() != 0 || today.Minute() != 0 || today.Second() != 0 {
		t.Errorf("Today() time = %v, want 00:00:00", today)
	}
}

func TestTomorrow(t *testing.T) {
	tomorrow := Tomorrow()
	today := Today()
	expected := today.AddDate(0, 0, 1)

	if !tomorrow.Equal(expected) {
		t.Errorf("Tomorrow() = %v, want %v", tomorrow, expected)
	}
}

func TestYesterday(t *testing.T) {
	yesterday := Yesterday()
	today := Today()
	expected := today.AddDate(0, 0, -1)

	if !yesterday.Equal(expected) {
		t.Errorf("Yesterday() = %v, want %v", yesterday, expected)
	}
}

func TestIsToday(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{
			name:     "今天",
			time:     time.Now(),
			expected: true,
		},
		{
			name:     "昨天",
			time:     time.Now().AddDate(0, 0, -1),
			expected: false,
		},
		{
			name:     "明天",
			time:     time.Now().AddDate(0, 0, 1),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsToday(tt.time)
			if result != tt.expected {
				t.Errorf("IsToday() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsThisWeek(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{
			name:     "本周",
			time:     time.Now(),
			expected: true,
		},
		{
			name:     "上周",
			time:     time.Now().AddDate(0, 0, -7),
			expected: false,
		},
		{
			name:     "下周",
			time:     time.Now().AddDate(0, 0, 7),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsThisWeek(tt.time)
			if result != tt.expected {
				t.Errorf("IsThisWeek() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsThisMonth(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{
			name:     "本月",
			time:     time.Now(),
			expected: true,
		},
		{
			name:     "上月",
			time:     time.Now().AddDate(0, -1, 0),
			expected: false,
		},
		{
			name:     "下月",
			time:     time.Now().AddDate(0, 1, 0),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsThisMonth(tt.time)
			if result != tt.expected {
				t.Errorf("IsThisMonth() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "小于1分钟",
			duration: 45 * time.Second,
			expected: "45s",
		},
		{
			name:     "分钟和秒",
			duration: 90 * time.Second,
			expected: "1m 30s",
		},
		{
			name:     "小时和分钟",
			duration: 150 * time.Minute,
			expected: "2h 30m",
		},
		{
			name:     "天和小时",
			duration: 50 * time.Hour,
			expected: "2d 2h",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsWeekend(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		expected bool
	}{
		{
			name:     "周六",
			time:     time.Date(2023, 6, 17, 0, 0, 0, 0, time.Local),
			expected: true,
		},
		{
			name:     "周日",
			time:     time.Date(2023, 6, 18, 0, 0, 0, 0, time.Local),
			expected: true,
		},
		{
			name:     "工作日",
			time:     time.Date(2023, 6, 16, 0, 0, 0, 0, time.Local),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWeekend(tt.time)
			if result != tt.expected {
				t.Errorf("IsWeekend() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAddWorkdays(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		days     int
		expected time.Time
	}{
		{
			name:     "添加工作日",
			start:    time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local), // 周四
			days:     5,
			expected: time.Date(2023, 6, 22, 0, 0, 0, 0, time.Local), // 下周四
		},
		{
			name:     "减少工作日",
			start:    time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local),
			days:     -5,
			expected: time.Date(2023, 6, 8, 0, 0, 0, 0, time.Local),
		},
		{
			name:     "零天数",
			start:    time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local),
			days:     0,
			expected: time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddWorkdays(tt.start, tt.days)
			if !result.Equal(tt.expected) {
				t.Errorf("AddWorkdays() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWorkdaysBetween(t *testing.T) {
	tests := []struct {
		name     string
		t1       time.Time
		t2       time.Time
		expected int
	}{
		{
			name:     "同一周工作日",
			t1:       time.Date(2023, 6, 12, 0, 0, 0, 0, time.Local), // 周一
			t2:       time.Date(2023, 6, 16, 0, 0, 0, 0, time.Local), // 周五
			expected: 5,
		},
		{
			name:     "跨周工作日",
			t1:       time.Date(2023, 6, 12, 0, 0, 0, 0, time.Local), // 周一
			t2:       time.Date(2023, 6, 19, 0, 0, 0, 0, time.Local), // 下周一
			expected: 6,
		},
		{
			name:     "相同日期",
			t1:       time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local),
			t2:       time.Date(2023, 6, 15, 0, 0, 0, 0, time.Local),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WorkdaysBetween(tt.t1, tt.t2)
			if result != tt.expected {
				t.Errorf("WorkdaysBetween() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTimeRange(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Hour)
	earlier := now.Add(-time.Hour)

	t.Run("Contains", func(t *testing.T) {
		tr := NewTimeRange(now, later)
		if !tr.Contains(now.Add(30 * time.Minute)) {
			t.Error("TimeRange.Contains() failed for time within range")
		}
		if tr.Contains(earlier) {
			t.Error("TimeRange.Contains() failed for time before range")
		}
	})

	t.Run("Overlaps", func(t *testing.T) {
		tr1 := NewTimeRange(now, later)
		tr2 := NewTimeRange(now.Add(30*time.Minute), later.Add(time.Hour))
		if !tr1.Overlaps(tr2) {
			t.Error("TimeRange.Overlaps() failed for overlapping ranges")
		}
	})

	t.Run("Duration", func(t *testing.T) {
		tr := NewTimeRange(now, later)
		if tr.Duration() != time.Hour {
			t.Error("TimeRange.Duration() returned incorrect duration")
		}
	})
}
