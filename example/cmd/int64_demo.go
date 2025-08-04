package main

import (
	"fmt"

	"github.com/stones-hub/taurus-pro-common/pkg/cmd"
)

// Int64DemoCommand 演示int值在int64选项中的行为
type Int64DemoCommand struct {
	*cmd.BaseCommand
}

// NewInt64DemoCommand 创建int64演示命令
func NewInt64DemoCommand() (*Int64DemoCommand, error) {
	options := []cmd.Option{
		{
			Name:        "small-number",
			Shorthand:   "s",
			Description: "小数值（在int范围内）",
			Type:        cmd.OptionTypeInt64,
			Required:    true,
		},
		{
			Name:        "large-number",
			Shorthand:   "l",
			Description: "大数值（超出int范围）",
			Type:        cmd.OptionTypeInt64,
			Required:    true,
		},
		{
			Name:        "int-max",
			Shorthand:   "m",
			Description: "int最大值",
			Type:        cmd.OptionTypeInt64,
			Required:    false,
			Default:     int64(2147483647),
		},
	}

	baseCmd, err := cmd.NewBaseCommand(
		"int64-demo",
		"演示int值在int64选项中的行为",
		"[--small-number <num>] [--large-number <num>] [--int-max <num>]",
		options,
	)
	if err != nil {
		return nil, err
	}

	return &Int64DemoCommand{BaseCommand: baseCmd}, nil
}

// Run 实现具体的命令逻辑
func (c *Int64DemoCommand) Run(args []string) error {
	// 解析选项
	ctx, err := c.ParseOptions(args)
	if err != nil {
		return err
	}

	// 获取选项值
	smallNum := ctx.Options["small-number"].(int64)
	largeNum := ctx.Options["large-number"].(int64)
	intMax := ctx.Options["int-max"].(int64)

	// 显示解析结果和类型信息
	fmt.Printf("=== int值在int64选项中的行为演示 ===\n")
	fmt.Printf("小数值: %d (类型: int64)\n", smallNum)
	fmt.Printf("大数值: %d (类型: int64)\n", largeNum)
	fmt.Printf("int最大值: %d (类型: int64)\n", intMax)

	// 显示数值范围信息
	fmt.Printf("\n=== 数值范围信息 ===\n")
	fmt.Printf("Go int 范围: -2147483648 到 2147483647 (32位系统)\n")
	fmt.Printf("Go int64 范围: -9223372036854775808 到 9223372036854775807\n")

	// 检查数值是否在int范围内
	if smallNum >= -2147483648 && smallNum <= 2147483647 {
		fmt.Printf("✓ 小数值 %d 在int范围内\n", smallNum)
	} else {
		fmt.Printf("✗ 小数值 %d 超出int范围\n", smallNum)
	}

	if largeNum >= -2147483648 && largeNum <= 2147483647 {
		fmt.Printf("✓ 大数值 %d 在int范围内\n", largeNum)
	} else {
		fmt.Printf("✗ 大数值 %d 超出int范围\n", largeNum)
	}

	if intMax >= -2147483648 && intMax <= 2147483647 {
		fmt.Printf("✓ int最大值 %d 在int范围内\n", intMax)
	} else {
		fmt.Printf("✗ int最大值 %d 超出int范围\n", intMax)
	}

	return nil
}

// 演示函数，展示如何使用
func demonstrateInt64Behavior() {
	fmt.Println("=== int64选项行为演示 ===")

	// 创建命令
	demoCmd, err := NewInt64DemoCommand()
	if err != nil {
		fmt.Printf("创建命令失败: %v\n", err)
		return
	}

	// 测试用例1：传入int范围内的值
	fmt.Println("测试1: 传入int范围内的值")
	args1 := []string{"--small-number", "12345", "--large-number", "2147483647"}
	ctx1, err := demoCmd.ParseOptions(args1)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
	} else {
		small := ctx1.Options["small-number"].(int64)
		large := ctx1.Options["large-number"].(int64)
		fmt.Printf("解析成功: small=%d, large=%d\n", small, large)
	}

	// 测试用例2：传入超出int范围的值
	fmt.Println("\n测试2: 传入超出int范围的值")
	args2 := []string{"--small-number", "12345", "--large-number", "9223372036854775807"}
	ctx2, err := demoCmd.ParseOptions(args2)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
	} else {
		small := ctx2.Options["small-number"].(int64)
		large := ctx2.Options["large-number"].(int64)
		fmt.Printf("解析成功: small=%d, large=%d\n", small, large)
	}

	// 测试用例3：使用短选项
	fmt.Println("\n测试3: 使用短选项")
	args3 := []string{"-s", "100", "-l", "5000000000"}
	ctx3, err := demoCmd.ParseOptions(args3)
	if err != nil {
		fmt.Printf("解析失败: %v\n", err)
	} else {
		small := ctx3.Options["small-number"].(int64)
		large := ctx3.Options["large-number"].(int64)
		fmt.Printf("解析成功: small=%d, large=%d\n", small, large)
	}
}

// 如果需要在main函数中调用演示，可以取消注释以下代码：
// func main() {
// 	// 创建命令管理器
// 	manager := cmd.NewManager()
//
// 	// 创建并注册int64演示命令
// 	demoCmd, err := NewInt64DemoCommand()
// 	if err != nil {
// 		fmt.Printf("创建命令失败: %v\n", err)
// 		os.Exit(1)
// 	}
//
// 	if err := manager.Register(demoCmd); err != nil {
// 		fmt.Printf("注册命令失败: %v\n", err)
// 		os.Exit(1)
// 	}
//
// 	// 运行命令管理器
// 	if err := manager.Run(); err != nil {
// 		fmt.Printf("命令执行失败: %v\n", err)
// 		os.Exit(1)
// 	}
// }
