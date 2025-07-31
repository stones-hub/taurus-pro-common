package main

import (
	"fmt"
	"os"

	"github.com/stones-hub/taurus-pro-common/pkg/cmd"
)

func main() {
	// 创建命令管理器
	cmdManager := cmd.NewManager()

	// 注册用户管理命令
	userCmd, err := cmd.NewBaseCommand(
		"user",
		"用户管理命令 - 创建、查询、更新用户信息，支持所有数据类型",
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
			{
				Name:        "roles",
				Shorthand:   "r",
				Description: "用户角色（逗号分隔）",
				Type:        cmd.OptionTypeString,
				Default:     "user",
			},
			{
				Name:        "department",
				Shorthand:   "d",
				Description: "所属部门",
				Type:        cmd.OptionTypeString,
				Default:     "技术部",
			},
			{
				Name:        "level",
				Shorthand:   "l",
				Description: "用户级别",
				Type:        cmd.OptionTypeInt,
				Default:     1,
			},
			{
				Name:        "verified",
				Shorthand:   "V",
				Description: "是否已验证",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "salary",
				Shorthand:   "S",
				Description: "薪资",
				Type:        cmd.OptionTypeFloat,
				Default:     15000.0,
			},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建用户命令失败: %v\n", err)
		os.Exit(1)
	}

	if err := cmdManager.Register(&UserCommand{BaseCommand: userCmd}); err != nil {
		fmt.Fprintf(os.Stderr, "注册用户命令失败: %v\n", err)
		os.Exit(1)
	}

	// 注册文件操作命令
	fileCmd, err := cmd.NewBaseCommand(
		"file",
		"文件操作命令 - 文件复制、移动、删除等操作，支持复杂选项",
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
			{
				Name:        "timeout",
				Shorthand:   "t",
				Description: "操作超时时间（秒）",
				Type:        cmd.OptionTypeInt,
				Default:     30,
			},
			{
				Name:        "compress",
				Shorthand:   "c",
				Description: "压缩文件",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "encrypt",
				Shorthand:   "E",
				Description: "加密文件",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "backup",
				Shorthand:   "b",
				Description: "创建备份",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "parallel",
				Shorthand:   "p",
				Description: "并行度",
				Type:        cmd.OptionTypeInt,
				Default:     1,
			},
			{
				Name:        "chunk_size",
				Shorthand:   "C",
				Description: "块大小（MB）",
				Type:        cmd.OptionTypeInt,
				Default:     10,
			},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建文件命令失败: %v\n", err)
		os.Exit(1)
	}

	if err := cmdManager.Register(&FileCommand{BaseCommand: fileCmd}); err != nil {
		fmt.Fprintf(os.Stderr, "注册文件命令失败: %v\n", err)
		os.Exit(1)
	}

	// 注册配置管理命令
	configCmd, err := cmd.NewBaseCommand(
		"config",
		"配置管理命令 - 查看、设置、导出配置，支持多种格式",
		"[options]",
		[]cmd.Option{
			{
				Name:        "key",
				Shorthand:   "k",
				Description: "配置键名",
				Type:        cmd.OptionTypeString,
				Default:     "",
			},
			{
				Name:        "value",
				Shorthand:   "v",
				Description: "配置值",
				Type:        cmd.OptionTypeString,
				Default:     "",
			},
			{
				Name:        "export",
				Shorthand:   "e",
				Description: "导出配置文件",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "format",
				Shorthand:   "f",
				Description: "导出格式 (json/yaml/xml)",
				Type:        cmd.OptionTypeString,
				Default:     "json",
			},
			{
				Name:        "pretty",
				Shorthand:   "p",
				Description: "美化输出",
				Type:        cmd.OptionTypeBool,
				Default:     true,
			},
			{
				Name:        "encrypt",
				Shorthand:   "E",
				Description: "加密配置",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "backup",
				Shorthand:   "b",
				Description: "备份配置",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "validate",
				Shorthand:   "V",
				Description: "验证配置",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "merge",
				Shorthand:   "m",
				Description: "合并配置",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "overwrite",
				Shorthand:   "o",
				Description: "覆盖配置",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建配置命令失败: %v\n", err)
		os.Exit(1)
	}

	if err := cmdManager.Register(&ConfigCommand{BaseCommand: configCmd}); err != nil {
		fmt.Fprintf(os.Stderr, "注册配置命令失败: %v\n", err)
		os.Exit(1)
	}

	// 注册数据库操作命令
	dbCmd, err := cmd.NewBaseCommand(
		"database",
		"数据库操作命令 - 连接、备份、恢复、迁移数据库",
		"[options]",
		[]cmd.Option{
			{
				Name:        "host",
				Shorthand:   "h",
				Description: "数据库主机",
				Type:        cmd.OptionTypeString,
				Default:     "localhost",
			},
			{
				Name:        "port",
				Shorthand:   "p",
				Description: "数据库端口",
				Type:        cmd.OptionTypeInt,
				Default:     3306,
			},
			{
				Name:        "database",
				Shorthand:   "d",
				Description: "数据库名称",
				Type:        cmd.OptionTypeString,
				Default:     "test",
			},
			{
				Name:        "username",
				Shorthand:   "u",
				Description: "用户名",
				Type:        cmd.OptionTypeString,
				Default:     "root",
			},
			{
				Name:        "password",
				Shorthand:   "P",
				Description: "密码",
				Type:        cmd.OptionTypeString,
				Default:     "",
			},
			{
				Name:        "ssl",
				Shorthand:   "s",
				Description: "启用 SSL",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "timeout",
				Shorthand:   "t",
				Description: "连接超时（秒）",
				Type:        cmd.OptionTypeInt,
				Default:     30,
			},
			{
				Name:        "pool_size",
				Shorthand:   "S",
				Description: "连接池大小",
				Type:        cmd.OptionTypeInt,
				Default:     10,
			},
			{
				Name:        "backup",
				Shorthand:   "b",
				Description: "备份数据库",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "restore",
				Shorthand:   "R",
				Description: "恢复数据库",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "migrate",
				Shorthand:   "m",
				Description: "执行迁移",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "optimize",
				Shorthand:   "o",
				Description: "优化数据库",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建数据库命令失败: %v\n", err)
		os.Exit(1)
	}

	if err := cmdManager.Register(&DatabaseCommand{BaseCommand: dbCmd}); err != nil {
		fmt.Fprintf(os.Stderr, "注册数据库命令失败: %v\n", err)
		os.Exit(1)
	}

	// 注册网络工具命令
	netCmd, err := cmd.NewBaseCommand(
		"network",
		"网络工具命令 - Ping、路由跟踪、端口扫描、文件传输",
		"[options]",
		[]cmd.Option{
			{
				Name:        "host",
				Shorthand:   "h",
				Description: "目标主机",
				Type:        cmd.OptionTypeString,
				Default:     "localhost",
			},
			{
				Name:        "port",
				Shorthand:   "p",
				Description: "目标端口",
				Type:        cmd.OptionTypeInt,
				Default:     80,
			},
			{
				Name:        "timeout",
				Shorthand:   "t",
				Description: "超时时间（秒）",
				Type:        cmd.OptionTypeInt,
				Default:     30,
			},
			{
				Name:        "protocol",
				Shorthand:   "P",
				Description: "协议 (tcp/udp/http/https)",
				Type:        cmd.OptionTypeString,
				Default:     "tcp",
			},
			{
				Name:        "ssl",
				Shorthand:   "s",
				Description: "启用 SSL",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "verbose",
				Shorthand:   "v",
				Description: "详细输出",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "ping",
				Shorthand:   "i",
				Description: "执行 Ping",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "trace",
				Shorthand:   "T",
				Description: "路由跟踪",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "scan",
				Shorthand:   "S",
				Description: "端口扫描",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "download",
				Shorthand:   "d",
				Description: "下载文件",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
			{
				Name:        "upload",
				Shorthand:   "u",
				Description: "上传文件",
				Type:        cmd.OptionTypeBool,
				Default:     false,
			},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建网络命令失败: %v\n", err)
		os.Exit(1)
	}

	if err := cmdManager.Register(&NetworkCommand{BaseCommand: netCmd}); err != nil {
		fmt.Fprintf(os.Stderr, "注册网络命令失败: %v\n", err)
		os.Exit(1)
	}

	// 运行命令管理器
	if err := cmdManager.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
