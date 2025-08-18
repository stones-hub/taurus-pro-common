package tslice

import (
	"reflect"
	"strings"
	"testing"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		arr      []interface{}
		item     interface{}
		expected bool
	}{
		{
			name:     "整数切片包含元素",
			arr:      []interface{}{1, 2, 3, 4, 5},
			item:     3,
			expected: true,
		},
		{
			name:     "整数切片不包含元素",
			arr:      []interface{}{1, 2, 3, 4, 5},
			item:     6,
			expected: false,
		},
		{
			name:     "字符串切片包含元素",
			arr:      []interface{}{"apple", "banana", "orange"},
			item:     "banana",
			expected: true,
		},
		{
			name:     "空切片",
			arr:      []interface{}{},
			item:     1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.arr, tt.item)
			if result != tt.expected {
				t.Errorf("Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsFunc(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	tests := []struct {
		name     string
		arr      []Person
		item     Person
		equals   func(Person, Person) bool
		expected bool
	}{
		{
			name: "按名字比较",
			arr: []Person{
				{Name: "John", Age: 30},
				{Name: "Jane", Age: 25},
			},
			item: Person{Name: "John", Age: 35}, // 年龄不同但名字相同
			equals: func(a, b Person) bool {
				return a.Name == b.Name
			},
			expected: true,
		},
		{
			name: "按年龄比较",
			arr: []Person{
				{Name: "John", Age: 30},
				{Name: "Jane", Age: 25},
			},
			item: Person{Name: "Bob", Age: 30}, // 名字不同但年龄相同
			equals: func(a, b Person) bool {
				return a.Age == b.Age
			},
			expected: true,
		},
		{
			name: "完全匹配",
			arr: []Person{
				{Name: "John", Age: 30},
				{Name: "Jane", Age: 25},
			},
			item: Person{Name: "John", Age: 30},
			equals: func(a, b Person) bool {
				return a.Name == b.Name && a.Age == b.Age
			},
			expected: true,
		},
		{
			name: "不匹配",
			arr: []Person{
				{Name: "John", Age: 30},
				{Name: "Jane", Age: 25},
			},
			item: Person{Name: "Bob", Age: 35},
			equals: func(a, b Person) bool {
				return a.Name == b.Name && a.Age == b.Age
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsFunc(tt.arr, tt.item, tt.equals)
			if result != tt.expected {
				t.Errorf("ContainsFunc() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	tests := []struct {
		name     string
		arr      []interface{}
		item     interface{}
		expected int
	}{
		{
			name:     "找到元素",
			arr:      []interface{}{1, 2, 3, 4, 5},
			item:     3,
			expected: 2,
		},
		{
			name:     "未找到元素",
			arr:      []interface{}{1, 2, 3, 4, 5},
			item:     6,
			expected: -1,
		},
		{
			name:     "空切片",
			arr:      []interface{}{},
			item:     1,
			expected: -1,
		},
		{
			name:     "字符串切片",
			arr:      []interface{}{"apple", "banana", "orange"},
			item:     "banana",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IndexOf(tt.arr, tt.item)
			if result != tt.expected {
				t.Errorf("IndexOf() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name     string
		arr      []interface{}
		item     interface{}
		expected []interface{}
	}{
		{
			name:     "移除单个元素",
			arr:      []interface{}{1, 2, 3, 4, 5},
			item:     3,
			expected: []interface{}{1, 2, 4, 5},
		},
		{
			name:     "移除多个相同元素",
			arr:      []interface{}{1, 2, 2, 3, 2, 4},
			item:     2,
			expected: []interface{}{1, 3, 4},
		},
		{
			name:     "移除不存在的元素",
			arr:      []interface{}{1, 2, 3},
			item:     4,
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "空切片",
			arr:      []interface{}{},
			item:     1,
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Remove(tt.arr, tt.item)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Remove() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name     string
		arr      []interface{}
		expected []interface{}
	}{
		{
			name:     "整数切片去重",
			arr:      []interface{}{1, 2, 2, 3, 3, 4, 5, 5},
			expected: []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:     "字符串切片去重",
			arr:      []interface{}{"apple", "banana", "apple", "orange", "banana"},
			expected: []interface{}{"apple", "banana", "orange"},
		},
		{
			name:     "空切片",
			arr:      []interface{}{},
			expected: []interface{}{},
		},
		{
			name:     "无重复元素",
			arr:      []interface{}{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Unique(tt.arr)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Unique() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name      string
		arr       []interface{}
		predicate func(interface{}) bool
		expected  []interface{}
	}{
		{
			name: "过滤偶数",
			arr:  []interface{}{1, 2, 3, 4, 5, 6},
			predicate: func(v interface{}) bool {
				return v.(int)%2 == 0
			},
			expected: []interface{}{2, 4, 6},
		},
		{
			name: "过滤字符串长度大于4",
			arr:  []interface{}{"apple", "banana", "kiwi", "orange"},
			predicate: func(v interface{}) bool {
				return len(v.(string)) > 4
			},
			expected: []interface{}{"apple", "banana", "orange"},
		},
		{
			name: "空切片",
			arr:  []interface{}{},
			predicate: func(v interface{}) bool {
				return true
			},
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(tt.arr, tt.predicate)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Filter() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMap(t *testing.T) {
	tests := []struct {
		name      string
		arr       []interface{}
		transform func(interface{}) interface{}
		expected  []interface{}
	}{
		{
			name: "整数平方",
			arr:  []interface{}{1, 2, 3, 4},
			transform: func(v interface{}) interface{} {
				return v.(int) * v.(int)
			},
			expected: []interface{}{1, 4, 9, 16},
		},
		{
			name: "字符串转大写",
			arr:  []interface{}{"apple", "banana", "orange"},
			transform: func(v interface{}) interface{} {
				return strings.ToUpper(v.(string))
			},
			expected: []interface{}{"APPLE", "BANANA", "ORANGE"},
		},
		{
			name: "空切片",
			arr:  []interface{}{},
			transform: func(v interface{}) interface{} {
				return v
			},
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.arr, tt.transform)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Map() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReduce(t *testing.T) {
	tests := []struct {
		name     string
		arr      []interface{}
		initial  interface{}
		reducer  func(interface{}, interface{}) interface{}
		expected interface{}
	}{
		{
			name:    "求和",
			arr:     []interface{}{1, 2, 3, 4, 5},
			initial: 0,
			reducer: func(acc, curr interface{}) interface{} {
				return acc.(int) + curr.(int)
			},
			expected: 15,
		},
		{
			name:    "字符串连接",
			arr:     []interface{}{"Hello", "World", "!"},
			initial: "",
			reducer: func(acc, curr interface{}) interface{} {
				if acc.(string) == "" {
					return curr.(string)
				}
				return acc.(string) + " " + curr.(string)
			},
			expected: "Hello World !",
		},
		{
			name:    "空切片",
			arr:     []interface{}{},
			initial: 0,
			reducer: func(acc, curr interface{}) interface{} {
				return acc.(int) + curr.(int)
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reduce(tt.arr, tt.initial, tt.reducer)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Reduce() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChunk(t *testing.T) {
	tests := []struct {
		name     string
		arr      []interface{}
		size     int
		expected [][]interface{}
	}{
		{
			name: "正常分块",
			arr:  []interface{}{1, 2, 3, 4, 5, 6, 7},
			size: 3,
			expected: [][]interface{}{
				{1, 2, 3},
				{4, 5, 6},
				{7},
			},
		},
		{
			name: "大小为1",
			arr:  []interface{}{1, 2, 3},
			size: 1,
			expected: [][]interface{}{
				{1}, {2}, {3},
			},
		},
		{
			name:     "空切片",
			arr:      []interface{}{},
			size:     2,
			expected: [][]interface{}{},
		},
		{
			name: "大小为0",
			arr:  []interface{}{1, 2, 3},
			size: 0,
			expected: [][]interface{}{
				{1, 2, 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Chunk(tt.arr, tt.size)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Chunk() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		arr      []interface{}
		expected []interface{}
	}{
		{
			name:     "整数切片",
			arr:      []interface{}{1, 2, 3, 4, 5},
			expected: []interface{}{5, 4, 3, 2, 1},
		},
		{
			name:     "字符串切片",
			arr:      []interface{}{"apple", "banana", "orange"},
			expected: []interface{}{"orange", "banana", "apple"},
		},
		{
			name:     "空切片",
			arr:      []interface{}{},
			expected: []interface{}{},
		},
		{
			name:     "单个元素",
			arr:      []interface{}{1},
			expected: []interface{}{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reverse(tt.arr)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Reverse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		name     string
		arr1     []interface{}
		arr2     []interface{}
		expected []interface{}
	}{
		{
			name:     "有交集",
			arr1:     []interface{}{1, 2, 3, 4, 5},
			arr2:     []interface{}{4, 5, 6, 7, 8},
			expected: []interface{}{4, 5},
		},
		{
			name:     "无交集",
			arr1:     []interface{}{1, 2, 3},
			arr2:     []interface{}{4, 5, 6},
			expected: []interface{}{},
		},
		{
			name:     "空切片",
			arr1:     []interface{}{},
			arr2:     []interface{}{1, 2, 3},
			expected: []interface{}{},
		},
		{
			name:     "完全相同",
			arr1:     []interface{}{1, 2, 3},
			arr2:     []interface{}{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Intersection(tt.arr1, tt.arr2)
			// 对于空切片，使用长度比较而不是 DeepEqual
			if len(result) != len(tt.expected) {
				t.Errorf("Intersection() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			if len(result) > 0 && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Intersection() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUnion(t *testing.T) {
	tests := []struct {
		name     string
		arr1     []interface{}
		arr2     []interface{}
		expected []interface{}
	}{
		{
			name:     "有重复元素",
			arr1:     []interface{}{1, 2, 3},
			arr2:     []interface{}{3, 4, 5},
			expected: []interface{}{1, 2, 3, 4, 5},
		},
		{
			name:     "无重复元素",
			arr1:     []interface{}{1, 2, 3},
			arr2:     []interface{}{4, 5, 6},
			expected: []interface{}{1, 2, 3, 4, 5, 6},
		},
		{
			name:     "空切片",
			arr1:     []interface{}{},
			arr2:     []interface{}{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "完全相同",
			arr1:     []interface{}{1, 2, 3},
			arr2:     []interface{}{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Union(tt.arr1, tt.arr2)
			// Union 函数使用 map 实现，返回的元素顺序不保证
			// 所以我们需要检查元素是否包含所有期望的元素，而不是严格的顺序比较
			if len(result) != len(tt.expected) {
				t.Errorf("Union() length = %d, want %d", len(result), len(tt.expected))
				return
			}

			// 检查结果是否包含所有期望的元素
			for _, expected := range tt.expected {
				found := false
				for _, actual := range result {
					if reflect.DeepEqual(actual, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Union() result %v missing expected element %v", result, expected)
					return
				}
			}
		})
	}
}

func TestDifference(t *testing.T) {
	tests := []struct {
		name     string
		arr1     []interface{}
		arr2     []interface{}
		expected []interface{}
	}{
		{
			name:     "有差集",
			arr1:     []interface{}{1, 2, 3, 4, 5},
			arr2:     []interface{}{4, 5, 6, 7},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "无差集",
			arr1:     []interface{}{1, 2, 3},
			arr2:     []interface{}{1, 2, 3},
			expected: []interface{}{},
		},
		{
			name:     "完全不同",
			arr1:     []interface{}{1, 2, 3},
			arr2:     []interface{}{4, 5, 6},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "空切片",
			arr1:     []interface{}{},
			arr2:     []interface{}{1, 2, 3},
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Difference(tt.arr1, tt.arr2)
			// 对于空切片，使用长度比较而不是 DeepEqual
			if len(result) != len(tt.expected) {
				t.Errorf("Difference() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			if len(result) > 0 && !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Difference() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestArrayToString(t *testing.T) {
	tests := []struct {
		name     string
		array    []interface{}
		expected string
	}{
		{
			name:     "基本类型",
			array:    []interface{}{1, "hello", true},
			expected: "1,hello,true",
		},
		{
			name:     "空切片",
			array:    []interface{}{},
			expected: "",
		},
		{
			name:     "混合类型",
			array:    []interface{}{42, "world", 3.14},
			expected: "42,world,3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrayToString(tt.array)
			if result != tt.expected {
				t.Errorf("ArrayToString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseStringSliceToUint64(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []uint64
	}{
		{
			name:     "有效数字",
			input:    []string{"123", "456", "789"},
			expected: []uint64{123, 456, 789},
		},
		{
			name:     "无效输入",
			input:    []string{"123", "abc", "456", "-789"},
			expected: []uint64{123, 0, 456, 0},
		},
		{
			name:     "空切片",
			input:    []string{},
			expected: []uint64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseStringSliceToUint64(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseStringSliceToUint64() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// 基准测试
func BenchmarkContains(b *testing.B) {
	arr := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	item := interface{}(5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Contains(arr, item)
	}
}

func BenchmarkIndexOf(b *testing.B) {
	arr := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	item := interface{}(5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IndexOf(arr, item)
	}
}

func BenchmarkRemove(b *testing.B) {
	arr := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	item := interface{}(5)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Remove(arr, item)
	}
}

func BenchmarkUnique(b *testing.B) {
	arr := []interface{}{1, 2, 2, 3, 3, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Unique(arr)
	}
}

func BenchmarkFilter(b *testing.B) {
	arr := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	predicate := func(v interface{}) bool {
		return v.(int)%2 == 0
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Filter(arr, predicate)
	}
}

func BenchmarkMap(b *testing.B) {
	arr := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	transform := func(v interface{}) interface{} {
		return v.(int) * v.(int)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Map(arr, transform)
	}
}

func BenchmarkReduce(b *testing.B) {
	arr := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	initial := interface{}(0)
	reducer := func(acc, curr interface{}) interface{} {
		return acc.(int) + curr.(int)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reduce(arr, initial, reducer)
	}
}

func BenchmarkChunk(b *testing.B) {
	arr := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	size := 3

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Chunk(arr, size)
	}
}

func BenchmarkReverse(b *testing.B) {
	arr := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reverse(arr)
	}
}

func BenchmarkIntersection(b *testing.B) {
	arr1 := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	arr2 := []interface{}{5, 6, 7, 8, 9, 10, 11, 12, 13, 14}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Intersection(arr1, arr2)
	}
}

func BenchmarkUnion(b *testing.B) {
	arr1 := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	arr2 := []interface{}{5, 6, 7, 8, 9, 10, 11, 12, 13, 14}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Union(arr1, arr2)
	}
}

func BenchmarkDifference(b *testing.B) {
	arr1 := []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	arr2 := []interface{}{5, 6, 7, 8, 9, 10, 11, 12, 13, 14}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Difference(arr1, arr2)
	}
}
