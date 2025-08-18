package ttable

import (
	"testing"
)

func TestPrintTable(t *testing.T) {
	tests := []struct {
		name string
		data TableData
	}{
		{
			name: "基本表格",
			data: TableData{
				{"Name", "Age", "City"},
				{"John", 30, "New York"},
				{"Jane", 25, "San Francisco"},
			},
		},
		{
			name: "空表格",
			data: TableData{},
		},
		{
			name: "单行表格",
			data: TableData{
				{"Header1", "Header2"},
			},
		},
		{
			name: "混合类型数据",
			data: TableData{
				{"Name", "Score", "Pass"},
				{"Alice", 95.5, true},
				{"Bob", 85.0, true},
				{"Charlie", 45.5, false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 由于PrintTable直接输出到标准输出，这里只测试它不会panic
			PrintTable(tt.data)
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		val      interface{}
		expected string
	}{
		{
			name:     "字符串",
			val:      "hello",
			expected: "hello",
		},
		{
			name:     "整数",
			val:      42,
			expected: "42",
		},
		{
			name:     "切片",
			val:      []interface{}{1, 2, 3},
			expected: "[1 2 3]",
		},
		{
			name:     "映射",
			val:      map[string]interface{}{"key": "value"},
			expected: "map[key:value]",
		},
		{
			name:     "布尔值",
			val:      true,
			expected: "true",
		},
		{
			name:     "nil值",
			val:      nil,
			expected: "<nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.val)
			if result != tt.expected {
				t.Errorf("formatValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSum(t *testing.T) {
	tests := []struct {
		name     string
		nums     []int
		expected int
	}{
		{
			name:     "正数",
			nums:     []int{1, 2, 3, 4, 5},
			expected: 15,
		},
		{
			name:     "包含负数",
			nums:     []int{-1, 0, 1},
			expected: 0,
		},
		{
			name:     "空切片",
			nums:     []int{},
			expected: 0,
		},
		{
			name:     "单个数字",
			nums:     []int{42},
			expected: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sum(tt.nums)
			if result != tt.expected {
				t.Errorf("sum() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRenderTable(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	data := [][]interface{}{
		{"John", 30, "New York"},
		{"Jane", 25, "San Francisco"},
	}

	tests := []struct {
		name   string
		format string
	}{
		{
			name:   "终端格式",
			format: "terminal",
		},
		{
			name:   "HTML格式",
			format: "html",
		},
		{
			name:   "Markdown格式",
			format: "markdown",
		},
		{
			name:   "无效格式",
			format: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 由于RenderTable直接输出到标准输出，这里只测试它不会panic
			RenderTable(headers, data, tt.format)
		})
	}
}

func TestRenderTableMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		lines    [][]interface{}
		expected string
	}{
		{
			name:    "基本表格",
			headers: []string{"Name", "Age", "City"},
			lines: [][]interface{}{
				{"John", 30, "New York"},
				{"Jane", 25, "San Francisco"},
			},
			expected: "| Name | Age | City |\n|" +
				" --- | --- | --- |\n" +
				"| John | 30 | New York |\n" +
				"| Jane | 25 | San Francisco |\n",
		},
		{
			name:    "空表格",
			headers: []string{},
			lines:   [][]interface{}{},
			expected: "| |\n" +
				"| --- |\n",
		},
		{
			name:    "只有表头",
			headers: []string{"Header1", "Header2"},
			lines:   [][]interface{}{},
			expected: "| Header1 | Header2 |\n" +
				"| --- | --- |\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderTableMarkdown(tt.headers, tt.lines)
			if result != tt.expected {
				t.Errorf("RenderTableMarkdown() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRenderTableHTML(t *testing.T) {
	tests := []struct {
		name     string
		headers  []string
		lines    [][]interface{}
		expected string
	}{
		{
			name:    "基本表格",
			headers: []string{"Name", "Age", "City"},
			lines: [][]interface{}{
				{"John", 30, "New York"},
				{"Jane", 25, "San Francisco"},
			},
			expected: "<table border=\"1\" cellspacing=\"0\" cellpadding=\"4\">\n" +
				"  <tr><th>Name</th><th>Age</th><th>City</th></tr>\n" +
				"  <tr><td>John</td><td>30</td><td>New York</td></tr>\n" +
				"  <tr><td>Jane</td><td>25</td><td>San Francisco</td></tr>\n" +
				"</table>\n",
		},
		{
			name:    "空表格",
			headers: []string{},
			lines:   [][]interface{}{},
			expected: "<table border=\"1\" cellspacing=\"0\" cellpadding=\"4\">\n" +
				"  <tr></tr>\n" +
				"</table>\n",
		},
		{
			name:    "只有表头",
			headers: []string{"Header1", "Header2"},
			lines:   [][]interface{}{},
			expected: "<table border=\"1\" cellspacing=\"0\" cellpadding=\"4\">\n" +
				"  <tr><th>Header1</th><th>Header2</th></tr>\n" +
				"</table>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderTableHTML(tt.headers, tt.lines)
			if result != tt.expected {
				t.Errorf("RenderTableHTML() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// 基准测试
func BenchmarkPrintTable(b *testing.B) {
	data := TableData{
		{"Name", "Age", "City", "Email", "Phone"},
		{"John", 30, "New York", "john@example.com", "123-456-7890"},
		{"Jane", 25, "San Francisco", "jane@example.com", "098-765-4321"},
		{"Bob", 35, "Chicago", "bob@example.com", "555-123-4567"},
		{"Alice", 28, "Boston", "alice@example.com", "777-888-9999"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		PrintTable(data)
	}
}

func BenchmarkRenderTableMarkdown(b *testing.B) {
	headers := []string{"Name", "Age", "City", "Email", "Phone"}
	lines := [][]interface{}{
		{"John", 30, "New York", "john@example.com", "123-456-7890"},
		{"Jane", 25, "San Francisco", "jane@example.com", "098-765-4321"},
		{"Bob", 35, "Chicago", "bob@example.com", "555-123-4567"},
		{"Alice", 28, "Boston", "alice@example.com", "777-888-9999"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderTableMarkdown(headers, lines)
	}
}

func BenchmarkRenderTableHTML(b *testing.B) {
	headers := []string{"Name", "Age", "City", "Email", "Phone"}
	lines := [][]interface{}{
		{"John", 30, "New York", "john@example.com", "123-456-7890"},
		{"Jane", 25, "San Francisco", "jane@example.com", "098-765-4321"},
		{"Bob", 35, "Chicago", "bob@example.com", "555-123-4567"},
		{"Alice", 28, "Boston", "alice@example.com", "777-888-9999"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderTableHTML(headers, lines)
	}
}

func BenchmarkFormatValue(b *testing.B) {
	testValues := []interface{}{
		"hello world",
		42,
		[]interface{}{1, 2, 3},
		map[string]interface{}{"key": "value"},
		true,
		nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, val := range testValues {
			formatValue(val)
		}
	}
}

func BenchmarkSum(b *testing.B) {
	testNums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum(testNums)
	}
}
