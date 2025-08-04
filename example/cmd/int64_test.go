package main

import (
	"fmt"
	"testing"

	"github.com/stones-hub/taurus-pro-common/pkg/cmd"
)

// TestInt64Command 测试int64选项功能
func TestInt64Command(t *testing.T) {
	// 创建命令管理器
	manager := cmd.NewManager()

	// 创建包含int64选项的命令
	options := []cmd.Option{
		{
			Name:        "user-id",
			Shorthand:   "u",
			Description: "用户ID（长整型）",
			Type:        cmd.OptionTypeInt64,
			Required:    true,
		},
		{
			Name:        "limit",
			Shorthand:   "l",
			Description: "限制数量（长整型，可选）",
			Type:        cmd.OptionTypeInt64,
			Required:    false,
			Default:     int64(100),
		},
	}

	// 创建测试命令
	testCmd, err := cmd.NewBaseCommand(
		"test-int64",
		"测试int64选项功能",
		"[--user-id <id>] [--limit <number>]",
		options,
	)
	if err != nil {
		t.Fatalf("创建命令失败: %v", err)
	}

	// 注册命令
	if err := manager.Register(testCmd); err != nil {
		t.Fatalf("注册命令失败: %v", err)
	}

	// 测试用例
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		checkResult func(*cmd.CommandContext) error
	}{
		{
			name:        "正常使用int64选项",
			args:        []string{"--user-id", "1234567890123456789", "--limit", "50"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				// 检查user-id
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != 1234567890123456789 {
					return fmt.Errorf("user-id值错误，期望: 1234567890123456789，实际: %v", userID)
				}

				// 检查limit
				limit, exists := ctx.Options["limit"]
				if !exists {
					return fmt.Errorf("limit选项不存在")
				}
				if l, ok := limit.(int64); !ok || l != 50 {
					return fmt.Errorf("limit值错误，期望: 50，实际: %v", limit)
				}

				return nil
			},
		},
		{
			name:        "使用短选项",
			args:        []string{"-u", "9876543210987654", "-l", "200"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				// 检查user-id
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != 9876543210987654 {
					return fmt.Errorf("user-id值错误，期望: 9876543210987654，实际: %v", userID)
				}

				// 检查limit
				limit, exists := ctx.Options["limit"]
				if !exists {
					return fmt.Errorf("limit选项不存在")
				}
				if l, ok := limit.(int64); !ok || l != 200 {
					return fmt.Errorf("limit值错误，期望: 200，实际: %v", limit)
				}

				return nil
			},
		},
		{
			name:        "使用默认值",
			args:        []string{"--user-id", "123"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				// 检查user-id
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != 123 {
					return fmt.Errorf("user-id值错误，期望: 123，实际: %v", userID)
				}

				// 检查limit应该使用默认值
				limit, exists := ctx.Options["limit"]
				if !exists {
					return fmt.Errorf("limit选项不存在")
				}
				if l, ok := limit.(int64); !ok || l != 100 {
					return fmt.Errorf("limit值错误，期望默认值: 100，实际: %v", limit)
				}

				return nil
			},
		},
		{
			name:        "缺少必填选项",
			args:        []string{"--limit", "50"},
			expectError: true,
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil // 这个测试用例期望出错，所以这里不会执行
			},
		},
	}

	// 运行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 解析选项
			ctx, err := testCmd.ParseOptions(tc.args)

			if tc.expectError {
				if err == nil {
					t.Errorf("期望出错，但没有错误")
				}
				return
			}

			if err != nil {
				t.Fatalf("解析选项失败: %v", err)
			}

			// 检查结果
			if err := tc.checkResult(ctx); err != nil {
				t.Errorf("结果验证失败: %v", err)
			}
		})
	}
}

// TestInt64WithLargeNumbers 测试大数值的int64选项
func TestInt64WithLargeNumbers(t *testing.T) {
	// 创建命令管理器
	manager := cmd.NewManager()

	// 创建包含int64选项的命令
	options := []cmd.Option{
		{
			Name:        "timestamp",
			Shorthand:   "t",
			Description: "时间戳（长整型）",
			Type:        cmd.OptionTypeInt64,
			Required:    true,
		},
	}

	// 创建测试命令
	testCmd, err := cmd.NewBaseCommand(
		"test-large-int64",
		"测试大数值int64选项",
		"[--timestamp <timestamp>]",
		options,
	)
	if err != nil {
		t.Fatalf("创建命令失败: %v", err)
	}

	// 注册命令
	if err := manager.Register(testCmd); err != nil {
		t.Fatalf("注册命令失败: %v", err)
	}

	// 测试大数值
	largeNumber := "9223372036854775807" // int64的最大值
	args := []string{"--timestamp", largeNumber}

	ctx, err := testCmd.ParseOptions(args)
	if err != nil {
		t.Fatalf("解析选项失败: %v", err)
	}

	// 检查结果
	timestamp, exists := ctx.Options["timestamp"]
	if !exists {
		t.Fatalf("timestamp选项不存在")
	}

	if ts, ok := timestamp.(int64); !ok || ts != 9223372036854775807 {
		t.Errorf("timestamp值错误，期望: 9223372036854775807，实际: %v", timestamp)
	}
}

// TestInt64WithIntValues 测试当传入int值但定义为int64类型时的行为
func TestInt64WithIntValues(t *testing.T) {
	// 创建命令管理器
	manager := cmd.NewManager()

	// 创建包含int64选项的命令
	options := []cmd.Option{
		{
			Name:        "user-id",
			Shorthand:   "u",
			Description: "用户ID（长整型）",
			Type:        cmd.OptionTypeInt64,
			Required:    true,
		},
		{
			Name:        "count",
			Shorthand:   "c",
			Description: "数量（长整型）",
			Type:        cmd.OptionTypeInt64,
			Required:    false,
			Default:     int64(10),
		},
	}

	// 创建测试命令
	testCmd, err := cmd.NewBaseCommand(
		"test-int64-with-int",
		"测试int值但定义为int64类型",
		"[--user-id <id>] [--count <num>]",
		options,
	)
	if err != nil {
		t.Fatalf("创建命令失败: %v", err)
	}

	// 注册命令
	if err := manager.Register(testCmd); err != nil {
		t.Fatalf("注册命令失败: %v", err)
	}

	// 测试用例：传入int范围内的值
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		checkResult func(*cmd.CommandContext) error
	}{
		{
			name:        "传入int范围内的正值",
			args:        []string{"--user-id", "12345", "--count", "100"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				// 检查user-id
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != 12345 {
					return fmt.Errorf("user-id值错误，期望: 12345，实际: %v (类型: %T)", userID, userID)
				}

				// 检查count
				count, exists := ctx.Options["count"]
				if !exists {
					return fmt.Errorf("count选项不存在")
				}
				if c, ok := count.(int64); !ok || c != 100 {
					return fmt.Errorf("count值错误，期望: 100，实际: %v (类型: %T)", count, count)
				}

				return nil
			},
		},
		{
			name:        "传入int范围内的负值",
			args:        []string{"--user-id", "-12345", "--count", "-50"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				// 检查user-id
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != -12345 {
					return fmt.Errorf("user-id值错误，期望: -12345，实际: %v (类型: %T)", userID, userID)
				}

				// 检查count
				count, exists := ctx.Options["count"]
				if !exists {
					return fmt.Errorf("count选项不存在")
				}
				if c, ok := count.(int64); !ok || c != -50 {
					return fmt.Errorf("count值错误，期望: -50，实际: %v (类型: %T)", count, count)
				}

				return nil
			},
		},
		{
			name:        "传入int最大值",
			args:        []string{"--user-id", "2147483647", "--count", "2147483647"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				// 检查user-id
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != 2147483647 {
					return fmt.Errorf("user-id值错误，期望: 2147483647，实际: %v (类型: %T)", userID, userID)
				}

				// 检查count
				count, exists := ctx.Options["count"]
				if !exists {
					return fmt.Errorf("count选项不存在")
				}
				if c, ok := count.(int64); !ok || c != 2147483647 {
					return fmt.Errorf("count值错误，期望: 2147483647，实际: %v (类型: %T)", count, count)
				}

				return nil
			},
		},
		{
			name:        "传入int最小值",
			args:        []string{"--user-id", "-2147483648", "--count", "-2147483648"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				// 检查user-id
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != -2147483648 {
					return fmt.Errorf("user-id值错误，期望: -2147483648，实际: %v (类型: %T)", userID, userID)
				}

				// 检查count
				count, exists := ctx.Options["count"]
				if !exists {
					return fmt.Errorf("count选项不存在")
				}
				if c, ok := count.(int64); !ok || c != -2147483648 {
					return fmt.Errorf("count值错误，期望: -2147483648，实际: %v (类型: %T)", count, count)
				}

				return nil
			},
		},
		{
			name:        "传入超出int范围但仍在int64范围内的值",
			args:        []string{"--user-id", "9223372036854775807", "--count", "9223372036854775806"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				// 检查user-id
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != 9223372036854775807 {
					return fmt.Errorf("user-id值错误，期望: 9223372036854775807，实际: %v (类型: %T)", userID, userID)
				}

				// 检查count
				count, exists := ctx.Options["count"]
				if !exists {
					return fmt.Errorf("count选项不存在")
				}
				if c, ok := count.(int64); !ok || c != 9223372036854775806 {
					return fmt.Errorf("count值错误，期望: 9223372036854775806，实际: %v (类型: %T)", count, count)
				}

				return nil
			},
		},
	}

	// 运行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 解析选项
			ctx, err := testCmd.ParseOptions(tc.args)

			if tc.expectError {
				if err == nil {
					t.Errorf("期望出错，但没有错误")
				}
				return
			}

			if err != nil {
				t.Fatalf("解析选项失败: %v", err)
			}

			// 检查结果
			if err := tc.checkResult(ctx); err != nil {
				t.Errorf("结果验证失败: %v", err)
			}
		})
	}
}

// TestExtraCommandLineOptions 测试传入多余命令行选项时的行为
func TestExtraCommandLineOptions(t *testing.T) {
	// 创建命令管理器
	manager := cmd.NewManager()

	// 创建只包含一个选项的命令
	options := []cmd.Option{
		{
			Name:        "user-id",
			Shorthand:   "u",
			Description: "用户ID（长整型）",
			Type:        cmd.OptionTypeInt64,
			Required:    true,
		},
	}

	// 创建测试命令
	testCmd, err := cmd.NewBaseCommand(
		"test-extra-options",
		"测试多余命令行选项",
		"[--user-id <id>]",
		options,
	)
	if err != nil {
		t.Fatalf("创建命令失败: %v", err)
	}

	// 注册命令
	if err := manager.Register(testCmd); err != nil {
		t.Fatalf("注册命令失败: %v", err)
	}

	// 测试用例
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		checkResult func(*cmd.CommandContext) error
	}{
		{
			name:        "传入未定义的选项（在定义的选项之前）",
			args:        []string{"--unknown-option", "123", "--user-id", "456"},
			expectError: true, // 应该出错，因为unknown-option未定义
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil // 这个测试用例期望出错，所以这里不会执行
			},
		},
		{
			name:        "传入未定义的选项（在定义的选项之后）",
			args:        []string{"--user-id", "456", "--unknown-option", "123"},
			expectError: true, // 应该出错，因为unknown-option未定义
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil // 这个测试用例期望出错，所以这里不会执行
			},
		},
		{
			name:        "传入未定义的短选项（在定义的选项之前）",
			args:        []string{"-x", "123", "--user-id", "456"},
			expectError: true, // 应该出错，因为-x未定义
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil // 这个测试用例期望出错，所以这里不会执行
			},
		},
		{
			name:        "传入未定义的短选项（在定义的选项之后）",
			args:        []string{"--user-id", "456", "-x", "123"},
			expectError: true, // 应该出错，因为-x未定义
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil // 这个测试用例期望出错，所以这里不会执行
			},
		},
		{
			name:        "传入未定义的选项但没有值",
			args:        []string{"--user-id", "456", "--unknown-flag"},
			expectError: true, // 应该出错，因为unknown-flag未定义
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil // 这个测试用例期望出错，所以这里不会执行
			},
		},
		{
			name:        "传入未定义的短选项但没有值",
			args:        []string{"--user-id", "456", "-f"},
			expectError: true, // 应该出错，因为-f未定义
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil // 这个测试用例期望出错，所以这里不会执行
			},
		},
		{
			name:        "正常情况：只传入定义的选项",
			args:        []string{"--user-id", "456"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != 456 {
					return fmt.Errorf("user-id值错误，期望: 456，实际: %v", userID)
				}
				return nil
			},
		},
		{
			name:        "传入位置参数",
			args:        []string{"--user-id", "456", "positional1", "positional2"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				// 检查选项
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != 456 {
					return fmt.Errorf("user-id值错误，期望: 456，实际: %v", userID)
				}

				// 检查位置参数
				if len(ctx.Args) != 2 {
					return fmt.Errorf("位置参数数量错误，期望: 2，实际: %d", len(ctx.Args))
				}
				if ctx.Args[0] != "positional1" || ctx.Args[1] != "positional2" {
					return fmt.Errorf("位置参数值错误，期望: [positional1 positional2]，实际: %v", ctx.Args)
				}

				return nil
			},
		},
	}

	// 运行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 解析选项
			ctx, err := testCmd.ParseOptions(tc.args)

			if tc.expectError {
				if err == nil {
					t.Errorf("期望出错，但没有错误")
				} else {
					t.Logf("预期的错误: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("解析选项失败: %v", err)
			}

			// 检查结果
			if err := tc.checkResult(ctx); err != nil {
				t.Errorf("结果验证失败: %v", err)
			}
		})
	}
}

// TestMultipleOptionsWithExtra 测试多个选项时传入多余选项的行为
func TestMultipleOptionsWithExtra(t *testing.T) {
	// 创建命令管理器
	manager := cmd.NewManager()

	// 创建包含多个选项的命令
	options := []cmd.Option{
		{
			Name:        "user-id",
			Shorthand:   "u",
			Description: "用户ID（长整型）",
			Type:        cmd.OptionTypeInt64,
			Required:    true,
		},
		{
			Name:        "limit",
			Shorthand:   "l",
			Description: "限制数量（长整型）",
			Type:        cmd.OptionTypeInt64,
			Required:    false,
			Default:     int64(100),
		},
	}

	// 创建测试命令
	testCmd, err := cmd.NewBaseCommand(
		"test-multiple-options",
		"测试多个选项时传入多余选项",
		"[--user-id <id>] [--limit <num>]",
		options,
	)
	if err != nil {
		t.Fatalf("创建命令失败: %v", err)
	}

	// 注册命令
	if err := manager.Register(testCmd); err != nil {
		t.Fatalf("注册命令失败: %v", err)
	}

	// 测试用例
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		checkResult func(*cmd.CommandContext) error
	}{
		{
			name:        "在开头传入未定义的选项",
			args:        []string{"--unknown", "123", "--user-id", "456", "--limit", "50"},
			expectError: true,
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil
			},
		},
		{
			name:        "在中间传入未定义的选项",
			args:        []string{"--user-id", "456", "--unknown", "123", "--limit", "50"},
			expectError: true,
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil
			},
		},
		{
			name:        "在结尾传入未定义的选项",
			args:        []string{"--user-id", "456", "--limit", "50", "--unknown", "123"},
			expectError: true,
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil
			},
		},
		{
			name:        "传入多个未定义的选项",
			args:        []string{"--unknown1", "123", "--user-id", "456", "--unknown2", "789", "--limit", "50"},
			expectError: true,
			checkResult: func(ctx *cmd.CommandContext) error {
				return nil
			},
		},
		{
			name:        "正常情况：只传入定义的选项",
			args:        []string{"--user-id", "456", "--limit", "50"},
			expectError: false,
			checkResult: func(ctx *cmd.CommandContext) error {
				// 检查user-id
				userID, exists := ctx.Options["user-id"]
				if !exists {
					return fmt.Errorf("user-id选项不存在")
				}
				if id, ok := userID.(int64); !ok || id != 456 {
					return fmt.Errorf("user-id值错误，期望: 456，实际: %v", userID)
				}

				// 检查limit
				limit, exists := ctx.Options["limit"]
				if !exists {
					return fmt.Errorf("limit选项不存在")
				}
				if l, ok := limit.(int64); !ok || l != 50 {
					return fmt.Errorf("limit值错误，期望: 50，实际: %v", limit)
				}

				return nil
			},
		},
	}

	// 运行测试用例
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 解析选项
			ctx, err := testCmd.ParseOptions(tc.args)

			if tc.expectError {
				if err == nil {
					t.Errorf("期望出错，但没有错误")
				} else {
					t.Logf("预期的错误: %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("解析选项失败: %v", err)
			}

			// 检查结果
			if err := tc.checkResult(ctx); err != nil {
				t.Errorf("结果验证失败: %v", err)
			}
		})
	}
}
