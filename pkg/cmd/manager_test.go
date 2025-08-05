package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestManager 测试命令管理器
func TestManager(t *testing.T) {
	t.Run("创建管理器", func(t *testing.T) {
		manager := NewManager()
		if manager == nil {
			t.Error("期望创建管理器成功，实际为nil")
		}
	})

	t.Run("注册命令", func(t *testing.T) {
		manager := NewManager()

		// 创建测试命令
		cmd, err := NewBaseCommand("test", "测试命令", "[options]", nil)
		if err != nil {
			t.Fatalf("创建命令失败: %v", err)
		}

		// 注册命令
		err = manager.Register(cmd)
		if err != nil {
			t.Errorf("期望注册成功，实际错误: %v", err)
		}

		// 验证命令已注册
		registeredCmd, exists := manager.GetCommand("test")
		if !exists {
			t.Error("期望命令存在，实际不存在")
		}
		if registeredCmd != cmd {
			t.Error("期望返回相同命令实例，实际不同")
		}
	})

	t.Run("注册空命令", func(t *testing.T) {
		manager := NewManager()

		err := manager.Register(nil)
		if err == nil {
			t.Error("期望注册失败，实际成功")
		}
		if !strings.Contains(err.Error(), "命令不能为空") {
			t.Errorf("期望错误包含'命令不能为空'，实际错误: %v", err)
		}
	})

	t.Run("重复注册命令", func(t *testing.T) {
		manager := NewManager()

		// 创建第一个命令
		cmd1, err := NewBaseCommand("test", "测试命令1", "[options]", nil)
		if err != nil {
			t.Fatalf("创建命令失败: %v", err)
		}

		// 创建第二个命令（同名）
		cmd2, err := NewBaseCommand("test", "测试命令2", "[options]", nil)
		if err != nil {
			t.Fatalf("创建命令失败: %v", err)
		}

		// 注册第一个命令
		err = manager.Register(cmd1)
		if err != nil {
			t.Errorf("期望第一次注册成功，实际错误: %v", err)
		}

		// 尝试注册同名命令
		err = manager.Register(cmd2)
		if err == nil {
			t.Error("期望重复注册失败，实际成功")
		}
		if !strings.Contains(err.Error(), "命令 'test' 已存在") {
			t.Errorf("期望错误包含'命令 'test' 已存在'，实际错误: %v", err)
		}
	})

	t.Run("获取不存在的命令", func(t *testing.T) {
		manager := NewManager()

		cmd, exists := manager.GetCommand("nonexistent")
		if exists {
			t.Error("期望命令不存在，实际存在")
		}
		if cmd != nil {
			t.Error("期望返回nil命令，实际返回非nil")
		}
	})

	t.Run("获取所有命令", func(t *testing.T) {
		manager := NewManager()

		// 创建多个命令
		cmd1, _ := NewBaseCommand("test1", "测试命令1", "[options]", nil)
		cmd2, _ := NewBaseCommand("test2", "测试命令2", "[options]", nil)
		cmd3, _ := NewBaseCommand("test3", "测试命令3", "[options]", nil)

		// 注册命令
		manager.Register(cmd1)
		manager.Register(cmd2)
		manager.Register(cmd3)

		// 获取所有命令
		commands := manager.GetCommands()

		if len(commands) != 3 {
			t.Errorf("期望3个命令，实际%d个", len(commands))
		}

		// 验证所有命令都存在
		if _, exists := commands["test1"]; !exists {
			t.Error("期望test1命令存在")
		}
		if _, exists := commands["test2"]; !exists {
			t.Error("期望test2命令存在")
		}
		if _, exists := commands["test3"]; !exists {
			t.Error("期望test3命令存在")
		}
	})

	t.Run("清空命令", func(t *testing.T) {
		manager := NewManager()

		// 注册命令
		cmd, _ := NewBaseCommand("test", "测试命令", "[options]", nil)
		manager.Register(cmd)

		// 验证命令存在
		if _, exists := manager.GetCommand("test"); !exists {
			t.Error("期望命令存在")
		}

		// 清空命令
		manager.Clear()

		// 验证命令不存在
		if _, exists := manager.GetCommand("test"); exists {
			t.Error("期望命令不存在")
		}

		// 验证命令列表为空
		commands := manager.GetCommands()
		if len(commands) != 0 {
			t.Errorf("期望0个命令，实际%d个", len(commands))
		}
	})
}

// TestManagerRun 测试管理器运行
func TestManagerRun(t *testing.T) {
	// 保存原始参数
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	t.Run("无参数时显示帮助", func(t *testing.T) {
		manager := NewManager()

		// 设置空参数
		os.Args = []string{"program"}

		// 直接运行，不捕获输出
		err := manager.Run()
		if err != nil {
			t.Errorf("期望运行成功，实际错误: %v", err)
		}
	})

	t.Run("help命令", func(t *testing.T) {
		manager := NewManager()

		// 设置help参数
		os.Args = []string{"program", "help"}

		// 直接运行，不捕获输出
		err := manager.Run()
		if err != nil {
			t.Errorf("期望运行成功，实际错误: %v", err)
		}
	})

	t.Run("未知命令", func(t *testing.T) {
		manager := NewManager()

		// 设置未知命令
		os.Args = []string{"program", "unknown"}

		// 直接运行，验证错误
		err := manager.Run()
		if err == nil {
			t.Error("期望运行失败，实际成功")
		}
		if !strings.Contains(err.Error(), "未知命令: unknown") {
			t.Errorf("期望错误包含'未知命令: unknown'，实际错误: %v", err)
		}
	})

	t.Run("命令帮助", func(t *testing.T) {
		manager := NewManager()

		// 创建测试命令
		cmd, _ := NewBaseCommand("test", "测试命令", "[options]", nil)
		manager.Register(cmd)

		// 设置命令帮助参数
		os.Args = []string{"program", "test", "--help"}

		// 直接运行，不捕获输出
		err := manager.Run()
		if err != nil {
			t.Errorf("期望运行成功，实际错误: %v", err)
		}
	})

	t.Run("执行命令", func(t *testing.T) {
		manager := NewManager()

		// 创建自定义命令
		customCmd := &CustomTestCommand{
			BaseCommand: &BaseCommand{
				name:        "custom",
				description: "自定义测试命令",
				usage:       "[options]",
			},
		}
		manager.Register(customCmd)

		// 设置命令参数
		os.Args = []string{"program", "custom", "arg1", "arg2"}

		// 直接运行
		err := manager.Run()
		if err != nil {
			t.Errorf("期望运行成功，实际错误: %v", err)
		}
	})
}

// TestHandleUnknownCommand 测试未知命令处理
func TestHandleUnknownCommand(t *testing.T) {
	manager := NewManager()

	// 注册一些命令
	cmd1, _ := NewBaseCommand("build", "构建项目", "[options]", nil)
	cmd2, _ := NewBaseCommand("test", "运行测试", "[options]", nil)
	cmd3, _ := NewBaseCommand("deploy", "部署应用", "[options]", nil)

	manager.Register(cmd1)
	manager.Register(cmd2)
	manager.Register(cmd3)

	t.Run("完全不同的命令", func(t *testing.T) {
		err := manager.handleUnknownCommand("completely-different")

		if err == nil {
			t.Error("期望返回错误，实际为nil")
		}

		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "未知命令: completely-different") {
			t.Errorf("期望错误包含'未知命令: completely-different'，实际为: %s", errorMsg)
		}

		// 应该没有建议
		if strings.Contains(errorMsg, "您可能想要运行以下命令之一") {
			t.Error("期望没有命令建议")
		}
	})

	t.Run("相似命令", func(t *testing.T) {
		err := manager.handleUnknownCommand("buil")

		if err == nil {
			t.Error("期望返回错误，实际为nil")
		}

		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "未知命令: buil") {
			t.Errorf("期望错误包含'未知命令: buil'，实际为: %s", errorMsg)
		}

		// 应该有建议
		if !strings.Contains(errorMsg, "您可能想要运行以下命令之一") {
			t.Error("期望有命令建议")
		}

		if !strings.Contains(errorMsg, "build") {
			t.Error("期望建议包含'build'命令")
		}
	})

	t.Run("大小写不敏感", func(t *testing.T) {
		err := manager.handleUnknownCommand("BUILD")

		if err == nil {
			t.Error("期望返回错误，实际为nil")
		}

		errorMsg := err.Error()
		if !strings.Contains(errorMsg, "您可能想要运行以下命令之一") {
			t.Error("期望有命令建议")
		}

		if !strings.Contains(errorMsg, "build") {
			t.Error("期望建议包含'build'命令")
		}
	})
}

// TestShowHelp 测试帮助显示
func TestShowHelp(t *testing.T) {
	// 保存原始参数
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	manager := NewManager()

	// 注册一些命令
	cmd1, _ := NewBaseCommand("build", "构建项目", "[options]", nil)
	cmd2, _ := NewBaseCommand("test", "运行测试", "[options]", nil)
	cmd3, _ := NewBaseCommand("deploy", "部署应用", "[options]", nil)

	manager.Register(cmd1)
	manager.Register(cmd2)
	manager.Register(cmd3)

	// 设置程序名
	os.Args = []string{"myapp"}

	// 直接运行，不验证输出内容
	err := manager.showHelp()
	if err != nil {
		t.Errorf("期望显示帮助成功，实际错误: %v", err)
	}
}

// CustomTestCommand 自定义测试命令
type CustomTestCommand struct {
	*BaseCommand
}

func (c *CustomTestCommand) Run(args []string) error {
	// 简单的测试实现
	if len(args) >= 2 {
		return nil
	}
	return fmt.Errorf("需要至少2个参数")
}
