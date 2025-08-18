package tstruct

import (
	"reflect"
	"testing"
)

// 测试用的结构体
type TestUser struct {
	Name     string `json:"name" mapstructure:"name"`
	Age      int    `json:"age" mapstructure:"age"`
	Email    string `json:"email,omitempty" mapstructure:"email"`
	Password string `json:"-" mapstructure:"-"`
}

type TestAddress struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

type TestPerson struct {
	Name    string      `json:"name"`
	Address TestAddress `json:"address"`
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name     string
		src      interface{}
		dst      interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name: "相同类型结构体复制",
			src: TestUser{
				Name:     "John",
				Age:      30,
				Email:    "john@example.com",
				Password: "secret",
			},
			dst: &TestUser{},
			expected: &TestUser{
				Name:     "John",
				Age:      30,
				Email:    "john@example.com",
				Password: "secret",
			},
			wantErr: false,
		},
		{
			name: "嵌套结构体复制",
			src: TestPerson{
				Name: "John",
				Address: TestAddress{
					City:    "New York",
					Country: "USA",
				},
			},
			dst: &TestPerson{},
			expected: &TestPerson{
				Name: "John",
				Address: TestAddress{
					City:    "New York",
					Country: "USA",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Copy(tt.src, tt.dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("Copy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.dst, tt.expected) {
				t.Errorf("Copy() = %v, want %v", tt.dst, tt.expected)
			}
		})
	}
}

func TestMapToStruct(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		target   interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name: "基本类型转换",
			data: map[string]interface{}{
				"name":  "John",
				"age":   30,
				"email": "john@example.com",
			},
			target: &TestUser{},
			expected: &TestUser{
				Name:  "John",
				Age:   30,
				Email: "john@example.com",
			},
			wantErr: false,
		},
		{
			name: "嵌套结构体转换",
			data: map[string]interface{}{
				"name": "John",
				"address": map[string]interface{}{
					"city":    "New York",
					"country": "USA",
				},
			},
			target: &TestPerson{},
			expected: &TestPerson{
				Name: "John",
				Address: TestAddress{
					City:    "New York",
					Country: "USA",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MapToStruct(tt.data, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapToStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.target, tt.expected) {
				t.Errorf("MapToStruct() = %v, want %v", tt.target, tt.expected)
			}
		})
	}
}

func TestMapToStructWithValidation(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]interface{}
		target         interface{}
		requiredFields []string
		wantErr        bool
	}{
		{
			name: "所有必需字段都存在",
			data: map[string]interface{}{
				"name":  "John",
				"age":   30,
				"email": "john@example.com",
			},
			target:         &TestUser{},
			requiredFields: []string{"name", "email"},
			wantErr:        false,
		},
		{
			name: "缺少必需字段",
			data: map[string]interface{}{
				"name": "John",
			},
			target:         &TestUser{},
			requiredFields: []string{"name", "email"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MapToStructWithValidation(tt.data, tt.target, tt.requiredFields...)
			if (err != nil) != tt.wantErr {
				t.Errorf("MapToStructWithValidation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractString(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]interface{}
		key     string
		want    string
		wantErr bool
	}{
		{
			name: "字符串字段",
			data: map[string]interface{}{
				"name": "John",
			},
			key:     "name",
			want:    "John",
			wantErr: false,
		},
		{
			name: "非字符串字段",
			data: map[string]interface{}{
				"age": 30,
			},
			key:     "age",
			want:    "",
			wantErr: true,
		},
		{
			name:    "字段不存在",
			data:    map[string]interface{}{},
			key:     "name",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractString(tt.data, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateStruct(t *testing.T) {
	type TestStruct struct {
		Name     string
		Age      int
		Email    string
		Optional string
	}

	tests := []struct {
		name           string
		target         interface{}
		requiredFields []string
		wantErr        bool
	}{
		{
			name: "所有必需字段都有值",
			target: TestStruct{
				Name:  "John",
				Email: "john@example.com",
			},
			requiredFields: []string{"Name", "Email"},
			wantErr:        false,
		},
		{
			name: "缺少必需字段值",
			target: TestStruct{
				Name: "John",
			},
			requiredFields: []string{"Name", "Email"},
			wantErr:        true,
		},
		{
			name:           "结构体为nil",
			target:         nil,
			requiredFields: []string{"Name"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStruct(tt.target, tt.requiredFields...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanPointer(t *testing.T) {
	tests := []struct {
		name    string
		dest    interface{}
		src     interface{}
		wantErr bool
	}{
		{
			name:    "整数赋值",
			dest:    new(int),
			src:     42,
			wantErr: false,
		},
		{
			name:    "字符串赋值",
			dest:    new(string),
			src:     "hello",
			wantErr: false,
		},
		{
			name:    "类型不匹配",
			dest:    new(int),
			src:     "not a number",
			wantErr: true,
		},
		{
			name:    "目标为nil",
			dest:    nil,
			src:     42,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ScanPointer(tt.dest, tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("ScanPointer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				switch d := tt.dest.(type) {
				case *int:
					if *d != tt.src.(int) {
						t.Errorf("ScanPointer() = %v, want %v", *d, tt.src)
					}
				case *string:
					if *d != tt.src.(string) {
						t.Errorf("ScanPointer() = %v, want %v", *d, tt.src)
					}
				}
			}
		})
	}
}

func TestTrimSpace(t *testing.T) {
	type TestStruct struct {
		Name  string
		Email string
		Age   int
	}

	tests := []struct {
		name     string
		input    *TestStruct
		expected *TestStruct
	}{
		{
			name: "去除空白字符",
			input: &TestStruct{
				Name:  "  John  ",
				Email: " john@example.com ",
				Age:   30,
			},
			expected: &TestStruct{
				Name:  "John",
				Email: "john@example.com",
				Age:   30,
			},
		},
		{
			name: "无空白字符",
			input: &TestStruct{
				Name:  "John",
				Email: "john@example.com",
				Age:   30,
			},
			expected: &TestStruct{
				Name:  "John",
				Email: "john@example.com",
				Age:   30,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TrimSpace(tt.input)
			if !reflect.DeepEqual(tt.input, tt.expected) {
				t.Errorf("TrimSpace() = %v, want %v", tt.input, tt.expected)
			}
		})
	}
}

func TestToStringWithFlag(t *testing.T) {
	tests := []struct {
		name     string
		v        interface{}
		flag     string
		expected string
		wantErr  bool
	}{
		{
			name: "结构体转换",
			v: struct {
				Name string
				Age  int
			}{
				Name: "John",
				Age:  30,
			},
			flag:     ", ",
			expected: "John, 30",
			wantErr:  false,
		},
		{
			name: "Map转换",
			v: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			flag:     "; ",
			expected: "age:30; name:John",
			wantErr:  false,
		},
		{
			name:     "切片转换",
			v:        []string{"apple", "banana", "orange"},
			flag:     " | ",
			expected: "apple | banana | orange",
			wantErr:  false,
		},
		{
			name:     "不支持的类型",
			v:        42,
			flag:     ", ",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToStringWithFlag(tt.v, tt.flag)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToStringWithFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ToStringWithFlag() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCheckType(t *testing.T) {
	tests := []struct {
		name     string
		v        interface{}
		expected string
	}{
		{
			name:     "整数",
			v:        42,
			expected: "int",
		},
		{
			name:     "字符串",
			v:        "hello",
			expected: "string",
		},
		{
			name:     "布尔值",
			v:        true,
			expected: "bool",
		},
		{
			name:     "切片",
			v:        []int{},
			expected: "slice",
		},
		{
			name:     "映射",
			v:        map[string]int{},
			expected: "map",
		},
		{
			name:     "结构体",
			v:        struct{}{},
			expected: "struct",
		},
		{
			name:     "指针",
			v:        new(int),
			expected: "pointer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckType(tt.v); got != tt.expected {
				t.Errorf("CheckType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConvertMapKeysToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "下划线转驼峰",
			input: map[string]interface{}{
				"user_name":     "John",
				"email_address": "john@example.com",
			},
			expected: map[string]interface{}{
				"userName":     "John",
				"emailAddress": "john@example.com",
			},
		},
		{
			name: "已经是驼峰格式",
			input: map[string]interface{}{
				"userName": "John",
				"age":      30,
			},
			expected: map[string]interface{}{
				"userName": "John",
				"age":      30,
			},
		},
		{
			name:     "空映射",
			input:    map[string]interface{}{},
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertMapKeysToCamelCase(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ConvertMapKeysToCamelCase() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStructToMap(t *testing.T) {
	type TestStruct struct {
		Name     string `json:"name"`
		Age      int    `json:"age"`
		Email    string `json:"email,omitempty"`
		Password string `json:"-"`
	}

	tests := []struct {
		name     string
		obj      interface{}
		tagName  string
		expected map[string]interface{}
	}{
		{
			name: "使用json标签",
			obj: TestStruct{
				Name:     "John",
				Age:      30,
				Email:    "john@example.com",
				Password: "secret",
			},
			tagName: "json",
			expected: map[string]interface{}{
				"name":  "John",
				"age":   30,
				"email": "john@example.com",
			},
		},
		{
			name: "不使用标签",
			obj: TestStruct{
				Name:     "John",
				Age:      30,
				Email:    "john@example.com",
				Password: "secret",
			},
			tagName: "",
			expected: map[string]interface{}{
				"Name":     "John",
				"Age":      30,
				"Email":    "john@example.com",
				"Password": "secret",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StructToMap(tt.obj, tt.tagName)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("StructToMap() = %v, want %v", got, tt.expected)
			}
		})
	}
}
