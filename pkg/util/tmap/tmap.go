package tmap

import (
	"reflect"
	"strconv"
	"time"
)

// GetString 安全地获取字符串值
func GetString(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case int, int8, int16, int32, int64:
			return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
		case uint, uint8, uint16, uint32, uint64:
			return strconv.FormatUint(reflect.ValueOf(v).Uint(), 10)
		case float32, float64:
			return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', -1, 64)
		case bool:
			return strconv.FormatBool(v)
		}
	}
	return defaultVal
}

// GetInt 安全地获取整数值
func GetInt(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int, int8, int16, int32, int64:
			return int(reflect.ValueOf(v).Int())
		case uint, uint8, uint16, uint32, uint64:
			return int(reflect.ValueOf(v).Uint())
		case float32, float64:
			return int(reflect.ValueOf(v).Float())
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		case bool:
			if v {
				return 1
			}
		}
	}
	return defaultVal
}

// GetInt64 安全地获取整数值，统一返回int64
func GetInt64(m map[string]interface{}, key string, defaultVal int64) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int, int8, int16, int32, int64:
			return reflect.ValueOf(v).Int()
		case uint, uint8, uint16, uint32, uint64:
			return int64(reflect.ValueOf(v).Uint())
		case float32, float64:
			return int64(reflect.ValueOf(v).Float())
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		case bool:
			if v {
				return 1
			}
		}
	}
	return defaultVal
}

// GetFloat64 安全地获取浮点数值
func GetFloat64(m map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float32, float64:
			return reflect.ValueOf(v).Float()
		case int, int8, int16, int32, int64:
			return float64(reflect.ValueOf(v).Int())
		case uint, uint8, uint16, uint32, uint64:
			return float64(reflect.ValueOf(v).Uint())
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		case bool:
			if v {
				return 1.0
			}
		}
	}
	return defaultVal
}

// GetBool 安全地获取布尔值
func GetBool(m map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case bool:
			return v
		case int, int8, int16, int32, int64:
			return reflect.ValueOf(v).Int() != 0
		case uint, uint8, uint16, uint32, uint64:
			return reflect.ValueOf(v).Uint() != 0
		case float32, float64:
			return reflect.ValueOf(v).Float() != 0
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b
			}
			// 检查常见的 true/false 值
			switch v {
			case "true", "1", "yes", "on", "TRUE", "True":
				return true
			case "false", "0", "no", "off", "FALSE", "False":
				return false
			}
		}
	}
	return defaultVal
}

// Get 从 map 中获取原始值
func Get(m map[string]interface{}, key string) (interface{}, bool) {
	if m == nil {
		return nil, false
	}
	val, exists := m[key]
	return val, exists
}

// Set 设置 map 中的值
func Set(m map[string]interface{}, key string, value interface{}) {
	if m == nil {
		return
	}
	m[key] = value
}

// GetTime 安全地获取时间值
func GetTime(m map[string]interface{}, key string, defaultVal time.Time) time.Time {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case time.Time:
			return v
		case int, int8, int16, int32, int64:
			// 时间戳转换（秒）
			timestamp := reflect.ValueOf(v).Int()
			return time.Unix(timestamp, 0)
		case uint, uint8, uint16, uint32, uint64:
			// 时间戳转换（秒）
			timestamp := int64(reflect.ValueOf(v).Uint())
			return time.Unix(timestamp, 0)
		case float32, float64:
			// 时间戳转换（秒，支持小数）
			timestamp := reflect.ValueOf(v).Float()
			return time.Unix(int64(timestamp), 0)
		case string:
			// 尝试解析时间字符串
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t
			}
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				return t
			}
			if t, err := time.Parse("2006-01-02", v); err == nil {
				return t
			}
			// 尝试解析时间戳字符串
			if timestamp, err := strconv.ParseInt(v, 10, 64); err == nil {
				return time.Unix(timestamp, 0)
			}
		}
	}
	return defaultVal
}

// GetTimestamp 安全地获取时间戳（秒）
func GetTimestamp(m map[string]interface{}, key string, defaultVal int64) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case time.Time:
			return v.Unix()
		case int, int8, int16, int32, int64:
			return reflect.ValueOf(v).Int()
		case uint, uint8, uint16, uint32, uint64:
			return int64(reflect.ValueOf(v).Uint())
		case float32, float64:
			return int64(reflect.ValueOf(v).Float())
		case string:
			// 尝试解析时间字符串
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t.Unix()
			}
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				return t.Unix()
			}
			if t, err := time.Parse("2006-01-02", v); err == nil {
				return t.Unix()
			}
			// 尝试解析时间戳字符串
			if timestamp, err := strconv.ParseInt(v, 10, 64); err == nil {
				return timestamp
			}
		}
	}
	return defaultVal
}

// GetTimestampMilli 安全地获取毫秒时间戳
func GetTimestampMilli(m map[string]interface{}, key string, defaultVal int64) int64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case time.Time:
			return v.UnixMilli()
		case int, int8, int16, int32, int64:
			timestamp := reflect.ValueOf(v).Int()
			// 如果时间戳小于 1e12，认为是秒级时间戳，需要转换为毫秒
			if timestamp < 1e12 {
				return timestamp * 1000
			}
			return timestamp
		case uint, uint8, uint16, uint32, uint64:
			timestamp := int64(reflect.ValueOf(v).Uint())
			// 如果时间戳小于 1e12，认为是秒级时间戳，需要转换为毫秒
			if timestamp < 1e12 {
				return timestamp * 1000
			}
			return timestamp
		case float32, float64:
			timestamp := int64(reflect.ValueOf(v).Float())
			// 如果时间戳小于 1e12，认为是秒级时间戳，需要转换为毫秒
			if timestamp < 1e12 {
				return timestamp * 1000
			}
			return timestamp
		case string:
			// 尝试解析时间字符串
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t.UnixMilli()
			}
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				return t.UnixMilli()
			}
			if t, err := time.Parse("2006-01-02", v); err == nil {
				return t.UnixMilli()
			}
			// 尝试解析时间戳字符串
			if timestamp, err := strconv.ParseInt(v, 10, 64); err == nil {
				// 如果时间戳小于 1e12，认为是秒级时间戳，需要转换为毫秒
				if timestamp < 1e12 {
					return timestamp * 1000
				}
				return timestamp
			}
		}
	}
	return defaultVal
}

// GetDateTime 安全地获取日期时间字符串
func GetDateTime(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case time.Time:
			return v.Format("2006-01-02 15:04:05")
		case int, int8, int16, int32, int64:
			// 时间戳转换
			timestamp := reflect.ValueOf(v).Int()
			return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
		case uint, uint8, uint16, uint32, uint64:
			// 时间戳转换
			timestamp := int64(reflect.ValueOf(v).Uint())
			return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
		case float32, float64:
			// 时间戳转换
			timestamp := reflect.ValueOf(v).Float()
			return time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
		case string:
			// 尝试解析时间字符串并重新格式化
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t.Format("2006-01-02 15:04:05")
			}
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				return t.Format("2006-01-02 15:04:05")
			}
			if t, err := time.Parse("2006-01-02", v); err == nil {
				return t.Format("2006-01-02 15:04:05")
			}
			// 尝试解析时间戳字符串
			if timestamp, err := strconv.ParseInt(v, 10, 64); err == nil {
				return time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
			}
			// 如果无法解析，直接返回原字符串
			return v
		}
	}
	return defaultVal
}

// Exists 检查 map 中是否存在指定键
func Exists(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	_, exists := m[key]
	return exists
}
