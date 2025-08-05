# 命令行工具使用指南

## 概述

本命令行工具提供了一个灵活、易用的命令行框架，支持多种数据类型、选项验证、帮助生成等功能。工具采用模块化设计，可以轻松扩展和定制。

## 核心功能

### 1. 命令接口 (Command Interface)

所有命令都必须实现 `Command` 接口：

```go
type Command interface {
    Name() string        // 返回命令名称
    Description() string // 返回命令描述
    Help() string        // 返回详细帮助信息
    Run(args []string) error // 执行命令
}
```

### 2. 选项类型支持

支持以下数据类型：
- `OptionTypeString` - 字符串类型
- `OptionTypeInt` - 整数类型
- `OptionTypeInt64` - 长整数类型
- `OptionTypeBool` - 布尔类型
- `OptionTypeFloat` - 浮点数类型

### 3. 命令管理器

提供命令注册、查找、执行等功能的统一管理。

## 基本用法

### 创建命令

```go
package main

import (
    "fmt"
    "github.com/your-project/pkg/cmd"
)

// 自定义命令
type MyCommand struct {
    *cmd.BaseCommand
}

func NewMyCommand() (*MyCommand, error) {
    options := []cmd.Option{
        {
            Name:        "name",
            Shorthand:   "n",
            Description: "指定名称",
            Type:        cmd.OptionTypeString,
            Required:    true,
        },
        {
            Name:        "count",
            Shorthand:   "c",
            Description: "指定数量",
            Type:        cmd.OptionTypeInt,
            Required:    false,
            Default:     10,
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

    baseCmd, err := cmd.NewBaseCommand("mycmd", "我的自定义命令", "[options] <file>", options)
    if err != nil {
        return nil, err
    }

    return &MyCommand{BaseCommand: baseCmd}, nil
}

func (c *MyCommand) Run(args []string) error {
    // 解析选项
    ctx, err := c.ParseOptions(args)
    if err != nil {
        return err
    }

    // 获取选项值
    name := ctx.Options["name"].(string)
    count := ctx.Options["count"].(int)
    verbose := ctx.Options["verbose"].(bool)

    // 执行命令逻辑
    if verbose {
        fmt.Printf("处理文件: %s, 数量: %d\n", name, count)
    }

    // 处理位置参数
    for _, arg := range ctx.Args {
        fmt.Printf("处理文件: %s\n", arg)
    }

    return nil
}

func main() {
    // 创建命令管理器
    manager := cmd.NewManager()

    // 创建并注册命令
    myCmd, err := NewMyCommand()
    if err != nil {
        panic(err)
    }

    err = manager.Register(myCmd)
    if err != nil {
        panic(err)
    }

    // 运行管理器
    if err := manager.Run(); err != nil {
        fmt.Printf("错误: %v\n", err)
        os.Exit(1)
    }
}
```

## 选项使用指南

### 1. 字符串选项

```bash
# 正确用法
myapp mycmd --name "test.txt"
myapp mycmd -n "test.txt"

# 错误用法
myapp mycmd --name test.txt  # 如果文件名包含空格，需要用引号
```

### 2. 整数选项

```bash
# 正确用法
myapp mycmd --count 5
myapp mycmd -c 5

# 错误用法
myapp mycmd --count "5"  # 不需要引号
myapp mycmd --count 5.5  # 不能使用浮点数
```

### 3. 布尔选项 ⭐ 重要

**布尔选项是最容易出错的地方！**

```bash
# 正确用法 - 布尔选项无需指定值
myapp mycmd --verbose
myapp mycmd -v

# 错误用法 - 不要为布尔选项指定值
myapp mycmd --verbose true   # ❌ 错误！
myapp mycmd --verbose false  # ❌ 错误！
myapp mycmd -v true          # ❌ 错误！
```

**布尔选项的工作原理：**
- 指定选项名表示 `true`
- 不指定选项名表示 `false`（或默认值）

### 4. 浮点数选项

```bash
# 正确用法
myapp mycmd --price 3.14
myapp mycmd --price 100.0

# 错误用法
myapp mycmd --price "3.14"  # 不需要引号
```

### 5. 长整数选项

```bash
# 正确用法
myapp mycmd --size 1234567890
myapp mycmd --size 0

# 错误用法
myapp mycmd --size 123.456  # 不能使用浮点数
```

## 常见错误和解决方案

### 1. 布尔选项错误

**错误示例：**
```bash
myapp mycmd --verbose true
# 错误: 未定义的选项 'true'
```

**原因：** 布尔选项不需要值，`true` 被当作未知选项处理。

**正确用法：**
```bash
myapp mycmd --verbose
```

### 2. 必填选项缺失

**错误示例：**
```bash
myapp mycmd file.txt
# 错误: 选项 --name 是必填的
```

**解决方案：**
```bash
myapp mycmd --name "test" file.txt
```

### 3. 未定义选项

**错误示例：**
```bash
myapp mycmd --unknown value
# 错误: 未定义的选项 '--unknown'
```

**解决方案：** 检查帮助信息，使用正确的选项名。

### 4. 选项值类型错误

**错误示例：**
```bash
myapp mycmd --count "abc"
# 错误: 参数解析错误: invalid value "abc" for flag -count: parse error
```

**解决方案：** 确保选项值的类型正确。

### 5. 短选项错误

**错误示例：**
```bash
myapp mycmd -ab  # 如果 -a 和 -b 是不同选项
# 错误: 未定义的选项 '-ab'
```

**正确用法：**
```bash
myapp mycmd -a -b
```

## 帮助系统

### 1. 查看所有命令

```bash
myapp
myapp help
```

输出示例：
```
使用方法:
  myapp <command> [arguments]

可用命令:
  build     构建项目
  test      运行测试
  deploy    部署应用

运行 'myapp <command> --help' 查看命令的详细信息
```

### 2. 查看命令详细帮助

```bash
myapp mycmd --help
myapp mycmd -h
```

输出示例：
```
我的自定义命令

用法:
  mycmd [options] <file>

选项:
  --name, -n        指定名称 (必填)
  --count, -c       指定数量 (默认: 10)
  --verbose, -v     详细输出 (默认: false)
```

### 3. 未知命令建议

当输入未知命令时，系统会提供相似命令建议：

```bash
myapp bui
# 未知命令: bui
#
# 您可能想要运行以下命令之一:
#   build
#
# 运行 'myapp help' 查看可用命令
```

## 最佳实践

### 1. 选项命名

- 使用描述性的选项名
- 短选项使用单个字符
- 保持命名一致性

```go
options := []cmd.Option{
    {
        Name:        "output",
        Shorthand:   "o",
        Description: "输出文件路径",
        Type:        cmd.OptionTypeString,
        Required:    true,
    },
    {
        Name:        "format",
        Shorthand:   "f",
        Description: "输出格式",
        Type:        cmd.OptionTypeString,
        Required:    false,
        Default:     "json",
    },
}
```

### 2. 错误处理

```go
func (c *MyCommand) Run(args []string) error {
    ctx, err := c.ParseOptions(args)
    if err != nil {
        // 错误信息已经格式化，直接返回
        return err
    }

    // 验证位置参数
    if len(ctx.Args) == 0 {
        return fmt.Errorf("请指定至少一个文件")
    }

    // 执行命令逻辑
    return nil
}
```

### 3. 帮助信息

提供清晰、详细的帮助信息：

```go
func (c *MyCommand) Help() string {
    return `我的自定义命令

这个命令用于处理文件，支持多种格式和选项。

用法:
  mycmd [options] <file1> [file2] ...

示例:
  mycmd --name "output.txt" --count 5 file1.txt file2.txt
  mycmd -n "output.txt" -c 5 -v file1.txt

更多信息请访问: https://example.com/docs/mycmd`
}
```

### 4. 默认值设置

为常用选项设置合理的默认值：

```go
options := []cmd.Option{
    {
        Name:        "timeout",
        Description: "操作超时时间（秒）",
        Type:        cmd.OptionTypeInt,
        Required:    false,
        Default:     30,  // 30秒默认超时
    },
    {
        Name:        "retry",
        Description: "重试次数",
        Type:        cmd.OptionTypeInt,
        Required:    false,
        Default:     3,   // 默认重试3次
    },
}
```

## 测试

运行测试：

```bash
# 运行所有测试
go test ./pkg/cmd/...

# 运行特定测试
go test ./pkg/cmd/ -run TestParseOptions

# 运行性能测试
go test ./pkg/cmd/ -bench=.
```

## 扩展指南

### 1. 添加新的选项类型

如果需要支持新的数据类型，可以扩展 `OptionType` 枚举和相关处理逻辑。

### 2. 自定义验证

可以在 `Run` 方法中添加自定义的选项验证逻辑。

### 3. 子命令支持

可以通过嵌套命令管理器来实现子命令功能。

## 故障排除

### 1. 选项解析失败

检查：
- 选项名是否正确
- 选项值类型是否匹配
- 是否缺少必填选项

### 2. 命令未找到

检查：
- 命令是否正确注册
- 命令名是否正确拼写
- 是否使用了正确的程序名

### 3. 帮助信息不显示

检查：
- 是否正确调用了 `--help` 或 `-h`
- 命令是否正确实现了 `Help()` 方法

## 版本兼容性

- Go 1.16+
- 支持所有主流操作系统
- 向后兼容，新版本不会破坏现有API

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个工具。

## 许可证

[许可证信息] 