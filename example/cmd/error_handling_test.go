package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stones-hub/taurus-pro-common/pkg/cmd"
)

// TestEnhancedErrorHandling 测试增强的错误处理机制
func TestEnhancedErrorHandling(t *testing.T) {
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
		{
			Name:        "verbose",
			Shorthand:   "v",
			Description: "详细输出",
			Type:        cmd.OptionTypeBool,
			Required:    false,
			Default:     false,
		},
	}

	// 创建测试命令
	testCmd, err := cmd.NewBaseCommand(
		"test-error-handling",
		"测试错误处理机制",
		"[--user-id <id>] [--limit <num>] [--verbose]",
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
		checkError  func(error) error
	}{
		{
			name:        "传入未定义的长选项",
			args:        []string{"--unknown-option", "123", "--user-id", "456"},
			expectError: true,
			checkError: func(err error) error {
				errMsg := err.Error()
				if !strings.Contains(errMsg, "错误: 未定义的选项") {
					return fmt.Errorf("错误信息不匹配，期望包含未定义选项信息，实际: %s", errMsg)
				}
				if !strings.Contains(errMsg, "可用的选项:") {
					return fmt.Errorf("错误信息不匹配，期望包含可用选项列表，实际: %s", errMsg)
				}
				if !strings.Contains(errMsg, "user-id, -u") {
					return fmt.Errorf("错误信息不匹配，期望包含user-id选项，实际: %s", errMsg)
				}
				if !strings.Contains(errMsg, "limit, -l") {
					return fmt.Errorf("错误信息不匹配，期望包含limit选项，实际: %s", errMsg)
				}
				if !strings.Contains(errMsg, "verbose, -v") {
					return fmt.Errorf("错误信息不匹配，期望包含verbose选项，实际: %s", errMsg)
				}
				return nil
			},
		},
		{
			name:        "传入未定义的短选项",
			args:        []string{"--user-id", "456", "-x", "123"},
			expectError: true,
			checkError: func(err error) error {
				errMsg := err.Error()
				if !strings.Contains(errMsg, "错误: 未定义的选项 '-x'") {
					return fmt.Errorf("错误信息不匹配，期望包含未定义选项信息，实际: %s", errMsg)
				}
				return nil
			},
		},
		{
			name:        "传入未定义的布尔选项",
			args:        []string{"--user-id", "456", "--debug"},
			expectError: true,
			checkError: func(err error) error {
				errMsg := err.Error()
				if !strings.Contains(errMsg, "错误: 未定义的选项 '--debug'") {
					return fmt.Errorf("错误信息不匹配，期望包含未定义选项信息，实际: %s", errMsg)
				}
				return nil
			},
		},
		{
			name:        "传入未定义的选项在结尾",
			args:        []string{"--user-id", "456", "--limit", "50", "--unknown"},
			expectError: true,
			checkError: func(err error) error {
				errMsg := err.Error()
				if !strings.Contains(errMsg, "错误: 未定义的选项 '--unknown'") {
					return fmt.Errorf("错误信息不匹配，期望包含未定义选项信息，实际: %s", errMsg)
				}
				return nil
			},
		},
		{
			name:        "传入多个未定义的选项",
			args:        []string{"--unknown1", "123", "--user-id", "456", "--unknown2", "789"},
			expectError: true,
			checkError: func(err error) error {
				errMsg := err.Error()
				// 应该报告第一个未定义的选项
				if !strings.Contains(errMsg, "错误: 未定义的选项 '--unknown1'") {
					return fmt.Errorf("错误信息不匹配，期望报告第一个未定义选项，实际: %s", errMsg)
				}
				return nil
			},
		},
		{
			name:        "正常情况：只传入定义的选项",
			args:        []string{"--user-id", "456", "--limit", "50", "--verbose"},
			expectError: false,
			checkError: func(err error) error {
				if err != nil {
					return fmt.Errorf("不应该有错误，但得到了: %v", err)
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
					return
				}

				// 检查错误信息
				if checkErr := tc.checkError(err); checkErr != nil {
					t.Errorf("错误检查失败: %v", checkErr)
				}

				// 打印错误信息以便查看
				t.Logf("错误信息:\n%s", err.Error())
			} else {
				if err != nil {
					t.Fatalf("解析选项失败: %v", err)
				}

				// 验证解析结果
				userID := ctx.Options["user-id"].(int64)
				if userID != 456 {
					t.Errorf("user-id值错误，期望: 456，实际: %v", userID)
				}

				limit := ctx.Options["limit"].(int64)
				if limit != 50 {
					t.Errorf("limit值错误，期望: 50，实际: %v", limit)
				}

				verbose := ctx.Options["verbose"].(bool)
				if !verbose {
					t.Errorf("verbose值错误，期望: true，实际: %v", verbose)
				}
			}
		})
	}
}

// TestErrorHandlingSingleOption 测试单个选项时的错误处理
func TestErrorHandlingSingleOption(t *testing.T) {
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
		"test-single-option",
		"测试单个选项的错误处理",
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
		checkError  func(error) error
	}{
		{
			name:        "传入未定义的选项",
			args:        []string{"--unknown", "123", "--user-id", "456"},
			expectError: true,
			checkError: func(err error) error {
				errMsg := err.Error()
				if !strings.Contains(errMsg, "错误: 未定义的选项 '--unknown'") {
					return fmt.Errorf("错误信息不匹配，期望包含未定义选项信息，实际: %s", errMsg)
				}
				if !strings.Contains(errMsg, "user-id, -u") {
					return fmt.Errorf("错误信息不匹配，期望包含user-id选项，实际: %s", errMsg)
				}
				if !strings.Contains(errMsg, "(必填)") {
					return fmt.Errorf("错误信息不匹配，期望标记必填选项，实际: %s", errMsg)
				}
				return nil
			},
		},
		{
			name:        "传入未定义的选项在结尾",
			args:        []string{"--user-id", "456", "--unknown", "123"},
			expectError: true,
			checkError: func(err error) error {
				errMsg := err.Error()
				if !strings.Contains(errMsg, "错误: 未定义的选项 '--unknown'") {
					return fmt.Errorf("错误信息不匹配，期望包含未定义选项信息，实际: %s", errMsg)
				}
				return nil
			},
		},
		{
			name:        "正常情况",
			args:        []string{"--user-id", "456"},
			expectError: false,
			checkError: func(err error) error {
				if err != nil {
					return fmt.Errorf("不应该有错误，但得到了: %v", err)
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
					return
				}

				// 检查错误信息
				if checkErr := tc.checkError(err); checkErr != nil {
					t.Errorf("错误检查失败: %v", checkErr)
				}

				// 打印错误信息以便查看
				t.Logf("错误信息:\n%s", err.Error())
			} else {
				if err != nil {
					t.Fatalf("解析选项失败: %v", err)
				}

				// 验证解析结果
				userID := ctx.Options["user-id"].(int64)
				if userID != 456 {
					t.Errorf("user-id值错误，期望: 456，实际: %v", userID)
				}
			}
		})
	}
}
