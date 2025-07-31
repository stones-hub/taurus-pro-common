package cmd

import (
	"flag"
	"fmt"
	"strings"
)

// Command 命令接口
type Command interface {
	// Name 返回命令名称
	Name() string

	// Description 返回命令描述
	Description() string

	// Help 返回命令的详细帮助信息
	Help() string

	// Run 执行命令
	Run(args []string) error
}

// Option 命令选项定义
type Option struct {
	Name        string      // 选项名称
	Shorthand   string      // 短选项名称
	Description string      // 选项描述
	Type        OptionType  // 选项类型
	Required    bool        // 是否必填
	Default     interface{} // 默认值
}

// OptionType 选项类型
type OptionType int

const (
	OptionTypeString OptionType = iota
	OptionTypeInt
	OptionTypeBool
	OptionTypeFloat
)

// CommandContext 命令执行上下文
type CommandContext struct {
	Args    []string
	Options map[string]interface{}
}

// BaseCommand 基础命令实现
type BaseCommand struct {
	name        string
	description string
	usage       string
	options     []Option
	helpText    string
}

// 安全的类型转换函数
func safeString(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}
	if s, ok := v.(string); ok {
		return s, true
	}
	return "", false
}

func safeInt(v interface{}) (int, bool) {
	if v == nil {
		return 0, false
	}
	if i, ok := v.(int); ok {
		return i, true
	}
	return 0, false
}

func safeBool(v interface{}) (bool, bool) {
	if v == nil {
		return false, false
	}
	if b, ok := v.(bool); ok {
		return b, true
	}
	return false, false
}

func safeFloat(v interface{}) (float64, bool) {
	if v == nil {
		return 0.0, false
	}
	if f, ok := v.(float64); ok {
		return f, true
	}
	return 0.0, false
}

// 验证选项配置的有效性
func validateOptions(options []Option) error {
	seenNames := make(map[string]bool)
	seenShorthands := make(map[string]bool)

	for _, opt := range options {
		// 验证选项名
		if opt.Name == "" {
			return fmt.Errorf("选项名不能为空")
		}
		if seenNames[opt.Name] {
			return fmt.Errorf("重复的选项名: %s", opt.Name)
		}
		seenNames[opt.Name] = true

		// 验证短选项名
		if opt.Shorthand != "" {
			if len(opt.Shorthand) != 1 {
				return fmt.Errorf("短选项名必须是单个字符: %s", opt.Shorthand)
			}
			if seenShorthands[opt.Shorthand] {
				return fmt.Errorf("重复的短选项名: %s", opt.Shorthand)
			}
			seenShorthands[opt.Shorthand] = true
		}

		// 验证默认值类型
		if opt.Default != nil {
			if err := validateDefaultValue(opt); err != nil {
				return fmt.Errorf("选项 %s 的默认值无效: %v", opt.Name, err)
			}
		}
	}

	return nil
}

// 验证默认值类型
func validateDefaultValue(opt Option) error {
	switch opt.Type {
	case OptionTypeString:
		if _, ok := opt.Default.(string); !ok {
			return fmt.Errorf("期望字符串类型，实际类型: %T", opt.Default)
		}
	case OptionTypeInt:
		if _, ok := opt.Default.(int); !ok {
			return fmt.Errorf("期望整数类型，实际类型: %T", opt.Default)
		}
	case OptionTypeBool:
		if _, ok := opt.Default.(bool); !ok {
			return fmt.Errorf("期望布尔类型，实际类型: %T", opt.Default)
		}
	case OptionTypeFloat:
		if _, ok := opt.Default.(float64); !ok {
			return fmt.Errorf("期望浮点数类型，实际类型: %T", opt.Default)
		}
	default:
		return fmt.Errorf("未知的选项类型: %v", opt.Type)
	}
	return nil
}

// NewBaseCommand 创建基础命令
func NewBaseCommand(name, description, usage string, options []Option) (*BaseCommand, error) {
	// 验证输入参数
	if name == "" {
		return nil, fmt.Errorf("命令名不能为空")
	}
	if description == "" {
		return nil, fmt.Errorf("命令描述不能为空")
	}

	// 验证选项配置
	if err := validateOptions(options); err != nil {
		return nil, err
	}

	cmd := &BaseCommand{
		name:        name,
		description: description,
		usage:       usage,
		options:     options,
	}
	cmd.helpText = cmd.generateHelp()
	return cmd, nil
}

// Name 返回命令名称
func (c *BaseCommand) Name() string {
	return c.name
}

// Description 返回命令描述
func (c *BaseCommand) Description() string {
	return c.description
}

// Help 返回帮助信息
func (c *BaseCommand) Help() string {
	return c.helpText
}

// generateHelp 生成帮助信息
func (c *BaseCommand) generateHelp() string {
	var help strings.Builder

	// 写入命令描述
	help.WriteString(c.description)
	help.WriteString("\n\n用法:\n  ")
	help.WriteString(c.name)
	help.WriteString(" ")
	help.WriteString(c.usage)
	help.WriteString("\n")

	// 如果有选项，写入选项说明
	if len(c.options) > 0 {
		help.WriteString("\n选项:\n")

		// 计算最长的选项名，用于对齐
		maxLen := 0
		for _, opt := range c.options {
			length := len(opt.Name)
			if opt.Shorthand != "" {
				length += 4 // 为 ", -x" 添加长度
			}
			if length > maxLen {
				maxLen = length
			}
		}

		// 写入每个选项的说明
		for _, opt := range c.options {
			help.WriteString("  --")
			help.WriteString(opt.Name)
			if opt.Shorthand != "" {
				help.WriteString(", -")
				help.WriteString(opt.Shorthand)
			}

			// 添加空格以对齐描述
			padding := maxLen - len(opt.Name)
			if opt.Shorthand != "" {
				padding -= 4
			}
			help.WriteString(strings.Repeat(" ", padding+2))

			help.WriteString(opt.Description)

			// 添加默认值或必填标记
			if opt.Required {
				help.WriteString(" (必填)")
			} else if opt.Default != nil {
				help.WriteString(" (默认: ")
				help.WriteString(fmt.Sprintf("%v", opt.Default))
				help.WriteString(")")
			}
			help.WriteString("\n")
		}
	}

	return help.String()
}

// ParseOptions 解析命令行选项
func (c *BaseCommand) ParseOptions(args []string) (*CommandContext, error) {
	// 创建 flag 集合
	flags := flag.NewFlagSet(c.name, flag.ExitOnError)

	// 创建选项值存储
	optionValues := make(map[string]interface{})
	optionPtrs := make(map[string]interface{})

	// 为每个选项创建对应的 flag
	for _, opt := range c.options {
		switch opt.Type {
		case OptionTypeString:
			defaultVal := ""
			if opt.Default != nil {
				if val, ok := safeString(opt.Default); ok {
					defaultVal = val
				} else {
					return nil, fmt.Errorf("选项 %s 的默认值类型错误", opt.Name)
				}
			}
			ptr := flags.String(opt.Name, defaultVal, opt.Description)
			optionPtrs[opt.Name] = ptr
			if opt.Shorthand != "" {
				flags.StringVar(ptr, opt.Shorthand, defaultVal, opt.Description)
			}

		case OptionTypeInt:
			defaultVal := 0
			if opt.Default != nil {
				if val, ok := safeInt(opt.Default); ok {
					defaultVal = val
				} else {
					return nil, fmt.Errorf("选项 %s 的默认值类型错误", opt.Name)
				}
			}
			ptr := flags.Int(opt.Name, defaultVal, opt.Description)
			optionPtrs[opt.Name] = ptr
			if opt.Shorthand != "" {
				flags.IntVar(ptr, opt.Shorthand, defaultVal, opt.Description)
			}

		case OptionTypeBool:
			defaultVal := false
			if opt.Default != nil {
				if val, ok := safeBool(opt.Default); ok {
					defaultVal = val
				} else {
					return nil, fmt.Errorf("选项 %s 的默认值类型错误", opt.Name)
				}
			}
			ptr := flags.Bool(opt.Name, defaultVal, opt.Description)
			optionPtrs[opt.Name] = ptr
			if opt.Shorthand != "" {
				flags.BoolVar(ptr, opt.Shorthand, defaultVal, opt.Description)
			}

		case OptionTypeFloat:
			defaultVal := 0.0
			if opt.Default != nil {
				if val, ok := safeFloat(opt.Default); ok {
					defaultVal = val
				} else {
					return nil, fmt.Errorf("选项 %s 的默认值类型错误", opt.Name)
				}
			}
			ptr := flags.Float64(opt.Name, defaultVal, opt.Description)
			optionPtrs[opt.Name] = ptr
			if opt.Shorthand != "" {
				flags.Float64Var(ptr, opt.Shorthand, defaultVal, opt.Description)
			}
		}
	}

	// 解析参数
	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	// 提取选项值
	for name, ptr := range optionPtrs {
		switch v := ptr.(type) {
		case *string:
			optionValues[name] = *v
		case *int:
			optionValues[name] = *v
		case *bool:
			optionValues[name] = *v
		case *float64:
			optionValues[name] = *v
		}
	}

	// 验证必填选项
	for _, opt := range c.options {
		if opt.Required {
			value, exists := optionValues[opt.Name]
			if !exists {
				return nil, fmt.Errorf("选项 --%s 是必填的", opt.Name)
			}

			switch opt.Type {
			case OptionTypeString:
				if str, ok := safeString(value); !ok || str == "" {
					return nil, fmt.Errorf("选项 --%s 是必填的", opt.Name)
				}
			case OptionTypeInt:
				if _, ok := safeInt(value); !ok {
					return nil, fmt.Errorf("选项 --%s 是必填的", opt.Name)
				}
			case OptionTypeFloat:
				if _, ok := safeFloat(value); !ok {
					return nil, fmt.Errorf("选项 --%s 是必填的", opt.Name)
				}
			}
		}
	}

	return &CommandContext{
		Args:    flags.Args(),
		Options: optionValues,
	}, nil
}
