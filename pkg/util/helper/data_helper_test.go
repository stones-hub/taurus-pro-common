package helper

import (
	"testing"
)

func TestMapToStruct(t *testing.T) {
	// 测试数据
	data := map[string]interface{}{
		"phone":       "13800138000",
		"type":        "code",
		"is_official": true,
		"content":     "123456",
		"params": map[string]interface{}{
			"extra": "value",
		},
	}

	// 目标结构体
	type TestStruct struct {
		Phone      string            `json:"phone"`
		Type       string            `json:"type"`
		IsOfficial bool              `json:"is_official"`
		Content    string            `json:"content"`
		Params     map[string]string `json:"params"`
	}

	var result TestStruct
	err := MapToStruct(data, &result)
	if err != nil {
		t.Fatalf("MapToStruct failed: %v", err)
	}

	// 验证结果
	if result.Phone != "13800138000" {
		t.Errorf("Expected phone '13800138000', got '%s'", result.Phone)
	}
	if result.Type != "code" {
		t.Errorf("Expected type 'code', got '%s'", result.Type)
	}
	if !result.IsOfficial {
		t.Errorf("Expected is_official true, got false")
	}
	if result.Content != "123456" {
		t.Errorf("Expected content '123456', got '%s'", result.Content)
	}
	if result.Params["extra"] != "value" {
		t.Errorf("Expected params['extra'] 'value', got '%s'", result.Params["extra"])
	}
}

func TestMapToStructWithValidation(t *testing.T) {
	// 测试数据
	data := map[string]interface{}{
		"phone":       "13800138000",
		"type":        "code",
		"is_official": true,
		"content":     "123456",
	}

	type TestStruct struct {
		Phone      string `json:"phone"`
		Type       string `json:"type"`
		IsOfficial bool   `json:"is_official"`
		Content    string `json:"content"`
	}

	var result TestStruct
	err := MapToStructWithValidation(data, &result, "phone", "type", "is_official", "content")
	if err != nil {
		t.Fatalf("MapToStructWithValidation failed: %v", err)
	}

	// 测试缺少必需字段的情况
	incompleteData := map[string]interface{}{
		"phone": "13800138000",
		"type":  "code",
		// 缺少 is_official 和 content
	}

	var incompleteResult TestStruct
	err = MapToStructWithValidation(incompleteData, &incompleteResult, "phone", "type", "is_official", "content")
	if err == nil {
		t.Error("Expected error for missing required field, but got none")
	}
}

func TestExtractString(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"age":  25,
	}

	// 测试正常情况
	name, err := ExtractString(data, "name")
	if err != nil {
		t.Fatalf("ExtractString failed: %v", err)
	}
	if name != "test" {
		t.Errorf("Expected 'test', got '%s'", name)
	}

	// 测试字段不存在
	_, err = ExtractString(data, "nonexistent")
	if err == nil {
		t.Error("Expected error for missing field, but got none")
	}

	// 测试类型错误
	_, err = ExtractString(data, "age")
	if err == nil {
		t.Error("Expected error for wrong type, but got none")
	}
}

func TestExtractBool(t *testing.T) {
	data := map[string]interface{}{
		"enabled": true,
		"name":    "test",
	}

	// 测试正常情况
	enabled, err := ExtractBool(data, "enabled")
	if err != nil {
		t.Fatalf("ExtractBool failed: %v", err)
	}
	if !enabled {
		t.Error("Expected true, got false")
	}

	// 测试类型错误
	_, err = ExtractBool(data, "name")
	if err == nil {
		t.Error("Expected error for wrong type, but got none")
	}
}

func TestExtractStringMap(t *testing.T) {
	data := map[string]interface{}{
		"params": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		"name": "test",
	}

	// 测试正常情况
	params, err := ExtractStringMap(data, "params")
	if err != nil {
		t.Fatalf("ExtractStringMap failed: %v", err)
	}
	if params["key1"] != "value1" {
		t.Errorf("Expected 'value1', got '%s'", params["key1"])
	}
	if params["key2"] != "value2" {
		t.Errorf("Expected 'value2', got '%s'", params["key2"])
	}

	// 测试类型错误
	_, err = ExtractStringMap(data, "name")
	if err == nil {
		t.Error("Expected error for wrong type, but got none")
	}
}

func TestValidateStruct(t *testing.T) {
	// 测试结构体
	type TestStruct struct {
		Name    string   `json:"name"`
		Age     int      `json:"age"`
		Hobbies []string `json:"hobbies"`
		Email   string   `json:"email"`
	}

	t.Run("验证成功 - 所有必需字段都有值", func(t *testing.T) {
		validStruct := TestStruct{
			Name:    "张三",
			Age:     25,
			Hobbies: []string{"读书", "游泳"},
			Email:   "zhangsan@example.com",
		}

		err := ValidateStruct(&validStruct, "Name", "Email", "Hobbies")
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})

	t.Run("验证失败 - 字符串字段为空", func(t *testing.T) {
		invalidStruct := TestStruct{
			Name:    "", // 空字符串
			Age:     25,
			Hobbies: []string{"读书", "游泳"},
			Email:   "zhangsan@example.com",
		}

		err := ValidateStruct(&invalidStruct, "Name", "Email", "Hobbies")
		if err == nil {
			t.Error("Expected error for empty Name field, but got none")
		}
		if err.Error() != "required field 'Name' is empty" {
			t.Errorf("Expected specific error message, but got: %v", err)
		}
	})

	t.Run("验证失败 - 切片字段为空", func(t *testing.T) {
		invalidStruct := TestStruct{
			Name:    "张三",
			Age:     25,
			Hobbies: []string{}, // 空切片
			Email:   "zhangsan@example.com",
		}

		err := ValidateStruct(&invalidStruct, "Name", "Email", "Hobbies")
		if err == nil {
			t.Error("Expected error for empty Hobbies field, but got none")
		}
		if err.Error() != "required field 'Hobbies' is empty" {
			t.Errorf("Expected specific error message, but got: %v", err)
		}
	})

	t.Run("验证失败 - 字段不存在", func(t *testing.T) {
		validStruct := TestStruct{
			Name:    "张三",
			Age:     25,
			Hobbies: []string{"读书", "游泳"},
			Email:   "zhangsan@example.com",
		}

		err := ValidateStruct(&validStruct, "Name", "NonExistentField")
		if err == nil {
			t.Error("Expected error for non-existent field, but got none")
		}
		if err.Error() != "field 'NonExistentField' does not exist" {
			t.Errorf("Expected specific error message, but got: %v", err)
		}
	})

	t.Run("验证失败 - 目标不是结构体", func(t *testing.T) {
		notStruct := "this is not a struct"

		err := ValidateStruct(notStruct, "Name")
		if err == nil {
			t.Error("Expected error for non-struct target, but got none")
		}
		if err.Error() != "target is not a struct" {
			t.Errorf("Expected specific error message, but got: %v", err)
		}
	})

	t.Run("验证成功 - 非指针结构体", func(t *testing.T) {
		validStruct := TestStruct{
			Name:    "张三",
			Age:     25,
			Hobbies: []string{"读书", "游泳"},
			Email:   "zhangsan@example.com",
		}

		err := ValidateStruct(validStruct, "Name", "Email", "Hobbies")
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})

	t.Run("验证成功 - 数字字段不为空", func(t *testing.T) {
		validStruct := TestStruct{
			Name:    "张三",
			Age:     0, // 数字类型的零值不算空
			Hobbies: []string{"读书", "游泳"},
			Email:   "zhangsan@example.com",
		}

		err := ValidateStruct(&validStruct, "Name", "Age", "Email", "Hobbies")
		if err != nil {
			t.Errorf("Expected no error for zero int value, but got: %v", err)
		}
	})

	t.Run("验证失败 - 多个字段为空", func(t *testing.T) {
		invalidStruct := TestStruct{
			Name:    "", // 空字符串
			Age:     25,
			Hobbies: []string{}, // 空切片
			Email:   "",         // 空字符串
		}

		err := ValidateStruct(&invalidStruct, "Name", "Email", "Hobbies")
		if err == nil {
			t.Error("Expected error for multiple empty fields, but got none")
		}
		// 应该返回第一个遇到的空字段错误
		if err.Error() != "required field 'Name' is empty" {
			t.Errorf("Expected first empty field error, but got: %v", err)
		}
	})
}
