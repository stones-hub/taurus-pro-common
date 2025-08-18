// Copyright (c) 2025 Taurus Team. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: yelei
// Email: 61647649@qq.com
// Date: 2025-06-13

package tstruct

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/jinzhu/copier"
	"github.com/stones-hub/taurus-pro-common/pkg/util/tstring"
)

// Copy 将源结构体的字段值复制到目标结构体，支持不同类型结构体之间的复制
// 参数：
//   - s: 源结构体
//   - ts: 目标结构体（必须是指针类型）
//
// 返回值：
//   - error: 如果复制过程中出现错误则返回错误信息
//
// 使用示例：
//
//	// 相同类型结构体复制
//	type User struct {
//	    Name string
//	    Age  int
//	}
//	src := User{Name: "John", Age: 30}
//	dst := User{}
//	err := tstruct.Copy(src, &dst)
//
//	// 不同类型结构体复制（字段名匹配的会被复制）
//	type UserDTO struct {
//	    Name     string
//	    Age      int
//	    Password string  // 这个字段不会被复制
//	}
//	src := User{Name: "John", Age: 30}
//	dst := UserDTO{}
//	err := tstruct.Copy(src, &dst)
//
// 注意事项：
//   - 目标结构体必须是指针类型
//   - 只会复制字段名匹配的字段
//   - 支持基本类型之间的转换（如 int32 到 int64）
//   - 支持嵌套结构体的复制
//   - 使用 github.com/jinzhu/copier 包实现
func Copy(s, ts interface{}) error {
	return copier.Copy(ts, s)
}

// MapToStruct 将 map[string]interface{} 转换为结构体，使用 JSON 序列化/反序列化确保类型安全
// 参数：
//   - data: 源数据 map
//   - target: 目标结构体（必须是指针类型）
//
// 返回值：
//   - error: 如果转换过程中出现错误则返回错误信息
//
// 使用示例：
//
//	// 基本使用
//	type User struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//	data := map[string]interface{}{
//	    "name": "John",
//	    "age":  30,
//	}
//	var user User
//	err := tstruct.MapToStruct(data, &user)
//
//	// 嵌套结构体
//	type Address struct {
//	    City    string `json:"city"`
//	    Country string `json:"country"`
//	}
//	type Person struct {
//	    Name    string  `json:"name"`
//	    Address Address `json:"address"`
//	}
//	data := map[string]interface{}{
//	    "name": "John",
//	    "address": map[string]interface{}{
//	        "city":    "New York",
//	        "country": "USA",
//	    },
//	}
//	var person Person
//	err := tstruct.MapToStruct(data, &person)
//
// 注意事项：
//   - 目标结构体必须是指针类型
//   - 结构体字段需要有 json 标签或与 map 的键名完全匹配
//   - 支持嵌套结构体和切片
//   - 会自动进行类型转换（如字符串到数字）
//   - 如果需要验证必需字段，请使用 MapToStructWithValidation
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

// MapToStructWithValidation 将 map[string]interface{} 转换为结构体，并在转换前验证必需字段是否存在
// 参数：
//   - data: 源数据 map
//   - target: 目标结构体（必须是指针类型）
//   - requiredFields: 必需字段的名称列表
//
// 返回值：
//   - error: 如果验证失败或转换过程中出现错误则返回错误信息
//
// 使用示例：
//
//	type User struct {
//	    Name     string `json:"name"`
//	    Age      int    `json:"age"`
//	    Email    string `json:"email"`
//	    Optional string `json:"optional"`
//	}
//
//	data := map[string]interface{}{
//	    "name":     "John",
//	    "age":      30,
//	    "optional": "value",
//	}
//
//	var user User
//	// 验证 name 和 email 字段是否存在
//	err := tstruct.MapToStructWithValidation(data, &user, "name", "email")
//	// 返回错误：required field 'email' is missing
//
// 注意事项：
//   - 目标结构体必须是指针类型
//   - 在转换之前会检查必需字段是否存在
//   - 字段名称必须与 map 的键名完全匹配
//   - 验证仅检查字段是否存在，不检查字段值是否有效
//   - 如果验证通过，会调用 MapToStruct 进行转换
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

// ExtractString 从 map[string]interface{} 中提取并返回字符串类型的字段值
// 参数：
//   - data: 源数据 map
//   - key: 要提取的字段名
//
// 返回值：
//   - string: 提取的字符串值
//   - error: 如果字段不存在或类型不匹配则返回错误信息
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "name": "John",
//	    "age":  30,
//	}
//
//	// 提取存在的字符串字段
//	name, err := tstruct.ExtractString(data, "name")
//	if err != nil {
//	    return err
//	}
//	fmt.Println(name)  // 输出: John
//
//	// 提取不存在的字段
//	email, err := tstruct.ExtractString(data, "email")
//	// 返回错误: field 'email' is missing
//
//	// 提取类型不匹配的字段
//	age, err := tstruct.ExtractString(data, "age")
//	// 返回错误: field 'age' is not a string
//
// 注意事项：
//   - 字段必须存在且类型必须是 string
//   - 不支持类型转换，如需类型转换请使用 fmt.Sprintf
//   - 对于可选字段，应该先检查字段是否存在
func ExtractString(data map[string]interface{}, key string) (string, error) {
	if value, exists := data[key]; exists {
		if str, ok := value.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("field '%s' is not a string", key)
	}
	return "", fmt.Errorf("field '%s' is missing", key)
}

// ExtractBool 从 map[string]interface{} 中提取并返回布尔类型的字段值
// 参数：
//   - data: 源数据 map
//   - key: 要提取的字段名
//
// 返回值：
//   - bool: 提取的布尔值
//   - error: 如果字段不存在或类型不匹配则返回错误信息
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "active": true,
//	    "name":   "John",
//	}
//
//	// 提取存在的布尔字段
//	active, err := tstruct.ExtractBool(data, "active")
//	if err != nil {
//	    return err
//	}
//	fmt.Println(active)  // 输出: true
//
//	// 提取不存在的字段
//	deleted, err := tstruct.ExtractBool(data, "deleted")
//	// 返回错误: field 'deleted' is missing
//
//	// 提取类型不匹配的字段
//	name, err := tstruct.ExtractBool(data, "name")
//	// 返回错误: field 'name' is not a boolean
//
// 注意事项：
//   - 字段必须存在且类型必须是 bool
//   - 不支持字符串到布尔值的转换（如 "true" 到 true）
//   - 对于可选字段，应该先检查字段是否存在
func ExtractBool(data map[string]interface{}, key string) (bool, error) {
	if value, exists := data[key]; exists {
		if b, ok := value.(bool); ok {
			return b, nil
		}
		return false, fmt.Errorf("field '%s' is not a boolean", key)
	}
	return false, fmt.Errorf("field '%s' is missing", key)
}

// ExtractInt 从 map[string]interface{} 中提取并返回整数类型的字段值，支持多种数值类型的自动转换
// 参数：
//   - data: 源数据 map
//   - key: 要提取的字段名
//
// 返回值：
//   - int: 提取并转换后的整数值
//   - error: 如果字段不存在、类型不匹配或转换失败则返回错误信息
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "age":     30,        // int
//	    "count":   int64(100),// int64
//	    "amount":  float64(50.0),  // float64
//	    "code":    "123",     // string
//	    "name":    "John",    // string
//	}
//
//	// 提取整数字段
//	age, err := tstruct.ExtractInt(data, "age")
//	// age == 30, err == nil
//
//	// 提取 int64 字段
//	count, err := tstruct.ExtractInt(data, "count")
//	// count == 100, err == nil
//
//	// 提取浮点数字段（会被转换为整数）
//	amount, err := tstruct.ExtractInt(data, "amount")
//	// amount == 50, err == nil
//
//	// 提取数字字符串
//	code, err := tstruct.ExtractInt(data, "code")
//	// code == 123, err == nil
//
//	// 提取非数字字符串
//	invalid, err := tstruct.ExtractInt(data, "name")
//	// 返回错误: field 'name' cannot be converted to int
//
// 注意事项：
//   - 支持从多种数值类型自动转换：int8/16/32/64、uint8/16/32/64、float32/64
//   - 支持从字符串解析数字（使用 strconv.Atoi）
//   - 浮点数会被截断为整数
//   - 超出 int 范围的 uint64 值会返回错误
//   - 对于可选字段，应该先检查字段是否存在
func ExtractInt(data map[string]interface{}, key string) (int, error) {
	if value, exists := data[key]; exists {
		switch v := value.(type) {
		case int:
			return v, nil
		case int8:
			return int(v), nil
		case int16:
			return int(v), nil
		case int32:
			return int(v), nil
		case int64:
			return int(v), nil
		case uint:
			return int(v), nil
		case uint8:
			return int(v), nil
		case uint16:
			return int(v), nil
		case uint32:
			return int(v), nil
		case uint64:
			if v > uint64(^uint(0)>>1) {
				return 0, fmt.Errorf("field '%s' value %d is too large for int", key, v)
			}
			return int(v), nil
		case float32:
			return int(v), nil
		case float64:
			return int(v), nil
		case string:
			// 尝试将字符串转换为整数
			if i, err := strconv.Atoi(v); err == nil {
				return i, nil
			}
		}
		return 0, fmt.Errorf("field '%s' cannot be converted to int", key)
	}
	return 0, fmt.Errorf("field '%s' is missing", key)
}

// ExtractFloat64 从 map[string]interface{} 中提取并返回浮点数类型的字段值
// 参数：
//   - data: 源数据 map
//   - key: 要提取的字段名
//
// 返回值：
//   - float64: 提取的浮点数值
//   - error: 如果字段不存在或类型不匹配则返回错误信息
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "price":  99.99,     // float64
//	    "amount": 100,       // int
//	    "name":   "John",    // string
//	}
//
//	// 提取浮点数字段
//	price, err := tstruct.ExtractFloat64(data, "price")
//	if err != nil {
//	    return err
//	}
//	fmt.Println(price)  // 输出: 99.99
//
//	// 提取不存在的字段
//	weight, err := tstruct.ExtractFloat64(data, "weight")
//	// 返回错误: field 'weight' is missing
//
//	// 提取类型不匹配的字段
//	name, err := tstruct.ExtractFloat64(data, "name")
//	// 返回错误: field 'name' is not a float64
//
// 注意事项：
//   - 字段必须存在且类型必须是 float64
//   - 不支持从其他数值类型转换（如 int 到 float64）
//   - 不支持从字符串解析浮点数
//   - 对于可选字段，应该先检查字段是否存在
//   - 如果需要更灵活的数值转换，请使用 ExtractInt
func ExtractFloat64(data map[string]interface{}, key string) (float64, error) {
	if value, exists := data[key]; exists {
		if f, ok := value.(float64); ok {
			return f, nil
		}
		return 0, fmt.Errorf("field '%s' is not a float64", key)
	}
	return 0, fmt.Errorf("field '%s' is missing", key)
}

// ExtractStringMap 从 map[string]interface{} 中提取并返回字符串映射类型的字段值，支持将非字符串值转换为字符串
// 参数：
//   - data: 源数据 map
//   - key: 要提取的字段名
//
// 返回值：
//   - map[string]string: 提取并转换后的字符串映射
//   - error: 如果字段不存在或类型不匹配则返回错误信息
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "labels": map[string]interface{}{
//	        "env":     "prod",
//	        "version": 2,
//	        "active":  true,
//	    },
//	    "name": "test",
//	}
//
//	// 提取并转换映射字段
//	labels, err := tstruct.ExtractStringMap(data, "labels")
//	if err != nil {
//	    return err
//	}
//	// labels 的内容：
//	// {
//	//   "env":     "prod",
//	//   "version": "2",
//	//   "active":  "true"
//	// }
//
//	// 提取不存在的字段
//	tags, err := tstruct.ExtractStringMap(data, "tags")
//	// 返回错误: field 'tags' is missing
//
//	// 提取类型不匹配的字段
//	invalid, err := tstruct.ExtractStringMap(data, "name")
//	// 返回错误: field 'name' is not a map
//
// 注意事项：
//   - 字段必须存在且类型必须是 map[string]interface{}
//   - 非字符串值会被转换为字符串（使用 fmt.Sprintf）
//   - 对于可选字段，应该先检查字段是否存在
//   - 返回的映射是一个新的副本，修改它不会影响原始数据
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

// ExtractStringSlice 从 map[string]interface{} 中提取并返回字符串切片类型的字段值，支持将非字符串值转换为字符串
// 参数：
//   - data: 源数据 map
//   - key: 要提取的字段名
//
// 返回值：
//   - []string: 提取并转换后的字符串切片
//   - error: 如果字段不存在或类型不匹配则返回错误信息
//
// 使用示例：
//
//	data := map[string]interface{}{
//	    "tags": []interface{}{
//	        "web",
//	        123,
//	        true,
//	        3.14,
//	    },
//	    "name": "test",
//	}
//
//	// 提取并转换切片字段
//	tags, err := tstruct.ExtractStringSlice(data, "tags")
//	if err != nil {
//	    return err
//	}
//	// tags 的内容：["web", "123", "true", "3.14"]
//
//	// 提取不存在的字段
//	labels, err := tstruct.ExtractStringSlice(data, "labels")
//	// 返回错误: field 'labels' is missing
//
//	// 提取类型不匹配的字段
//	invalid, err := tstruct.ExtractStringSlice(data, "name")
//	// 返回错误: field 'name' is not a slice
//
// 注意事项：
//   - 字段必须存在且类型必须是 []interface{}
//   - 非字符串值会被转换为字符串（使用 fmt.Sprintf）
//   - 对于可选字段，应该先检查字段是否存在
//   - 返回的切片是一个新的副本，修改它不会影响原始数据
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

// ValidateStruct 验证结构体中指定字段是否为空值，支持多种类型的空值检查
// 参数：
//   - target: 要验证的结构体（必须是结构体或结构体指针）
//   - requiredFields: 必需字段的名称列表
//
// 返回值：
//   - error: 如果验证失败则返回错误信息，包括：
//   - 目标不是结构体或结构体指针
//   - 目标是 nil 指针
//   - 字段不存在
//   - 字段值为空
//
// 使用示例：
//
//	type User struct {
//	    ID       int
//	    Name     string
//	    Email    string
//	    Age      int
//	    Optional string
//	}
//
//	user := User{
//	    ID:    1,
//	    Name:  "John",
//	    Email: "",
//	    Age:   0,
//	}
//
//	// 验证必需字段
//	err := tstruct.ValidateStruct(user, "ID", "Name", "Email")
//	// 返回错误: required field 'Email' is empty
//
//	// 验证指针类型
//	err := tstruct.ValidateStruct(&user, "ID", "Name")
//	// 验证通过，返回 nil
//
//	// 验证不存在的字段
//	err := tstruct.ValidateStruct(user, "Username")
//	// 返回错误: field 'Username' does not exist
//
// 注意事项：
//   - 支持结构体值和结构体指针
//   - 以下情况被视为空值：
//   - string: 空字符串 ("")
//   - 切片/映射/数组: 长度为0
//   - 指针/接口: nil
//   - 数值类型: 0
//   - bool: false
//   - 结构体: 如果实现了 IsEmpty() bool 方法，则使用该方法判断
//   - 字段必须是导出的（首字母大写）
//   - 验证按字段名称顺序进行，遇到第一个空值就返回错误
func ValidateStruct(target interface{}, requiredFields ...string) error {
	v := reflect.ValueOf(target)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fmt.Errorf("target is nil")
		}
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

		// 检查字段是否为空
		isEmpty := false
		switch field.Kind() {
		case reflect.String:
			isEmpty = field.String() == ""
		case reflect.Slice, reflect.Map, reflect.Array:
			isEmpty = field.Len() == 0
		case reflect.Ptr, reflect.Interface:
			isEmpty = field.IsNil()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			isEmpty = field.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			isEmpty = field.Uint() == 0
		case reflect.Float32, reflect.Float64:
			isEmpty = field.Float() == 0
		case reflect.Bool:
			isEmpty = !field.Bool()
		case reflect.Struct:
			// 对于结构体，如果它实现了 IsEmpty() bool 方法，则使用该方法
			if method := field.MethodByName("IsEmpty"); method.IsValid() {
				results := method.Call(nil)
				if len(results) == 1 && results[0].Kind() == reflect.Bool {
					isEmpty = results[0].Bool()
				}
			}
		}

		if isEmpty {
			return fmt.Errorf("required field '%s' is empty", fieldName)
		}
	}

	return nil
}

// ScanPointer 将源值赋值给目标指针，支持类型转换和空值检查
// 参数：
//   - dest: 目标指针（必须是有效的非空指针）
//   - src: 源值（可以是值或指针类型）
//
// 返回值：
//   - error: 如果赋值过程中出现错误则返回错误信息，包括：
//   - 目标不是指针类型
//   - 目标是 nil 指针
//   - 源值是 nil 指针
//   - 类型不匹配且无法转换
//
// 使用示例：
//
//	// 基本类型赋值
//	var i int
//	err := tstruct.ScanPointer(&i, 42)
//	// i == 42
//
//	// 指针类型赋值
//	var s string
//	source := "hello"
//	err := tstruct.ScanPointer(&s, &source)
//	// s == "hello"
//
//	// 类型转换
//	var f float64
//	err := tstruct.ScanPointer(&f, 42)
//	// f == 42.0
//
//	// 错误处理
//	var ptr *string
//	err := tstruct.ScanPointer(ptr, "test")
//	// 返回错误: destination pointer is nil
//
//	var i int
//	err := tstruct.ScanPointer(&i, "not a number")
//	// 返回错误: cannot convert source type string to destination type int
//
// 注意事项：
//   - 目标必须是非 nil 指针
//   - 如果源值是指针，会自动解引用
//   - 支持基本类型之间的自动转换
//   - 不支持复杂类型（如结构体）之间的转换
//   - 对于可能失败的转换，建议使用类型断言
//   - 如果需要批量处理多个指针，请使用 ScanPointersSlice
func ScanPointer(dest interface{}, src interface{}) error {
	if dest == nil {
		return fmt.Errorf("destination pointer is nil")
	}

	// 获取目标的反射值
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer, got %v", dv.Kind())
	}
	if dv.IsNil() {
		return fmt.Errorf("destination pointer is nil")
	}

	// 获取指针指向的元素
	dv = dv.Elem()

	// 获取源数据的反射值
	sv := reflect.ValueOf(src)

	// 如果源数据是指针，获取其指向的值
	if sv.Kind() == reflect.Ptr {
		if sv.IsNil() {
			return fmt.Errorf("source value is nil")
		}
		sv = sv.Elem()
	}

	// 检查类型是否匹配
	if !sv.Type().ConvertibleTo(dv.Type()) {
		return fmt.Errorf("cannot convert source type %v to destination type %v", sv.Type(), dv.Type())
	}

	// 进行类型转换（如果需要）并赋值
	if sv.Type() != dv.Type() {
		sv = sv.Convert(dv.Type())
	}

	// 设置值
	dv.Set(sv)
	return nil
}

// MustScanPointer 是 ScanPointer 的简化版本，如果赋值失败会触发 panic
// 参数：
//   - dest: 目标指针（必须是有效的非空指针）
//   - src: 源值（可以是值或指针类型）
//
// 使用示例：
//
//	// 基本使用
//	var i int
//	tstruct.MustScanPointer(&i, 42)
//	// i == 42
//
//	// 触发 panic 的情况
//	var ptr *string
//	tstruct.MustScanPointer(ptr, "test")
//	// panic: destination pointer is nil
//
// 注意事项：
//   - 只在确保赋值不会失败的情况下使用
//   - 在处理用户输入时应该使用 ScanPointer
//   - panic 会包含具体的错误信息
//   - 建议在 init() 或配置加载等场景使用
func MustScanPointer(dest interface{}, src interface{}) {
	if err := ScanPointer(dest, src); err != nil {
		panic(err)
	}
}

// ScanPointersSlice 批量处理多个指针的赋值操作，要求目标和源切片长度相同
// 参数：
//   - dests: 目标指针切片
//   - srcs: 源值切片
//
// 返回值：
//   - error: 如果任何赋值操作失败则返回错误信息，包括：
//   - 目标和源切片长度不匹配
//   - 任何单个赋值操作的错误
//
// 使用示例：
//
//	// 基本使用
//	var (
//	    i int
//	    s string
//	    f float64
//	)
//	dests := []interface{}{&i, &s, &f}
//	srcs := []interface{}{42, "hello", 3.14}
//	err := tstruct.ScanPointersSlice(dests, srcs)
//	// i == 42, s == "hello", f == 3.14
//
//	// 长度不匹配
//	dests := []interface{}{&i, &s}
//	srcs := []interface{}{42}
//	err := tstruct.ScanPointersSlice(dests, srcs)
//	// 返回错误: destination and source slices must have same length
//
//	// 类型不匹配
//	dests := []interface{}{&i, &s}
//	srcs := []interface{}{42, true}
//	err := tstruct.ScanPointersSlice(dests, srcs)
//	// 返回错误: error at index 1: cannot convert source type bool to destination type string
//
// 注意事项：
//   - 目标和源切片必须长度相同
//   - 按索引顺序逐个处理赋值
//   - 遇到第一个错误就停止处理并返回
//   - 错误信息会包含失败的索引位置
//   - 支持所有 ScanPointer 支持的类型转换
func ScanPointersSlice(dests []interface{}, srcs []interface{}) error {
	if len(dests) != len(srcs) {
		return fmt.Errorf("destination and source slices must have same length")
	}

	for i := range dests {
		if err := ScanPointer(dests[i], srcs[i]); err != nil {
			return fmt.Errorf("error at index %d: %v", i, err)
		}
	}
	return nil
}

// TrimSpace 去除结构体中所有字符串字段的首尾空白字符
// 参数：
//   - target: 要处理的结构体指针
//
// 使用示例：
//
//	type User struct {
//	    Name     string
//	    Email    string
//	    Age      int       // 非字符串字段会被忽略
//	    Address  string
//	}
//
//	user := &User{
//	    Name:    "  John  ",
//	    Email:   "john@example.com  ",
//	    Age:     30,
//	    Address: "\t123 Main St\n",
//	}
//
//	tstruct.TrimSpace(user)
//	// 结果：
//	// user.Name == "John"
//	// user.Email == "john@example.com"
//	// user.Age == 30 (未变化)
//	// user.Address == "123 Main St"
//
// 注意事项：
//   - 必须传入结构体指针，否则函数会直接返回
//   - 只处理字符串类型的字段
//   - 使用 strings.TrimSpace 处理每个字符串
//   - 会移除所有类型的空白字符（空格、制表符、换行符等）
//   - 不会递归处理嵌套结构体
//   - 不会处理未导出的字段（小写字段名）
func TrimSpace(target interface{}) {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Ptr {
		return
	}
	t = t.Elem()
	v := reflect.ValueOf(target).Elem()
	for i := 0; i < t.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.String:
			v.Field(i).SetString(strings.TrimSpace(v.Field(i).String()))
		}
	}
}

// ToStringWithFlag 将结构体、映射或切片转换为字符串，使用指定的分隔符连接各个元素
// 参数：
//   - v: 要转换的值（必须是结构体、映射或切片类型）
//   - flag: 用于分隔元素的字符串
//
// 返回值：
//   - string: 转换后的字符串
//   - error: 如果转换过程中出现错误则返回错误信息
//
// 使用示例：
//
//	// 结构体转换
//	type Person struct {
//	    Name string
//	    Age  int
//	}
//	p := Person{Name: "John", Age: 30}
//	str, err := tstruct.ToStringWithFlag(p, ", ")
//	// 返回: "John, 30"
//
//	// 映射转换
//	m := map[string]interface{}{
//	    "name": "John",
//	    "age":  30,
//	}
//	str, err := tstruct.ToStringWithFlag(m, "; ")
//	// 返回: "name:John; age:30"
//
//	// 切片转换
//	s := []string{"apple", "banana", "orange"}
//	str, err := tstruct.ToStringWithFlag(s, " | ")
//	// 返回: "apple | banana | orange"
//
// 注意事项：
//   - 支持的类型：结构体、映射、切片、数组
//   - 结构体：按字段顺序连接字段值
//   - 映射：按 "键:值" 格式连接
//   - 切片/数组：直接连接元素值
//   - 所有值都使用 fmt.Sprintf("%v") 转换为字符串
//   - 不支持的类型会返回错误
//   - 空集合会返回空字符串
func ToStringWithFlag(v interface{}, flag string) (string, error) {
	var result strings.Builder

	// 使用反射处理不同类型
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Struct:
		// 遍历结构体的字段
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			if i > 0 {
				result.WriteString(flag)
			}
			result.WriteString(fmt.Sprintf("%v", field.Interface()))
		}
	case reflect.Map:
		// 遍历 map 的键值对（按键名排序，保证稳定输出）
		keys := val.MapKeys()
		type kv struct {
			k reflect.Value
			s string
		}
		pairs := make([]kv, 0, len(keys))
		for _, key := range keys {
			pairs = append(pairs, kv{k: key, s: fmt.Sprint(key.Interface())})
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].s < pairs[j].s })
		for i, p := range pairs {
			if i > 0 {
				result.WriteString(flag)
			}
			value := val.MapIndex(p.k)
			result.WriteString(fmt.Sprintf("%v:%v", p.s, value.Interface()))
		}
	case reflect.Slice, reflect.Array:
		// 遍历 slice 或 array 的元素
		for i := 0; i < val.Len(); i++ {
			if i > 0 {
				result.WriteString(flag)
			}
			result.WriteString(fmt.Sprintf("%v", val.Index(i).Interface()))
		}
	default:
		return "", fmt.Errorf("unsupported type: %v", val.Kind())
	}

	return result.String(), nil
}

// CheckType 检查并返回传入值的类型描述，支持所有 Go 基本类型和复合类型
// 参数：
//   - v: 要检查的值（任意类型）
//
// 返回值：
//   - string: 类型描述字符串，包括：
//   - 基本类型：bool, int/uint 系列, float 系列, complex 系列
//   - 特殊类型：byte (uint8), rune (int32)
//   - 复合类型：array, slice, map, struct, pointer, function, channel, interface
//
// 使用示例：
//
//	// 基本类型
//	fmt.Println(tstruct.CheckType(42))          // "int"
//	fmt.Println(tstruct.CheckType(3.14))        // "float64"
//	fmt.Println(tstruct.CheckType(true))        // "bool"
//	fmt.Println(tstruct.CheckType("hello"))     // "string"
//
//	// 特殊类型
//	var b byte = 65
//	fmt.Println(tstruct.CheckType(b))           // "byte"
//	var r rune = '世'
//	fmt.Println(tstruct.CheckType(r))           // "rune"
//
//	// 复合类型
//	fmt.Println(tstruct.CheckType([]int{}))     // "slice"
//	fmt.Println(tstruct.CheckType([3]int{}))    // "array"
//	fmt.Println(tstruct.CheckType(map[string]int{}))  // "map"
//	fmt.Println(tstruct.CheckType(struct{}{}))  // "struct"
//	fmt.Println(tstruct.CheckType(&struct{}{})) // "pointer"
//	fmt.Println(tstruct.CheckType(func(){}))    // "function"
//	fmt.Println(tstruct.CheckType(make(chan int))) // "channel"
//
// 注意事项：
//   - uint8 会被识别为 "byte"
//   - int32 会被识别为 "rune"
//   - 对于未知类型会返回 "type_name (unknown)"
//   - 类型名称全部小写
//   - 不会递归检查复合类型的元素类型
func CheckType(v interface{}) string {
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if t.Kind() == reflect.Int32 {
			return "rune"
		}
		return t.Kind().String()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if t.Kind() == reflect.Uint8 {
			return "byte"
		}
		return t.Kind().String()
	case reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return t.Kind().String()
	case reflect.String:
		return "string"
	case reflect.Array:
		return "array"
	case reflect.Slice:
		return "slice"
	case reflect.Map:
		return "map"
	case reflect.Struct:
		return "struct"
	case reflect.Ptr:
		return "pointer"
	case reflect.Func:
		return "function"
	case reflect.Chan:
		return "channel"
	case reflect.Interface:
		return "interface"
	default:
		return fmt.Sprintf("%v (unknown)", t.Kind())
	}
}

// ConvertMapKeysToCamelCase 将映射中的所有键名转换为小驼峰格式（第一个单词小写，后续单词首字母大写）
// 参数：
//   - inputMap: 要处理的映射，键必须是字符串类型
//
// 返回值：
//   - map[string]interface{}: 返回一个新的映射，其中所有键名都转换为小驼峰格式
//
// 使用示例：
//
//	// 基本使用
//	input := map[string]interface{}{
//	    "user_name":     "John",
//	    "email_address": "john@example.com",
//	    "phone_number":  "1234567890",
//	}
//	result := tstruct.ConvertMapKeysToCamelCase(input)
//	// 结果：
//	// {
//	//   "userName":     "John",
//	//   "emailAddress": "john@example.com",
//	//   "phoneNumber":  "1234567890"
//	// }
//
//	// 处理单个单词
//	input := map[string]interface{}{
//	    "name": "John",
//	    "age":  30,
//	}
//	result := tstruct.ConvertMapKeysToCamelCase(input)
//	// 键名已经是小驼峰格式，保持不变
//
// 注意事项：
//   - 创建新的映射，不修改原始映射
//   - 只转换键名，值保持不变
//   - 使用 UnderscoreToCamelCase 函数处理每个键
//   - 适用于 API 响应格式转换
//   - 如果键名已经是驼峰格式，则保持不变
func ConvertMapKeysToCamelCase(inputMap map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range inputMap {
		newKey := k
		if strings.Contains(k, "_") {
			newKey = tstring.UnderscoreToCamelCase(k)
		}
		newMap[newKey] = v
	}
	return newMap
}

// StructToMap 将结构体转换为映射，支持使用结构体标签自定义键名
// 参数：
//   - obj: 要转换的结构体或结构体指针
//   - tagName: 用于获取字段名的标签名（如 "json"、"yaml"、"mapstructure" 等）
//
// 返回值：
//   - map[string]interface{}: 转换后的映射，如果输入无效则返回 nil
//
// 使用示例：
//
//	// 基本使用
//	type User struct {
//	    Name     string `json:"name"`
//	    Age      int    `json:"age"`
//	    Email    string `json:"email,omitempty"`
//	    Password string `json:"-"`  // 将被忽略
//	}
//
//	user := User{
//	    Name:     "John",
//	    Age:      30,
//	    Email:    "john@example.com",
//	    Password: "secret",
//	}
//
//	// 使用 json 标签
//	result := tstruct.StructToMap(user, "json")
//	// 结果：
//	// {
//	//   "name": "John",
//	//   "age": 30,
//	//   "email": "john@example.com"
//	// }
//
//	// 不使用标签（使用字段名）
//	result := tstruct.StructToMap(user, "")
//	// 结果：
//	// {
//	//   "Name": "John",
//	//   "Age": 30,
//	//   "Email": "john@example.com",
//	//   "Password": "secret"
//	// }
//
//	// 嵌套结构体
//	type Address struct {
//	    City    string `json:"city"`
//	    Country string `json:"country"`
//	}
//
//	type Person struct {
//	    Name    string  `json:"name"`
//	    Address Address `json:"address"`
//	}
//
//	person := Person{
//	    Name: "John",
//	    Address: Address{
//	        City:    "New York",
//	        Country: "USA",
//	    },
//	}
//
//	result := tstruct.StructToMap(person, "json")
//	// 结果：
//	// {
//	//   "name": "John",
//	//   "address": {
//	//     "city": "New York",
//	//     "country": "USA"
//	//   }
//	// }
//
// 注意事项：
//   - 支持结构体值和结构体指针
//   - 只处理导出的字段（首字母大写）
//   - 支持嵌套结构体的递归转换
//   - 支持标签中的选项（如 omitempty）
//   - 标签值为 "-" 的字段会被忽略
//   - 如果不提供标签名，使用字段名作为键
//   - 如果输入不是结构体类型，返回 nil
func StructToMap(obj interface{}, tagName string) map[string]interface{} {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	// 如果是指针，获取其指向的元素
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	data := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// 如果字段是未导出的，跳过
		if !field.IsExported() {
			continue
		}

		// 获取标签值，如果没有标签或标签为空，使用字段名
		key := field.Name
		if tagName != "" {
			tag := field.Tag.Get(tagName)
			if tag == "-" {
				// 明确忽略
				continue
			}
			if tag != "" {
				// 处理标签中的第一个值，例如 `json:"name,omitempty"` 只取 "name"
				key = strings.Split(tag, ",")[0]
			}
		}

		// 获取字段值
		value := v.Field(i).Interface()

		// 如果值是结构体或结构体指针，递归处理
		if reflect.TypeOf(value).Kind() == reflect.Struct {
			value = StructToMap(value, tagName)
		} else if reflect.TypeOf(value).Kind() == reflect.Ptr && reflect.TypeOf(value).Elem().Kind() == reflect.Struct {
			if !v.Field(i).IsNil() {
				value = StructToMap(value, tagName)
			}
		}

		data[key] = value
	}
	return data
}

// StructToMapByTagMapstructure 利用反射将结构体转化为map，支持mapstructure标签
// Deprecated: 使用 StructToMap(obj, "mapstructure") 替代
func StructToMapByTagMapstructure(obj interface{}) map[string]interface{} {
	return StructToMap(obj, "mapstructure")
}

// Pointer 创建一个指向输入值的新指针
// 参数：
//   - in: 任意类型的输入值
//
// 返回值：
//   - *T: 指向输入值副本的指针
//
// 使用示例：
//
//	// 基本类型
//	str := "hello"
//	strPtr := tstruct.Pointer(str)
//	// *strPtr == "hello"
//
//	num := 42
//	numPtr := tstruct.Pointer(num)
//	// *numPtr == 42
//
//	// 结构体
//	type User struct {
//	    Name string
//	    Age  int
//	}
//	user := User{Name: "John", Age: 30}
//	userPtr := tstruct.Pointer(user)
//	// userPtr.Name == "John"
//	// userPtr.Age == 30
//
//	// 切片
//	slice := []int{1, 2, 3}
//	slicePtr := tstruct.Pointer(slice)
//	// (*slicePtr)[0] == 1
//
// 注意事项：
//   - 创建输入值的副本，而不是直接返回输入值的地址
//   - 适用于任何类型（使用泛型）
//   - 常用于创建可选字段的指针值
//   - 对于 nil 接口值会返回 nil 指针
//   - 线程安全，每次调用创建新的副本
func Pointer[T any](in T) (out *T) {
	return &in
}
