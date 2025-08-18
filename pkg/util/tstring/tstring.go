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

package tstring

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/language"
	"golang.org/x/text/transform"
)

const (
	defaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// Reverse 将输入字符串按字节顺序反转
// 参数：
//   - s: 要反转的字符串
//
// 返回值：
//   - string: 反转后的字符串
//
// 使用示例：
//
//	reversed := tstring.Reverse("Hello")  // 返回 "olleH"
//
// 注意事项：
//   - 此函数按字节处理，对于多字节字符（如中文）可能会产生乱码
//   - 如果需要处理Unicode字符，应该使用 []rune 进行处理
func Reverse(s string) string {
	l := len(s)
	buf := make([]byte, l)
	for i := 0; i < len(s); i++ {
		buf[l-1] = s[i]
		l--
	}
	return string(buf)
}

// Random 生成随机字符串
// 这是一个向后兼容的函数，内部使用 RandomString 实现
// 如果需要更多控制，请直接使用 RandomString 函数
func Random(n int) string {
	return RandomString(n)
}

// Normalize 标准化字符串，移除多余的空白字符和换行符
// 参数：
//   - s: 要标准化的字符串
//
// 返回值：
//   - string: 标准化后的字符串，所有连续的空白字符会被替换为单个空格，
//     字符串首尾的空白字符会被移除
//
// 使用示例：
//
//	text := "  Hello   World\n\r  "
//	normalized := tstring.Normalize(text)  // 返回 "Hello World"
//
// 注意事项：
//   - 会处理所有类型的空白字符，包括空格、制表符、换行符等
//   - 如果输入字符串只包含空白字符，返回空字符串
func Normalize(s string) string {
	nlRe := regexp.MustCompile(`\r?\r`)
	re := regexp.MustCompile(`\s+`)

	s = nlRe.ReplaceAllString(s, " ")
	s = re.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}

// ChunkSplit 将字符串按指定长度分割成多个块，并在每个块之间添加分隔符
// 参数：
//   - input: 要分割的字符串
//   - chunkSize: 每个块的大小（字符数）
//   - delimiter: 块之间的分隔符
//
// 返回值：
//   - string: 分割后的字符串，每个块之间包含指定的分隔符
//
// 使用示例：
//
//	// 每4个字符分割，使用连字符作为分隔符
//	result := tstring.ChunkSplit("HelloWorld", 4, "-")  // 返回 "Hell-oWor-ld"
//
//	// 用于格式化长字符串
//	text := "1234567890abcdefghij"
//	formatted := tstring.ChunkSplit(text, 5, "\n")  // 每5个字符换行
//
// 注意事项：
//   - 如果 chunkSize 小于等于0，返回原始字符串
//   - 最后一个块如果长度小于 chunkSize，也会添加分隔符
//   - 分割基于字节而不是字符，对于多字节字符可能会出现意外的分割
func ChunkSplit(input string, chunkSize int, delimiter string) string {
	// 如果 chunkSize 小于等于 0，直接返回原字符串
	if chunkSize <= 0 {
		return input
	}

	var chunks []string
	for i := 0; i < len(input); i += chunkSize {
		end := i + chunkSize
		if end > len(input) {
			end = len(input)
		}
		chunks = append(chunks, input[i:end])
	}

	// 在每个块后面添加分隔符
	var result strings.Builder
	for _, chunk := range chunks {
		result.WriteString(chunk)
		result.WriteString(delimiter)
	}

	return result.String()
}

// GBKToUTF8 将GBK编码的字节序列转换为UTF-8编码的字符串
// 参数：
//   - b: GBK编码的字节序列
//
// 返回值：
//   - string: 转换后的UTF-8字符串
//   - error: 如果转换过程中出现错误则返回错误信息
//
// 使用示例：
//
//	gbkBytes := []byte{0xB2, 0xE2, 0xCA, 0xD4}  // GBK编码的"测试"
//	utf8Str, err := tstring.GBKToUTF8(gbkBytes)
//	if err != nil {
//	    return err
//	}
//	fmt.Println(utf8Str)  // 输出: 测试
//
// 注意事项：
//   - 输入必须是有效的GBK编码字节序列
//   - 如果输入包含非GBK编码的字节，会返回错误
//   - 建议配合 UTF8ToGBK 函数使用，实现编码的双向转换
func GBKToUTF8(b []byte) (string, error) {
	decoder := simplifiedchinese.GBK.NewDecoder()
	utf8String, _, err := transform.String(decoder, string(b))
	if err != nil {
		return "", err
	}
	return utf8String, nil
}

// ToCamelCase 将分隔符分隔的字符串转换为小驼峰格式（第一个单词首字母小写，其他单词首字母大写）
// 参数：
//   - s: 要转换的字符串
//   - delimiter: 单词之间的分隔符
//
// 返回值：
//   - string: 转换后的小驼峰格式字符串
//
// 使用示例：
//
//	// 使用下划线分隔的字符串
//	result := tstring.ToCamelCase("hello_world_test", "_")  // 返回 "helloWorldTest"
//
//	// 使用空格分隔的字符串
//	result := tstring.ToCamelCase("hello world test", " ")  // 返回 "helloWorldTest"
//
//	// 使用连字符分隔的字符串
//	result := tstring.ToCamelCase("hello-world-test", "-")  // 返回 "helloWorldTest"
//
// 注意事项：
//   - 如果输入字符串为空，返回空字符串
//   - 连续的分隔符会被视为一个分隔符
//   - 分隔符前后的空白字符会被忽略
//   - 第一个单词会被转换为全小写，之后的单词首字母大写
func ToCamelCase(s string, delimiter string) string {
	// 将分隔符替换为空格，这样可以处理连续的分隔符
	s = strings.ReplaceAll(s, delimiter, " ")
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	titleCaser := cases.Title(language.Und)
	var result strings.Builder

	// 第一个单词保持原样
	result.WriteString(strings.ToLower(words[0]))

	// 其余单词首字母大写
	for _, word := range words[1:] {
		result.WriteString(titleCaser.String(word))
	}

	return result.String()
}

// UnderscoreToCamelCase 将下划线格式的字符串转换为驼峰格式
func UnderscoreToCamelCase(s string) string {
	return ToCamelCase(s, "_")
}

// MakeHump 将字符串转换为驼峰命名，支持自定义分隔符
// Deprecated: 使用 ToCamelCase 替代，它提供了相同的功能
func MakeHump(s, delimiter string) string {
	return ToCamelCase(s, delimiter)
}

// FirstUpper 将字符串的首字母转换为大写（标题格式）
// 参数：
//   - s: 要转换的字符串
//
// 返回值：
//   - string: 首字母大写的字符串
//
// 使用示例：
//
//	result := tstring.FirstUpper("hello")  // 返回 "Hello"
//	result := tstring.FirstUpper("world")  // 返回 "World"
//
// 注意事项：
//   - 如果输入字符串为空，返回空字符串
//   - 使用 Unicode 标题格式规则进行转换，支持各种语言的字符
//   - 只转换第一个字符，其他字符保持不变
func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	// 只转换第一个字符，其他字符保持不变
	runes := []rune(s)
	if len(runes) > 0 && runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] = runes[0] - 32 // 转换为大写
	}
	return string(runes)
}

// FirstLower 将字符串的首字母转换为小写
// 参数：
//   - s: 要转换的字符串
//
// 返回值：
//   - string: 首字母小写的字符串
//
// 使用示例：
//
//	result := tstring.FirstLower("Hello")  // 返回 "hello"
//	result := tstring.FirstLower("World")  // 返回 "world"
//
// 注意事项：
//   - 如果输入字符串为空，返回空字符串
//   - 只转换第一个字符，其他字符保持不变
//   - 此函数通常用于生成符合小驼峰命名规范的标识符
func FirstLower(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// RandomString 生成指定长度的随机字符串，支持自定义字符集
// 参数：
//   - n: 要生成的字符串长度
//   - charset: 可选的字符集，如果不提供则使用默认字符集（字母和数字）
//
// 返回值：
//   - string: 生成的随机字符串
//
// 使用示例：
//
//	// 使用默认字符集（字母和数字）
//	str := tstring.RandomString(8)  // 可能返回 "a1B2c3D4"
//
//	// 使用自定义字符集
//	str := tstring.RandomString(6, "ABCDEF0123456789")  // 生成6位16进制字符
//
//	// 生成纯数字随机串
//	str := tstring.RandomString(4, "0123456789")  // 生成4位数字验证码
//
// 注意事项：
//   - 使用 crypto/rand 生成安全的随机数
//   - 默认字符集包含大小写字母和数字
//   - 如果生成随机数失败，会使用最后一个字符（极少发生）
//   - 适用于生成验证码、临时密码等场景
//   - 生成的字符串长度严格等于指定的长度
func RandomString(n int, charset ...string) string {
	var letters []rune
	if len(charset) > 0 {
		letters = []rune(charset[0])
	} else {
		letters = []rune(defaultCharset)
	}

	// 使用 crypto/rand 生成更安全的随机数
	b := make([]rune, n)
	max := big.NewInt(int64(len(letters)))
	for i := range b {
		// 生成随机索引
		index, err := rand.Int(rand.Reader, max)
		if err != nil {
			// 如果生成随机数失败，使用最后一个字符（这种情况极少发生）
			b[i] = letters[len(letters)-1]
			continue
		}
		b[i] = letters[index.Int64()]
	}
	return string(b)
}

// FilterInvisibleChars 从字符串中过滤掉所有不可见字符（控制字符和特殊字符）
// 参数：
//   - s: 要过滤的字符串
//
// 返回值：
//   - string: 只包含可见字符的字符串
//
// 使用示例：
//
//	// 过滤控制字符
//	text := "Hello\x00World\x1F"
//	clean := tstring.FilterInvisibleChars(text)  // 返回 "HelloWorld"
//
//	// 过滤空格和制表符
//	text := "Hello\t \nWorld"
//	clean := tstring.FilterInvisibleChars(text)  // 返回 "HelloWorld"
//
// 注意事项：
//   - ASCII码 32（空格）以下的字符会被移除
//   - ASCII码 127（DEL）及以上的字符会被移除
//   - 此函数主要用于清理可能包含控制字符的用户输入
//   - 如果需要保留空格，应该使用 Normalize 函数
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

// ReplacePlaceholders 按顺序将字符串中的问号(?)替换为提供的替换值
// 参数：
//   - query: 包含问号占位符的字符串
//   - replacements: 要替换的值列表，按顺序替换每个问号
//
// 返回值：
//   - string: 完成替换后的字符串
//
// 使用示例：
//
//	// 替换SQL查询中的占位符
//	query := "SELECT * FROM users WHERE age > ? AND status = ?"
//	sql := tstring.ReplacePlaceholders(query, []string{"18", "'active'"})
//	// 返回: "SELECT * FROM users WHERE age > 18 AND status = 'active'"
//
//	// 替换消息模板中的占位符
//	template := "Hello ?, your code is ?"
//	msg := tstring.ReplacePlaceholders(template, []string{"John", "123456"})
//	// 返回: "Hello John, your code is 123456"
//
// 注意事项：
//   - 如果替换值数量少于问号数量，剩余的问号将保持不变
//   - 如果替换值数量多于问号数量，多余的替换值将被忽略
//   - 替换是按顺序进行的，从左到右替换每个问号
//   - 此函数不处理转义的问号（如 \?）
func ReplacePlaceholders(query string, replacements []string) string {
	for _, replacement := range replacements {
		query = strings.Replace(query, "?", replacement, 1)
	}
	return query
}

// FilterNumber 从字符串中移除所有数字字符
// 参数：
//   - str: 要处理的字符串
//
// 返回值：
//   - string: 移除所有数字后的字符串
//
// 使用示例：
//
//	// 移除版本号中的数字
//	version := "v1.2.3"
//	result := tstring.FilterNumber(version)  // 返回 "v.."
//
//	// 从产品编号中提取字母部分
//	code := "ABC123DEF456"
//	letters := tstring.FilterNumber(code)  // 返回 "ABCDEF"
//
// 注意事项：
//   - 会移除所有阿拉伯数字（0-9）
//   - 不会移除其他类型的数字（如罗马数字、中文数字等）
//   - 保留所有其他字符，包括标点符号和空白字符
//   - 如果字符串只包含数字，将返回空字符串
func FilterNumber(str string) string {
	re := regexp.MustCompile(`\d+`)
	return re.ReplaceAllString(str, "")
}

// UTF8ToGBK 将UTF-8编码的字符串转换为GBK编码的字节序列
// 参数：
//   - s: UTF-8编码的字符串
//
// 返回值：
//   - []byte: GBK编码的字节序列
//   - error: 如果转换过程中出现错误则返回错误信息
//
// 使用示例：
//
//	// 转换中文字符串
//	utf8Str := "测试"
//	gbkBytes, err := tstring.UTF8ToGBK(utf8Str)
//	if err != nil {
//	    return err
//	}
//	// gbkBytes 现在包含GBK编码的字节序列
//
//	// 写入GBK编码的文件
//	gbkBytes, err := tstring.UTF8ToGBK("你好，世界")
//	if err != nil {
//	    return err
//	}
//	err = os.WriteFile("hello.txt", gbkBytes, 0644)
//
// 注意事项：
//   - 输入必须是有效的UTF-8编码字符串
//   - 如果字符在GBK字符集中不存在，会返回错误
//   - 建议配合 GBKToUTF8 函数使用，实现编码的双向转换
//   - 主要用于处理需要GBK编码的遗留系统或中文Windows系统
func UTF8ToGBK(s string) ([]byte, error) {
	encoder := simplifiedchinese.GBK.NewEncoder()
	gbkBytes, _, err := transform.Bytes(encoder, []byte(s))
	if err != nil {
		return nil, err
	}
	return gbkBytes, nil
}

// Truncate 将字符串截断到指定长度，并可选地在末尾添加省略号或其他标记
// 参数：
//   - s: 要截断的字符串
//   - length: 目标长度（字符数）
//   - omission: 可选的省略标记，默认为"..."
//
// 返回值：
//   - string: 截断后的字符串，如果需要则包含省略标记
//
// 使用示例：
//
//	// 使用默认省略号
//	text := "Hello, World!"
//	result := tstring.Truncate(text, 8)  // 返回 "Hello..."
//
//	// 使用自定义省略标记
//	text := "Hello, World!"
//	result := tstring.Truncate(text, 8, ">>>")  // 返回 "Hello>>>"
//
//	// 不需要截断的情况
//	text := "Hello"
//	result := tstring.Truncate(text, 10)  // 返回 "Hello"
//
// 注意事项：
//   - 如果 length 小于等于0，返回空字符串
//   - 如果原字符串长度小于等于目标长度，返回原字符串
//   - 省略标记的长度会计入目标长度中
//   - 如果省略标记长度大于等于目标长度，返回省略标记的前 length 个字符
//   - 此函数按 Unicode 字符处理，支持中文等多字节字符
func Truncate(s string, length int, omission ...string) string {
	if length <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= length {
		return s
	}

	omissionStr := "..."
	if len(omission) > 0 {
		omissionStr = omission[0]
	}

	// 如果省略号长度大于等于截断长度，直接返回省略号的前n个字符
	omissionRunes := []rune(omissionStr)
	if len(omissionRunes) >= length {
		return string(omissionRunes[:length])
	}

	// 返回截断后的字符串加省略号
	return string(runes[:length-len(omissionRunes)]) + omissionStr
}

// EqualFold 比较两个字符串是否相等，忽略大小写差异
// 参数：
//   - s1: 第一个字符串
//   - s2: 第二个字符串
//
// 返回值：
//   - bool: 如果两个字符串在忽略大小写的情况下相等，返回 true；否则返回 false
//
// 使用示例：
//
//	// 比较不同大小写的相同单词
//	equal := tstring.EqualFold("hello", "HELLO")  // 返回 true
//	equal := tstring.EqualFold("World", "world")  // 返回 true
//
//	// 比较不同的单词
//	equal := tstring.EqualFold("hello", "world")  // 返回 false
//
// 注意事项：
//   - 使用 Unicode 大小写折叠规则进行比较
//   - 支持多语言字符的大小写比较
//   - 空字符串只与空字符串相等
//   - 此函数比直接转换后比较更高效
func EqualFold(s1, s2 string) bool {
	return strings.EqualFold(s1, s2)
}

// PadLeft 在字符串左侧填充指定字符，直到字符串达到目标长度
// 参数：
//   - s: 要填充的字符串
//   - length: 目标长度
//   - pad: 用于填充的字符
//
// 返回值：
//   - string: 左侧填充后的字符串
//
// 使用示例：
//
//	// 数字前补零
//	num := "42"
//	result := tstring.PadLeft(num, 5, '0')  // 返回 "00042"
//
//	// 文本左对齐
//	text := "Hello"
//	result := tstring.PadLeft(text, 10, ' ')  // 返回 "     Hello"
//
// 注意事项：
//   - 如果原字符串长度大于或等于目标长度，返回原字符串
//   - 如果目标长度小于等于0，返回原字符串
//   - 填充字符使用 rune 类型，支持 Unicode 字符
//   - 常用于格式化输出，如数字前补零
func PadLeft(s string, length int, pad rune) string {
	if length <= 0 {
		return s
	}

	runes := []rune(s)
	if len(runes) >= length {
		return s
	}

	padding := make([]rune, length-len(runes))
	for i := range padding {
		padding[i] = pad
	}

	return string(padding) + s
}

// PadRight 在字符串右侧填充指定字符，直到字符串达到目标长度
// 参数：
//   - s: 要填充的字符串
//   - length: 目标长度
//   - pad: 用于填充的字符
//
// 返回值：
//   - string: 右侧填充后的字符串
//
// 使用示例：
//
//	// 文本右对齐
//	text := "Hello"
//	result := tstring.PadRight(text, 10, '-')  // 返回 "Hello-----"
//
//	// 固定宽度格式化
//	name := "John"
//	result := tstring.PadRight(name, 8, ' ')  // 返回 "John    "
//
// 注意事项：
//   - 如果原字符串长度大于或等于目标长度，返回原字符串
//   - 如果目标长度小于等于0，返回原字符串
//   - 填充字符使用 rune 类型，支持 Unicode 字符
//   - 常用于固定宽度的文本格式化
func PadRight(s string, length int, pad rune) string {
	if length <= 0 {
		return s
	}

	runes := []rune(s)
	if len(runes) >= length {
		return s
	}

	padding := make([]rune, length-len(runes))
	for i := range padding {
		padding[i] = pad
	}

	return s + string(padding)
}

// IsEmpty 检查字符串是否为空或仅包含空白字符
// 参数：
//   - s: 要检查的字符串
//
// 返回值：
//   - bool: 如果字符串为空或仅包含空白字符返回 true，否则返回 false
//
// 使用示例：
//
//	// 检查空字符串
//	empty := tstring.IsEmpty("")        // 返回 true
//	empty := tstring.IsEmpty("  \t\n")  // 返回 true
//	empty := tstring.IsEmpty("hello")   // 返回 false
//
// 注意事项：
//   - 会移除字符串首尾的空白字符后再判断
//   - 空白字符包括空格、制表符、换行符等
//   - 常用于验证用户输入是否有效
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty 检查字符串是否非空且包含非空白字符
// 参数：
//   - s: 要检查的字符串
//
// 返回值：
//   - bool: 如果字符串非空且包含非空白字符返回 true，否则返回 false
//
// 使用示例：
//
//	// 检查非空字符串
//	notEmpty := tstring.IsNotEmpty("hello")   // 返回 true
//	notEmpty := tstring.IsNotEmpty("")        // 返回 false
//	notEmpty := tstring.IsNotEmpty("  \t\n")  // 返回 false
//
// 注意事项：
//   - 是 IsEmpty 函数的逻辑取反
//   - 会移除字符串首尾的空白字符后再判断
//   - 空白字符包括空格、制表符、换行符等
//   - 常用于验证必填字段
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// DefaultIfEmpty 如果字符串为空或仅包含空白字符，则返回默认值
// 参数：
//   - s: 要检查的字符串
//   - defaultValue: 当字符串为空时返回的默认值
//
// 返回值：
//   - string: 如果输入字符串非空则返回原字符串，否则返回默认值
//
// 使用示例：
//
//	// 处理空字符串
//	result := tstring.DefaultIfEmpty("", "N/A")        // 返回 "N/A"
//	result := tstring.DefaultIfEmpty("  \t\n", "N/A")  // 返回 "N/A"
//	result := tstring.DefaultIfEmpty("hello", "N/A")   // 返回 "hello"
//
//	// 用于提供默认显示文本
//	name := getUserName()  // 可能返回空字符串
//	display := tstring.DefaultIfEmpty(name, "Anonymous")
//
// 注意事项：
//   - 使用 IsEmpty 函数判断字符串是否为空
//   - 会移除字符串首尾的空白字符后再判断
//   - 如果默认值也是空字符串，仍然会返回默认值
//   - 常用于处理可选的显示文本
func DefaultIfEmpty(s string, defaultValue string) string {
	if IsEmpty(s) {
		return defaultValue
	}
	return s
}
