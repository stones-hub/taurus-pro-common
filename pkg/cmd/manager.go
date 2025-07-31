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
type Manager struct {
	commands map[string]Command
	mu       sync.RWMutex
}

// NewManager 创建命令管理器
func NewManager() *Manager {
	return &Manager{
		commands: make(map[string]Command),
	}
}

// Register 注册命令
func (m *Manager) Register(cmd Command) error {
	if cmd == nil {
		return fmt.Errorf("命令不能为空")
	}

	name := cmd.Name()
	if name == "" {
		return fmt.Errorf("命令名不能为空")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查命令是否已存在
	if _, exists := m.commands[name]; exists {
		return fmt.Errorf("命令 '%s' 已存在", name)
	}

	m.commands[name] = cmd
	return nil
}

// GetCommand 获取命令（线程安全）
func (m *Manager) GetCommand(name string) (Command, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cmd, exists := m.commands[name]
	return cmd, exists
}

// GetCommands 获取所有命令（线程安全）
func (m *Manager) GetCommands() map[string]Command {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本以避免外部修改
	commands := make(map[string]Command)
	for name, cmd := range m.commands {
		commands[name] = cmd
	}
	return commands
}

// Run 运行命令
func (m *Manager) Run() error {
	// 设置 panic 恢复
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("命令执行出错: %v\n", err)
			fmt.Printf("堆栈信息:\n%s\n", debug.Stack())
			os.Exit(1)
		}
	}()

	if len(os.Args) < 2 {
		return m.showHelp()
	}

	cmdName := os.Args[1]
	if cmdName == "help" {
		return m.showHelp()
	}

	cmd, ok := m.GetCommand(cmdName)
	if !ok {
		return m.handleUnknownCommand(cmdName)
	}

	// 如果是请求帮助，显示命令的详细帮助信息
	if len(os.Args) > 2 && (os.Args[2] == "-h" || os.Args[2] == "--help") {
		fmt.Print(cmd.Help())
		return nil
	}

	// 执行命令
	return cmd.Run(os.Args[2:])
}

// handleUnknownCommand 处理未知命令，提供相似命令建议
func (m *Manager) handleUnknownCommand(cmdName string) error {
	commands := m.GetCommands()

	// 寻找相似命令
	var suggestions []string
	for name := range commands {
		if strings.Contains(strings.ToLower(name), strings.ToLower(cmdName)) ||
			strings.Contains(strings.ToLower(cmdName), strings.ToLower(name)) {
			suggestions = append(suggestions, name)
		}
	}

	// 构建错误信息
	var errMsg strings.Builder
	errMsg.WriteString(fmt.Sprintf("未知命令: %s\n\n", cmdName))

	if len(suggestions) > 0 {
		errMsg.WriteString("您可能想要运行以下命令之一:\n")
		for _, suggestion := range suggestions {
			errMsg.WriteString(fmt.Sprintf("  %s\n", suggestion))
		}
		errMsg.WriteString("\n")
	}

	errMsg.WriteString(fmt.Sprintf("运行 '%s help' 查看可用命令", os.Args[0]))

	return fmt.Errorf("%s", errMsg.String())
}

// showHelp 显示帮助信息
func (m *Manager) showHelp() error {
	fmt.Println("使用方法:")
	fmt.Printf("  %s <command> [arguments]\n\n", os.Args[0])
	fmt.Println("可用命令:")

	commands := m.GetCommands()

	// 获取所有命令并排序
	var names []string
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	// 计算最长命令名
	maxLen := 0
	for _, name := range names {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	// 显示命令列表
	for _, name := range names {
		cmd := commands[name]
		padding := strings.Repeat(" ", maxLen-len(name)+2)
		fmt.Printf("  %s%s%s\n", name, padding, cmd.Description())
	}

	fmt.Printf("\n运行 '%s <command> --help' 查看命令的详细信息\n", os.Args[0])
	return nil
}
