package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/cmd"
)

// ==================== 基础测试 ====================

// TestCommandCreation 测试命令创建
func TestCommandCreation(t *testing.T) {
	t.Run("创建用户命令", func(t *testing.T) {
		userCmd, err := cmd.NewBaseCommand(
			"user",
			"用户管理命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name",
					Shorthand:   "n",
					Description: "用户名",
					Type:        cmd.OptionTypeString,
					Required:    true,
				},
				{
					Name:        "age",
					Shorthand:   "a",
					Description: "年龄",
					Type:        cmd.OptionTypeInt,
					Default:     25,
				},
			},
		)
		if err != nil {
			t.Errorf("创建用户命令失败: %v", err)
		}
		if userCmd.Name() != "user" {
			t.Errorf("命令名称不正确: 期望 'user', 实际 '%s'", userCmd.Name())
		}
	})

	t.Run("创建文件命令", func(t *testing.T) {
		fileCmd, err := cmd.NewBaseCommand(
			"file",
			"文件操作命令",
			"[options] [files...]",
			[]cmd.Option{
				{
					Name:        "source",
					Shorthand:   "s",
					Description: "源文件",
					Type:        cmd.OptionTypeString,
					Required:    true,
				},
				{
					Name:        "force",
					Shorthand:   "f",
					Description: "强制执行",
					Type:        cmd.OptionTypeBool,
					Default:     false,
				},
			},
		)
		if err != nil {
			t.Errorf("创建文件命令失败: %v", err)
		}
		if fileCmd.Name() != "file" {
			t.Errorf("命令名称不正确: 期望 'file', 实际 '%s'", fileCmd.Name())
		}
	})
}

// TestOptionParsing 测试选项解析
func TestOptionParsing(t *testing.T) {
	t.Run("解析字符串选项", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name",
					Description: "名称",
					Type:        cmd.OptionTypeString,
					Default:     "default",
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{"--name", "test_value"})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		value := ctx.Options["name"].(string)
		if value != "test_value" {
			t.Errorf("选项值不正确: 期望 'test_value', 实际 '%s'", value)
		}
	})

	t.Run("解析整数选项", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "count",
					Description: "数量",
					Type:        cmd.OptionTypeInt,
					Default:     10,
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{"--count", "42"})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		value := ctx.Options["count"].(int)
		if value != 42 {
			t.Errorf("选项值不正确: 期望 42, 实际 %d", value)
		}
	})

	t.Run("解析布尔选项", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "verbose",
					Description: "详细输出",
					Type:        cmd.OptionTypeBool,
					Default:     false,
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{"--verbose"})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		value := ctx.Options["verbose"].(bool)
		if !value {
			t.Errorf("选项值不正确: 期望 true, 实际 %v", value)
		}
	})

	t.Run("解析浮点数选项", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "score",
					Description: "评分",
					Type:        cmd.OptionTypeFloat,
					Default:     0.0,
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{"--score", "95.5"})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		value := ctx.Options["score"].(float64)
		if value != 95.5 {
			t.Errorf("选项值不正确: 期望 95.5, 实际 %f", value)
		}
	})
}

// TestRequiredOptions 测试必填选项
func TestRequiredOptions(t *testing.T) {
	t.Run("缺少必填选项", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name",
					Description: "名称（必填）",
					Type:        cmd.OptionTypeString,
					Required:    true,
				},
			},
		)

		_, err := cmd.ParseOptions([]string{})
		if err == nil {
			t.Error("应该返回错误，因为缺少必填选项")
		}
		if !strings.Contains(err.Error(), "必填") {
			t.Errorf("错误信息不正确: %v", err)
		}
	})

	t.Run("提供必填选项", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name",
					Description: "名称（必填）",
					Type:        cmd.OptionTypeString,
					Required:    true,
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{"--name", "test"})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		value := ctx.Options["name"].(string)
		if value != "test" {
			t.Errorf("选项值不正确: 期望 'test', 实际 '%s'", value)
		}
	})
}

// TestDefaultValues 测试默认值
func TestDefaultValues(t *testing.T) {
	t.Run("字符串默认值", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name",
					Description: "名称",
					Type:        cmd.OptionTypeString,
					Default:     "default_name",
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		value := ctx.Options["name"].(string)
		if value != "default_name" {
			t.Errorf("默认值不正确: 期望 'default_name', 实际 '%s'", value)
		}
	})

	t.Run("整数默认值", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "count",
					Description: "数量",
					Type:        cmd.OptionTypeInt,
					Default:     100,
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		value := ctx.Options["count"].(int)
		if value != 100 {
			t.Errorf("默认值不正确: 期望 100, 实际 %d", value)
		}
	})

	t.Run("布尔默认值", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "debug",
					Description: "调试模式",
					Type:        cmd.OptionTypeBool,
					Default:     true,
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		value := ctx.Options["debug"].(bool)
		if !value {
			t.Errorf("默认值不正确: 期望 true, 实际 %v", value)
		}
	})

	t.Run("浮点数默认值", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "ratio",
					Description: "比率",
					Type:        cmd.OptionTypeFloat,
					Default:     3.14,
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		value := ctx.Options["ratio"].(float64)
		if value != 3.14 {
			t.Errorf("默认值不正确: 期望 3.14, 实际 %f", value)
		}
	})
}

// TestShorthandOptions 测试短选项
func TestShorthandOptions(t *testing.T) {
	t.Run("短选项解析", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name",
					Shorthand:   "n",
					Description: "名称",
					Type:        cmd.OptionTypeString,
					Default:     "",
				},
				{
					Name:        "verbose",
					Shorthand:   "v",
					Description: "详细输出",
					Type:        cmd.OptionTypeBool,
					Default:     false,
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{"-n", "test", "-v"})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		name := ctx.Options["name"].(string)
		verbose := ctx.Options["verbose"].(bool)

		if name != "test" {
			t.Errorf("短选项值不正确: 期望 'test', 实际 '%s'", name)
		}
		if !verbose {
			t.Errorf("短选项值不正确: 期望 true, 实际 %v", verbose)
		}
	})
}

// TestPositionalArgs 测试位置参数
func TestPositionalArgs(t *testing.T) {
	t.Run("位置参数解析", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options] [files...]",
			[]cmd.Option{
				{
					Name:        "verbose",
					Description: "详细输出",
					Type:        cmd.OptionTypeBool,
					Default:     false,
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{"--verbose", "file1.txt", "file2.txt", "file3.txt"})
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		verbose := ctx.Options["verbose"].(bool)
		if !verbose {
			t.Errorf("选项值不正确: 期望 true, 实际 %v", verbose)
		}

		expectedArgs := []string{"file1.txt", "file2.txt", "file3.txt"}
		if len(ctx.Args) != len(expectedArgs) {
			t.Errorf("位置参数数量不正确: 期望 %d, 实际 %d", len(expectedArgs), len(ctx.Args))
		}

		for i, arg := range expectedArgs {
			if ctx.Args[i] != arg {
				t.Errorf("位置参数不正确: 期望 '%s', 实际 '%s'", arg, ctx.Args[i])
			}
		}
	})
}

// ==================== 命令管理器测试 ====================

// TestCommandManager 测试命令管理器
func TestCommandManager(t *testing.T) {
	t.Run("注册命令", func(t *testing.T) {
		manager := cmd.NewManager()

		// 创建测试命令
		testCmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{},
		)

		// 注册命令
		err := manager.Register(&TestCommand{BaseCommand: testCmd})
		if err != nil {
			t.Errorf("注册命令失败: %v", err)
		}

		// 验证命令已注册
		cmd, exists := manager.GetCommand("test")
		if !exists {
			t.Error("命令应该存在")
		}
		if cmd.Name() != "test" {
			t.Errorf("命令名称不正确: 期望 'test', 实际 '%s'", cmd.Name())
		}
	})

	t.Run("重复注册命令", func(t *testing.T) {
		manager := cmd.NewManager()

		// 创建测试命令
		testCmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{},
		)

		// 第一次注册
		err := manager.Register(&TestCommand{BaseCommand: testCmd})
		if err != nil {
			t.Errorf("第一次注册失败: %v", err)
		}

		// 第二次注册应该失败
		err = manager.Register(&TestCommand{BaseCommand: testCmd})
		if err == nil {
			t.Error("重复注册应该失败")
		}
		if !strings.Contains(err.Error(), "已存在") {
			t.Errorf("错误信息不正确: %v", err)
		}
	})

	t.Run("获取所有命令", func(t *testing.T) {
		manager := cmd.NewManager()

		// 注册多个命令
		cmd1, _ := cmd.NewBaseCommand("cmd1", "命令1", "[options]", []cmd.Option{})
		cmd2, _ := cmd.NewBaseCommand("cmd2", "命令2", "[options]", []cmd.Option{})
		cmd3, _ := cmd.NewBaseCommand("cmd3", "命令3", "[options]", []cmd.Option{})

		manager.Register(&TestCommand{BaseCommand: cmd1})
		manager.Register(&TestCommand{BaseCommand: cmd2})
		manager.Register(&TestCommand{BaseCommand: cmd3})

		// 获取所有命令
		commands := manager.GetCommands()
		if len(commands) != 3 {
			t.Errorf("命令数量不正确: 期望 3, 实际 %d", len(commands))
		}

		// 验证命令存在
		expectedNames := []string{"cmd1", "cmd2", "cmd3"}
		for _, name := range expectedNames {
			if _, exists := commands[name]; !exists {
				t.Errorf("命令 '%s' 应该存在", name)
			}
		}
	})
}

// ==================== 集成测试 ====================

// TestUserCommand 测试用户命令
func TestUserCommand(t *testing.T) {
	t.Run("用户命令基本功能", func(t *testing.T) {
		// 创建用户命令
		userCmd, _ := cmd.NewBaseCommand(
			"user",
			"用户管理命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name",
					Shorthand:   "n",
					Description: "用户名（必填）",
					Type:        cmd.OptionTypeString,
					Required:    true,
				},
				{
					Name:        "email",
					Shorthand:   "e",
					Description: "邮箱地址",
					Type:        cmd.OptionTypeString,
					Default:     "user@example.com",
				},
				{
					Name:        "age",
					Shorthand:   "a",
					Description: "年龄",
					Type:        cmd.OptionTypeInt,
					Default:     25,
				},
				{
					Name:        "active",
					Shorthand:   "A",
					Description: "是否激活",
					Type:        cmd.OptionTypeBool,
					Default:     true,
				},
				{
					Name:        "score",
					Shorthand:   "s",
					Description: "用户评分",
					Type:        cmd.OptionTypeFloat,
					Default:     85.5,
				},
				{
					Name:        "verbose",
					Shorthand:   "v",
					Description: "详细输出",
					Type:        cmd.OptionTypeBool,
					Default:     false,
				},
			},
		)

		// 创建用户命令实例
		userCommand := &UserCommand{BaseCommand: userCmd}

		// 测试参数解析
		args := []string{
			"--name", "张三",
			"--email", "zhangsan@example.com",
			"--age", "30",
			"--score", "92.5",
			"--verbose",
		}

		ctx, err := userCommand.ParseOptions(args)
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		// 验证选项值
		if ctx.Options["name"].(string) != "张三" {
			t.Errorf("姓名不正确: %v", ctx.Options["name"])
		}
		if ctx.Options["email"].(string) != "zhangsan@example.com" {
			t.Errorf("邮箱不正确: %v", ctx.Options["email"])
		}
		if ctx.Options["age"].(int) != 30 {
			t.Errorf("年龄不正确: %v", ctx.Options["age"])
		}
		if ctx.Options["score"].(float64) != 92.5 {
			t.Errorf("评分不正确: %v", ctx.Options["score"])
		}
		if !ctx.Options["verbose"].(bool) {
			t.Errorf("详细模式不正确: %v", ctx.Options["verbose"])
		}
		if !ctx.Options["active"].(bool) {
			t.Errorf("激活状态不正确: %v", ctx.Options["active"])
		}
	})

	t.Run("用户命令短选项", func(t *testing.T) {
		userCmd, _ := cmd.NewBaseCommand(
			"user",
			"用户管理命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name",
					Shorthand:   "n",
					Description: "用户名（必填）",
					Type:        cmd.OptionTypeString,
					Required:    true,
				},
				{
					Name:        "email",
					Shorthand:   "e",
					Description: "邮箱地址",
					Type:        cmd.OptionTypeString,
					Default:     "user@example.com",
				},
			},
		)

		userCommand := &UserCommand{BaseCommand: userCmd}

		// 使用短选项
		args := []string{"-n", "李四", "-e", "lisi@example.com"}

		ctx, err := userCommand.ParseOptions(args)
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		if ctx.Options["name"].(string) != "李四" {
			t.Errorf("姓名不正确: %v", ctx.Options["name"])
		}
		if ctx.Options["email"].(string) != "lisi@example.com" {
			t.Errorf("邮箱不正确: %v", ctx.Options["email"])
		}
	})
}

// TestFileCommand 测试文件命令
func TestFileCommand(t *testing.T) {
	t.Run("文件命令基本功能", func(t *testing.T) {
		fileCmd, _ := cmd.NewBaseCommand(
			"file",
			"文件操作命令",
			"[options] [files...]",
			[]cmd.Option{
				{
					Name:        "source",
					Shorthand:   "s",
					Description: "源文件路径（必填）",
					Type:        cmd.OptionTypeString,
					Required:    true,
				},
				{
					Name:        "destination",
					Shorthand:   "d",
					Description: "目标路径",
					Type:        cmd.OptionTypeString,
					Default:     "./",
				},
				{
					Name:        "force",
					Shorthand:   "f",
					Description: "强制执行",
					Type:        cmd.OptionTypeBool,
					Default:     false,
				},
				{
					Name:        "recursive",
					Shorthand:   "r",
					Description: "递归处理",
					Type:        cmd.OptionTypeBool,
					Default:     false,
				},
			},
		)

		fileCommand := &FileCommand{BaseCommand: fileCmd}

		// 测试参数解析
		args := []string{
			"--source", "/path/to/source",
			"--destination", "/path/to/dest",
			"--force",
			"--recursive",
			"file1.txt",
			"file2.txt",
		}

		ctx, err := fileCommand.ParseOptions(args)
		if err != nil {
			t.Errorf("解析选项失败: %v", err)
		}

		// 验证选项值
		if ctx.Options["source"].(string) != "/path/to/source" {
			t.Errorf("源路径不正确: %v", ctx.Options["source"])
		}
		if ctx.Options["destination"].(string) != "/path/to/dest" {
			t.Errorf("目标路径不正确: %v", ctx.Options["destination"])
		}
		if !ctx.Options["force"].(bool) {
			t.Errorf("强制执行不正确: %v", ctx.Options["force"])
		}
		if !ctx.Options["recursive"].(bool) {
			t.Errorf("递归处理不正确: %v", ctx.Options["recursive"])
		}

		// 验证位置参数
		expectedArgs := []string{"file1.txt", "file2.txt"}
		if len(ctx.Args) != len(expectedArgs) {
			t.Errorf("位置参数数量不正确: 期望 %d, 实际 %d", len(expectedArgs), len(ctx.Args))
		}
		for i, arg := range expectedArgs {
			if ctx.Args[i] != arg {
				t.Errorf("位置参数不正确: 期望 '%s', 实际 '%s'", arg, ctx.Args[i])
			}
		}
	})
}

// ==================== 边界条件测试 ====================

// TestEdgeCases 测试边界条件
func TestEdgeCases(t *testing.T) {
	t.Run("空参数", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{},
		)

		ctx, err := cmd.ParseOptions([]string{})
		if err != nil {
			t.Errorf("解析空参数失败: %v", err)
		}

		if len(ctx.Args) != 0 {
			t.Errorf("空参数解析结果不正确: %v", ctx.Args)
		}
	})

	t.Run("特殊字符选项名", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "opt-with-dash",
					Description: "带连字符的选项",
					Type:        cmd.OptionTypeString,
					Default:     "",
				},
				{
					Name:        "opt_with_underscore",
					Description: "带下划线的选项",
					Type:        cmd.OptionTypeString,
					Default:     "",
				},
			},
		)

		args := []string{"--opt-with-dash", "value1", "--opt_with_underscore", "value2"}
		ctx, err := cmd.ParseOptions(args)
		if err != nil {
			t.Errorf("解析特殊字符选项失败: %v", err)
		}

		if ctx.Options["opt-with-dash"].(string) != "value1" {
			t.Errorf("带连字符选项值不正确: %v", ctx.Options["opt-with-dash"])
		}
		if ctx.Options["opt_with_underscore"].(string) != "value2" {
			t.Errorf("带下划线选项值不正确: %v", ctx.Options["opt_with_underscore"])
		}
	})

	t.Run("Unicode支持", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"测试",
			"测试命令",
			"[选项]",
			[]cmd.Option{
				{
					Name:        "选项名",
					Description: "中文选项",
					Type:        cmd.OptionTypeString,
					Default:     "",
				},
			},
		)

		args := []string{"--选项名", "中文值"}
		ctx, err := cmd.ParseOptions(args)
		if err != nil {
			t.Errorf("解析Unicode选项失败: %v", err)
		}

		if ctx.Options["选项名"].(string) != "中文值" {
			t.Errorf("Unicode选项值不正确: %v", ctx.Options["选项名"])
		}
	})

	t.Run("数值边界", func(t *testing.T) {
		cmd, _ := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "max_int",
					Description: "最大整数",
					Type:        cmd.OptionTypeInt,
					Default:     2147483647,
				},
				{
					Name:        "max_float",
					Description: "最大浮点数",
					Type:        cmd.OptionTypeFloat,
					Default:     1.7976931348623157e+308,
				},
			},
		)

		ctx, err := cmd.ParseOptions([]string{})
		if err != nil {
			t.Errorf("解析数值边界失败: %v", err)
		}

		if ctx.Options["max_int"].(int) != 2147483647 {
			t.Errorf("最大整数值不正确: %v", ctx.Options["max_int"])
		}
		if ctx.Options["max_float"].(float64) != 1.7976931348623157e+308 {
			t.Errorf("最大浮点数值不正确: %v", ctx.Options["max_float"])
		}
	})
}

// ==================== 性能测试 ====================

// TestPerformance 测试性能
func TestPerformance(t *testing.T) {
	t.Run("大量选项解析", func(t *testing.T) {
		// 创建大量选项
		options := make([]cmd.Option, 100)
		for i := 0; i < 100; i++ {
			options[i] = cmd.Option{
				Name:        fmt.Sprintf("option%d", i),
				Description: fmt.Sprintf("选项 %d", i),
				Type:        cmd.OptionTypeString,
				Default:     fmt.Sprintf("value%d", i),
			}
		}

		cmd, _ := cmd.NewBaseCommand("test", "测试命令", "[options]", options)

		// 构建参数
		args := make([]string, 200)
		for i := 0; i < 100; i++ {
			args[i*2] = fmt.Sprintf("--option%d", i)
			args[i*2+1] = fmt.Sprintf("test_value_%d", i)
		}

		start := time.Now()
		ctx, err := cmd.ParseOptions(args)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("解析大量选项失败: %v", err)
		}

		if len(ctx.Options) != 100 {
			t.Errorf("选项数量不正确: 期望 100, 实际 %d", len(ctx.Options))
		}

		// 性能检查：应该在合理时间内完成
		if duration > time.Second {
			t.Errorf("解析性能太慢: %v", duration)
		}
	})

	t.Run("命令注册性能", func(t *testing.T) {
		manager := cmd.NewManager()

		start := time.Now()
		for i := 0; i < 1000; i++ {
			cmd, _ := cmd.NewBaseCommand(
				fmt.Sprintf("cmd%d", i),
				fmt.Sprintf("命令 %d", i),
				"[options]",
				[]cmd.Option{},
			)
			manager.Register(&TestCommand{BaseCommand: cmd})
		}
		duration := time.Since(start)

		// 性能检查：应该在合理时间内完成
		if duration > time.Second {
			t.Errorf("命令注册性能太慢: %v", duration)
		}

		commands := manager.GetCommands()
		if len(commands) != 1000 {
			t.Errorf("命令数量不正确: 期望 1000, 实际 %d", len(commands))
		}
	})
}

// ==================== 并发测试 ====================

// TestConcurrency 测试并发安全
func TestConcurrency(t *testing.T) {
	t.Run("并发命令注册", func(t *testing.T) {
		manager := cmd.NewManager()
		done := make(chan bool, 10)

		// 启动多个goroutine并发注册命令
		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()
				for j := 0; j < 100; j++ {
					cmd, _ := cmd.NewBaseCommand(
						fmt.Sprintf("cmd%d_%d", id, j),
						fmt.Sprintf("命令 %d_%d", id, j),
						"[options]",
						[]cmd.Option{},
					)
					manager.Register(&TestCommand{BaseCommand: cmd})
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 验证结果
		commands := manager.GetCommands()
		if len(commands) != 1000 {
			t.Errorf("并发注册后命令数量不正确: 期望 1000, 实际 %d", len(commands))
		}
	})

	t.Run("并发命令访问", func(t *testing.T) {
		manager := cmd.NewManager()

		// 先注册一些命令
		for i := 0; i < 100; i++ {
			cmd, _ := cmd.NewBaseCommand(
				fmt.Sprintf("cmd%d", i),
				fmt.Sprintf("命令 %d", i),
				"[options]",
				[]cmd.Option{},
			)
			manager.Register(&TestCommand{BaseCommand: cmd})
		}

		done := make(chan bool, 10)

		// 启动多个goroutine并发访问命令
		for i := 0; i < 10; i++ {
			go func() {
				defer func() { done <- true }()
				for j := 0; j < 1000; j++ {
					manager.GetCommands()
					manager.GetCommand("cmd0")
				}
			}()
		}

		// 等待所有goroutine完成
		for i := 0; i < 10; i++ {
			<-done
		}

		// 验证数据完整性
		commands := manager.GetCommands()
		if len(commands) != 100 {
			t.Errorf("并发访问后命令数量不正确: 期望 100, 实际 %d", len(commands))
		}
	})
}

// ==================== 错误处理测试 ====================

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	t.Run("无效选项类型", func(t *testing.T) {
		// 这个测试应该在命令创建时就失败，而不是在解析时
		_, err := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "count",
					Description: "数量",
					Type:        cmd.OptionTypeInt,
					Default:     "invalid", // 错误的默认值类型
				},
			},
		)
		if err == nil {
			t.Error("应该返回错误，因为默认值类型不正确")
		}
	})

	t.Run("重复选项名", func(t *testing.T) {
		_, err := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name",
					Description: "名称",
					Type:        cmd.OptionTypeString,
				},
				{
					Name:        "name", // 重复的选项名
					Description: "名称2",
					Type:        cmd.OptionTypeString,
				},
			},
		)

		if err == nil {
			t.Error("应该返回错误，因为选项名重复")
		}
		if !strings.Contains(err.Error(), "重复的选项名") {
			t.Errorf("错误信息不正确: %v", err)
		}
	})

	t.Run("重复短选项名", func(t *testing.T) {
		_, err := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name1",
					Shorthand:   "n",
					Description: "名称1",
					Type:        cmd.OptionTypeString,
				},
				{
					Name:        "name2",
					Shorthand:   "n", // 重复的短选项名
					Description: "名称2",
					Type:        cmd.OptionTypeString,
				},
			},
		)

		if err == nil {
			t.Error("应该返回错误，因为短选项名重复")
		}
		if !strings.Contains(err.Error(), "重复的短选项名") {
			t.Errorf("错误信息不正确: %v", err)
		}
	})

	t.Run("无效短选项名长度", func(t *testing.T) {
		_, err := cmd.NewBaseCommand(
			"test",
			"测试命令",
			"[options]",
			[]cmd.Option{
				{
					Name:        "name",
					Shorthand:   "ab", // 长度不为1
					Description: "名称",
					Type:        cmd.OptionTypeString,
				},
			},
		)

		if err == nil {
			t.Error("应该返回错误，因为短选项名长度不正确")
		}
		if !strings.Contains(err.Error(), "单个字符") {
			t.Errorf("错误信息不正确: %v", err)
		}
	})
}

// ==================== 辅助类型和函数 ====================

// TestCommand 测试命令
type TestCommand struct {
	*cmd.BaseCommand
}

func (c *TestCommand) Run(args []string) error {
	return nil
}

// ==================== 基准测试 ====================

// BenchmarkOptionParsing 基准测试选项解析
func BenchmarkOptionParsing(b *testing.B) {
	cmd, _ := cmd.NewBaseCommand(
		"test",
		"测试命令",
		"[options]",
		[]cmd.Option{
			{
				Name:        "name",
				Description: "名称",
				Type:        cmd.OptionTypeString,
				Default:     "default",
			},
			{
				Name:        "count",
				Description: "数量",
				Type:        cmd.OptionTypeInt,
				Default:     10,
			},
			{
				Name:        "verbose",
				Description: "详细输出",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
		},
	)

	args := []string{"--name", "test_value", "--count", "42", "--verbose"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cmd.ParseOptions(args)
		if err != nil {
			b.Fatalf("解析选项失败: %v", err)
		}
	}
}

// BenchmarkCommandRegistration 基准测试命令注册
func BenchmarkCommandRegistration(b *testing.B) {
	manager := cmd.NewManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd, _ := cmd.NewBaseCommand(
			fmt.Sprintf("cmd%d", i),
			fmt.Sprintf("命令 %d", i),
			"[options]",
			[]cmd.Option{},
		)
		manager.Register(&TestCommand{BaseCommand: cmd})
	}
}

// BenchmarkCommandLookup 基准测试命令查找
func BenchmarkCommandLookup(b *testing.B) {
	manager := cmd.NewManager()

	// 注册一些命令
	for i := 0; i < 100; i++ {
		cmd, _ := cmd.NewBaseCommand(
			fmt.Sprintf("cmd%d", i),
			fmt.Sprintf("命令 %d", i),
			"[options]",
			[]cmd.Option{},
		)
		manager.Register(&TestCommand{BaseCommand: cmd})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.GetCommand("cmd0")
	}
}
