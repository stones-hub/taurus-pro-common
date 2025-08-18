# Taurus Pro Common

[![Go Version](https://img.shields.io/badge/Go-1.24.2+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/stones-hub/taurus-pro-common)](https://goreportcard.com/report/github.com/stones-hub/taurus-pro-common)

一个功能丰富的Go语言通用组件库，为Taurus Pro项目提供核心功能支持。

## 📋 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [安装](#安装)
- [核心组件](#核心组件)
- [使用示例](#使用示例)
- [API文档](#api文档)
- [贡献指南](#贡献指南)
- [许可证](#许可证)

## ✨ 功能特性

- 🚀 **高性能**: 基于Go 1.24.2+，充分利用Go语言的并发特性
- 🛠️ **模块化设计**: 每个功能模块独立，可按需引入
- 📝 **完善的日志系统**: 支持多种输出格式、日志轮转、级别控制
- ⏰ **定时任务管理**: 基于cron的灵活定时任务调度器
- 💬 **企业微信集成**: 完整的企业微信应用对接功能
- 📧 **邮件服务**: 支持SMTP邮件发送，支持附件
- 🔐 **安全工具**: RSA、AES加密，JWT令牌等
- 📊 **命令行工具**: 灵活的命令行框架，支持多种数据类型
- 🔄 **错误恢复**: 优雅的panic恢复机制
- 📁 **文件上传**: 支持阿里云OSS、腾讯云COS等云存储

## 🚀 快速开始

### 基本使用

```go
package main

import (
    "github.com/stones-hub/taurus-pro-common/pkg/logx"
    "github.com/stones-hub/taurus-pro-common/pkg/cron"
)

func main() {
    // 创建日志记录器
    logger, err := logx.New(logx.LoggerOptions{
        Name:   "myapp",
        Level:  logx.Info,
        Output: logx.Console,
    })
    if err != nil {
        panic(err)
    }

    // 创建定时任务管理器
    cronManager := cron.New()
    
    // 添加定时任务
    cronManager.AddFunc("*/5 * * * *", func() {
        logger.Info("定时任务执行中...")
    })

    // 启动定时任务
    cronManager.Start()
    defer cronManager.Stop()

    // 保持程序运行
    select {}
}
```

## 📦 安装

```bash
go get github.com/stones-hub/taurus-pro-common
```

## 🏗️ 核心组件

### 1. 日志系统 (pkg/logx)

灵活的日志记录系统，支持多种输出格式和配置选项。

```go
import "github.com/stones-hub/taurus-pro-common/pkg/logx"

// 创建文件日志记录器
logger, err := logx.New(logx.LoggerOptions{
    Name:       "app",
    Level:      logx.Info,
    Output:     logx.File,
    FilePath:   "./logs/app.log",
    MaxSize:    100,    // MB
    MaxBackups: 10,
    MaxAge:     30,     // 天
    Compress:   true,
})
```

**特性:**
- 支持控制台和文件输出
- 自动日志轮转
- 多种日志级别
- 自定义格式化器
- 线程安全

### 2. 定时任务管理 (pkg/cron)

基于cron表达式的定时任务调度器，支持任务分组和监控。

```go
import "github.com/stones-hub/taurus-pro-common/pkg/cron"

// 创建定时任务管理器
manager := cron.New(cron.WithConcurrencyMode(cron.SkipIfRunning))

// 添加定时任务
taskID := manager.AddFunc("0 */6 * * *", func() {
    // 每6小时执行一次
    log.Println("执行定时任务")
})

// 启动管理器
manager.Start()
defer manager.Stop()
```

**特性:**
- 支持标准cron表达式
- 多种并发控制模式
- 任务分组管理
- 执行指标监控
- 优雅停止支持

### 3. 命令行工具 (pkg/cmd)

灵活的命令行框架，支持多种数据类型和选项验证。

```go
import "github.com/stones-hub/taurus-pro-common/pkg/cmd"

// 创建命令
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
            Default:     10,
        },
    }

    baseCmd, err := cmd.NewBaseCommand("mycmd", "我的命令", "[options] <file>", options)
    if err != nil {
        return nil, err
    }

    return &MyCommand{BaseCommand: baseCmd}, nil
}
```

**特性:**
- 支持字符串、整数、布尔、浮点数等类型
- 自动选项验证
- 内置帮助系统
- 短选项支持
- 位置参数处理

### 4. 企业微信集成 (pkg/util/enterprise_wechat)

完整的企业微信应用对接功能，支持消息发送、素材上传等。

```go
import "github.com/stones-hub/taurus-pro-common/pkg/util"

// 创建企业微信实例
wechat := util.NewEnterWechat(
    "your_corpid",
    1000001,
    "应用名称",
    "your_secret",
    []string{"user1", "user2"},
)

// 发送文本消息
err := wechat.SendTextMessage("Hello, World!")
```

**特性:**
- 自动access_token管理
- 支持多种消息类型
- 素材上传功能
- 用户和部门管理
- 错误处理和重试

### 5. 邮件服务 (pkg/util/email)

支持SMTP的邮件发送服务，支持HTML内容和附件。

```go
import "github.com/stones-hub/taurus-pro-common/pkg/util"

// 配置邮件信息
emailInfo := util.EmailInfo{
    Username:       "your_email@example.com",
    Password:       "your_password",
    Host:           "smtp.example.com",
    Port:           587,
    Encryption:     mail.EncryptionSTARTTLS,
    ConnectTimeout: 10 * time.Second,
    SendTimeout:    30 * time.Second,
}

// 创建邮件内容
content := &util.EmailContent{
    From:    "Your Name <your_email@example.com>",
    Subject: "测试邮件",
    Body:    "<h1>Hello World</h1>",
    File:    []string{"./attachment.pdf"},
}

// 发送邮件
err := util.SendEmail(content, "recipient@example.com", emailInfo)
```

**特性:**
- 支持多种加密方式
- 支持附件发送
- 超时控制
- 长连接支持
- 多种认证方式

### 6. 安全工具 (pkg/util/secure)

提供RSA、AES加密、哈希等安全相关功能。

```go
import "github.com/stones-hub/taurus-pro-common/pkg/util/secure"

// RSA加密
encrypted, err := secure.RSAEncrypt([]byte("hello"), publicKey)

// AES加密
encrypted, err := secure.AESEncrypt([]byte("hello"), key)

// 生成哈希
hash := secure.GenerateHash([]byte("data"), "sha256")
```

**特性:**
- RSA加密/解密
- AES加密/解密
- 多种哈希算法
- Base64编码/解码
- 数字签名

### 7. 文件上传 (pkg/util/upload)

支持多种云存储服务的文件上传功能。

```go
import "github.com/stones-hub/taurus-pro-common/pkg/util/upload"

// 阿里云OSS上传
ossUploader := upload.NewAliyunOSSUploader(
    "your_endpoint",
    "your_access_key",
    "your_secret_key",
    "your_bucket",
)

err := ossUploader.UploadFile("local_file.txt", "remote_file.txt")
```

**特性:**
- 阿里云OSS支持
- 腾讯云COS支持
- 本地文件存储
- 统一上传接口
- 错误重试机制

### 8. 错误恢复 (pkg/recovery)

优雅的panic恢复机制，防止程序崩溃。

```go
import "github.com/stones-hub/taurus-pro-common/pkg/recovery"

// 使用恢复装饰器
func myFunction() {
    defer recovery.Recover()
    
    // 可能发生panic的代码
    panic("something went wrong")
}
```

**特性:**
- 自动panic恢复
- 错误日志记录
- 可配置恢复行为
- 支持自定义处理器

### 9. 通用工具函数 (pkg/utils)

按功能分类组织的通用工具函数集合。

```go
import "github.com/stones-hub/taurus-pro-common/pkg/utils"

// 字符串工具
reversed := utils.ReverseString("Hello")
randomStr := utils.RandString(10)

// 时间工具
timestamp := utils.GetUnixMilliSeconds()
formatted := utils.TimeFormatter(time.Now())

// 验证工具
isValid := utils.CheckEmail("user@example.com")
isValidID := utils.CheckIDCard("110101199001011234")

// 网络工具
localIP, _ := utils.GetLocalIP()
isOpen := utils.IsPortOpen("localhost", 8080, 5*time.Second)
```

**特性:**
- 按功能分类组织
- 完整的函数注释
- 支持中文场景
- 线程安全设计
- 丰富的验证功能

## 📚 使用示例

### 完整的Web应用示例

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "github.com/stones-hub/taurus-pro-common/pkg/logx"
    "github.com/stones-hub/taurus-pro-common/pkg/cron"
    "github.com/stones-hub/taurus-pro-common/pkg/recovery"
)

func main() {
    // 初始化日志
    logger, err := logx.New(logx.LoggerOptions{
        Name:       "webapp",
        Level:      logx.Info,
        Output:     logx.File,
        FilePath:   "./logs/webapp.log",
        MaxSize:    100,
        MaxBackups: 10,
    })
    if err != nil {
        log.Fatal(err)
    }

    // 初始化定时任务
    cronManager := cron.New()
    cronManager.AddFunc("0 0 * * *", func() {
        logger.Info("执行每日清理任务")
    })
    cronManager.Start()
    defer cronManager.Stop()

    // 设置HTTP路由
    http.HandleFunc("/", recovery.RecoverHandler(func(w http.ResponseWriter, r *http.Request) {
        logger.Info("收到请求: " + r.URL.Path)
        w.Write([]byte("Hello, Taurus Pro!"))
    }))

    // 启动HTTP服务
    logger.Info("启动HTTP服务在端口8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 命令行工具示例

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/stones-hub/taurus-pro-common/pkg/cmd"
)

type BuildCommand struct {
    *cmd.BaseCommand
}

func NewBuildCommand() (*BuildCommand, error) {
    options := []cmd.Option{
        {
            Name:        "output",
            Shorthand:   "o",
            Description: "输出文件路径",
            Type:        cmd.OptionTypeString,
            Required:    true,
        },
        {
            Name:        "verbose",
            Shorthand:   "v",
            Description: "详细输出",
            Type:        cmd.OptionTypeBool,
        },
    }

    baseCmd, err := cmd.NewBaseCommand("build", "构建项目", "[options]", options)
    if err != nil {
        return nil, err
    }

    return &BuildCommand{BaseCommand: baseCmd}, nil
}

func (c *BuildCommand) Run(args []string) error {
    ctx, err := c.ParseOptions(args)
    if err != nil {
        return err
    }

    output := ctx.Options["output"].(string)
    verbose := ctx.Options["verbose"].(bool)

    if verbose {
        fmt.Printf("开始构建项目，输出到: %s\n", output)
    }

    fmt.Println("构建完成!")
    return nil
}

func main() {
    manager := cmd.NewManager()
    
    buildCmd, err := NewBuildCommand()
    if err != nil {
        panic(err)
    }

    err = manager.Register(buildCmd)
    if err != nil {
        panic(err)
    }

    if err := manager.Run(); err != nil {
        fmt.Printf("错误: %v\n", err)
        os.Exit(1)
    }
}
```

## 🔧 配置选项

### 日志配置

```go
type LoggerOptions struct {
    Name       string // 日志记录器名称
    Prefix     string // 日志前缀
    Level      Level  // 日志级别
    Output     Output // 输出方式
    FilePath   string // 文件路径
    MaxSize    int    // 最大文件大小(MB)
    MaxBackups int    // 最大备份文件数
    MaxAge     int    // 最大保留天数
    Compress   bool   // 是否压缩
    Formatter  string // 格式化器
}
```

### 定时任务配置

```go
type Options struct {
    ConcurrencyMode ConcurrencyMode // 并发控制模式
    EnableSeconds   bool            // 是否启用秒级精度
    Location        *time.Location  // 时区设置
    Logger          *log.Logger     // 日志记录器
}
```

## 🧪 测试

运行测试套件：

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/logx/...
go test ./pkg/cron/...
go test ./pkg/cmd/...

# 运行性能测试
go test -bench=. ./pkg/...

# 生成测试覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📖 API文档

详细的API文档请参考各个包的GoDoc：

- [logx包](https://pkg.go.dev/github.com/stones-hub/taurus-pro-common/pkg/logx)
- [cron包](https://pkg.go.dev/github.com/stones-hub/taurus-pro-common/pkg/cron)
- [cmd包](https://pkg.go.dev/github.com/stones-hub/taurus-pro-common/pkg/cmd)
- [util包](https://pkg.go.dev/github.com/stones-hub/taurus-pro-common/pkg/util)

## 🤝 贡献指南

我们欢迎所有形式的贡献！请查看以下指南：

### 提交Issue

- 使用清晰的标题描述问题
- 提供详细的复现步骤
- 包含环境信息和错误日志

### 提交Pull Request

1. Fork项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建Pull Request

### 代码规范

- 遵循Go语言官方代码规范
- 添加适当的测试用例
- 更新相关文档
- 确保所有测试通过

## 📄 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。

## 📞 联系方式

- 作者: yelei
- 邮箱: 61647649@qq.com
- 项目地址: https://github.com/stones-hub/taurus-pro-common

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者和用户！

---

**Taurus Pro Common** - 让Go开发更简单、更高效！ 
