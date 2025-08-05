package jsonutil

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// 注意：这个测试文件需要在安装了json-iterator/go依赖后才能运行
// go get github.com/json-iterator/go

func TestNew(t *testing.T) {
	// 测试默认配置
	util := New()
	if util == nil {
		t.Fatal("New() 返回了 nil")
	}

	// 测试自定义配置
	config := Config{
		UseNumber:              true,
		TagKey:                 "json",
		OnlyTaggedField:        false,
		ValidateJsonRawMessage: true,
		CaseSensitive:          true,
		DisallowUnknownFields:  false,
	}
	util2 := New(config)
	if util2 == nil {
		t.Fatal("New(config) 返回了 nil")
	}
}

func TestParse(t *testing.T) {
	util := New()

	// 测试基本JSON解析
	jsonData := `{"id": 123, "name": "test", "active": true}`
	data, err := util.Parse([]byte(jsonData))
	if err != nil {
		t.Fatalf("Parse 失败: %v", err)
	}

	if len(data) != 3 {
		t.Errorf("期望3个字段，实际得到 %d 个", len(data))
	}

	// 测试空JSON
	emptyData, err := util.Parse([]byte("{}"))
	if err != nil {
		t.Fatalf("解析空JSON失败: %v", err)
	}
	if len(emptyData) != 0 {
		t.Errorf("期望0个字段，实际得到 %d 个", len(emptyData))
	}

	// 测试无效JSON
	_, err = util.Parse([]byte(`{"id": 123, "name": "test"`))
	if err == nil {
		t.Error("期望解析无效JSON时返回错误")
	}
}

func TestParseStringFunc(t *testing.T) {
	util := New()

	// 测试基本字符串解析
	jsonStr := `{"id": 123, "name": "test"}`
	data, err := util.ParseString(jsonStr)
	if err != nil {
		t.Fatalf("ParseString 失败: %v", err)
	}

	if len(data) != 2 {
		t.Errorf("期望2个字段，实际得到 %d 个", len(data))
	}

	// 测试空字符串
	_, err = util.ParseString("")
	if err == nil {
		t.Error("期望解析空字符串时返回错误")
	}
}

func TestParseReader(t *testing.T) {
	util := New()

	// 测试从Reader解析
	jsonStr := `{"id": 123, "name": "test"}`
	reader := strings.NewReader(jsonStr)
	data, err := util.ParseReader(reader)
	if err != nil {
		t.Fatalf("ParseReader 失败: %v", err)
	}

	if len(data) != 2 {
		t.Errorf("期望2个字段，实际得到 %d 个", len(data))
	}
}

func TestParseRequest(t *testing.T) {
	util := New()

	// 测试HTTP请求解析
	jsonStr := `{"id": 123, "name": "test"}`
	req, _ := http.NewRequest("POST", "/test", strings.NewReader(jsonStr))
	data, err := util.ParseRequest(req)
	if err != nil {
		t.Fatalf("ParseRequest 失败: %v", err)
	}

	if len(data) != 2 {
		t.Errorf("期望2个字段，实际得到 %d 个", len(data))
	}
}

func TestNumberConversion(t *testing.T) {
	util := New()

	// 测试各种数字类型的转换
	testCases := []struct {
		name     string
		jsonStr  string
		expected map[string]interface{}
	}{
		{
			name:    "整数",
			jsonStr: `{"id": 123}`,
			expected: map[string]interface{}{
				"id": int64(123),
			},
		},
		{
			name:    "浮点数",
			jsonStr: `{"score": 99.5}`,
			expected: map[string]interface{}{
				"score": float64(99.5),
			},
		},
		{
			name:    "整数值的浮点数",
			jsonStr: `{"count": 100.0}`,
			expected: map[string]interface{}{
				"count": int64(100),
			},
		},
		{
			name:    "数字字符串",
			jsonStr: `{"age": "25"}`,
			expected: map[string]interface{}{
				"age": "25", // 字符串保持为字符串
			},
		},
		{
			name:    "浮点数字符串",
			jsonStr: `{"price": "99.99"}`,
			expected: map[string]interface{}{
				"price": "99.99", // 字符串保持为字符串
			},
		},
		{
			name:    "混合类型",
			jsonStr: `{"id": 123, "score": 99.5, "age": "25", "price": "99.99"}`,
			expected: map[string]interface{}{
				"id":    int64(123),
				"score": float64(99.5),
				"age":   "25",    // 字符串保持为字符串
				"price": "99.99", // 字符串保持为字符串
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := util.ParseString(tc.jsonStr)
			if err != nil {
				t.Fatalf("解析失败: %v", err)
			}

			for key, expectedValue := range tc.expected {
				actualValue, exists := data[key]
				if !exists {
					t.Errorf("字段 %s 不存在", key)
					continue
				}

				if !reflect.DeepEqual(actualValue, expectedValue) {
					t.Errorf("字段 %s: 期望 %v (%T), 实际得到 %v (%T)",
						key, expectedValue, expectedValue, actualValue, actualValue)
				}
			}
		})
	}
}

func TestGetMethods(t *testing.T) {
	util := New()

	// 准备测试数据
	jsonStr := `{
		"id": 123,
		"name": "test",
		"score": 99.5,
		"active": true,
		"tags": ["tag1", "tag2"],
		"profile": {"age": 25, "height": 175.5},
		"count": 100.0,
		"price": "99.99",
		"enabled": "true",
		"disabled": "false",
		"zero": 0,
		"one": 1
	}`

	data, err := util.ParseString(jsonStr)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	// 测试 GetInt64
	t.Run("GetInt64", func(t *testing.T) {
		if id, ok := util.GetInt64(data, "id"); !ok || id != 123 {
			t.Errorf("GetInt64(id): 期望 123, 实际得到 %v, ok=%v", id, ok)
		}

		if count, ok := util.GetInt64(data, "count"); !ok || count != 100 {
			t.Errorf("GetInt64(count): 期望 100, 实际得到 %v, ok=%v", count, ok)
		}

		if _, ok := util.GetInt64(data, "score"); ok {
			t.Error("GetInt64(score): 期望失败，因为score是浮点数")
		}

		if _, ok := util.GetInt64(data, "missing"); ok {
			t.Error("GetInt64(missing): 期望失败，因为字段不存在")
		}
	})

	// 测试 GetInt
	t.Run("GetInt", func(t *testing.T) {
		if id, ok := util.GetInt(data, "id"); !ok || id != 123 {
			t.Errorf("GetInt(id): 期望 123, 实际得到 %v, ok=%v", id, ok)
		}
	})

	// 测试 GetFloat64
	t.Run("GetFloat64", func(t *testing.T) {
		if score, ok := util.GetFloat64(data, "score"); !ok || score != 99.5 {
			t.Errorf("GetFloat64(score): 期望 99.5, 实际得到 %v, ok=%v", score, ok)
		}

		if price, ok := util.GetFloat64(data, "price"); !ok || price != 99.99 {
			t.Errorf("GetFloat64(price): 期望 99.99, 实际得到 %v, ok=%v", price, ok)
		}
	})

	// 测试 GetString
	t.Run("GetString", func(t *testing.T) {
		if name, ok := util.GetString(data, "name"); !ok || name != "test" {
			t.Errorf("GetString(name): 期望 test, 实际得到 %v, ok=%v", name, ok)
		}

		if idStr, ok := util.GetString(data, "id"); !ok || idStr != "123" {
			t.Errorf("GetString(id): 期望 123, 实际得到 %v, ok=%v", idStr, ok)
		}
	})

	// 测试 GetBool
	t.Run("GetBool", func(t *testing.T) {
		if active, ok := util.GetBool(data, "active"); !ok || !active {
			t.Errorf("GetBool(active): 期望 true, 实际得到 %v, ok=%v", active, ok)
		}

		if enabled, ok := util.GetBool(data, "enabled"); !ok || !enabled {
			t.Errorf("GetBool(enabled): 期望 true, 实际得到 %v, ok=%v", enabled, ok)
		}

		if disabled, ok := util.GetBool(data, "disabled"); !ok || disabled {
			t.Errorf("GetBool(disabled): 期望 false, 实际得到 %v, ok=%v", disabled, ok)
		}

		if zero, ok := util.GetBool(data, "zero"); !ok || zero {
			t.Errorf("GetBool(zero): 期望 false, 实际得到 %v, ok=%v", zero, ok)
		}

		if one, ok := util.GetBool(data, "one"); !ok || !one {
			t.Errorf("GetBool(one): 期望 true, 实际得到 %v, ok=%v", one, ok)
		}
	})

	// 测试 GetMap
	t.Run("GetMap", func(t *testing.T) {
		if profile, ok := util.GetMap(data, "profile"); !ok {
			t.Error("GetMap(profile): 期望成功")
		} else {
			if len(profile) != 2 {
				t.Errorf("GetMap(profile): 期望2个字段，实际得到 %d 个", len(profile))
			}
		}

		if _, ok := util.GetMap(data, "id"); ok {
			t.Error("GetMap(id): 期望失败，因为id不是map")
		}
	})

	// 测试 GetArray
	t.Run("GetArray", func(t *testing.T) {
		if tags, ok := util.GetArray(data, "tags"); !ok {
			t.Error("GetArray(tags): 期望成功")
		} else {
			if len(tags) != 2 {
				t.Errorf("GetArray(tags): 期望2个元素，实际得到 %d 个", len(tags))
			}
		}

		if _, ok := util.GetArray(data, "id"); ok {
			t.Error("GetArray(id): 期望失败，因为id不是数组")
		}
	})
}

func TestGetStruct(t *testing.T) {

	util := New()

	type Profile struct {
		Age    int     `json:"age"`
		Height float64 `json:"height"`
	}

	jsonStr := `{
		"user": {
			"id": 123,
			"profile": {
				"age": 25,
				"height": 175.5
			}
		}
	}`

	data, err := util.ParseString(jsonStr)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	if user, ok := util.GetMap(data, "user"); ok {
		var profileStruct Profile
		if !util.GetStruct(user, "profile", &profileStruct) {
			t.Error("GetStruct 失败")
		} else {
			if profileStruct.Age != 25 {
				t.Errorf("期望 Age=25, 实际得到 %d", profileStruct.Age)
			}
			if profileStruct.Height != 175.5 {
				t.Errorf("期望 Height=175.5, 实际得到 %f", profileStruct.Height)
			}
		}
	}
}

func TestNestedStructures(t *testing.T) {
	util := New()

	jsonStr := `{
		"user": {
			"id": 123,
			"profile": {
				"age": 25,
				"height": 175.5,
				"preferences": {
					"theme": "dark",
					"notifications": true
				}
			},
			"tags": ["tag1", "tag2", 123],
			"scores": [99.5, 88.0, 92.5]
		}
	}`

	data, err := util.ParseString(jsonStr)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}

	// 测试嵌套map
	if user, ok := util.GetMap(data, "user"); ok {
		if id, ok := util.GetInt64(user, "id"); !ok || id != 123 {
			t.Errorf("嵌套map获取id失败: %v", id)
		}

		if profile, ok := util.GetMap(user, "profile"); ok {
			if age, ok := util.GetInt64(profile, "age"); !ok || age != 25 {
				t.Errorf("嵌套map获取age失败: %v", age)
			}

			if preferences, ok := util.GetMap(profile, "preferences"); ok {
				if theme, ok := util.GetString(preferences, "theme"); !ok || theme != "dark" {
					t.Errorf("嵌套map获取theme失败: %v", theme)
				}
			}
		}
	}

	// 测试嵌套数组
	if user, ok := util.GetMap(data, "user"); ok {
		if tags, ok := util.GetArray(user, "tags"); ok {
			if len(tags) != 3 {
				t.Errorf("期望tags有3个元素，实际得到 %d 个", len(tags))
			}

			// 检查数组中的混合类型
			if tag1, ok := tags[0].(string); !ok || tag1 != "tag1" {
				t.Errorf("tags[0] 期望 tag1, 实际得到 %v", tags[0])
			}

			if tag3, ok := tags[2].(int64); !ok || tag3 != 123 {
				t.Errorf("tags[2] 期望 123, 实际得到 %v", tags[2])
			}
		}

		if scores, ok := util.GetArray(user, "scores"); ok {
			if len(scores) != 3 {
				t.Errorf("期望scores有3个元素，实际得到 %d 个", len(scores))
			}

			// 检查数组中的浮点数
			if score1, ok := scores[0].(float64); !ok || score1 != 99.5 {
				t.Errorf("scores[0] 期望 99.5, 实际得到 %v", scores[0])
			}
		}
	}
}

func TestValidation(t *testing.T) {
	util := New()

	// 测试JSON格式验证
	t.Run("Validate", func(t *testing.T) {
		validJSON := []byte(`{"id": 123, "name": "test"}`)
		if err := util.Validate(validJSON); err != nil {
			t.Errorf("验证有效JSON失败: %v", err)
		}

		invalidJSON := []byte(`{"id": 123, "name": "test"`)
		if err := util.Validate(invalidJSON); err == nil {
			t.Error("验证无效JSON应该返回错误")
		}
	})

	// 测试字符串验证
	t.Run("ValidateString", func(t *testing.T) {
		validStr := `{"id": 123, "name": "test"}`
		if err := util.ValidateString(validStr); err != nil {
			t.Errorf("验证有效字符串失败: %v", err)
		}

		invalidStr := `{"id": 123, "name": "test"`
		if err := util.ValidateString(invalidStr); err == nil {
			t.Error("验证无效字符串应该返回错误")
		}
	})

	// 测试必需字段验证
	t.Run("ValidateRequired", func(t *testing.T) {
		data := map[string]interface{}{
			"id":   int64(123),
			"name": "test",
		}

		if err := util.ValidateRequired(data, "id", "name"); err != nil {
			t.Errorf("验证必需字段失败: %v", err)
		}

		if err := util.ValidateRequired(data, "id", "name", "missing"); err == nil {
			t.Error("验证缺失字段应该返回错误")
		}
	})

	// 测试类型验证
	t.Run("ValidateType", func(t *testing.T) {
		data := map[string]interface{}{
			"id":    int64(123),
			"name":  "test",
			"score": float64(99.5),
		}

		if err := util.ValidateType(data, "id", reflect.Int64); err != nil {
			t.Errorf("验证int64类型失败: %v", err)
		}

		if err := util.ValidateType(data, "name", reflect.String); err != nil {
			t.Errorf("验证string类型失败: %v", err)
		}

		if err := util.ValidateType(data, "score", reflect.Float64); err != nil {
			t.Errorf("验证float64类型失败: %v", err)
		}

		if err := util.ValidateType(data, "id", reflect.String); err == nil {
			t.Error("验证错误类型应该返回错误")
		}

		if err := util.ValidateType(data, "missing", reflect.Int64); err == nil {
			t.Error("验证缺失字段应该返回错误")
		}
	})
}

func TestMarshal(t *testing.T) {

	util := New()

	// 测试序列化map
	t.Run("MarshalMap", func(t *testing.T) {
		data := map[string]interface{}{
			"id":     int64(123),
			"name":   "test",
			"score":  99.5,
			"active": true,
		}

		bytes, err := util.Marshal(data)
		if err != nil {
			t.Fatalf("序列化失败: %v", err)
		}

		// 验证序列化结果
		var result map[string]interface{}
		if err := json.Unmarshal(bytes, &result); err != nil {
			t.Fatalf("反序列化失败: %v", err)
		}

		if len(result) != 4 {
			t.Errorf("期望4个字段，实际得到 %d 个", len(result))
		}
	})

	// 测试序列化为字符串
	t.Run("MarshalToString", func(t *testing.T) {
		data := map[string]interface{}{
			"id":   int64(123),
			"name": "test",
		}

		str, err := util.MarshalToString(data)
		if err != nil {
			t.Fatalf("序列化为字符串失败: %v", err)
		}

		if !strings.Contains(str, `"id":123`) {
			t.Errorf("序列化结果不包含id字段: %s", str)
		}

		if !strings.Contains(str, `"name":"test"`) {
			t.Errorf("序列化结果不包含name字段: %s", str)
		}
	})
}

func TestStructUnmarshal(t *testing.T) {
	util := New()

	type User struct {
		ID     int64   `json:"id"`
		Name   string  `json:"name"`
		Age    int     `json:"age"`
		Score  float64 `json:"score"`
		Active bool    `json:"active"`
	}

	// 测试字节数组解析
	t.Run("Unmarshal", func(t *testing.T) {
		jsonData := []byte(`{"id": 123, "name": "test", "age": 25, "score": 99.5, "active": true}`)
		var user User

		err := util.Unmarshal(jsonData, &user)
		if err != nil {
			t.Fatalf("解析结构体失败: %v", err)
		}

		if user.ID != 123 {
			t.Errorf("期望 ID=123, 实际得到 %d", user.ID)
		}
		if user.Name != "test" {
			t.Errorf("期望 Name=test, 实际得到 %s", user.Name)
		}
		if user.Age != 25 {
			t.Errorf("期望 Age=25, 实际得到 %d", user.Age)
		}
		if user.Score != 99.5 {
			t.Errorf("期望 Score=99.5, 实际得到 %f", user.Score)
		}
		if !user.Active {
			t.Errorf("期望 Active=true, 实际得到 %v", user.Active)
		}
	})

	// 测试字符串解析
	t.Run("UnmarshalString", func(t *testing.T) {
		jsonStr := `{"id": 456, "name": "张三", "age": 30, "score": 88.5, "active": false}`
		var user User

		err := util.UnmarshalString(jsonStr, &user)
		if err != nil {
			t.Fatalf("解析字符串到结构体失败: %v", err)
		}

		if user.ID != 456 {
			t.Errorf("期望 ID=456, 实际得到 %d", user.ID)
		}
		if user.Name != "张三" {
			t.Errorf("期望 Name=张三, 实际得到 %s", user.Name)
		}
	})
}

func TestEdgeCases(t *testing.T) {
	util := New()

	// 测试空值
	t.Run("EmptyValues", func(t *testing.T) {
		jsonStr := `{"null": null, "empty": "", "zero": 0, "false": false}`
		data, err := util.ParseString(jsonStr)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		// 测试null值
		if value, exists := util.Get(data, "null"); !exists {
			t.Error("null字段应该存在")
		} else if value != nil {
			t.Errorf("null字段应该为nil，实际得到 %v", value)
		}

		// 测试空字符串
		if value, ok := util.GetString(data, "empty"); !ok || value != "" {
			t.Errorf("empty字段期望空字符串，实际得到 %v", value)
		}

		// 测试零值
		if value, ok := util.GetInt64(data, "zero"); !ok || value != 0 {
			t.Errorf("zero字段期望0，实际得到 %v", value)
		}

		// 测试false
		if value, ok := util.GetBool(data, "false"); !ok || value {
			t.Errorf("false字段期望false，实际得到 %v", value)
		}
	})

	// 测试大数字
	t.Run("LargeNumbers", func(t *testing.T) {
		jsonStr := `{"bigInt": 9223372036854775807, "bigFloat": 1.7976931348623157e+308}`
		data, err := util.ParseString(jsonStr)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if value, ok := util.GetInt64(data, "bigInt"); !ok || value != 9223372036854775807 {
			t.Errorf("bigInt字段期望 9223372036854775807，实际得到 %v", value)
		}

		if value, ok := util.GetFloat64(data, "bigFloat"); !ok || value != 1.7976931348623157e+308 {
			t.Errorf("bigFloat字段期望 1.7976931348623157e+308，实际得到 %v", value)
		}
	})

	// 测试特殊字符
	t.Run("SpecialCharacters", func(t *testing.T) {
		jsonStr := `{"unicode": "中文测试", "special": "!@#$%^&*()", "newline": "line1\nline2"}`
		data, err := util.ParseString(jsonStr)
		if err != nil {
			t.Fatalf("解析失败: %v", err)
		}

		if value, ok := util.GetString(data, "unicode"); !ok || value != "中文测试" {
			t.Errorf("unicode字段期望 中文测试，实际得到 %v", value)
		}

		if value, ok := util.GetString(data, "special"); !ok || value != "!@#$%^&*()" {
			t.Errorf("special字段期望 !@#$%%^&*()，实际得到 %v", value)
		}

		if value, ok := util.GetString(data, "newline"); !ok || value != "line1\nline2" {
			t.Errorf("newline字段期望 line1\\nline2，实际得到 %v", value)
		}
	})
}

func TestErrorHandling(t *testing.T) {
	util := New()

	// 测试各种错误情况
	testCases := []struct {
		name        string
		jsonStr     string
		expectError bool
	}{
		{"无效JSON", `{"id": 123, "name": "test"`, true},
		{"空字符串", "", true},
		{"null", "null", false},
		{"空对象", "{}", false},
		{"空数组", "[]", true},
		{"单个值", "123", true},
		{"字符串", `"test"`, true},
		{"布尔值", "true", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := util.ParseString(tc.jsonStr)
			if tc.expectError && err == nil {
				t.Error("期望返回错误，但没有")
			}
			if !tc.expectError && err != nil {
				t.Errorf("不期望返回错误，但得到了: %v", err)
			}
		})
	}
}

func TestConcurrentUsage(t *testing.T) {
	util := New()
	jsonStr := `{"id": 123, "name": "test", "score": 99.5}`

	// 测试并发使用
	t.Run("ConcurrentParse", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- true }()

				data, err := util.ParseString(jsonStr)
				if err != nil {
					t.Errorf("并发解析失败: %v", err)
					return
				}

				if id, ok := util.GetInt64(data, "id"); !ok || id != 123 {
					t.Errorf("并发获取id失败: %v", id)
				}
			}()
		}

		for i := 0; i < numGoroutines; i++ {
			<-done
		}
	})
}

func TestPerformance(t *testing.T) {
	util := New()

	// 创建大量测试数据
	largeJSON := `{"data": [`
	for i := 0; i < 10; i++ {
		if i > 0 {
			largeJSON += ","
		}
		largeJSON += fmt.Sprintf(`{"id": %d, "name": "user%d", "score": %d.5}`, i, i, i)
	}
	largeJSON += `]}`

	// 测试解析性能
	t.Run("ParsePerformance", func(t *testing.T) {
		start := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := util.ParseString(largeJSON)
				if err != nil {
					b.Fatalf("解析失败: %v", err)
				}
			}
		})

		t.Logf("解析性能: %s", start.String())
	})

	log.Println(largeJSON)

	// 测试序列化性能
	t.Run("MarshalPerformance", func(t *testing.T) {
		data, _ := util.ParseString(largeJSON)

		start := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := util.Marshal(data)
				if err != nil {
					b.Fatalf("序列化失败: %v", err)
				}
			}
		})

		t.Logf("序列化性能: %s", start.String())
	})
}

// 基准测试
func BenchmarkParse(b *testing.B) {
	util := New()
	jsonStr := `{"id": 123, "name": "test", "score": 99.5, "active": true, "tags": ["tag1", "tag2"], "profile": {"age": 25, "height": 175.5}}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := util.ParseString(jsonStr)
		if err != nil {
			b.Fatalf("解析失败: %v", err)
		}
	}
}

func BenchmarkGetInt64(b *testing.B) {
	util := New()
	jsonStr := `{"id": 123, "count": 100, "score": 99.5}`
	data, _ := util.ParseString(jsonStr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		util.GetInt64(data, "id")
		util.GetInt64(data, "count")
	}
}

func BenchmarkMarshal(b *testing.B) {
	util := New()
	data := map[string]interface{}{
		"id":     int64(123),
		"name":   "test",
		"score":  99.5,
		"active": true,
		"tags":   []string{"tag1", "tag2"},
		"profile": map[string]interface{}{
			"age":    25,
			"height": 175.5,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := util.Marshal(data)
		if err != nil {
			b.Fatalf("序列化失败: %v", err)
		}
	}
}
