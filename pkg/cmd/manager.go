// Package cmd 提供命令行工具的核心功能
// 包含命令管理器、命令注册、执行等功能
package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
)

// Manager 命令管理器
// 负责管理所有注册的命令，提供命令注册、查找、执行等功能
// 支持线程安全的并发操作
type Manager struct {
	commands map[string]Command // 命令映射表，键为命令名，值为命令实例
	mu       sync.RWMutex       // 读写锁，保证并发安全
}

// NewManager 创建新的命令管理器
// 返回一个初始化的命令管理器实例
func NewManager() *Manager {
	return &Manager{
		commands: make(map[string]Command),
	}
}

// Register 注册命令到管理器
// cmd: 要注册的命令实例
// 返回注册结果，nil表示成功
// 线程安全：使用写锁保护命令映射表
func (m *Manager) Register(cmd Command) error {
	// 验证命令实例
	if cmd == nil {
		return fmt.Errorf("命令不能为空")
	}

	// 获取命令名称
	name := cmd.Name()
	if name == "" {
		return fmt.Errorf("命令名不能为空")
	}

	// 获取写锁，保护命令映射表
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查命令是否已存在
	if _, exists := m.commands[name]; exists {
		return fmt.Errorf("命令 '%s' 已存在", name)
	}

	// 注册命令
	m.commands[name] = cmd
	return nil
}

// GetCommand 获取指定名称的命令
// name: 命令名称
// 返回值：(命令实例, 是否存在)
// 线程安全：使用读锁保护命令映射表
func (m *Manager) GetCommand(name string) (Command, bool) {
	// 获取读锁，保护命令映射表
	m.mu.RLock()
	defer m.mu.RUnlock()

	cmd, exists := m.commands[name]
	return cmd, exists
}

// GetCommands 获取所有注册的命令
// 返回命令映射表的副本，避免外部修改影响内部状态
// 线程安全：使用读锁保护命令映射表
func (m *Manager) GetCommands() map[string]Command {
	// 获取读锁，保护命令映射表
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本以避免外部修改
	commands := make(map[string]Command)
	for name, cmd := range m.commands {
		commands[name] = cmd
	}
	return commands
}

// Run 运行命令管理器
// 解析命令行参数，查找并执行对应的命令
// 返回执行结果，nil表示成功
// 包含panic恢复机制，确保程序稳定性
func (m *Manager) Run() error {
	// 设置 panic 恢复，确保程序不会因为命令执行错误而崩溃
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("命令执行出错: %v\n", err)
			fmt.Printf("堆栈信息:\n%s\n", debug.Stack())
			os.Exit(1)
		}
	}()

	// 检查命令行参数
	if len(os.Args) < 2 {
		// 没有提供命令名，显示帮助信息
		return m.showHelp()
	}

	// 获取命令名
	cmdName := os.Args[1]

	// 处理帮助命令
	if cmdName == "help" {
		return m.showHelp()
	}

	// 查找命令
	cmd, ok := m.GetCommand(cmdName)
	if !ok {
		// 命令不存在，提供相似命令建议
		return m.handleUnknownCommand(cmdName)
	}

	// 检查是否请求帮助
	if len(os.Args) > 2 && (os.Args[2] == "-h" || os.Args[2] == "--help") {
		// 显示命令的详细帮助信息
		fmt.Print(cmd.Help())
		return nil
	}

	// 执行命令，传递剩余的参数
	return cmd.Run(os.Args[2:])
}

// handleUnknownCommand 处理未知命令
// cmdName: 未知的命令名
// 提供相似命令建议，帮助用户找到正确的命令
// 返回格式化的错误信息
func (m *Manager) handleUnknownCommand(cmdName string) error {
	// 获取所有注册的命令
	commands := m.GetCommands()

	// 寻找相似命令
	var suggestions []string
	for name := range commands {
		// 使用不区分大小写的字符串匹配
		// 检查命令名是否包含用户输入的部分，或用户输入是否包含命令名
		if strings.Contains(strings.ToLower(name), strings.ToLower(cmdName)) ||
			strings.Contains(strings.ToLower(cmdName), strings.ToLower(name)) {
			suggestions = append(suggestions, name)
		}
	}

	// 构建错误信息
	var errMsg strings.Builder
	errMsg.WriteString(fmt.Sprintf("未知命令: %s\n\n", cmdName))

	// 如果有相似命令，提供建议
	if len(suggestions) > 0 {
		errMsg.WriteString("您可能想要运行以下命令之一:\n")
		for _, suggestion := range suggestions {
			errMsg.WriteString(fmt.Sprintf("  %s\n", suggestion))
		}
		errMsg.WriteString("\n")
	}

	// 提供帮助命令提示
	errMsg.WriteString(fmt.Sprintf("运行 '%s help' 查看可用命令", os.Args[0]))

	return fmt.Errorf("%s", errMsg.String())
}

// showHelp 显示帮助信息
// 列出所有可用的命令及其描述
// 按命令名排序，提供整齐的显示格式
func (m *Manager) showHelp() error {
	// 显示使用方法
	fmt.Println("使用方法:")
	fmt.Printf("  %s <command> [arguments]\n\n", os.Args[0])
	fmt.Println("可用命令:")

	// 获取所有命令
	commands := m.GetCommands()

	// 获取所有命令名并排序
	var names []string
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	// 计算最长命令名，用于对齐显示
	maxLen := 0
	for _, name := range names {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	// 显示命令列表
	for _, name := range names {
		cmd := commands[name]
		// 计算填充空格，确保描述对齐
		padding := strings.Repeat(" ", maxLen-len(name)+2)
		fmt.Printf("  %s%s%s\n", name, padding, cmd.Description())
	}

	// 提供详细帮助的提示
	fmt.Printf("\n运行 '%s <command> --help' 查看命令的详细信息\n", os.Args[0])
	return nil
}

// Clear 清空所有命令
// 用于测试或重置命令管理器
// 线程安全：使用写锁保护命令映射表
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.commands = make(map[string]Command)
}
