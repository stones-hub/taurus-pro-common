# 命令行工具示例

这是一个基于 `taurus-pro-common` 命令行模块的完整示例，展示了如何构建功能丰富的命令行工具。

## 🚀 快速开始

### 构建工具

```bash
# 使用构建脚本
./build.sh

# 或手动构建
go build -o example-cli .
```

### 运行工具

```bash
# 显示帮助
./example-cli help

# 显示特定命令的帮助
./example-cli user --help
./example-cli file --help
./example-cli config --help
./example-cli database --help
./example-cli network --help
```

## 📋 功能特性

### 1. 用户管理命令 (`user`)

支持完整的用户信息管理，包括所有数据类型。

**选项:**
- `--name, -n` (必填): 用户名
- `--email, -e`: 邮箱地址 (默认: user@example.com)
- `--age, -a`: 年龄 (默认: 25)
- `--active, -A`: 是否激活 (默认: true)
- `--score, -s`: 用户评分 (默认: 85.5)
- `--verbose, -v`: 详细输出 (默认: false)
- `--roles, -r`: 用户角色 (默认: user)
- `--department, -d`: 所属部门 (默认: 技术部)
- `--level, -l`: 用户级别 (默认: 1)
- `--verified, -V`: 是否已验证 (默认: false)
- `--salary, -S`: 薪资 (默认: 15000.0)

**示例:**
```bash
# 创建用户
./example-cli user --name 张三 --email zhangsan@example.com --age 30 --verbose

# 使用短选项
./example-cli user -n 李四 -e lisi@example.com -a 28 -v

# 设置薪资和评分
./example-cli user --name 王五 --salary 25000 --score 95.5
```

### 2. 文件操作命令 (`file`)

支持复杂的文件操作，包括位置参数。

**选项:**
- `--source, -s` (必填): 源文件路径
- `--destination, -d`: 目标路径 (默认: ./)
- `--force, -f`: 强制执行 (默认: false)
- `--recursive, -r`: 递归处理 (默认: false)
- `--timeout, -t`: 操作超时时间 (默认: 30)
- `--compress, -c`: 压缩文件 (默认: false)
- `--encrypt, -E`: 加密文件 (默认: false)
- `--backup, -b`: 创建备份 (默认: false)
- `--parallel, -p`: 并行度 (默认: 1)
- `--chunk_size, -C`: 块大小 (默认: 10)

**示例:**
```bash
# 基本文件操作
./example-cli file --source /path/to/source --destination /path/to/dest

# 复杂操作
./example-cli file -s /tmp/data -d /backup -f -r -c -b --parallel 4

# 带位置参数
./example-cli file --source /tmp/file1.txt file2.txt file3.txt
```

### 3. 配置管理命令 (`config`)

支持配置的查看、设置和导出。

**选项:**
- `--key, -k`: 配置键名
- `--value, -v`: 配置值
- `--export, -e`: 导出配置文件 (默认: false)
- `--format, -f`: 导出格式 (默认: json)
- `--pretty, -p`: 美化输出 (默认: true)
- `--encrypt, -E`: 加密配置 (默认: false)
- `--backup, -b`: 备份配置 (默认: false)
- `--validate, -V`: 验证配置 (默认: false)
- `--merge, -m`: 合并配置 (默认: false)
- `--overwrite, -o`: 覆盖配置 (默认: false)

**示例:**
```bash
# 查看所有配置
./example-cli config

# 设置配置
./example-cli config --key app.debug --value true

# 查询配置
./example-cli config --key database.host

# 导出配置
./example-cli config --export --format yaml --pretty
```

### 4. 数据库操作命令 (`database`)

支持数据库连接、备份、恢复等操作。

**选项:**
- `--host, -h`: 数据库主机 (默认: localhost)
- `--port, -p`: 数据库端口 (默认: 3306)
- `--database, -d`: 数据库名称 (默认: test)
- `--username, -u`: 用户名 (默认: root)
- `--password, -P`: 密码 (默认: "")
- `--ssl, -s`: 启用 SSL (默认: false)
- `--timeout, -t`: 连接超时 (默认: 30)
- `--pool_size, -S`: 连接池大小 (默认: 10)
- `--backup, -b`: 备份数据库 (默认: false)
- `--restore, -R`: 恢复数据库 (默认: false)
- `--migrate, -m`: 执行迁移 (默认: false)
- `--optimize, -o`: 优化数据库 (默认: false)

**示例:**
```bash
# 连接数据库
./example-cli database --host localhost --port 3306 --database mydb

# 备份数据库
./example-cli database --backup --host db.example.com --username admin

# 执行迁移
./example-cli database --migrate --ssl --timeout 60
```

### 5. 网络工具命令 (`network`)

支持网络诊断和文件传输。

**选项:**
- `--host, -h`: 目标主机 (默认: localhost)
- `--port, -p`: 目标端口 (默认: 80)
- `--timeout, -t`: 超时时间 (默认: 30)
- `--protocol, -P`: 协议 (默认: tcp)
- `--ssl, -s`: 启用 SSL (默认: false)
- `--verbose, -v`: 详细输出 (默认: false)
- `--ping, -i`: 执行 Ping (默认: false)
- `--trace, -T`: 路由跟踪 (默认: false)
- `--scan, -S`: 端口扫描 (默认: false)
- `--download, -d`: 下载文件 (默认: false)
- `--upload, -u`: 上传文件 (默认: false)

**示例:**
```bash
# Ping 测试
./example-cli network --ping --host google.com

# 端口扫描
./example-cli network --scan --port 443 --host example.com

# 路由跟踪
./example-cli network --trace --host github.com --verbose
```

## 🧪 测试

### 运行所有测试

```bash
go test -v .
```

### 运行特定测试

```bash
# 基础功能测试
go test -v -run TestCommandCreation
go test -v -run TestOptionParsing
go test -v -run TestRequiredOptions

# 集成测试
go test -v -run TestUserCommand
go test -v -run TestFileCommand

# 边界条件测试
go test -v -run TestEdgeCases

# 性能测试
go test -v -run TestPerformance

# 并发测试
go test -v -run TestConcurrency

# 错误处理测试
go test -v -run TestErrorHandling
```

### 运行基准测试

```bash
# 所有基准测试
go test -bench=.

# 特定基准测试
go test -bench=BenchmarkOptionParsing
go test -bench=BenchmarkCommandRegistration
go test -bench=BenchmarkCommandLookup
```

### 测试覆盖率

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out

# 查看覆盖率报告
go tool cover -html=coverage.out -o coverage.html
```

## 📁 项目结构

```
example/cmd/
├── main.go                 # 主入口文件
├── commands.go             # 命令实现
├── comprehensive_test.go   # 全面测试用例
├── build.sh               # 构建脚本
└── README.md              # 项目文档
```

## 🔧 开发指南

### 添加新命令

1. 在 `commands.go` 中定义命令结构体：

```go
type NewCommand struct {
    *cmd.BaseCommand
}
```

2. 实现 `Run` 方法：

```go
func (c *NewCommand) Run(args []string) error {
    ctx, err := c.ParseOptions(args)
    if err != nil {
        return err
    }
    
    // 实现命令逻辑
    return nil
}
```

3. 在 `main.go` 中注册命令：

```go
newCmd, err := cmd.NewBaseCommand(
    "new",
    "新命令描述",
    "[options]",
    []cmd.Option{
        // 定义选项
    },
)
if err != nil {
    // 处理错误
}
cmdManager.Register(&NewCommand{BaseCommand: newCmd})
```

### 选项类型

支持以下选项类型：

- `cmd.OptionTypeString`: 字符串类型
- `cmd.OptionTypeInt`: 整数类型
- `cmd.OptionTypeBool`: 布尔类型
- `cmd.OptionTypeFloat`: 浮点数类型

### 选项属性

每个选项支持以下属性：

- `Name`: 选项名称（必填）
- `Shorthand`: 短选项名（可选）
- `Description`: 选项描述
- `Type`: 选项类型
- `Required`: 是否必填（默认: false）
- `Default`: 默认值

## 🎯 最佳实践

### 1. 命令设计

- 使用清晰的命令名称
- 提供详细的描述信息
- 合理使用必填和可选选项
- 支持短选项以提高用户体验

### 2. 选项设计

- 使用有意义的选项名称
- 提供合理的默认值
- 使用短选项简化输入
- 添加详细的描述信息

### 3. 错误处理

- 提供清晰的错误信息
- 验证输入参数
- 处理边界情况
- 使用适当的退出码

### 4. 测试

- 编写全面的单元测试
- 测试边界条件
- 测试错误情况
- 编写基准测试

## 🚨 注意事项

1. **短选项限制**: 短选项名只能是单个字符，最多支持26个带短选项的参数
2. **并发安全**: 命令管理器是线程安全的，支持并发访问
3. **内存管理**: 大量选项时注意内存使用
4. **性能考虑**: 复杂命令可能需要优化性能

## 📄 许可证

本项目基于 MIT 许可证开源。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📞 支持

如有问题，请通过以下方式联系：

- 提交 Issue
- 发送邮件
- 查看文档

---

**享受使用这个强大的命令行工具框架！** 🎉 