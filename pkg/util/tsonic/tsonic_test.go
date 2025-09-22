package tsonic

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name:   "默认配置",
			config: DefaultConfig,
		},
		{
			name: "自定义配置",
			config: Config{
				UseNumber:             false,
				TagKey:                "custom",
				CaseSensitive:         false,
				DisallowUnknownFields: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			util := New(tt.config)
			assert.Equal(t, tt.config, util.config)
		})
	}
}

func TestJSONUtil_Parse(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "基本类型",
			json: `{
				"string": "hello",
				"number": 123,
				"float": 123.45,
				"bool": true,
				"null": null
			}`,
			want: map[string]interface{}{
				"string": "hello",
				"number": float64(123),
				"float":  123.45,
				"bool":   true,
				"null":   nil,
			},
			wantErr: false,
		},
		{
			name: "嵌套对象",
			json: `{
				"object": {
					"key": "value"
				},
				"array": [1, 2, 3]
			}`,
			want: map[string]interface{}{
				"object": map[string]interface{}{
					"key": "value",
				},
				"array": []interface{}{float64(1), float64(2), float64(3)},
			},
			wantErr: false,
		},
		{
			name:    "无效JSON",
			json:    `{"invalid": }`,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "空JSON",
			json:    "",
			want:    nil,
			wantErr: true,
		},
	}

	util := Default
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.Parse([]byte(tt.json))
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestJSONUtil_ParseString(t *testing.T) {
	util := Default
	result, err := util.ParseString(`{"name": "张三", "age": 25}`)
	assert.NoError(t, err)
	assert.Equal(t, "张三", result["name"])
	assert.Equal(t, float64(25), result["age"])
}

func TestJSONUtil_ParseReader(t *testing.T) {
	util := Default
	reader := strings.NewReader(`{"name": "张三", "age": 25}`)
	result, err := util.ParseReader(reader)
	assert.NoError(t, err)
	assert.Equal(t, "张三", result["name"])
	assert.Equal(t, float64(25), result["age"])
}

func TestJSONUtil_ParseRequest(t *testing.T) {
	util := Default
	jsonStr := `{"name": "张三", "age": 25}`
	req, _ := http.NewRequest("POST", "/test", bytes.NewBufferString(jsonStr))
	result, err := util.ParseRequest(req)
	assert.NoError(t, err)
	assert.Equal(t, "张三", result["name"])
	assert.Equal(t, float64(25), result["age"])
}

type testStruct struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Hobbies []string `json:"hobbies"`
}

func TestJSONUtil_UnmarshalAndMarshal(t *testing.T) {
	util := Default
	jsonStr := `{
		"name": "张三",
		"age": 25,
		"hobbies": ["读书", "运动"]
	}`

	// 测试 Unmarshal
	var data testStruct
	err := util.UnmarshalString(jsonStr, &data)
	assert.NoError(t, err)
	assert.Equal(t, "张三", data.Name)
	assert.Equal(t, 25, data.Age)
	assert.Equal(t, []string{"读书", "运动"}, data.Hobbies)

	// 测试 Marshal
	bytes, err := util.Marshal(data)
	assert.NoError(t, err)
	var result testStruct
	err = json.Unmarshal(bytes, &result)
	assert.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestJSONUtil_GetMethods(t *testing.T) {
	util := Default
	data := map[string]interface{}{
		"string":  "hello",
		"int":     123,
		"float":   123.45,
		"bool":    true,
		"array":   []interface{}{1, 2, 3},
		"object":  map[string]interface{}{"key": "value"},
		"null":    nil,
		"invalid": make(chan int),
	}

	// 测试 GetString
	str, ok := util.GetString(data, "string")
	assert.True(t, ok)
	assert.Equal(t, "hello", str)

	// 测试 GetInt
	num, ok := util.GetInt(data, "int")
	assert.True(t, ok)
	assert.Equal(t, 123, num)

	// 测试 GetFloat64
	float, ok := util.GetFloat64(data, "float")
	assert.True(t, ok)
	assert.Equal(t, 123.45, float)

	// 测试 GetBool
	boolean, ok := util.GetBool(data, "bool")
	assert.True(t, ok)
	assert.Equal(t, true, boolean)

	// 测试 GetArray
	array, ok := util.GetArray(data, "array")
	assert.True(t, ok)
	assert.Equal(t, []interface{}{1, 2, 3}, array)

	// 测试 GetMap
	object, ok := util.GetMap(data, "object")
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"key": "value"}, object)

	// 测试不存在的键
	_, ok = util.GetString(data, "nonexistent")
	assert.False(t, ok)

	// 测试类型不匹配
	_, ok = util.GetString(data, "array")
	assert.False(t, ok)
}

func TestJSONUtil_Validate(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "有效JSON",
			json:    `{"name": "张三", "age": 25}`,
			wantErr: false,
		},
		{
			name:    "无效JSON",
			json:    `{"name": "张三", age: 25}`,
			wantErr: true,
		},
		{
			name:    "空JSON",
			json:    "",
			wantErr: true,
		},
		{
			name:    "复杂JSON",
			json:    `{"array": [1, 2, 3], "object": {"key": "value"}, "null": null}`,
			wantErr: false,
		},
	}

	util := Default
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := util.ValidateString(tt.json)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJSONUtil_ValidateRequired(t *testing.T) {
	util := Default
	data := map[string]interface{}{
		"name":  "张三",
		"age":   25,
		"email": "zhangsan@example.com",
	}

	// 测试存在的字段
	err := util.ValidateRequired(data, "name", "age")
	assert.NoError(t, err)

	// 测试不存在的字段
	err = util.ValidateRequired(data, "name", "phone")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "phone")
}

func TestJSONUtil_ValidateType(t *testing.T) {
	util := Default
	data := map[string]interface{}{
		"string": "hello",
		"int":    123,
		"bool":   true,
		"array":  []interface{}{1, 2, 3},
		"object": map[string]interface{}{"key": "value"},
	}

	tests := []struct {
		name         string
		field        string
		expectedType reflect.Kind
		wantErr      bool
	}{
		{"字符串类型", "string", reflect.String, false},
		{"整数类型", "int", reflect.Int, false},
		{"布尔类型", "bool", reflect.Bool, false},
		{"数组类型", "array", reflect.Slice, false},
		{"对象类型", "object", reflect.Map, false},
		{"不存在的字段", "nonexistent", reflect.String, true},
		{"类型不匹配", "string", reflect.Int, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := util.ValidateType(data, tt.field, tt.expectedType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTypeConversions(t *testing.T) {
	util := Default
	data := map[string]interface{}{
		"int_str":   "123",
		"float_str": "123.45",
		"bool_str":  "true",
		"bool_int":  1,
		"bool_zero": 0,
		"float_int": 123.0,
		"invalid":   make(chan int),
	}

	// 测试字符串转整数
	num, ok := util.GetInt(data, "int_str")
	assert.True(t, ok)
	assert.Equal(t, 123, num)

	// 测试字符串转浮点数
	float, ok := util.GetFloat64(data, "float_str")
	assert.True(t, ok)
	assert.Equal(t, 123.45, float)

	// 测试字符串转布尔值
	boolean, ok := util.GetBool(data, "bool_str")
	assert.True(t, ok)
	assert.Equal(t, true, boolean)

	// 测试整数转布尔值
	boolean, ok = util.GetBool(data, "bool_int")
	assert.True(t, ok)
	assert.Equal(t, true, boolean)

	// 测试零值转布尔值
	boolean, ok = util.GetBool(data, "bool_zero")
	assert.True(t, ok)
	assert.Equal(t, false, boolean)

	// 测试浮点数转整数
	num, ok = util.GetInt(data, "float_int")
	assert.True(t, ok)
	assert.Equal(t, 123, num)

	// 测试无效类型转换
	_, ok = util.GetInt(data, "invalid")
	assert.False(t, ok)
}
