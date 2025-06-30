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

package util

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ReplacePlaceholders 按顺序替换字符串中的?为提供的替换值
func ReplacePlaceholders(query string, replacements []string) string {
	for _, replacement := range replacements {
		query = strings.Replace(query, "?", replacement, 1)
	}
	return query
}

// underscoreToCamelCase 将下划线格式的字符串转换为驼峰格式
func underscoreToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	var result strings.Builder
	titleCaser := cases.Title(language.Und)
	for i, part := range parts {
		if i > 0 {
			part = titleCaser.String(part)
		}
		result.WriteString(part)
	}
	return result.String()
}

// convertMapKeysToCamelCase 将 map 中的键转换为驼峰格式
func ConvertMapKeysToCamelCase(inputMap map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range inputMap {
		newKey := underscoreToCamelCase(k)
		newMap[newKey] = v
	}
	return newMap
}

// checkType 函数用于检查传入值的类型
func CheckType(v interface{}) {
	t := reflect.TypeOf(v)
	switch t.Kind() {
	case reflect.Bool:
		fmt.Println("The type is bool.")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if t.Kind() == reflect.Int32 {
			fmt.Println("The type is rune.")
		} else {
			fmt.Printf("The type is %v.\n", t.Kind())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		if t.Kind() == reflect.Uint8 {
			fmt.Println("The type is byte.")
		} else {
			fmt.Printf("The type is %v.\n", t.Kind())
		}
	case reflect.Float32, reflect.Float64:
		fmt.Printf("The type is %v.\n", t.Kind())
	case reflect.Complex64, reflect.Complex128:
		fmt.Printf("The type is %v.\n", t.Kind())
	case reflect.String:
		fmt.Println("The type is string.")
	case reflect.Array:
		fmt.Println("The type is array.")
	case reflect.Slice:
		fmt.Println("The type is slice.")
	case reflect.Map:
		fmt.Println("The type is map.")
	case reflect.Struct:
		fmt.Println("The type is struct.")
	case reflect.Ptr:
		fmt.Println("The type is pointer.")
	case reflect.Func:
		fmt.Println("The type is function.")
	case reflect.Chan:
		fmt.Println("The type is channel.")
	case reflect.Interface:
		fmt.Println("The type is interface.")
	default:
		fmt.Printf("The type is %v (unknown).\n", t.Kind())
	}
}

// ToJsonString 将任意类型转换为json字符串
func ToJsonString(v interface{}) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}

// ToStringWithFlag 将结构体、map、slice 的每个数据用 flag 分割，最终返回字符串
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
		// 遍历 map 的键值对
		keys := val.MapKeys()
		for i, key := range keys {
			if i > 0 {
				result.WriteString(flag)
			}
			value := val.MapIndex(key)
			result.WriteString(fmt.Sprintf("%v:%v", key.Interface(), value.Interface()))
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

// 过滤字符串中的所有数字
func FilterNumber(str string) string {
	re := regexp.MustCompile(`\d+`)
	return re.ReplaceAllString(str, "")
}

// @description: 去除结构体空格
// @param: target interface (target: 目标结构体,传入必须是指针类型)
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
