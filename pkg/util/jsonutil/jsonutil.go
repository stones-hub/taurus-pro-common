package jsonutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// Config JSON解析配置
type Config struct {
	UseNumber              bool   // 使用json.Number类型
	TagKey                 string // 结构体标签键
	OnlyTaggedField        bool   // 只处理有标签的字段
	ValidateJsonRawMessage bool   // 验证RawMessage
	CaseSensitive          bool   // 大小写敏感
	DisallowUnknownFields  bool   // 禁止未知字段
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	UseNumber:              true,
	TagKey:                 "json",
	OnlyTaggedField:        false,
	ValidateJsonRawMessage: true,
	CaseSensitive:          true,
	DisallowUnknownFields:  false,
}

// JSONUtil JSON工具实例
type JSONUtil struct {
	config Config
	api    jsoniter.API
}

// New 创建新的JSON工具实例
func New(config ...Config) *JSONUtil {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	// 构建json-iterator配置
	jsonConfig := jsoniter.Config{
		UseNumber:              cfg.UseNumber,
		TagKey:                 cfg.TagKey,
		OnlyTaggedField:        cfg.OnlyTaggedField,
		ValidateJsonRawMessage: cfg.ValidateJsonRawMessage,
		CaseSensitive:          cfg.CaseSensitive,
		DisallowUnknownFields:  cfg.DisallowUnknownFields,
	}

	return &JSONUtil{
		config: cfg,
		api:    jsonConfig.Froze(),
	}
}

// Default 默认JSON工具实例
var Default = New()

// Parse 解析JSON到map
func (j *JSONUtil) Parse(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := j.api.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	if j.config.UseNumber {
		return j.convertNumbers(result), nil
	}
	return result, nil
}

// ParseString 解析JSON字符串
func (j *JSONUtil) ParseString(data string) (map[string]interface{}, error) {
	return j.Parse([]byte(data))
}

// ParseReader 从io.Reader解析JSON
func (j *JSONUtil) ParseReader(reader io.Reader) (map[string]interface{}, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read data failed: %w", err)
	}
	return j.Parse(data)
}

// ParseRequest 解析HTTP请求的JSON
func (j *JSONUtil) ParseRequest(r *http.Request) (map[string]interface{}, error) {
	return j.ParseReader(r.Body)
}

// Unmarshal 解析JSON到结构体
func (j *JSONUtil) Unmarshal(data []byte, v interface{}) error {
	return j.api.Unmarshal(data, v)
}

// UnmarshalString 解析JSON字符串到结构体
func (j *JSONUtil) UnmarshalString(data string, v interface{}) error {
	return j.Unmarshal([]byte(data), v)
}

// Marshal 将数据序列化为JSON
func (j *JSONUtil) Marshal(v interface{}) ([]byte, error) {
	cleaned := cleanForMarshal(v)
	return json.Marshal(cleaned)
}

// MarshalToString 将数据序列化为JSON字符串
func (j *JSONUtil) MarshalToString(v interface{}) (string, error) {
	cleaned := cleanForMarshal(v)
	bytes, err := json.Marshal(cleaned)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Get 获取任意类型的值
func (j *JSONUtil) Get(data map[string]interface{}, key string) (interface{}, bool) {
	value, exists := data[key]
	return value, exists
}

// GetInt64 获取int64值
func (j *JSONUtil) GetInt64(data map[string]interface{}, key string) (int64, bool) {
	value, exists := data[key]
	if !exists {
		return 0, false
	}
	return j.toInt64(value)
}

// GetInt 获取int值
func (j *JSONUtil) GetInt(data map[string]interface{}, key string) (int, bool) {
	value, exists := data[key]
	if !exists {
		return 0, false
	}
	return j.toInt(value)
}

// GetFloat64 获取float64值
func (j *JSONUtil) GetFloat64(data map[string]interface{}, key string) (float64, bool) {
	value, exists := data[key]
	if !exists {
		return 0, false
	}
	return j.toFloat64(value)
}

// GetString 获取字符串值
func (j *JSONUtil) GetString(data map[string]interface{}, key string) (string, bool) {
	value, exists := data[key]
	if !exists {
		return "", false
	}
	return j.toString(value)
}

// GetBool 获取布尔值
func (j *JSONUtil) GetBool(data map[string]interface{}, key string) (bool, bool) {
	value, exists := data[key]
	if !exists {
		return false, false
	}
	return j.toBool(value)
}

// GetMap 获取map值
func (j *JSONUtil) GetMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	value, exists := data[key]
	if !exists {
		return nil, false
	}
	return j.toMap(value)
}

// GetArray 获取数组值
func (j *JSONUtil) GetArray(data map[string]interface{}, key string) ([]interface{}, bool) {
	value, exists := data[key]
	if !exists {
		return nil, false
	}
	return j.toArray(value)
}

// GetStruct 获取并转换为结构体
func (j *JSONUtil) GetStruct(data map[string]interface{}, key string, v interface{}) bool {
	value, exists := data[key]
	if !exists {
		return false
	}
	return j.toStruct(value, v)
}

// Validate 验证JSON格式
func (j *JSONUtil) Validate(data []byte) error {
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON format")
	}
	return nil
}

// ValidateString 验证JSON字符串格式
func (j *JSONUtil) ValidateString(data string) error {
	return j.Validate([]byte(data))
}

// ValidateRequired 验证你想要的字段有没有
func (j *JSONUtil) ValidateRequired(data map[string]interface{}, fields ...string) error {
	for _, field := range fields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}
	return nil
}

// ValidateType 验证你想要的字段类型是否正确
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

// convertNumbers 智能转换数字类型
func (j *JSONUtil) convertNumbers(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		result[key] = j.convertValue(value)
	}
	return result
}

// convertValue 转换单个值
func (j *JSONUtil) convertValue(value interface{}) interface{} {
	switch v := value.(type) {
	case json.Number:
		return j.convertNumber(v)
	case map[string]interface{}:
		return j.convertNumbers(v)
	case []interface{}:
		return j.convertArray(v)
	default:
		return v
	}
}

// convertNumber 转换Number类型
func (j *JSONUtil) convertNumber(num json.Number) interface{} {
	// 优先尝试int64
	if intVal, err := num.Int64(); err == nil {
		return intVal
	}

	// 再尝试float64
	if floatVal, err := num.Float64(); err == nil {
		// 检查是否是整数
		if floatVal == float64(int64(floatVal)) {
			return int64(floatVal)
		}
		return floatVal
	}

	// 最后返回字符串
	return num.String()
}

// convertArray 转换数组
func (j *JSONUtil) convertArray(arr []interface{}) []interface{} {
	result := make([]interface{}, len(arr))
	for i, value := range arr {
		result[i] = j.convertValue(value)
	}
	return result
}

// toInt64 转换为int64
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
	case json.Number:
		if intVal, err := v.Int64(); err == nil {
			return intVal, true
		}
		return 0, false
	case string:
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toInt 转换为int
func (j *JSONUtil) toInt(value interface{}) (int, bool) {
	if int64Val, ok := j.toInt64(value); ok {
		return int(int64Val), true
	}
	return 0, false
}

// toFloat64 转换为float64
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
	case json.Number:
		if floatVal, err := v.Float64(); err == nil {
			return floatVal, true
		}
		return 0, false
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toString 转换为字符串
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
	case json.Number:
		return v.String(), true
	case bool:
		return strconv.FormatBool(v), true
	default:
		return "", false
	}
}

// toBool 转换为布尔值
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

// toMap 转换为map
func (j *JSONUtil) toMap(value interface{}) (map[string]interface{}, bool) {
	if mapVal, ok := value.(map[string]interface{}); ok {
		return mapVal, true
	}
	return nil, false
}

// toArray 转换为数组
func (j *JSONUtil) toArray(value interface{}) ([]interface{}, bool) {
	if arrVal, ok := value.([]interface{}); ok {
		return arrVal, true
	}
	return nil, false
}

// toStruct 转换为结构体
func (j *JSONUtil) toStruct(value interface{}, v interface{}) bool {
	cleaned := cleanForMarshal(value)
	bytes, err := json.Marshal(cleaned)
	if err != nil {
		return false
	}
	return j.api.Unmarshal(bytes, v) == nil
}

// cleanForMarshal 清理数据以避免 json-iterator 的 panic
func cleanForMarshal(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case json.Number:
		if intVal, err := val.Int64(); err == nil {
			return intVal
		}
		if floatVal, err := val.Float64(); err == nil {
			return floatVal
		}
		return val.String()
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range val {
			if v != nil {
				result[k] = cleanForMarshal(v)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, 0, len(val))
		for _, v := range val {
			if v != nil {
				result = append(result, cleanForMarshal(v))
			}
		}
		return result
	default:
		return val
	}
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
