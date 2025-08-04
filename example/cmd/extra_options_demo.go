package main

import (
	"fmt"
	"os"

	"github.com/stones-hub/taurus-pro-common/pkg/cmd"
)

// ExtraOptionsDemoCommand 演示传入多余命令行选项时的行为
type ExtraOptionsDemoCommand struct {
	*cmd.BaseCommand
}

// NewExtraOptionsDemoCommand 创建多余选项演示命令
func NewExtraOptionsDemoCommand() (*ExtraOptionsDemoCommand, error) {
	options := []cmd.Option{
		{
			Name:        "user-id",
			Shorthand:   "u",
			Description: "用户ID（长整型）",
			Type:        cmd.OptionTypeInt64,
			Required:    true,
		},
	}

	baseCmd, err := cmd.NewBaseCommand(
		"extra-options-demo",
		"演示传入多余命令行选项时的行为",
		"[--user-id <id>]",
		options,
	)
	if err != nil {
		return nil, err
	}

	return &ExtraOptionsDemoCommand{BaseCommand: baseCmd}, nil
}

// Run 实现具体的命令逻辑
func (c *ExtraOptionsDemoCommand) Run(args []string) error {
	// 解析选项
	ctx, err := c.ParseOptions(args)
	if err != nil {
		return err
	}

	// 获取选项值
	userID := ctx.Options["user-id"].(int64)

	// 显示解析结果
	fmt.Printf("=== 多余命令行选项演示 ===\n")
	fmt.Printf("用户ID: %d\n", userID)
	fmt.Printf("位置参数: %v\n", ctx.Args)

	return nil
}

// 演示函数，展示各种情况
func demonstrateExtraOptionsBehavior() {
	fmt.Println("=== 多余命令行选项行为演示 ===")

	// 创建命令
	demoCmd, err := NewExtraOptionsDemoCommand()
	if err != nil {
		fmt.Printf("创建命令失败: %v\n", err)
		return
	}

	// 测试用例1：正常情况
	fmt.Println("测试1: 正常情况 - 只传入定义的选项")
	args1 := []string{"--user-id", "12345"}
	ctx1, err := demoCmd.ParseOptions(args1)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
	} else {
		userID := ctx1.Options["user-id"].(int64)
		fmt.Printf("解析成功: userID=%d\n", userID)
	}

	// 测试用例2：传入位置参数
	fmt.Println("\n测试2: 传入位置参数")
	args2 := []string{"--user-id", "12345", "file1.txt", "file2.txt"}
	ctx2, err := demoCmd.ParseOptions(args2)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
	} else {
		userID := ctx2.Options["user-id"].(int64)
		fmt.Printf("解析成功: userID=%d, 位置参数=%v\n", userID, ctx2.Args)
	}

	// 测试用例3：传入未定义的选项（会失败）
	fmt.Println("\n测试3: 传入未定义的选项（在定义的选项之前）")
	fmt.Println("命令行: --unknown-option 123 --user-id 456")
	fmt.Println("预期结果: 解析失败，显示错误信息")

	// 测试用例4：传入未定义的选项（会失败）
	fmt.Println("\n测试4: 传入未定义的选项（在定义的选项之后）")
	fmt.Println("命令行: --user-id 456 --unknown-option 123")
	fmt.Println("预期结果: 解析失败，显示错误信息")

	// 测试用例5：传入未定义的短选项（会失败）
	fmt.Println("\n测试5: 传入未定义的短选项")
	fmt.Println("命令行: --user-id 456 -x 123")
	fmt.Println("预期结果: 解析失败，显示错误信息")

	// 测试用例6：传入未定义的布尔选项（会失败）
	fmt.Println("\n测试6: 传入未定义的布尔选项")
	fmt.Println("命令行: --user-id 456 --verbose")
	fmt.Println("预期结果: 解析失败，显示错误信息")
}

// 实际运行演示的命令
func runExtraOptionsDemo() {
	// 创建命令管理器
	manager := cmd.NewManager()

	// 创建并注册多余选项演示命令
	demoCmd, err := NewExtraOptionsDemoCommand()
	if err != nil {
		fmt.Printf("创建命令失败: %v\n", err)
		os.Exit(1)
	}

	if err := manager.Register(demoCmd); err != nil {
		fmt.Printf("注册命令失败: %v\n", err)
		os.Exit(1)
	}

	// 运行命令管理器
	if err := manager.Run(); err != nil {
		fmt.Printf("命令执行失败: %v\n", err)
		os.Exit(1)
	}
}

// 使用说明
func printUsageExamples() {
	fmt.Println("=== 使用示例 ===")
	fmt.Println()
	fmt.Println("正常使用:")
	fmt.Println("  ./program extra-options-demo --user-id 12345")
	fmt.Println("  ./program extra-options-demo --user-id 12345 file1.txt file2.txt")
	fmt.Println()
	fmt.Println("错误使用（会失败）:")
	fmt.Println("  ./program extra-options-demo --unknown-option 123 --user-id 456")
	fmt.Println("  ./program extra-options-demo --user-id 456 --unknown-option 123")
	fmt.Println("  ./program extra-options-demo --user-id 456 -x 123")
	fmt.Println("  ./program extra-options-demo --user-id 456 --verbose")
	fmt.Println()
	fmt.Println("错误信息示例:")
	fmt.Println("  flag provided but not defined: -unknown-option")
	fmt.Println("  Usage of extra-options-demo:")
	fmt.Println("    -u int")
	fmt.Println("          用户ID（长整型）")
	fmt.Println("    -user-id int")
	fmt.Println("          用户ID（长整型）")
}
