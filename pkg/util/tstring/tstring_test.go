package tstring

import (
	"strings"
	"testing"
)

// TestBasicStringOperations 测试基本的字符串操作，避免复杂的依赖
func TestBasicStringOperations(t *testing.T) {
	// 测试基本的字符串反转
	t.Run("BasicReverse", func(t *testing.T) {
		result := Reverse("Hello")
		if result != "olleH" {
			t.Errorf("Reverse('Hello') = %v, want 'olleH'", result)
		}
	})

	// 测试空字符串
	t.Run("EmptyString", func(t *testing.T) {
		result := Reverse("")
		if result != "" {
			t.Errorf("Reverse('') = %v, want ''", result)
		}
	})

	// 测试单个字符
	t.Run("SingleChar", func(t *testing.T) {
		result := Reverse("a")
		if result != "a" {
			t.Errorf("Reverse('a') = %v, want 'a'", result)
		}
	})
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "单个字符",
			input:    "a",
			expected: "a",
		},
		{
			name:     "英文字符串",
			input:    "Hello",
			expected: "olleH",
		},
		{
			name:     "带空格的字符串",
			input:    "Hello World",
			expected: "dlroW olleH",
		},
		{
			name:     "数字字符串",
			input:    "12345",
			expected: "54321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reverse(tt.input)
			if result != tt.expected {
				t.Errorf("Reverse() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRandomString(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		charset     string
		checkLength bool
		checkChars  bool
	}{
		{
			name:        "默认字符集",
			length:      10,
			charset:     "",
			checkLength: true,
			checkChars:  true,
		},
		{
			name:        "自定义字符集",
			length:      8,
			charset:     "ABC123",
			checkLength: true,
			checkChars:  true,
		},
		{
			name:        "长度为0",
			length:      0,
			charset:     "",
			checkLength: true,
			checkChars:  false,
		},
		{
			name:        "长度为1",
			length:      1,
			charset:     "",
			checkLength: true,
			checkChars:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.charset == "" {
				result = RandomString(tt.length)
			} else {
				result = RandomString(tt.length, tt.charset)
			}

			// 检查长度
			if tt.checkLength && len(result) != tt.length {
				t.Errorf("RandomString() length = %v, want %v", len(result), tt.length)
			}

			// 检查字符是否在字符集中
			if tt.checkChars {
				charset := tt.charset
				if charset == "" {
					charset = defaultCharset
				}
				for _, c := range result {
					if !strings.ContainsRune(charset, c) {
						t.Errorf("RandomString() contains invalid character: %c", c)
					}
				}
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "多个空格",
			input:    "Hello   World",
			expected: "Hello World",
		},
		{
			name:     "首尾空格",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
		{
			name:     "换行符",
			input:    "Hello\nWorld",
			expected: "Hello World",
		},
		{
			name:     "制表符",
			input:    "Hello\tWorld",
			expected: "Hello World",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "只有空白字符",
			input:    "   \n\t   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input)
			if result != tt.expected {
				t.Errorf("Normalize() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestChunkSplit(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		chunkSize int
		delimiter string
		expected  string
	}{
		{
			name:      "正常分割",
			input:     "HelloWorld",
			chunkSize: 5,
			delimiter: "-",
			expected:  "Hello-World-",
		},
		{
			name:      "不等长分割",
			input:     "Hello",
			chunkSize: 2,
			delimiter: "/",
			expected:  "He/ll/o/",
		},
		{
			name:      "空字符串",
			input:     "",
			chunkSize: 2,
			delimiter: "-",
			expected:  "",
		},
		{
			name:      "chunkSize为0",
			input:     "Hello",
			chunkSize: 0,
			delimiter: "-",
			expected:  "Hello",
		},
		{
			name:      "chunkSize大于字符串长度",
			input:     "Hello",
			chunkSize: 10,
			delimiter: "-",
			expected:  "Hello-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ChunkSplit(tt.input, tt.chunkSize, tt.delimiter)
			if result != tt.expected {
				t.Errorf("ChunkSplit() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		delimiter string
		expected  string
	}{
		{
			name:      "下划线分隔",
			input:     "hello_world_test",
			delimiter: "_",
			expected:  "helloWorldTest",
		},
		{
			name:      "空格分隔",
			input:     "hello world test",
			delimiter: " ",
			expected:  "helloWorldTest",
		},
		{
			name:      "连字符分隔",
			input:     "hello-world-test",
			delimiter: "-",
			expected:  "helloWorldTest",
		},
		{
			name:      "空字符串",
			input:     "",
			delimiter: "_",
			expected:  "",
		},
		{
			name:      "单个单词",
			input:     "hello",
			delimiter: "_",
			expected:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToCamelCase(tt.input, tt.delimiter)
			if result != tt.expected {
				t.Errorf("ToCamelCase() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFirstUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "小写字母开头",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "大写字母开头",
			input:    "Hello",
			expected: "Hello",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "单个字符",
			input:    "h",
			expected: "H",
		},
		{
			name:     "数字开头",
			input:    "123hello",
			expected: "123hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstUpper(tt.input)
			if result != tt.expected {
				t.Errorf("FirstUpper() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFirstLower(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "大写字母开头",
			input:    "Hello",
			expected: "hello",
		},
		{
			name:     "小写字母开头",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "单个字符",
			input:    "H",
			expected: "h",
		},
		{
			name:     "数字开头",
			input:    "123Hello",
			expected: "123Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FirstLower(tt.input)
			if result != tt.expected {
				t.Errorf("FirstLower() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterInvisibleChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "包含控制字符",
			input:    "Hello\x00World\x1F",
			expected: "HelloWorld",
		},
		{
			name:     "包含空格和制表符",
			input:    "Hello\t \nWorld",
			expected: "HelloWorld",
		},
		{
			name:     "正常字符串",
			input:    "Hello World",
			expected: "HelloWorld",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "只有不可见字符",
			input:    "\x00\x01\x02",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterInvisibleChars(tt.input)
			if result != tt.expected {
				t.Errorf("FilterInvisibleChars() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReplacePlaceholders(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		replacements []string
		expected     string
	}{
		{
			name:         "SQL查询占位符",
			query:        "SELECT * FROM users WHERE age > ? AND status = ?",
			replacements: []string{"18", "'active'"},
			expected:     "SELECT * FROM users WHERE age > 18 AND status = 'active'",
		},
		{
			name:         "消息模板",
			query:        "Hello ?, your code is ?",
			replacements: []string{"John", "123456"},
			expected:     "Hello John, your code is 123456",
		},
		{
			name:         "替换值不足",
			query:        "Hello ? ?",
			replacements: []string{"World"},
			expected:     "Hello World ?",
		},
		{
			name:         "替换值过多",
			query:        "Hello ?",
			replacements: []string{"World", "Extra"},
			expected:     "Hello World",
		},
		{
			name:         "无占位符",
			query:        "Hello World",
			replacements: []string{"Test"},
			expected:     "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplacePlaceholders(tt.query, tt.replacements)
			if result != tt.expected {
				t.Errorf("ReplacePlaceholders() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFilterNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "包含数字",
			input:    "abc123def456",
			expected: "abcdef",
		},
		{
			name:     "只有数字",
			input:    "12345",
			expected: "",
		},
		{
			name:     "没有数字",
			input:    "abcdef",
			expected: "abcdef",
		},
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "混合字符",
			input:    "test123test456test",
			expected: "testtesttest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterNumber(tt.input)
			if result != tt.expected {
				t.Errorf("FilterNumber() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		length   int
		omission []string
		expected string
	}{
		{
			name:     "使用默认省略号",
			s:        "Hello, World!",
			length:   8,
			expected: "Hello...",
		},
		{
			name:     "使用自定义省略标记",
			s:        "Hello, World!",
			length:   8,
			omission: []string{">>>"},
			expected: "Hello>>>",
		},
		{
			name:     "字符串长度小于目标长度",
			s:        "Hello",
			length:   10,
			expected: "Hello",
		},
		{
			name:     "长度为0",
			s:        "Hello",
			length:   0,
			expected: "",
		},
		{
			name:     "省略标记长度大于目标长度",
			s:        "Hello",
			length:   3,
			omission: []string{"...."},
			expected: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if len(tt.omission) > 0 {
				result = Truncate(tt.s, tt.length, tt.omission[0])
			} else {
				result = Truncate(tt.s, tt.length)
			}
			if result != tt.expected {
				t.Errorf("Truncate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "空字符串",
			input:    "",
			expected: true,
		},
		{
			name:     "只有空格",
			input:    "   ",
			expected: true,
		},
		{
			name:     "只有制表符和换行符",
			input:    "\t\n",
			expected: true,
		},
		{
			name:     "非空字符串",
			input:    "Hello",
			expected: false,
		},
		{
			name:     "包含空格的非空字符串",
			input:    " Hello ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmpty(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmpty() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultIfEmpty(t *testing.T) {
	tests := []struct {
		name         string
		s            string
		defaultValue string
		expected     string
	}{
		{
			name:         "空字符串",
			s:            "",
			defaultValue: "N/A",
			expected:     "N/A",
		},
		{
			name:         "只有空格",
			s:            "   ",
			defaultValue: "N/A",
			expected:     "N/A",
		},
		{
			name:         "非空字符串",
			s:            "Hello",
			defaultValue: "N/A",
			expected:     "Hello",
		},
		{
			name:         "默认值为空",
			s:            "",
			defaultValue: "",
			expected:     "",
		},
		{
			name:         "包含空格的非空字符串",
			s:            " Hello ",
			defaultValue: "N/A",
			expected:     " Hello ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultIfEmpty(tt.s, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("DefaultIfEmpty() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// 基准测试
func BenchmarkReverse(b *testing.B) {
	testString := "Hello, World! This is a test string for benchmarking."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reverse(testString)
	}
}

func BenchmarkRandomString(b *testing.B) {
	b.Run("Length10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RandomString(10)
		}
	})

	b.Run("Length20", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RandomString(20)
		}
	})
}

func BenchmarkNormalize(b *testing.B) {
	testString := "  Hello   World\n\r  \t  Test  "

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Normalize(testString)
	}
}

func BenchmarkChunkSplit(b *testing.B) {
	testString := "HelloWorldThisIsATestStringForBenchmarking"

	b.Run("ChunkSize5", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ChunkSplit(testString, 5, "-")
		}
	})

	b.Run("ChunkSize10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ChunkSplit(testString, 10, "-")
		}
	})
}

func BenchmarkToCamelCase(b *testing.B) {
	testString := "hello_world_test_string"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToCamelCase(testString, "_")
	}
}

func BenchmarkFirstUpper(b *testing.B) {
	testStrings := []string{"hello", "world", "test", "string"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			FirstUpper(s)
		}
	}
}

func BenchmarkFirstLower(b *testing.B) {
	testStrings := []string{"Hello", "World", "Test", "String"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			FirstLower(s)
		}
	}
}

func BenchmarkFilterInvisibleChars(b *testing.B) {
	testString := "Hello\x00World\x1F\t \nTest\x7F"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterInvisibleChars(testString)
	}
}

func BenchmarkReplacePlaceholders(b *testing.B) {
	query := "SELECT * FROM users WHERE age > ? AND status = ? AND name LIKE ?"
	replacements := []string{"18", "'active'", "'%John%'"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ReplacePlaceholders(query, replacements)
	}
}

func BenchmarkFilterNumber(b *testing.B) {
	testString := "abc123def456ghi789jkl"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FilterNumber(testString)
	}
}

func BenchmarkTruncate(b *testing.B) {
	testString := "This is a very long string that needs to be truncated for testing purposes"

	b.Run("Length20", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Truncate(testString, 20)
		}
	})

	b.Run("Length50", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			Truncate(testString, 50)
		}
	})
}

func BenchmarkIsEmpty(b *testing.B) {
	testStrings := []string{"", "   ", "\t\n", "Hello", " World "}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			IsEmpty(s)
		}
	}
}
