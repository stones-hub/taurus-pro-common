// Package cmd 提供命令行工具的核心功能
// 包含命令接口定义、选项类型、基础命令实现等
package cmd

import (
	"flag"
	"fmt"
	"strings"
)

// Command 命令接口
// 定义了命令行工具中所有命令必须实现的方法
type Command interface {
	// Name 返回命令名称
	// 用于在命令行中识别和调用命令
	Name() string

	// Description 返回命令描述
	// 用于在帮助信息中显示命令的简要说明
	Description() string

	// Help 返回命令的详细帮助信息
	// 包含命令的完整用法说明、选项列表等
	Help() string

	// Run 执行命令
	// args: 命令行参数，不包含命令名本身
	// 返回执行结果，nil表示成功
	Run(args []string) error
}

// Option 命令选项定义
// 描述命令支持的配置选项
type Option struct {
	Name        string      // 选项名称，如 "--verbose"
	Shorthand   string      // 短选项名称，如 "-v"
	Description string      // 选项描述，用于帮助信息
	Type        OptionType  // 选项数据类型
	Required    bool        // 是否必填选项
	Default     interface{} // 默认值，nil表示无默认值
}

// OptionType 选项类型枚举
// 定义支持的数据类型
type OptionType int

const (
	OptionTypeString OptionType = iota // 字符串类型
	OptionTypeInt                      // 整数类型
	OptionTypeInt64                    // 长整数类型
	OptionTypeBool                     // 布尔类型
	OptionTypeFloat                    // 浮点数类型
)

// CommandContext 命令执行上下文
// 包含解析后的参数和选项值
type CommandContext struct {
	Args    []string               // 位置参数（非选项参数）
	Options map[string]interface{} // 选项值映射，键为选项名，值为解析后的值
}

// BaseCommand 基础命令实现
// 提供命令接口的默认实现，包含选项解析、帮助生成等功能
type BaseCommand struct {
	name        string   // 命令名称
	description string   // 命令描述
	usage       string   // 使用说明
	options     []Option // 选项列表
	helpText    string   // 缓存的帮助文本
}

// safeString 安全的字符串类型转换
// 将interface{}转换为string，避免panic
// 返回值：(转换后的字符串, 是否转换成功)
func safeString(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}
	if s, ok := v.(string); ok {
		return s, true
	}
	return "", false
}

// safeInt 安全的整数类型转换
// 将interface{}转换为int，避免panic
// 返回值：(转换后的整数, 是否转换成功)
func safeInt(v interface{}) (int, bool) {
	if v == nil {
		return 0, false
	}
	if i, ok := v.(int); ok {
		return i, true
	}
	return 0, false
}

// safeBool 安全的布尔类型转换
// 将interface{}转换为bool，避免panic
// 返回值：(转换后的布尔值, 是否转换成功)
func safeBool(v interface{}) (bool, bool) {
	if v == nil {
		return false, false
	}
	if b, ok := v.(bool); ok {
		return b, true
	}
	return false, false
}

// safeFloat 安全的浮点数类型转换
// 将interface{}转换为float64，避免panic
// 返回值：(转换后的浮点数, 是否转换成功)
func safeFloat(v interface{}) (float64, bool) {
	if v == nil {
		return 0.0, false
	}
	if f, ok := v.(float64); ok {
		return f, true
	}
	return 0.0, false
}

// safeInt64 安全的长整数类型转换
// 将interface{}转换为int64，避免panic
// 返回值：(转换后的长整数, 是否转换成功)
func safeInt64(v interface{}) (int64, bool) {
	if v == nil {
		return 0, false
	}
	if i, ok := v.(int64); ok {
		return i, true
	}
	return 0, false
}

// validateOptions 验证选项配置的有效性
// 检查选项名、短选项名、默认值类型等
// 返回验证错误，nil表示验证通过
func validateOptions(options []Option) error {
	seenNames := make(map[string]bool)      // 已见过的选项名
	seenShorthands := make(map[string]bool) // 已见过的短选项名
	var errors []string                     // 收集所有错误信息

	for i, opt := range options {
		// 验证选项名
		if opt.Name == "" {
			errors = append(errors, fmt.Sprintf("选项[%d]: 选项名不能为空", i+1))
		} else if seenNames[opt.Name] {
			errors = append(errors, fmt.Sprintf("选项[%d]: 重复的选项名 '%s'", i+1, opt.Name))
		} else {
			seenNames[opt.Name] = true
		}

		// 验证短选项名
		if opt.Shorthand != "" {
			if len(opt.Shorthand) != 1 {
				errors = append(errors, fmt.Sprintf("选项[%d] '%s': 短选项名必须是单个字符，当前为 '%s' (长度: %d)",
					i+1, opt.Name, opt.Shorthand, len(opt.Shorthand)))
			} else if !isValidShortOption(opt.Shorthand[0]) {
				errors = append(errors, fmt.Sprintf("选项[%d] '%s': 短选项名必须是字母或数字，当前为 '%s'",
					i+1, opt.Name, opt.Shorthand))
			} else if seenShorthands[opt.Shorthand] {
				errors = append(errors, fmt.Sprintf("选项[%d] '%s': 重复的短选项名 '%s'", i+1, opt.Name, opt.Shorthand))
			} else {
				seenShorthands[opt.Shorthand] = true
			}
		}

		// 验证默认值类型
		if opt.Default != nil {
			if err := validateDefaultValue(opt); err != nil {
				errors = append(errors, fmt.Sprintf("选项[%d] '%s': 默认值无效 - %v", i+1, opt.Name, err))
			}
		}
	}

	// 如果有错误，返回格式化的错误信息
	if len(errors) > 0 {
		var errMsg strings.Builder
		errMsg.WriteString("选项配置验证失败:\n")
		for _, err := range errors {
			errMsg.WriteString("  • ")
			errMsg.WriteString(err)
			errMsg.WriteString("\n")
		}
		errMsg.WriteString("\n请检查选项配置并修复上述问题。")
		return fmt.Errorf("%s", errMsg.String())
	}

	return nil
}

// validateDefaultValue 验证选项默认值的类型是否正确
// 确保默认值的类型与选项声明的类型一致
func validateDefaultValue(opt Option) error {
	switch opt.Type {
	case OptionTypeString:
		if _, ok := safeString(opt.Default); !ok {
			return fmt.Errorf("期望字符串类型，实际类型: %T", opt.Default)
		}
	case OptionTypeInt:
		if _, ok := safeInt(opt.Default); !ok {
			return fmt.Errorf("期望整数类型，实际类型: %T", opt.Default)
		}
	case OptionTypeInt64:
		if _, ok := safeInt64(opt.Default); !ok {
			return fmt.Errorf("期望长整数类型，实际类型: %T", opt.Default)
		}
	case OptionTypeBool:
		if _, ok := safeBool(opt.Default); !ok {
			return fmt.Errorf("期望布尔类型，实际类型: %T", opt.Default)
		}
	case OptionTypeFloat:
		if _, ok := safeFloat(opt.Default); !ok {
			return fmt.Errorf("期望浮点数类型，实际类型: %T", opt.Default)
		}
	default:
		return fmt.Errorf("未知的选项类型: %v", opt.Type)
	}
	return nil
}

// isValidShortOption 检查短选项字符是否有效
// 只允许字母和数字作为短选项
func isValidShortOption(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// NewBaseCommand 创建新的基础命令
// name: 命令名称
// description: 命令描述
// usage: 使用说明
// options: 选项列表
// 返回命令实例和错误信息
func NewBaseCommand(name, description, usage string, options []Option) (*BaseCommand, error) {
	// 验证输入参数
	var errors []string

	if name == "" {
		errors = append(errors, "命令名不能为空")
	}
	if description == "" {
		errors = append(errors, "命令描述不能为空")
	}
	if usage == "" {
		errors = append(errors, "使用说明不能为空")
	}

	// 如果有基本参数错误，直接返回
	if len(errors) > 0 {
		var errMsg strings.Builder
		errMsg.WriteString("命令创建失败:\n")
		for _, err := range errors {
			errMsg.WriteString("  • ")
			errMsg.WriteString(err)
			errMsg.WriteString("\n")
		}
		return nil, fmt.Errorf("%s", errMsg.String())
	}

	// 验证选项配置
	if err := validateOptions(options); err != nil {
		return nil, fmt.Errorf("命令 '%s' 创建失败: %v", name, err)
	}

	cmd := &BaseCommand{
		name:        name,
		description: description,
		usage:       usage,
		options:     options,
	}

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

// Run 执行命令的默认实现
// 子类可以重写此方法以提供具体的命令逻辑
func (c *BaseCommand) Run(args []string) error {
	// 默认实现：解析选项并显示帮助信息
	ctx, err := c.ParseOptions(args)
	if err != nil {
		return err
	}

	// 如果没有位置参数，显示帮助信息
	if len(ctx.Args) == 0 {
		fmt.Print(c.Help())
		return nil
	}

	return fmt.Errorf("命令 '%s' 未实现具体的执行逻辑", c.name)
}

// Help 返回命令的详细帮助信息
// 如果帮助文本未生成，则先生成再返回
func (c *BaseCommand) Help() string {
	if c.helpText == "" {
		c.helpText = c.generateHelp()
	}
	return c.helpText
}

// generateHelp 生成详细的帮助信息
// 包含命令描述、用法说明、选项列表等
func (c *BaseCommand) generateHelp() string {
	var help strings.Builder

	// 命令描述
	help.WriteString(c.description)
	help.WriteString("\n\n")

	// 用法说明
	help.WriteString("用法:\n")
	help.WriteString(fmt.Sprintf("  %s %s\n\n", c.name, c.usage))

	// 选项列表
	if len(c.options) > 0 {
		help.WriteString("选项:\n")

		// 计算最长选项名用于对齐
		maxLen := 0
		for _, opt := range c.options {
			optLen := len(opt.Name)
			if opt.Shorthand != "" {
				optLen += 4 // "name, -s" 格式的长度
			}
			if optLen > maxLen {
				maxLen = optLen
			}
		}

		// 生成选项列表
		for _, opt := range c.options {
			// 选项名和短选项
			optName := opt.Name
			if opt.Shorthand != "" {
				optName = fmt.Sprintf("%s, -%s", opt.Name, opt.Shorthand)
			}

			// 对齐和描述
			padding := strings.Repeat(" ", maxLen-len(optName)+2)
			help.WriteString(fmt.Sprintf("  %s%s%s", optName, padding, opt.Description))

			// 必填标记
			if opt.Required {
				help.WriteString(" (必填)")
			}

			// 默认值
			if opt.Default != nil {
				help.WriteString(fmt.Sprintf(" (默认: %v)", opt.Default))
			}

			help.WriteString("\n")
		}
	}

	return help.String()
}

// ParseOptions 解析命令行选项和参数
// args: 原始命令行参数（不包含命令名）
// 返回解析后的上下文和错误信息
func (c *BaseCommand) ParseOptions(args []string) (*CommandContext, error) {
	// 创建 flag 集合用于参数解析，使用ContinueOnError以便捕获错误
	flags := flag.NewFlagSet(c.name, flag.ContinueOnError)

	// 创建选项值存储
	optionValues := make(map[string]interface{}) // 最终返回的选项值
	optionPtrs := make(map[string]interface{})   // flag指针，用于获取解析后的值
	providedOptions := make(map[string]bool)     // 跟踪用户实际提供的选项

	// 为每个选项创建对应的 flag
	for _, opt := range c.options {
		switch opt.Type {
		case OptionTypeString:
			// 处理字符串类型选项
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
			// 注册短选项
			if opt.Shorthand != "" {
				flags.StringVar(ptr, opt.Shorthand, defaultVal, opt.Description)
			}

		case OptionTypeInt:
			// 处理整数类型选项
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
			// 注册短选项
			if opt.Shorthand != "" {
				flags.IntVar(ptr, opt.Shorthand, defaultVal, opt.Description)
			}

		case OptionTypeInt64:
			// 处理长整数类型选项
			defaultVal := int64(0)
			if opt.Default != nil {
				if val, ok := safeInt64(opt.Default); ok {
					defaultVal = val
				} else {
					return nil, fmt.Errorf("选项 %s 的默认值类型错误", opt.Name)
				}
			}
			ptr := flags.Int64(opt.Name, defaultVal, opt.Description)
			optionPtrs[opt.Name] = ptr
			// 注册短选项
			if opt.Shorthand != "" {
				flags.Int64Var(ptr, opt.Shorthand, defaultVal, opt.Description)
			}

		case OptionTypeBool:
			// 处理布尔类型选项
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
			// 注册短选项
			if opt.Shorthand != "" {
				flags.BoolVar(ptr, opt.Shorthand, defaultVal, opt.Description)
			}

		case OptionTypeFloat:
			// 处理浮点数类型选项
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
			// 注册短选项
			if opt.Shorthand != "" {
				flags.Float64Var(ptr, opt.Shorthand, defaultVal, opt.Description)
			}
		}
	}

	// 解析命令行参数
	if err := flags.Parse(args); err != nil {
		// 提供更友好的错误信息
		return nil, c.formatParseError(err, args)
	}

	// 检查用户实际提供了哪些选项
	// 通过检查参数中是否包含选项名来判断
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") {
			optionName := strings.TrimPrefix(arg, "--")
			providedOptions[optionName] = true

			// 检查布尔选项是否被错误地指定了值
			for _, opt := range c.options {
				if opt.Name == optionName && opt.Type == OptionTypeBool {
					// 检查下一个参数是否是值（不是选项）
					if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
						// 检查下一个参数是否是布尔值字符串
						nextArg := args[i+1]
						if nextArg == "true" || nextArg == "false" {
							return nil, fmt.Errorf("布尔选项 --%s 不需要值，请移除 '%s'", opt.Name, nextArg)
						}
					}
				}
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) == 2 {
			// 短选项，需要找到对应的长选项名
			shortOpt := arg[1:]
			for _, opt := range c.options {
				if opt.Shorthand == shortOpt {
					providedOptions[opt.Name] = true

					// 检查布尔选项是否被错误地指定了值
					if opt.Type == OptionTypeBool {
						// 检查下一个参数是否是值（不是选项）
						if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
							// 检查下一个参数是否是布尔值字符串
							nextArg := args[i+1]
							if nextArg == "true" || nextArg == "false" {
								return nil, fmt.Errorf("布尔选项 -%s 不需要值，请移除 '%s'", opt.Shorthand, nextArg)
							}
						}
					}
					break
				}
			}
		}
	}

	// 提取选项值到结果映射中
	for name, ptr := range optionPtrs {
		switch v := ptr.(type) {
		case *string:
			optionValues[name] = *v
		case *int:
			optionValues[name] = *v
		case *int64:
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
			// 检查用户是否提供了该选项
			if !providedOptions[opt.Name] {
				return nil, fmt.Errorf("选项 --%s 是必填的", opt.Name)
			}

			value := optionValues[opt.Name]
			// 根据类型进行额外的验证
			switch opt.Type {
			case OptionTypeString:
				// 字符串类型：检查是否为空字符串
				if str, ok := safeString(value); !ok || str == "" {
					return nil, fmt.Errorf("选项 --%s 是必填的", opt.Name)
				}
			case OptionTypeInt:
				// 整数类型：检查类型转换是否成功
				if _, ok := safeInt(value); !ok {
					return nil, fmt.Errorf("选项 --%s 是必填的", opt.Name)
				}
			case OptionTypeInt64:
				// 长整数类型：检查类型转换是否成功
				if _, ok := safeInt64(value); !ok {
					return nil, fmt.Errorf("选项 --%s 是必填的", opt.Name)
				}
			case OptionTypeFloat:
				// 浮点数类型：检查类型转换是否成功
				if _, ok := safeFloat(value); !ok {
					return nil, fmt.Errorf("选项 --%s 是必填的", opt.Name)
				}
			}
		}
	}

	// 返回解析结果
	return &CommandContext{
		Args:    flags.Args(), // 位置参数（非选项参数）
		Options: optionValues, // 选项值映射
	}, nil
}

// formatParseError 格式化解析错误，提供友好的用户提示
// err: 原始解析错误
// args: 原始命令行参数
// 返回格式化的错误信息
func (c *BaseCommand) formatParseError(err error, args []string) error {
	// 检查是否是未定义选项的错误
	if strings.Contains(err.Error(), "flag provided but not defined") {
		// 从错误信息中提取未定义的选项名
		errText := err.Error()
		// 错误格式通常是: "flag provided but not defined: -option-name"
		parts := strings.Split(errText, ": ")
		var unknownFlag string
		if len(parts) >= 2 {
			unknownFlag = strings.TrimSpace(parts[1])
		} else {
			// 如果无法从错误信息中提取，则从参数中查找
			for _, arg := range args {
				if strings.HasPrefix(arg, "-") {
					unknownFlag = arg
					break
				}
			}
		}

		// 构建友好的错误信息
		var errMsg strings.Builder
		errMsg.WriteString(fmt.Sprintf("错误: 未定义的选项 '%s'\n\n", unknownFlag))
		errMsg.WriteString("可用的选项:\n")

		// 计算最长选项名用于对齐
		maxLen := 0
		for _, opt := range c.options {
			optLen := len(opt.Name)
			if opt.Shorthand != "" {
				optLen += 4 // "name, -s" 格式的长度
			}
			if optLen > maxLen {
				maxLen = optLen
			}
		}

		// 显示所有可用选项
		for _, opt := range c.options {
			optName := opt.Name
			if opt.Shorthand != "" {
				optName = fmt.Sprintf("%s, -%s", opt.Name, opt.Shorthand)
			}

			padding := strings.Repeat(" ", maxLen-len(optName)+2)
			errMsg.WriteString(fmt.Sprintf("  %s%s%s", optName, padding, opt.Description))

			if opt.Required {
				errMsg.WriteString(" (必填)")
			}

			if opt.Default != nil {
				errMsg.WriteString(fmt.Sprintf(" (默认: %v)", opt.Default))
			}

			errMsg.WriteString("\n")
		}

		errMsg.WriteString(fmt.Sprintf("\n使用方法: %s %s\n", c.name, c.usage))
		errMsg.WriteString(fmt.Sprintf("运行 '%s --help' 查看详细帮助", c.name))

		return fmt.Errorf("%s", errMsg.String())
	}

	// 其他类型的错误，返回原始错误信息
	return fmt.Errorf("参数解析错误: %v", err)
}
