package tjson

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestParseAndMarshal(t *testing.T) {
	// 测试数据
	jsonData := []byte(`{
		"string": "hello",
		"number": 123,
		"float": 123.45,
		"bool": true,
		"array": [1, 2, 3],
		"object": {
			"key": "value"
		},
		"null": null
	}`)

	// 测试Parse
	parsed, err := Default.Parse(jsonData)
	if err != nil {
		t.Errorf("Parse() error = %v", err)
		return
	}

	// 验证解析结果
	tests := []struct {
		name     string
		key      string
		wantType reflect.Kind
		want     interface{}
	}{
		{"string", "string", reflect.String, "hello"},
		{"number", "number", reflect.Int64, int64(123)},
		{"float", "float", reflect.Float64, 123.45},
		{"bool", "bool", reflect.Bool, true},
		{"array", "array", reflect.Slice, []interface{}{int64(1), int64(2), int64(3)}},
		{"object", "object", reflect.Map, map[string]interface{}{"key": "value"}},
		{"null", "null", reflect.Invalid, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, exists := parsed[tt.key]
			if !exists {
				t.Errorf("key %s not found", tt.key)
				return
			}

			if got == nil {
				if tt.wantType != reflect.Invalid {
					t.Errorf("got nil, want type %v", tt.wantType)
				}
				return
			}

			gotType := reflect.TypeOf(got).Kind()
			if gotType != tt.wantType {
				t.Errorf("type = %v, want %v", gotType, tt.wantType)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("value = %v, want %v", got, tt.want)
			}
		})
	}

	// 测试Marshal
	marshaled, err := Default.Marshal(parsed)
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
		return
	}

	// 验证序列化结果
	var reparsed map[string]interface{}
	if err := json.Unmarshal(marshaled, &reparsed); err != nil {
		t.Errorf("json.Unmarshal() error = %v", err)
		return
	}

	// 比较关键字段的值，而不是严格比较整个结构
	// 因为Marshal过程中类型可能发生变化
	for key, originalValue := range parsed {
		if originalValue == nil {
			if reparsed[key] != nil {
				t.Errorf("Marshal() result for key '%s' should be nil", key)
			}
			continue
		}

		reparsedValue := reparsed[key]
		if reparsedValue == nil {
			t.Errorf("Marshal() result for key '%s' is nil, want %v", key, originalValue)
			continue
		}

		// 对于数字类型，比较数值而不是类型
		switch v := originalValue.(type) {
		case int64:
			if float64(v) != reparsedValue.(float64) {
				t.Errorf("Marshal() result for key '%s' = %v, want %v", key, reparsedValue, v)
			}
		case float64:
			if v != reparsedValue.(float64) {
				t.Errorf("Marshal() result for key '%s' = %v, want %v", key, reparsedValue, v)
			}
		case []interface{}:
			// 对数组进行特殊处理
			if reparsedArray, ok := reparsedValue.([]interface{}); ok {
				if len(v) != len(reparsedArray) {
					t.Errorf("Marshal() result for key '%s' array length = %d, want %d", key, len(reparsedArray), len(v))
					continue
				}
				for i, originalItem := range v {
					reparsedItem := reparsedArray[i]
					// 比较数组元素的值，考虑类型转换
					if !reflect.DeepEqual(originalItem, reparsedItem) {
						// 尝试类型转换后比较
						converted := false
						switch orig := originalItem.(type) {
						case int64:
							if reparsedFloat, ok := reparsedItem.(float64); ok {
								if float64(orig) == reparsedFloat {
									converted = true
								}
							}
						case float64:
							if reparsedFloat, ok := reparsedItem.(float64); ok {
								if orig == reparsedFloat {
									converted = true
								}
							}
						}
						if !converted {
							t.Errorf("Marshal() result for key '%s'[%d] = %v (type: %T), want %v (type: %T)",
								key, i, reparsedItem, reparsedItem, originalItem, originalItem)
						}
					}
				}
			} else {
				t.Errorf("Marshal() result for key '%s' is not an array", key)
			}
		default:
			if !reflect.DeepEqual(originalValue, reparsedValue) {
				t.Errorf("Marshal() result for key '%s' = %v, want %v", key, reparsedValue, originalValue)
			}
		}
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid json",
			input:   `{"name": "test"}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{"name": "test"`,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   \n\t  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Default.ParseString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseString() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid json",
			input:   `{"name": "test"}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{"name": "test"`,
			wantErr: true,
		},
		{
			name:    "read error",
			input:   "error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reader io.Reader
			if tt.input == "error" {
				reader = &errorReader{}
			} else {
				reader = strings.NewReader(tt.input)
			}

			_, err := Default.ParseReader(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestParseRequest(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantErr  bool
		wantData map[string]interface{}
	}{
		{
			name:    "valid json",
			body:    `{"name": "test"}`,
			wantErr: false,
			wantData: map[string]interface{}{
				"name": "test",
			},
		},
		{
			name:     "invalid json",
			body:     `{"name": "test"`,
			wantErr:  true,
			wantData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", strings.NewReader(tt.body))
			got, err := Default.ParseRequest(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.wantData) {
				t.Errorf("ParseRequest() = %v, want %v", got, tt.wantData)
			}
		})
	}
}

func TestGetMethods(t *testing.T) {
	// 测试数据
	data := map[string]interface{}{
		"int":     123,
		"float":   123.45,
		"string":  "hello",
		"bool":    true,
		"array":   []interface{}{1, 2, 3},
		"map":     map[string]interface{}{"key": "value"},
		"invalid": make(chan int),
	}

	// 测试GetInt
	if val, ok := Default.GetInt(data, "int"); !ok || val != 123 {
		t.Errorf("GetInt() = %v, %v, want 123, true", val, ok)
	}

	// 测试GetFloat64
	if val, ok := Default.GetFloat64(data, "float"); !ok || val != 123.45 {
		t.Errorf("GetFloat64() = %v, %v, want 123.45, true", val, ok)
	}

	// 测试GetString
	if val, ok := Default.GetString(data, "string"); !ok || val != "hello" {
		t.Errorf("GetString() = %v, %v, want hello, true", val, ok)
	}

	// 测试GetBool
	if val, ok := Default.GetBool(data, "bool"); !ok || !val {
		t.Errorf("GetBool() = %v, %v, want true, true", val, ok)
	}

	// 测试GetArray
	if val, ok := Default.GetArray(data, "array"); !ok || !reflect.DeepEqual(val, []interface{}{1, 2, 3}) {
		t.Errorf("GetArray() = %v, %v, want [1 2 3], true", val, ok)
	}

	// 测试GetMap
	if val, ok := Default.GetMap(data, "map"); !ok || !reflect.DeepEqual(val, map[string]interface{}{"key": "value"}) {
		t.Errorf("GetMap() = %v, %v, want map[key:value], true", val, ok)
	}

	// 测试不存在的键
	if _, ok := Default.GetInt(data, "nonexistent"); ok {
		t.Error("GetInt() succeeded for nonexistent key")
	}

	// 测试类型不匹配
	if _, ok := Default.GetInt(data, "string"); ok {
		t.Error("GetInt() succeeded for string value")
	}
}

func TestTypeConversion(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		wantInt  int64
		wantOk   bool
		wantBool bool
	}{
		{"int", 123, 123, true, true},
		{"int64", int64(123), 123, true, true},
		{"float64", 123.0, 123, true, true},
		{"string number", "123", 123, true, true},
		{"bool true", true, 1, true, true},
		{"bool false", false, 0, true, false},
		{"invalid type", "hello", 0, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]interface{}{"value": tt.value}

			// 测试整数转换
			if got, ok := Default.GetInt64(data, "value"); ok != tt.wantOk || (ok && got != tt.wantInt) {
				t.Errorf("GetInt64() = %v, %v, want %v, %v", got, ok, tt.wantInt, tt.wantOk)
			}

			// 测试布尔值转换 - 对于字符串数字，布尔值转换应该失败
			wantBoolOk := tt.wantOk
			if tt.name == "string number" {
				wantBoolOk = false // 字符串数字不能直接转换为布尔值
			}
			if got, ok := Default.GetBool(data, "value"); ok != wantBoolOk || (ok && got != tt.wantBool) {
				t.Errorf("GetBool() = %v, %v, want %v, %v", got, ok, tt.wantBool, wantBoolOk)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	// 测试有效的JSON
	validJSON := []byte(`{"name": "test", "age": 25}`)
	if err := Default.Validate(validJSON); err != nil {
		t.Errorf("Validate() error = %v", err)
	}

	// 测试无效的JSON
	invalidJSON := []byte(`{"name": "test", age: 25}`)
	if err := Default.Validate(invalidJSON); err == nil {
		t.Error("Validate() succeeded for invalid JSON")
	}

	// 测试必需字段验证
	data := map[string]interface{}{
		"name": "test",
		"age":  25,
	}

	if err := Default.ValidateRequired(data, "name", "age"); err != nil {
		t.Errorf("ValidateRequired() error = %v", err)
	}

	if err := Default.ValidateRequired(data, "name", "nonexistent"); err == nil {
		t.Error("ValidateRequired() succeeded for missing field")
	}

	// 测试类型验证
	if err := Default.ValidateType(data, "name", reflect.String); err != nil {
		t.Errorf("ValidateType() error = %v", err)
	}

	if err := Default.ValidateType(data, "age", reflect.Int); err != nil {
		t.Errorf("ValidateType() error = %v", err)
	}

	if err := Default.ValidateType(data, "name", reflect.Int); err == nil {
		t.Error("ValidateType() succeeded for wrong type")
	}
}

func TestJSONFormatting(t *testing.T) {
	// 测试数据
	data := map[string]interface{}{
		"name": "test",
		"age":  25,
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	// 测试压缩
	jsonBytes, _ := json.Marshal(data)
	compacted, err := Default.CompactJSON(jsonBytes)
	if err != nil {
		t.Errorf("CompactJSON() error = %v", err)
	}

	if bytes.Contains(compacted, []byte{' ', '\n', '\t'}) {
		t.Error("CompactJSON() result contains whitespace")
	}

	// 测试格式化
	formatted, err := Default.FormatJSON(jsonBytes, "", "  ")
	if err != nil {
		t.Errorf("FormatJSON() error = %v", err)
	}

	if !bytes.Contains(formatted, []byte{'\n'}) {
		t.Error("FormatJSON() result doesn't contain newlines")
	}
}

func TestJSONMerge(t *testing.T) {
	// 测试数据
	json1 := map[string]interface{}{
		"name": "test1",
		"nested": map[string]interface{}{
			"key1": "value1",
		},
	}

	json2 := map[string]interface{}{
		"age": 25,
		"nested": map[string]interface{}{
			"key2": "value2",
		},
	}

	// 测试合并
	merged := Default.MergeJSON(json1, json2)

	// 验证顶层字段
	if merged["name"] != "test1" || merged["age"] != 25 {
		t.Error("MergeJSON() failed to merge top-level fields")
	}

	// 验证嵌套字段
	if nested, ok := merged["nested"].(map[string]interface{}); ok {
		if nested["key1"] != "value1" || nested["key2"] != "value2" {
			t.Error("MergeJSON() failed to merge nested fields")
		}
	} else {
		t.Error("MergeJSON() failed to preserve nested structure")
	}

	// 测试空对象合并
	emptyMerged := Default.MergeJSON()
	if len(emptyMerged) != 0 {
		t.Error("MergeJSON() with no arguments should return empty map")
	}

	// 测试nil值合并
	nilMerged := Default.MergeJSON(nil, json1)
	if !reflect.DeepEqual(nilMerged, json1) {
		t.Error("MergeJSON() with nil should ignore nil values")
	}
}

func TestGetByPath(t *testing.T) {
	// 测试数据
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "test",
			"address": map[string]interface{}{
				"city": "Beijing",
			},
		},
		"scores": []interface{}{
			map[string]interface{}{
				"subject": "math",
				"score":   95,
			},
			map[string]interface{}{
				"subject": "english",
				"score":   85,
			},
		},
	}

	tests := []struct {
		name     string
		path     string
		want     interface{}
		wantBool bool
	}{
		{"simple path", "user.name", "test", true},
		{"nested path", "user.address.city", "Beijing", true},
		{"array index", "scores[0].subject", "math", true},
		{"array score", "scores[1].score", 85, true}, // 使用int类型，让测试更灵活
		{"invalid path", "user.invalid", nil, false},
		{"invalid index", "scores[2]", nil, false},
		{"invalid array path", "invalid[0]", nil, false},
		{"empty path", "", nil, false},
		{"root path", ".", nil, false},
		{"double dots", "user..name", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Default.GetByPath(data, tt.path)
			if ok != tt.wantBool {
				t.Errorf("GetByPath() ok = %v, want %v", ok, tt.wantBool)
				return
			}
			if ok {
				// 对于数字类型，进行更灵活的比较
				if tt.name == "array score" {
					// 检查是否为数字类型且值相等
					switch v := got.(type) {
					case int64:
						if int(v) != tt.want.(int) {
							t.Errorf("GetByPath() = %v, want %v", got, tt.want)
						}
					case int:
						if v != tt.want.(int) {
							t.Errorf("GetByPath() = %v, want %v", got, tt.want)
						}
					case float64:
						if int(v) != tt.want.(int) {
							t.Errorf("GetByPath() = %v, want %v", got, tt.want)
						}
					default:
						if !reflect.DeepEqual(got, tt.want) {
							t.Errorf("GetByPath() = %v, want %v", got, tt.want)
						}
					}
				} else {
					if !reflect.DeepEqual(got, tt.want) {
						t.Errorf("GetByPath() = %v, want %v", got, tt.want)
					}
				}
			}
		})
	}
}

type TestStruct struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Hobbies []string `json:"hobbies"`
}

func TestStructOperations(t *testing.T) {
	// 测试结构体序列化
	test := TestStruct{
		Name:    "test",
		Age:     25,
		Hobbies: []string{"reading", "coding"},
	}

	// 测试Marshal
	jsonBytes, err := Default.Marshal(test)
	if err != nil {
		t.Errorf("Marshal() error = %v", err)
	}

	// 测试Unmarshal
	var decoded TestStruct
	if err := Default.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Errorf("Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(test, decoded) {
		t.Error("Marshal/Unmarshal cycle failed to preserve data")
	}

	// 测试GetStruct
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name":    "test",
			"age":     25,
			"hobbies": []string{"reading", "coding"},
		},
	}

	var user TestStruct
	if ok := Default.GetStruct(data, "user", &user); !ok {
		t.Error("GetStruct() failed")
	}

	if !reflect.DeepEqual(test, user) {
		t.Error("GetStruct() failed to decode correctly")
	}

	// 测试无效的结构体指针
	if ok := Default.GetStruct(data, "user", test); ok {
		t.Error("GetStruct() succeeded with non-pointer value")
	}

	// 测试字段类型不匹配
	invalidData := map[string]interface{}{
		"user": map[string]interface{}{
			"name":    123,       // 应该是字符串
			"age":     "25",      // 应该是整数
			"hobbies": "reading", // 应该是切片
		},
	}

	var invalidUser TestStruct
	if ok := Default.GetStruct(invalidData, "user", &invalidUser); ok {
		t.Error("GetStruct() succeeded with invalid field types")
	}
}

func TestJSONKeys(t *testing.T) {
	jsonStr := `{
		"name": "test",
		"age": 25,
		"address": {
			"city": "Beijing",
			"street": "Main St"
		}
	}`

	keys, err := Default.GetJSONKeys(jsonStr)
	if err != nil {
		t.Errorf("GetJSONKeys() error = %v", err)
		return
	}

	expectedKeys := []string{"name", "age", "address"}
	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("GetJSONKeys() = %v, want %v", keys, expectedKeys)
	}

	// 测试无效的JSON
	_, err = Default.GetJSONKeys("{invalid json}")
	if err == nil {
		t.Error("GetJSONKeys() succeeded with invalid JSON")
	}

	// 测试非对象JSON
	_, err = Default.GetJSONKeys(`["array"]`)
	if err == nil {
		t.Error("GetJSONKeys() succeeded with non-object JSON")
	}
}

func TestToJSONString(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"age":  25,
	}

	jsonStr := Default.ToJSONString(data)
	expected := `{"age":25,"name":"test"}`
	if jsonStr != expected {
		t.Errorf("ToJSONString() = %v, want %v", jsonStr, expected)
	}

	// 测试无效数据
	invalidData := make(chan int)
	jsonStr = Default.ToJSONString(invalidData)
	if jsonStr != "" {
		t.Error("ToJSONString() should return empty string for invalid data")
	}
}
