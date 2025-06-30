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
	"log"
	"math/rand"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//@author: [piexlmax](https://github.com/piexlmax)
//@function: StructToMap
//@description: 利用反射将结构体转化为map
//@param: obj interface{}
//@return: map[string]interface{}

func StructToMapByeTagmapstructure(obj interface{}) map[string]interface{} {
	obj1 := reflect.TypeOf(obj)
	obj2 := reflect.ValueOf(obj)

	data := make(map[string]interface{})
	for i := 0; i < obj1.NumField(); i++ {
		if obj1.Field(i).Tag.Get("mapstructure") != "" {
			data[obj1.Field(i).Tag.Get("mapstructure")] = obj2.Field(i).Interface()
		} else {
			data[obj1.Field(i).Name] = obj2.Field(i).Interface()
		}
	}
	return data
}

//@author: [piexlmax](https://github.com/piexlmax)
//@function: ArrayToString
//@description: 将数组格式化为字符串
//@param: array []interface{}
//@return: string

func ArrayToString(array []interface{}) string {
	return strings.Replace(strings.Trim(fmt.Sprint(array), "[]"), " ", ",", -1)
}

// Pointer 将值转换为指针
func Pointer[T any](in T) (out *T) {
	return &in
}

// FirstUpper 将字符串的首字母大写
func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// FirstLower 将字符串的首字母小写
func FirstLower(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// MaheHump 将字符串转换为驼峰命名，支持自定义分隔符
func MaheHump(s, delimiter string) string {
	// Replace the custom delimiter with a space
	s = strings.ReplaceAll(s, delimiter, " ")
	words := strings.Fields(s)
	c := cases.Title(language.Und)

	for i := 1; i < len(words); i++ {
		words[i] = c.String(words[i])
	}

	return strings.Join(words, "")
}

// 随机字符串
func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:',.<>?/~`")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[RandomInt(0, len(letters))]
	}
	return string(b)
}

// RandomInt 生成一个随机整数
func RandomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

// FilterInvisibleChars 过滤掉不可见字符
func FilterInvisibleChars(s string) string {
	resRunes := []rune{}
	for _, r := range s {
		// ASCII码，通常小于等于32或者大于等于127的都属于不可见字符
		if r > 32 && r < 127 {
			resRunes = append(resRunes, r)
		}
	}
	return string(resRunes)
}

// 将结构体转换为map,  obj 结构体或结构体指针
func StructToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	objValue := reflect.ValueOf(obj)

	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		fmt.Println("输入不是结构体类型")
		return result
	}

	typeOfObj := objValue.Type()
	for i := 0; i < objValue.NumField(); i++ {
		field := objValue.Field(i)
		fieldType := typeOfObj.Field(i)
		tag := fieldType.Tag.Get("json")
		if tag == "" {
			tag = fieldType.Name
		}
		result[tag] = field.Interface()
	}
	return result
}

// 将结构体转换为map[string]string
// obj 结构体或结构体指针
func StructToStringMap(obj interface{}) map[string]string {
	result := make(map[string]string)
	objValue := reflect.ValueOf(obj)

	if objValue.Kind() == reflect.Ptr {
		objValue = objValue.Elem()
	}

	if objValue.Kind() != reflect.Struct {
		log.Printf("输入不是结构体类型: %v \n", objValue.Kind())
		return result
	}

	typeOfObj := objValue.Type()
	for i := 0; i < objValue.NumField(); i++ {
		field := objValue.Field(i)
		fieldType := typeOfObj.Field(i)
		tag := fieldType.Tag.Get("json")
		if tag == "" {
			tag = fieldType.Name
		}
		var valueStr string
		switch field.Kind() {
		case reflect.String:
			valueStr = field.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			valueStr = strconv.FormatInt(field.Int(), 10)
		case reflect.Float32, reflect.Float64:
			valueStr = strconv.FormatFloat(field.Float(), 'f', -1, 64)
		default:
			valueStr = fmt.Sprintf("%v", field.Interface())
		}
		result[tag] = valueStr
	}
	return result
}

// MapToStruct 将 map 转换为结构体
// m map[string]interface{} key 是string，value 是任意类型的map
// s interface{} 结构体指针 &sample
func MapToStruct(m map[string]interface{}, s interface{}) error {
	// 将 map 转换为 JSON
	jsonData, err := json.Marshal(m)
	if err != nil {
		return err
	}

	// 将 JSON 转换为结构体
	return json.Unmarshal(jsonData, s)
}

// map -> struct
// map -> json -> struct

// struct -> map
// 反射最方便
