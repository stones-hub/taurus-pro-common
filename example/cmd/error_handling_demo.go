package main

import (
	"fmt"

	"github.com/stones-hub/taurus-pro-common/pkg/cmd"
)

// ErrorHandlingDemoCommand 演示错误处理机制
type ErrorHandlingDemoCommand struct {
	*cmd.BaseCommand
}

// NewErrorHandlingDemoCommand 创建错误处理演示命令
func NewErrorHandlingDemoCommand() (*ErrorHandlingDemoCommand, error) {
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

	baseCmd, err := cmd.NewBaseCommand(
		"error-handling-demo",
		"演示错误处理机制",
		"[--user-id <id>] [--limit <num>] [--verbose]",
		options,
	)
	if err != nil {
		return nil, err
	}

	return &ErrorHandlingDemoCommand{BaseCommand: baseCmd}, nil
}

// Run 实现具体的命令逻辑
func (c *ErrorHandlingDemoCommand) Run(args []string) error {
	// 解析选项
	ctx, err := c.ParseOptions(args)
	if err != nil {
		return err
	}

	// 获取选项值
	userID := ctx.Options["user-id"].(int64)
	limit := ctx.Options["limit"].(int64)
	verbose := ctx.Options["verbose"].(bool)

	// 显示解析结果
	fmt.Printf("=== 错误处理演示 ===\n")
	fmt.Printf("用户ID: %d\n", userID)
	fmt.Printf("限制数量: %d\n", limit)
	fmt.Printf("详细输出: %t\n", verbose)
	fmt.Printf("位置参数: %v\n", ctx.Args)

	return nil
}

// 演示函数，展示各种错误情况
func demonstrateErrorHandling() {
	fmt.Println("=== 错误处理机制演示 ===\n")

	// 创建命令
	demoCmd, err := NewErrorHandlingDemoCommand()
	if err != nil {
		fmt.Printf("创建命令失败: %v\n", err)
		return
	}

	// 测试用例1：正常情况
	fmt.Println("测试1: 正常情况 - 只传入定义的选项")
	args1 := []string{"--user-id", "12345", "--limit", "50", "--verbose"}
	ctx1, err := demoCmd.ParseOptions(args1)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
	} else {
		userID := ctx1.Options["user-id"].(int64)
		limit := ctx1.Options["limit"].(int64)
		verbose := ctx1.Options["verbose"].(bool)
		fmt.Printf("解析成功: userID=%d, limit=%d, verbose=%t\n", userID, limit, verbose)
	}

	// 测试用例2：传入未定义的选项
	fmt.Println("\n测试2: 传入未定义的选项")
	fmt.Println("命令行: --unknown-option 123 --user-id 456")
	fmt.Println("预期结果: 显示友好的错误信息，包含可用选项列表")

	// 测试用例3：传入未定义的短选项
	fmt.Println("\n测试3: 传入未定义的短选项")
	fmt.Println("命令行: --user-id 456 -x 123")
	fmt.Println("预期结果: 显示友好的错误信息，指出未定义的选项")

	// 测试用例4：传入未定义的布尔选项
	fmt.Println("\n测试4: 传入未定义的布尔选项")
	fmt.Println("命令行: --user-id 456 --debug")
	fmt.Println("预期结果: 显示友好的错误信息，指出未定义的选项")

	// 测试用例5：传入未定义的选项在结尾
	fmt.Println("\n测试5: 传入未定义的选项在结尾")
	fmt.Println("命令行: --user-id 456 --limit 50 --unknown")
	fmt.Println("预期结果: 显示友好的错误信息，指出未定义的选项")
}

// 使用说明
func printErrorHandlingExamples() {
	fmt.Println("=== 错误处理示例 ===")
	fmt.Println()
	fmt.Println("正常使用:")
	fmt.Println("  ./program error-handling-demo --user-id 12345 --limit 50 --verbose")
	fmt.Println()
	fmt.Println("错误使用示例（会显示友好错误信息）:")
	fmt.Println("  ./program error-handling-demo --unknown-option 123 --user-id 456")
	fmt.Println("  ./program error-handling-demo --user-id 456 -x 123")
	fmt.Println("  ./program error-handling-demo --user-id 456 --debug")
	fmt.Println("  ./program error-handling-demo --user-id 456 --limit 50 --unknown")
	fmt.Println()
	fmt.Println("错误信息特点:")
	fmt.Println("  ✓ 明确指出未定义的选项")
	fmt.Println("  ✓ 列出所有可用的选项")
	fmt.Println("  ✓ 显示选项的必填/可选状态")
	fmt.Println("  ✓ 显示默认值信息")
	fmt.Println("  ✓ 提供使用方法和帮助提示")
	fmt.Println()
	fmt.Println("错误信息示例:")
	fmt.Println("  错误: 未定义的选项 '--unknown-option'")
	fmt.Println("")
	fmt.Println("  可用的选项:")
	fmt.Println("    user-id, -u  用户ID（长整型） (必填)")
	fmt.Println("    limit, -l    限制数量（长整型） (默认: 100)")
	fmt.Println("    verbose, -v  详细输出 (默认: false)")
	fmt.Println("")
	fmt.Println("  使用方法: error-handling-demo [--user-id <id>] [--limit <num>] [--verbose]")
	fmt.Println("  运行 'error-handling-demo --help' 查看详细帮助")
}
