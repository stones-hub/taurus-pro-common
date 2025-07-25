package helper

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// MapToStruct 将 map[string]interface{} 转换为结构体
// 使用 JSON 序列化/反序列化的方式，确保类型安全
func MapToStruct(data map[string]interface{}, target interface{}) error {
	// 将 map 转换为 JSON 字节
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal map to json failed: %w", err)
	}

	// 将 JSON 字节反序列化为目标结构体
	err = json.Unmarshal(jsonBytes, target)
	if err != nil {
		return fmt.Errorf("unmarshal json to struct failed: %w", err)
	}

	return nil
}

// MapToStructWithValidation 将 map[string]interface{} 转换为结构体，并验证必需字段
func MapToStructWithValidation(data map[string]interface{}, target interface{}, requiredFields ...string) error {
	// 检查必需字段
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("required field '%s' is missing", field)
		}
	}

	// 转换为结构体
	return MapToStruct(data, target)
}

// ExtractString 从 map[string]interface{} 中提取字符串字段
func ExtractString(data map[string]interface{}, key string) (string, error) {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("field '%s' is not a string", key)
	}
	return "", fmt.Errorf("field '%s' is missing", key)
}

// ExtractBool 从 map[string]interface{} 中提取布尔字段
func ExtractBool(data map[string]interface{}, key string) (bool, error) {
	if value, exists := data[key]; exists {
		if b, ok := value.(bool); ok {
			return b, nil
		}
		return false, fmt.Errorf("field '%s' is not a boolean", key)
	}
	return false, fmt.Errorf("field '%s' is missing", key)
}

// ExtractInt 从 map[string]interface{} 中提取整数字段
func ExtractInt(data map[string]interface{}, key string) (int, error) {
	if value, exists := data[key]; exists {
		switch v := value.(type) {
		case int:
			return v, nil
		case float64:
			return int(v), nil
		default:
			return 0, fmt.Errorf("field '%s' is not a number", key)
		}
	}
	return 0, fmt.Errorf("field '%s' is missing", key)
}

// ExtractFloat64 从 map[string]interface{} 中提取浮点数字段
func ExtractFloat64(data map[string]interface{}, key string) (float64, error) {
	if value, exists := data[key]; exists {
		if f, ok := value.(float64); ok {
			return f, nil
		}
		return 0, fmt.Errorf("field '%s' is not a float64", key)
	}
	return 0, fmt.Errorf("field '%s' is missing", key)
}

// ExtractStringMap 从 map[string]interface{} 中提取字符串映射字段
func ExtractStringMap(data map[string]interface{}, key string) (map[string]string, error) {
	if value, exists := data[key]; exists {
		if mapValue, ok := value.(map[string]interface{}); ok {
			result := make(map[string]string)
			for k, v := range mapValue {
				if strVal, ok := v.(string); ok {
					result[k] = strVal
				} else {
					// 尝试将其他类型转换为字符串
					result[k] = fmt.Sprintf("%v", v)
				}
			}
			return result, nil
		}
		return nil, fmt.Errorf("field '%s' is not a map", key)
	}
	return nil, fmt.Errorf("field '%s' is missing", key)
}

// ExtractStringSlice 从 map[string]interface{} 中提取字符串切片字段
func ExtractStringSlice(data map[string]interface{}, key string) ([]string, error) {
	if value, exists := data[key]; exists {
		if sliceValue, ok := value.([]interface{}); ok {
			result := make([]string, len(sliceValue))
			for i, v := range sliceValue {
				if strVal, ok := v.(string); ok {
					result[i] = strVal
				} else {
					result[i] = fmt.Sprintf("%v", v)
				}
			}
			return result, nil
		}
		return nil, fmt.Errorf("field '%s' is not a slice", key)
	}
	return nil, fmt.Errorf("field '%s' is missing", key)
}

// ValidateStruct 验证结构体的必需字段是否为空
func ValidateStruct(target interface{}, requiredFields ...string) error {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("target is not a struct")
	}

	for _, fieldName := range requiredFields {
		field := v.FieldByName(fieldName)
		if !field.IsValid() {
			return fmt.Errorf("field '%s' does not exist", fieldName)
		}

		if field.Kind() == reflect.String && field.String() == "" {
			return fmt.Errorf("required field '%s' is empty", fieldName)
		}

		if field.Kind() == reflect.Slice && field.Len() == 0 {
			return fmt.Errorf("required field '%s' is empty", fieldName)
		}
	}

	return nil
}
