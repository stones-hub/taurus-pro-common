package cmd

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// TestOptionType 测试选项类型枚举
func TestOptionType(t *testing.T) {
	tests := []struct {
		name     string
		optType  OptionType
		expected string
	}{
		{"String类型", OptionTypeString, "string"},
		{"Int类型", OptionTypeInt, "int"},
		{"Int64类型", OptionTypeInt64, "int64"},
		{"Bool类型", OptionTypeBool, "bool"},
		{"Float类型", OptionTypeFloat, "float64"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证选项类型值
			switch tt.optType {
			case OptionTypeString, OptionTypeInt, OptionTypeInt64, OptionTypeBool, OptionTypeFloat:
				// 类型有效
			default:
				t.Errorf("无效的选项类型: %v", tt.optType)
			}
		})
	}
}

// TestOption 测试选项结构体
func TestOption(t *testing.T) {
	opt := Option{
		Name:        "test",
		Shorthand:   "t",
		Description: "测试选项",
		Type:        OptionTypeString,
		Required:    false,
		Default:     "default",
	}

	if opt.Name != "test" {
		t.Errorf("期望选项名为 'test', 实际为 '%s'", opt.Name)
	}

	if opt.Shorthand != "t" {
		t.Errorf("期望短选项为 't', 实际为 '%s'", opt.Shorthand)
	}

	if opt.Type != OptionTypeString {
		t.Errorf("期望选项类型为 OptionTypeString, 实际为 %v", opt.Type)
	}
}

// TestSafeTypeConversions 测试安全的类型转换函数
func TestSafeTypeConversions(t *testing.T) {
	t.Run("safeString", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			expected string
			ok       bool
		}{
			{"hello", "hello", true},
			{123, "", false},
			{nil, "", false},
			{"", "", true},
		}

		for _, tt := range tests {
			result, ok := safeString(tt.input)
			if result != tt.expected || ok != tt.ok {
				t.Errorf("safeString(%v) = (%s, %t), 期望 (%s, %t)",
					tt.input, result, ok, tt.expected, tt.ok)
			}
		}
	})

	t.Run("safeInt", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			expected int
			ok       bool
		}{
			{123, 123, true},
			{"hello", 0, false},
			{nil, 0, false},
			{0, 0, true},
		}

		for _, tt := range tests {
			result, ok := safeInt(tt.input)
			if result != tt.expected || ok != tt.ok {
				t.Errorf("safeInt(%v) = (%d, %t), 期望 (%d, %t)",
					tt.input, result, ok, tt.expected, tt.ok)
			}
		}
	})

	t.Run("safeBool", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			expected bool
			ok       bool
		}{
			{true, true, true},
			{false, false, true},
			{"true", false, false},
			{nil, false, false},
		}

		for _, tt := range tests {
			result, ok := safeBool(tt.input)
			if result != tt.expected || ok != tt.ok {
				t.Errorf("safeBool(%v) = (%t, %t), 期望 (%t, %t)",
					tt.input, result, ok, tt.expected, tt.ok)
			}
		}
	})

	t.Run("safeFloat", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			expected float64
			ok       bool
		}{
			{3.14, 3.14, true},
			{123, 0.0, false},
			{"3.14", 0.0, false},
			{nil, 0.0, false},
		}

		for _, tt := range tests {
			result, ok := safeFloat(tt.input)
			if result != tt.expected || ok != tt.ok {
				t.Errorf("safeFloat(%v) = (%f, %t), 期望 (%f, %t)",
					tt.input, result, ok, tt.expected, tt.ok)
			}
		}
	})

	t.Run("safeInt64", func(t *testing.T) {
		tests := []struct {
			input    interface{}
			expected int64
			ok       bool
		}{
			{int64(123), int64(123), true},
			{123, int64(0), false},
			{"123", int64(0), false},
			{nil, int64(0), false},
		}

		for _, tt := range tests {
			result, ok := safeInt64(tt.input)
			if result != tt.expected || ok != tt.ok {
				t.Errorf("safeInt64(%v) = (%d, %t), 期望 (%d, %t)",
					tt.input, result, ok, tt.expected, tt.ok)
			}
		}
	})
}

// TestValidateOptions 测试选项验证
func TestValidateOptions(t *testing.T) {
	t.Run("有效选项", func(t *testing.T) {
		options := []Option{
			{
				Name:        "name",
				Shorthand:   "n",
				Description: "名称",
				Type:        OptionTypeString,
				Required:    false,
				Default:     "default",
			},
			{
				Name:        "count",
				Shorthand:   "c",
				Description: "数量",
				Type:        OptionTypeInt,
				Required:    true,
				Default:     nil,
			},
		}

		err := validateOptions(options)
		if err != nil {
			t.Errorf("期望验证通过，实际错误: %v", err)
		}
	})

	t.Run("空选项名", func(t *testing.T) {
		options := []Option{
			{
				Name:        "",
				Description: "空名称",
				Type:        OptionTypeString,
			},
		}

		err := validateOptions(options)
		if err == nil {
			t.Error("期望验证失败，实际通过")
		}
		if !strings.Contains(err.Error(), "选项名不能为空") {
			t.Errorf("期望错误包含'选项名不能为空'，实际错误: %v", err)
		}
	})

	t.Run("重复选项名", func(t *testing.T) {
		options := []Option{
			{
				Name:        "name",
				Description: "名称1",
				Type:        OptionTypeString,
			},
			{
				Name:        "name",
				Description: "名称2",
				Type:        OptionTypeString,
			},
		}

		err := validateOptions(options)
		if err == nil {
			t.Error("期望验证失败，实际通过")
		}
		if !strings.Contains(err.Error(), "重复的选项名") {
			t.Errorf("期望错误包含'重复的选项名'，实际错误: %v", err)
		}
	})

	t.Run("无效短选项", func(t *testing.T) {
		options := []Option{
			{
				Name:        "name",
				Shorthand:   "ab", // 长度大于1
				Description: "名称",
				Type:        OptionTypeString,
			},
		}

		err := validateOptions(options)
		if err == nil {
			t.Error("期望验证失败，实际通过")
		}
		if !strings.Contains(err.Error(), "短选项名必须是单个字符") {
			t.Errorf("期望错误包含'短选项名必须是单个字符'，实际错误: %v", err)
		}
	})

	t.Run("重复短选项", func(t *testing.T) {
		options := []Option{
			{
				Name:        "name1",
				Shorthand:   "n",
				Description: "名称1",
				Type:        OptionTypeString,
			},
			{
				Name:        "name2",
				Shorthand:   "n",
				Description: "名称2",
				Type:        OptionTypeString,
			},
		}

		err := validateOptions(options)
		if err == nil {
			t.Error("期望验证失败，实际通过")
		}
		if !strings.Contains(err.Error(), "重复的短选项名") {
			t.Errorf("期望错误包含'重复的短选项名'，实际错误: %v", err)
		}
	})

	t.Run("默认值类型错误", func(t *testing.T) {
		options := []Option{
			{
				Name:        "count",
				Description: "数量",
				Type:        OptionTypeInt,
				Default:     "not a number", // 字符串类型，但选项是int类型
			},
		}

		err := validateOptions(options)
		if err == nil {
			t.Error("期望验证失败，实际通过")
		}
		if !strings.Contains(err.Error(), "默认值无效") {
			t.Errorf("期望错误包含'默认值无效'，实际错误: %v", err)
		}
	})
}

// TestNewBaseCommand 测试创建基础命令
func TestNewBaseCommand(t *testing.T) {
	t.Run("有效命令", func(t *testing.T) {
		options := []Option{
			{
				Name:        "name",
				Description: "名称",
				Type:        OptionTypeString,
			},
		}

		cmd, err := NewBaseCommand("test", "测试命令", "[options]", options)
		if err != nil {
			t.Errorf("期望创建成功，实际错误: %v", err)
		}

		if cmd.Name() != "test" {
			t.Errorf("期望命令名为 'test', 实际为 '%s'", cmd.Name())
		}

		if cmd.Description() != "测试命令" {
			t.Errorf("期望描述为 '测试命令', 实际为 '%s'", cmd.Description())
		}
	})

	t.Run("空命令名", func(t *testing.T) {
		_, err := NewBaseCommand("", "描述", "用法", nil)
		if err == nil {
			t.Error("期望创建失败，实际成功")
		}
		if !strings.Contains(err.Error(), "命令名不能为空") {
			t.Errorf("期望错误包含'命令名不能为空'，实际错误: %v", err)
		}
	})

	t.Run("空描述", func(t *testing.T) {
		_, err := NewBaseCommand("test", "", "用法", nil)
		if err == nil {
			t.Error("期望创建失败，实际成功")
		}
		if !strings.Contains(err.Error(), "命令描述不能为空") {
			t.Errorf("期望错误包含'命令描述不能为空'，实际错误: %v", err)
		}
	})

	t.Run("空用法", func(t *testing.T) {
		_, err := NewBaseCommand("test", "描述", "", nil)
		if err == nil {
			t.Error("期望创建失败，实际成功")
		}
		if !strings.Contains(err.Error(), "使用说明不能为空") {
			t.Errorf("期望错误包含'使用说明不能为空'，实际错误: %v", err)
		}
	})
}

// TestBaseCommandHelp 测试帮助信息生成
func TestBaseCommandHelp(t *testing.T) {
	options := []Option{
		{
			Name:        "name",
			Shorthand:   "n",
			Description: "名称",
			Type:        OptionTypeString,
			Required:    true,
			Default:     nil,
		},
		{
			Name:        "count",
			Shorthand:   "c",
			Description: "数量",
			Type:        OptionTypeInt,
			Required:    false,
			Default:     10,
		},
		{
			Name:        "verbose",
			Shorthand:   "v",
			Description: "详细输出",
			Type:        OptionTypeBool,
			Required:    false,
			Default:     false,
		},
	}

	cmd, err := NewBaseCommand("test", "测试命令", "[options] <file>", options)
	if err != nil {
		t.Fatalf("创建命令失败: %v", err)
	}

	help := cmd.Help()

	// 验证帮助信息包含必要内容
	if !strings.Contains(help, "测试命令") {
		t.Error("帮助信息应包含命令描述")
	}

	if !strings.Contains(help, "test [options] <file>") {
		t.Error("帮助信息应包含用法说明")
	}

	if !strings.Contains(help, "name") {
		t.Error("帮助信息应包含选项名")
	}

	if !strings.Contains(help, "(必填)") {
		t.Error("帮助信息应标记必填选项")
	}

	if !strings.Contains(help, "(默认: 10)") {
		t.Error("帮助信息应显示默认值")
	}
}

// TestParseOptions 测试选项解析
func TestParseOptions(t *testing.T) {
	options := []Option{
		{
			Name:        "name",
			Shorthand:   "n",
			Description: "名称",
			Type:        OptionTypeString,
			Required:    true,
		},
		{
			Name:        "count",
			Shorthand:   "c",
			Description: "数量",
			Type:        OptionTypeInt,
			Required:    false,
			Default:     10,
		},
		{
			Name:        "verbose",
			Shorthand:   "v",
			Description: "详细输出",
			Type:        OptionTypeBool,
			Required:    false,
			Default:     false,
		},
		{
			Name:        "price",
			Shorthand:   "p",
			Description: "价格",
			Type:        OptionTypeFloat,
			Required:    false,
			Default:     0.0,
		},
	}

	cmd, err := NewBaseCommand("test", "测试命令", "[options]", options)
	if err != nil {
		t.Fatalf("创建命令失败: %v", err)
	}

	t.Run("有效参数", func(t *testing.T) {
		args := []string{"--name", "testfile", "--count", "5", "--verbose", "file1.txt", "file2.txt"}

		ctx, err := cmd.ParseOptions(args)
		if err != nil {
			t.Errorf("期望解析成功，实际错误: %v", err)
		}

		// 验证选项值
		if name, ok := ctx.Options["name"].(string); !ok || name != "testfile" {
			t.Errorf("期望name为'testfile'，实际为%v", ctx.Options["name"])
		}

		if count, ok := ctx.Options["count"].(int); !ok || count != 5 {
			t.Errorf("期望count为5，实际为%v", ctx.Options["count"])
		}

		if verbose, ok := ctx.Options["verbose"].(bool); !ok || !verbose {
			t.Errorf("期望verbose为true，实际为%v", ctx.Options["verbose"])
		}

		// 验证位置参数
		expectedArgs := []string{"file1.txt", "file2.txt"}
		if !reflect.DeepEqual(ctx.Args, expectedArgs) {
			t.Errorf("期望位置参数为%v，实际为%v", expectedArgs, ctx.Args)
		}
	})

	t.Run("短选项", func(t *testing.T) {
		args := []string{"-n", "testfile", "-c", "5", "-v", "file.txt"}

		ctx, err := cmd.ParseOptions(args)
		if err != nil {
			t.Errorf("期望解析成功，实际错误: %v", err)
		}

		if name, ok := ctx.Options["name"].(string); !ok || name != "testfile" {
			t.Errorf("期望name为'testfile'，实际为%v", ctx.Options["name"])
		}
		_ = ctx // 避免未使用变量警告
	})

	t.Run("缺少必填选项", func(t *testing.T) {
		args := []string{"--count", "5", "file.txt"}

		ctx, err := cmd.ParseOptions(args)
		if err == nil {
			t.Error("期望解析失败，实际成功")
		}
		if ctx != nil {
			t.Error("期望返回nil上下文，实际返回非nil")
		}
		if !strings.Contains(err.Error(), "选项 --name 是必填的") {
			t.Errorf("期望错误包含'选项 --name 是必填的'，实际错误: %v", err)
		}
	})

	t.Run("未定义选项", func(t *testing.T) {
		args := []string{"--unknown", "value", "file.txt"}

		_, err := cmd.ParseOptions(args)
		if err == nil {
			t.Error("期望解析失败，实际成功")
		}
		if !strings.Contains(err.Error(), "未定义的选项") {
			t.Errorf("期望错误包含'未定义的选项'，实际错误: %v", err)
		}
	})

	t.Run("布尔选项无需值", func(t *testing.T) {
		args := []string{"--name", "testfile", "--verbose", "file.txt"}

		ctx, err := cmd.ParseOptions(args)
		if err != nil {
			t.Errorf("期望解析成功，实际错误: %v", err)
		}

		if verbose, ok := ctx.Options["verbose"].(bool); !ok || !verbose {
			t.Errorf("期望verbose为true，实际为%v", ctx.Options["verbose"])
		}
	})

	t.Run("浮点数选项", func(t *testing.T) {
		args := []string{"--name", "testfile", "--price", "3.14", "file.txt"}

		ctx, err := cmd.ParseOptions(args)
		if err != nil {
			t.Errorf("期望解析成功，实际错误: %v", err)
		}

		if price, ok := ctx.Options["price"].(float64); !ok || price != 3.14 {
			t.Errorf("期望price为3.14，实际为%v", ctx.Options["price"])
		}
		_ = ctx // 避免未使用变量警告
	})

	t.Run("布尔选项错误使用", func(t *testing.T) {
		args := []string{"--name", "testfile", "--verbose", "true", "file.txt"}

		ctx, err := cmd.ParseOptions(args)
		if err == nil {
			t.Error("期望解析失败，实际成功")
		}
		if !strings.Contains(err.Error(), "布尔选项 --verbose 不需要值") {
			t.Errorf("期望错误包含'布尔选项 --verbose 不需要值'，实际错误: %v", err)
		}
		if ctx != nil {
			t.Error("期望返回nil上下文，实际返回非nil")
		}
	})

	t.Run("布尔选项正确使用", func(t *testing.T) {
		args := []string{"--name", "testfile", "--verbose", "file.txt"}

		ctx, err := cmd.ParseOptions(args)
		if err != nil {
			t.Errorf("期望解析成功，实际错误: %v", err)
		}

		if verbose, ok := ctx.Options["verbose"].(bool); !ok || !verbose {
			t.Errorf("期望verbose为true，实际为%v", ctx.Options["verbose"])
		}
		_ = ctx // 避免未使用变量警告
	})
}

// TestCommandRun 测试命令执行
func TestCommandRun(t *testing.T) {
	options := []Option{
		{
			Name:        "name",
			Description: "名称",
			Type:        OptionTypeString,
			Required:    false,
		},
	}

	cmd, err := NewBaseCommand("test", "测试命令", "[options]", options)
	if err != nil {
		t.Fatalf("创建命令失败: %v", err)
	}

	t.Run("无参数时显示帮助", func(t *testing.T) {
		err := cmd.Run([]string{})
		if err != nil {
			t.Errorf("期望执行成功，实际错误: %v", err)
		}
	})

	t.Run("有参数时返回未实现错误", func(t *testing.T) {
		err := cmd.Run([]string{"--name", "test", "file.txt"})
		if err == nil {
			t.Error("期望返回错误，实际成功")
		}
		if !strings.Contains(err.Error(), "未实现具体的执行逻辑") {
			t.Errorf("期望错误包含'未实现具体的执行逻辑'，实际错误: %v", err)
		}
	})
}

// TestFormatParseError 测试解析错误格式化
func TestFormatParseError(t *testing.T) {
	options := []Option{
		{
			Name:        "name",
			Description: "名称",
			Type:        OptionTypeString,
		},
	}

	cmd, err := NewBaseCommand("test", "测试命令", "[options]", options)
	if err != nil {
		t.Fatalf("创建命令失败: %v", err)
	}

	// 模拟未定义选项的错误
	args := []string{"--unknown", "value"}
	err = cmd.formatParseError(fmt.Errorf("flag provided but not defined: -unknown"), args)

	if err == nil {
		t.Error("期望返回错误，实际为nil")
	}

	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "未定义的选项") {
		t.Errorf("期望错误信息包含'未定义的选项'，实际为: %s", errorMsg)
	}

	if !strings.Contains(errorMsg, "可用的选项") {
		t.Errorf("期望错误信息包含'可用的选项'，实际为: %s", errorMsg)
	}
}

// TestIsValidShortOption 测试短选项字符验证
func TestIsValidShortOption(t *testing.T) {
	tests := []struct {
		char     byte
		expected bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'-', false},
		{'_', false},
		{' ', false},
		{'\t', false},
	}

	for _, tt := range tests {
		result := isValidShortOption(tt.char)
		if result != tt.expected {
			t.Errorf("isValidShortOption('%c') = %t, 期望 %t", tt.char, result, tt.expected)
		}
	}
}

// TestValidateDefaultValue 测试默认值验证
func TestValidateDefaultValue(t *testing.T) {
	t.Run("有效默认值", func(t *testing.T) {
		tests := []struct {
			opt      Option
			expected error
		}{
			{Option{Type: OptionTypeString, Default: "hello"}, nil},
			{Option{Type: OptionTypeInt, Default: 123}, nil},
			{Option{Type: OptionTypeInt64, Default: int64(123)}, nil},
			{Option{Type: OptionTypeBool, Default: true}, nil},
			{Option{Type: OptionTypeFloat, Default: 3.14}, nil},
		}

		for _, tt := range tests {
			err := validateDefaultValue(tt.opt)
			if err != tt.expected {
				t.Errorf("validateDefaultValue(%v) = %v, 期望 %v", tt.opt, err, tt.expected)
			}
		}
	})

	t.Run("无效默认值", func(t *testing.T) {
		tests := []struct {
			opt      Option
			expected string
		}{
			{Option{Type: OptionTypeString, Default: 123}, "期望字符串类型"},
			{Option{Type: OptionTypeInt, Default: "123"}, "期望整数类型"},
			{Option{Type: OptionTypeBool, Default: "true"}, "期望布尔类型"},
			{Option{Type: OptionTypeFloat, Default: "3.14"}, "期望浮点数类型"},
		}

		for _, tt := range tests {
			err := validateDefaultValue(tt.opt)
			if err == nil {
				t.Errorf("期望validateDefaultValue(%v)返回错误，实际为nil", tt.opt)
			}
			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("期望错误包含'%s'，实际错误: %v", tt.expected, err)
			}
		}
	})
}

// BenchmarkParseOptions 性能测试
func BenchmarkParseOptions(b *testing.B) {
	options := []Option{
		{
			Name:        "name",
			Shorthand:   "n",
			Description: "名称",
			Type:        OptionTypeString,
			Required:    true,
		},
		{
			Name:        "count",
			Shorthand:   "c",
			Description: "数量",
			Type:        OptionTypeInt,
			Required:    false,
			Default:     10,
		},
		{
			Name:        "verbose",
			Shorthand:   "v",
			Description: "详细输出",
			Type:        OptionTypeBool,
			Required:    false,
			Default:     false,
		},
	}

	cmd, err := NewBaseCommand("test", "测试命令", "[options]", options)
	if err != nil {
		b.Fatalf("创建命令失败: %v", err)
	}

	args := []string{"--name", "testfile", "--count", "5", "--verbose", "file.txt"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cmd.ParseOptions(args)
		if err != nil {
			b.Fatalf("解析失败: %v", err)
		}
	}
}
