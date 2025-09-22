package tsonic

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/bytedance/sonic"
)

// Config 定义JSON解析器的配置选项
type Config struct {
	UseNumber             bool   // 使用 sonic.Number 类型处理数字
	TagKey                string // 结构体标签键
	CaseSensitive         bool   // 大小写敏感
	DisallowUnknownFields bool   // 禁止未知字段
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	UseNumber:             true,
	TagKey:                "json",
	CaseSensitive:         true,
	DisallowUnknownFields: false,
}

// JSONUtil 提供JSON处理的核心功能
type JSONUtil struct {
	config Config
}

// New 创建一个新的JSON工具实例
func New(config ...Config) *JSONUtil {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return &JSONUtil{
		config: cfg,
	}
}

// Default 默认JSON工具实例
var Default = New()

// Parse 将JSON字节数据解析为map
func (j *JSONUtil) Parse(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := sonic.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}
	return result, nil
}

// ParseString 将JSON字符串解析为map
func (j *JSONUtil) ParseString(data string) (map[string]interface{}, error) {
	return j.Parse([]byte(data))
}

// ParseReader 从io.Reader读取并解析JSON数据
func (j *JSONUtil) ParseReader(reader io.Reader) (map[string]interface{}, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read data failed: %w", err)
	}
	return j.Parse(data)
}

// ParseRequest 解析HTTP请求体中的JSON数据
func (j *JSONUtil) ParseRequest(r *http.Request) (map[string]interface{}, error) {
	return j.ParseReader(r.Body)
}

// Unmarshal 将JSON数据解析到指定的结构体
func (j *JSONUtil) Unmarshal(data []byte, v interface{}) error {
	return sonic.Unmarshal(data, v)
}

// UnmarshalString 将JSON字符串解析到指定的结构体
func (j *JSONUtil) UnmarshalString(data string, v interface{}) error {
	return j.Unmarshal([]byte(data), v)
}

// Marshal 将任意数据序列化为JSON字节数组
func (j *JSONUtil) Marshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

// MarshalToString 将任意数据序列化为JSON字符串
func (j *JSONUtil) MarshalToString(v interface{}) (string, error) {
	bytes, err := j.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Get 从map中获取指定键的值
func (j *JSONUtil) Get(data map[string]interface{}, key string) (interface{}, bool) {
	value, exists := data[key]
	return value, exists
}

// GetInt64 从map中获取int64类型的值
func (j *JSONUtil) GetInt64(data map[string]interface{}, key string) (int64, bool) {
	value, exists := data[key]
	if !exists {
		return 0, false
	}
	return j.toInt64(value)
}

// GetInt 从map中获取int类型的值
func (j *JSONUtil) GetInt(data map[string]interface{}, key string) (int, bool) {
	value, exists := data[key]
	if !exists {
		return 0, false
	}
	return j.toInt(value)
}

// GetFloat64 从map中获取float64类型的值
func (j *JSONUtil) GetFloat64(data map[string]interface{}, key string) (float64, bool) {
	value, exists := data[key]
	if !exists {
		return 0, false
	}
	return j.toFloat64(value)
}

// GetString 从map中获取string类型的值
func (j *JSONUtil) GetString(data map[string]interface{}, key string) (string, bool) {
	value, exists := data[key]
	if !exists {
		return "", false
	}
	return j.toString(value)
}

// GetBool 从map中获取bool类型的值
func (j *JSONUtil) GetBool(data map[string]interface{}, key string) (bool, bool) {
	value, exists := data[key]
	if !exists {
		return false, false
	}
	return j.toBool(value)
}

// GetMap 从map中获取嵌套的map值
func (j *JSONUtil) GetMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	value, exists := data[key]
	if !exists {
		return nil, false
	}
	return j.toMap(value)
}

// GetArray 从map中获取数组值
func (j *JSONUtil) GetArray(data map[string]interface{}, key string) ([]interface{}, bool) {
	value, exists := data[key]
	if !exists {
		return nil, false
	}
	return j.toArray(value)
}

// GetStruct 从map中获取值并转换为指定的结构体
func (j *JSONUtil) GetStruct(data map[string]interface{}, key string, v interface{}) bool {
	value, exists := data[key]
	if !exists {
		return false
	}
	return j.toStruct(value, v)
}

// Validate 验证JSON数据的格式是否正确
func (j *JSONUtil) Validate(data []byte) error {
	if !sonic.Valid(data) {
		return fmt.Errorf("invalid JSON format")
	}
	return nil
}

// ValidateString 验证JSON字符串的格式是否正确
func (j *JSONUtil) ValidateString(data string) error {
	return j.Validate([]byte(data))
}

// ValidateRequired 验证map中是否包含所有必需的字段
func (j *JSONUtil) ValidateRequired(data map[string]interface{}, fields ...string) error {
	for _, field := range fields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}
	return nil
}

// ValidateType 验证map中指定字段的值类型是否符合预期
func (j *JSONUtil) ValidateType(data map[string]interface{}, field string, expectedType reflect.Kind) error {
	value, exists := data[field]
	if !exists {
		return fmt.Errorf("field '%s' is missing", field)
	}

	actualType := reflect.TypeOf(value).Kind()
	if actualType != expectedType {
		return fmt.Errorf("field '%s' expected type %v, got %v", field, expectedType, actualType)
	}

	return nil
}

// 类型转换辅助函数
func (j *JSONUtil) toInt64(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case int32:
		return int64(v), true
	case int16:
		return int64(v), true
	case int8:
		return int64(v), true
	case uint64:
		return int64(v), true
	case uint:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint8:
		return int64(v), true
	case float64:
		if v == float64(int64(v)) {
			return int64(v), true
		}
		return 0, false
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func (j *JSONUtil) toInt(value interface{}) (int, bool) {
	if int64Val, ok := j.toInt64(value); ok {
		return int(int64Val), true
	}
	return 0, false
}

func (j *JSONUtil) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func (j *JSONUtil) toString(value interface{}) (string, bool) {
	switch v := value.(type) {
	case string:
		return v, true
	case int64:
		return strconv.FormatInt(v, 10), true
	case int:
		return strconv.Itoa(v), true
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(v), true
	default:
		return "", false
	}
}

func (j *JSONUtil) toBool(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		if strings.ToLower(v) == "true" {
			return true, true
		}
		if strings.ToLower(v) == "false" {
			return false, true
		}
		if v == "1" {
			return true, true
		}
		if v == "0" {
			return false, true
		}
		return false, false
	case int64:
		return v != 0, true
	case int:
		return v != 0, true
	case float64:
		return v != 0, true
	default:
		return false, false
	}
}

func (j *JSONUtil) toMap(value interface{}) (map[string]interface{}, bool) {
	if mapVal, ok := value.(map[string]interface{}); ok {
		return mapVal, true
	}
	return nil, false
}

func (j *JSONUtil) toArray(value interface{}) ([]interface{}, bool) {
	if arrVal, ok := value.([]interface{}); ok {
		return arrVal, true
	}
	return nil, false
}

func (j *JSONUtil) toStruct(value interface{}, v interface{}) bool {
	bytes, err := sonic.Marshal(value)
	if err != nil {
		return false
	}
	return sonic.Unmarshal(bytes, v) == nil
}

// 便捷函数
func Parse(data []byte) (map[string]interface{}, error) {
	return Default.Parse(data)
}

func ParseString(data string) (map[string]interface{}, error) {
	return Default.ParseString(data)
}

func ParseReader(reader io.Reader) (map[string]interface{}, error) {
	return Default.ParseReader(reader)
}

func ParseRequest(r *http.Request) (map[string]interface{}, error) {
	return Default.ParseRequest(r)
}

func Unmarshal(data []byte, v interface{}) error {
	return Default.Unmarshal(data, v)
}

func UnmarshalString(data string, v interface{}) error {
	return Default.UnmarshalString(data, v)
}

func Marshal(v interface{}) ([]byte, error) {
	return Default.Marshal(v)
}

func MarshalToString(v interface{}) (string, error) {
	return Default.MarshalToString(v)
}

func Get(data map[string]interface{}, key string) (interface{}, bool) {
	return Default.Get(data, key)
}

func GetInt64(data map[string]interface{}, key string) (int64, bool) {
	return Default.GetInt64(data, key)
}

func GetInt(data map[string]interface{}, key string) (int, bool) {
	return Default.GetInt(data, key)
}

func GetFloat64(data map[string]interface{}, key string) (float64, bool) {
	return Default.GetFloat64(data, key)
}

func GetString(data map[string]interface{}, key string) (string, bool) {
	return Default.GetString(data, key)
}

func GetBool(data map[string]interface{}, key string) (bool, bool) {
	return Default.GetBool(data, key)
}

func GetMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	return Default.GetMap(data, key)
}

func GetArray(data map[string]interface{}, key string) ([]interface{}, bool) {
	return Default.GetArray(data, key)
}

func GetStruct(data map[string]interface{}, key string, v interface{}) bool {
	return Default.GetStruct(data, key, v)
}

func Validate(data []byte) error {
	return Default.Validate(data)
}

func ValidateString(data string) error {
	return Default.ValidateString(data)
}

func ValidateRequired(data map[string]interface{}, fields ...string) error {
	return Default.ValidateRequired(data, fields...)
}

func ValidateType(data map[string]interface{}, field string, expectedType reflect.Kind) error {
	return Default.ValidateType(data, field, expectedType)
}
