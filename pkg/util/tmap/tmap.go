package tmap

import (
	"reflect"
	"strconv"
)

// GetString 安全地获取字符串值
func GetString(m map[string]interface{}, key string, defaultVal string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
		case float32, float64:
			return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', -1, 64)
		case bool:
			return strconv.FormatBool(v)
		default:
			return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
		}
	}
	return defaultVal
}

// GetInt 安全地获取整数值
func GetInt(m map[string]interface{}, key string, defaultVal int) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return int(reflect.ValueOf(v).Int())
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
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return reflect.ValueOf(v).Int()
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
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return float64(reflect.ValueOf(v).Int())
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
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return reflect.ValueOf(v).Int() != 0
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

// Exists 检查 map 中是否存在指定键
func Exists(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	_, exists := m[key]
	return exists
}
