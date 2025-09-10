package tmap

import (
	"testing"
	"time"
)

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"str":   "hello",
		"int":   123,
		"float": 45.67,
		"bool":  true,
	}

	// 测试字符串类型
	if result := GetString(m, "str", ""); result != "hello" {
		t.Errorf("Expected 'hello', got '%s'", result)
	}

	// 测试 int 转 string
	if result := GetString(m, "int", ""); result != "123" {
		t.Errorf("Expected '123', got '%s'", result)
	}

	// 测试 float 转 string
	if result := GetString(m, "float", ""); result != "45.67" {
		t.Errorf("Expected '45.67', got '%s'", result)
	}

	// 测试 bool 转 string
	if result := GetString(m, "bool", ""); result != "true" {
		t.Errorf("Expected 'true', got '%s'", result)
	}

	// 测试不存在的键
	if result := GetString(m, "nonexistent", ""); result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}

	// 测试默认值
	if result := GetString(m, "nonexistent", "default"); result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}

	// 测试 nil map
	if result := GetString(nil, "key", "default"); result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}

func TestGetInt(t *testing.T) {
	m := map[string]interface{}{
		"int":    123,
		"str":    "456",
		"float":  78.9,
		"bool":   true,
		"bool_f": false,
	}

	// 测试 int 类型
	if result := GetInt(m, "int", 0); result != 123 {
		t.Errorf("Expected 123, got %d", result)
	}

	// 测试 string 转 int
	if result := GetInt(m, "str", 0); result != 456 {
		t.Errorf("Expected 456, got %d", result)
	}

	// 测试 float 转 int
	if result := GetInt(m, "float", 0); result != 78 {
		t.Errorf("Expected 78, got %d", result)
	}

	// 测试 bool 转 int
	if result := GetInt(m, "bool", 0); result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}

	if result := GetInt(m, "bool_f", 0); result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}

	// 测试不存在的键
	if result := GetInt(m, "nonexistent", 0); result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}

	// 测试默认值
	if result := GetInt(m, "nonexistent", 999); result != 999 {
		t.Errorf("Expected 999, got %d", result)
	}
}

func TestGetInt64(t *testing.T) {
	m := map[string]interface{}{
		"int":    123,
		"str":    "456",
		"float":  78.9,
		"bool":   true,
		"bool_f": false,
	}

	// 测试 int 类型
	if result := GetInt64(m, "int", 0); result != 123 {
		t.Errorf("Expected 123, got %d", result)
	}

	// 测试 string 转 int64
	if result := GetInt64(m, "str", 0); result != 456 {
		t.Errorf("Expected 456, got %d", result)
	}

	// 测试 float 转 int64
	if result := GetInt64(m, "float", 0); result != 78 {
		t.Errorf("Expected 78, got %d", result)
	}

	// 测试 bool 转 int64
	if result := GetInt64(m, "bool", 0); result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}

	if result := GetInt64(m, "bool_f", 0); result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}

	// 测试默认值
	if result := GetInt64(m, "nonexistent", 999); result != 999 {
		t.Errorf("Expected 999, got %d", result)
	}
}

func TestGetFloat64(t *testing.T) {
	m := map[string]interface{}{
		"int":   123,
		"str":   "45.67",
		"float": 78.9,
		"bool":  true,
	}

	// 测试 int 转 float64
	if result := GetFloat64(m, "int", 0); result != 123.0 {
		t.Errorf("Expected 123.0, got %f", result)
	}

	// 测试 string 转 float64
	if result := GetFloat64(m, "str", 0); result != 45.67 {
		t.Errorf("Expected 45.67, got %f", result)
	}

	// 测试 float64 类型
	if result := GetFloat64(m, "float", 0); result != 78.9 {
		t.Errorf("Expected 78.9, got %f", result)
	}

	// 测试 bool 转 float64
	if result := GetFloat64(m, "bool", 0); result != 1.0 {
		t.Errorf("Expected 1.0, got %f", result)
	}

	// 测试默认值
	if result := GetFloat64(m, "nonexistent", 99.9); result != 99.9 {
		t.Errorf("Expected 99.9, got %f", result)
	}
}

func TestGetBool(t *testing.T) {
	m := map[string]interface{}{
		"bool_t": true,
		"bool_f": false,
		"int_1":  1,
		"int_0":  0,
		"str_t":  "true",
		"str_f":  "false",
		"str_1":  "1",
		"str_0":  "0",
	}

	// 测试 bool 类型
	if result := GetBool(m, "bool_t", false); result != true {
		t.Errorf("Expected true, got %v", result)
	}

	if result := GetBool(m, "bool_f", true); result != false {
		t.Errorf("Expected false, got %v", result)
	}

	// 测试 int 转 bool
	if result := GetBool(m, "int_1", false); result != true {
		t.Errorf("Expected true, got %v", result)
	}

	if result := GetBool(m, "int_0", true); result != false {
		t.Errorf("Expected false, got %v", result)
	}

	// 测试 string 转 bool
	if result := GetBool(m, "str_t", false); result != true {
		t.Errorf("Expected true, got %v", result)
	}

	if result := GetBool(m, "str_f", true); result != false {
		t.Errorf("Expected false, got %v", result)
	}

	if result := GetBool(m, "str_1", false); result != true {
		t.Errorf("Expected true, got %v", result)
	}

	if result := GetBool(m, "str_0", true); result != false {
		t.Errorf("Expected false, got %v", result)
	}

	// 测试默认值
	if result := GetBool(m, "nonexistent", true); result != true {
		t.Errorf("Expected true, got %v", result)
	}
}

func TestGet(t *testing.T) {
	m := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
	}

	// 测试存在的键
	val, exists := Get(m, "key1")
	if !exists {
		t.Error("Expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("Expected 'value1', got '%v'", val)
	}

	// 测试不存在的键
	_, exists = Get(m, "nonexistent")
	if exists {
		t.Error("Expected key to not exist")
	}

	// 测试 nil map
	_, exists = Get(nil, "key")
	if exists {
		t.Error("Expected key to not exist in nil map")
	}
}

func TestSet(t *testing.T) {
	m := make(map[string]interface{})

	// 测试设置值
	Set(m, "key1", "value1")
	if m["key1"] != "value1" {
		t.Errorf("Expected 'value1', got '%v'", m["key1"])
	}

	// 测试 nil map
	Set(nil, "key", "value") // 应该不会 panic
}

func TestExists(t *testing.T) {
	m := map[string]interface{}{
		"key1": "value1",
	}

	// 测试存在的键
	if !Exists(m, "key1") {
		t.Error("Expected key1 to exist")
	}

	// 测试不存在的键
	if Exists(m, "nonexistent") {
		t.Error("Expected key to not exist")
	}

	// 测试 nil map
	if Exists(nil, "key") {
		t.Error("Expected key to not exist in nil map")
	}
}

func TestGetTime(t *testing.T) {
	now := time.Now()
	m := map[string]interface{}{
		"time":          now,
		"timestamp":     now.Unix(),
		"timestamp_str": "1640995200", // 2022-01-01 00:00:00
		"datetime_str":  "2022-01-01 12:30:45",
		"date_str":      "2022-01-01",
		"rfc3339_str":   "2022-01-01T12:30:45Z",
	}

	// 测试 time.Time 类型
	if result := GetTime(m, "time", time.Time{}); !result.Equal(now) {
		t.Errorf("Expected %v, got %v", now, result)
	}

	// 测试时间戳转换
	expectedTime := time.Unix(1640995200, 0)
	if result := GetTime(m, "timestamp_str", time.Time{}); !result.Equal(expectedTime) {
		t.Errorf("Expected %v, got %v", expectedTime, result)
	}

	// 测试日期时间字符串转换
	expectedDateTime, _ := time.Parse("2006-01-02 15:04:05", "2022-01-01 12:30:45")
	if result := GetTime(m, "datetime_str", time.Time{}); !result.Equal(expectedDateTime) {
		t.Errorf("Expected %v, got %v", expectedDateTime, result)
	}

	// 测试默认值
	defaultTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if result := GetTime(m, "nonexistent", defaultTime); !result.Equal(defaultTime) {
		t.Errorf("Expected %v, got %v", defaultTime, result)
	}
}

func TestGetTimestamp(t *testing.T) {
	now := time.Now()
	m := map[string]interface{}{
		"time":          now,
		"timestamp":     int64(1640995200),
		"timestamp_str": "1640995200",
		"datetime_str":  "2022-01-01 12:30:45",
	}

	// 测试 time.Time 类型
	if result := GetTimestamp(m, "time", 0); result != now.Unix() {
		t.Errorf("Expected %d, got %d", now.Unix(), result)
	}

	// 测试 int64 类型
	if result := GetTimestamp(m, "timestamp", 0); result != 1640995200 {
		t.Errorf("Expected 1640995200, got %d", result)
	}

	// 测试字符串时间戳
	if result := GetTimestamp(m, "timestamp_str", 0); result != 1640995200 {
		t.Errorf("Expected 1640995200, got %d", result)
	}

	// 测试默认值
	if result := GetTimestamp(m, "nonexistent", 999); result != 999 {
		t.Errorf("Expected 999, got %d", result)
	}
}

func TestGetTimestampMilli(t *testing.T) {
	now := time.Now()
	m := map[string]interface{}{
		"time":            now,
		"timestamp_sec":   int64(1640995200),    // 秒级时间戳
		"timestamp_milli": int64(1640995200000), // 毫秒级时间戳
		"timestamp_str":   "1640995200",
	}

	// 测试 time.Time 类型
	if result := GetTimestampMilli(m, "time", 0); result != now.UnixMilli() {
		t.Errorf("Expected %d, got %d", now.UnixMilli(), result)
	}

	// 测试秒级时间戳转换为毫秒
	if result := GetTimestampMilli(m, "timestamp_sec", 0); result != 1640995200000 {
		t.Errorf("Expected 1640995200000, got %d", result)
	}

	// 测试毫秒级时间戳
	if result := GetTimestampMilli(m, "timestamp_milli", 0); result != 1640995200000 {
		t.Errorf("Expected 1640995200000, got %d", result)
	}

	// 测试字符串时间戳
	if result := GetTimestampMilli(m, "timestamp_str", 0); result != 1640995200000 {
		t.Errorf("Expected 1640995200000, got %d", result)
	}

	// 测试默认值
	if result := GetTimestampMilli(m, "nonexistent", 999); result != 999 {
		t.Errorf("Expected 999, got %d", result)
	}
}

func TestGetDateTime(t *testing.T) {
	now := time.Now()
	m := map[string]interface{}{
		"time":         now,
		"timestamp":    int64(1640995200), // 2022-01-01 00:00:00
		"datetime_str": "2022-01-01 12:30:45",
		"date_str":     "2022-01-01",
		"rfc3339_str":  "2022-01-01T12:30:45Z",
	}

	// 测试 time.Time 类型
	expected := now.Format("2006-01-02 15:04:05")
	if result := GetDateTime(m, "time", ""); result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试时间戳转换
	expectedTime := time.Unix(1640995200, 0)
	expected = expectedTime.Format("2006-01-02 15:04:05")
	if result := GetDateTime(m, "timestamp", ""); result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试日期时间字符串
	expected = "2022-01-01 12:30:45"
	if result := GetDateTime(m, "datetime_str", ""); result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试默认值
	if result := GetDateTime(m, "nonexistent", "default"); result != "default" {
		t.Errorf("Expected 'default', got '%s'", result)
	}
}
