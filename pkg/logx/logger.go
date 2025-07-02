// Copyright (c) 2025 Taurus Team. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package logx provides a flexible and extensible logging system.
// It supports multiple named loggers, different output formats,
// log rotation, and various log levels.
package logx

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/natefinch/lumberjack"
)

var defaultConfig = LoggerOptions{
	Name:       "default",
	Prefix:     "",
	Level:      Info,
	Output:     Console,
	FilePath:   "",
	MaxSize:    100,
	MaxBackups: 10,
	MaxAge:     30,
	Compress:   false,
	Formatter:  "default",
}

// Logger represents a single logger instance
type Logger struct {
	config LoggerOptions
	writer io.Writer
	mu     sync.RWMutex // 保护 config 和 writer 的并发访问
}

// New creates a new Logger instance
func New(cfg LoggerOptions) (*Logger, error) {
	var writer io.Writer

	if cfg.Output == File {
		if cfg.FilePath == "" {
			return nil, fmt.Errorf("file path cannot be empty for file output")
		}

		// 解析日志文件的绝对路径
		logPath, err := resolveLogPath(cfg.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve log path: %w", err)
		}

		// 检查日志文件路径是否是目录
		if info, err := os.Stat(logPath); err == nil && info.IsDir() {
			return nil, fmt.Errorf("log path is a directory: %s", logPath)
		}

		// 获取日志目录
		dir := filepath.Dir(logPath)

		// 1. 目录不存在就创建
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create log directory: %w", err)
			}
		} else if err == nil {
			// 2. 目录存在，检查是否是目录
			info, err := os.Stat(dir)
			if err != nil {
				return nil, fmt.Errorf("failed to access directory: %w", err)
			}
			if !info.IsDir() {
				return nil, fmt.Errorf("path is not a directory: %s", dir)
			}
		} else {
			return nil, fmt.Errorf("failed to access directory: %w", err)
		}

		// 3. 检查目录权限
		if err := ensureDirectoryExists(dir); err != nil {
			return nil, err
		}

		writer = &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
	} else {
		writer = os.Stdout
	}

	return &Logger{
		config: cfg,
		writer: writer,
		mu:     sync.RWMutex{},
	}, nil
}

// log formats and writes a log message
func (l *Logger) log(level Level, extraSkip int, format string, args ...interface{}) {
	l.mu.RLock()
	if level < l.config.Level {
		l.mu.RUnlock()
		return
	}

	_, file, line, ok := runtime.Caller(2 + extraSkip)
	if !ok {
		file = "unknown"
		line = 0
	}

	msg := fmt.Sprintf(format, args...)
	formatted := GetFormatter(l.config.Formatter).Format(level, l.config.Prefix, file, line, msg)
	l.mu.RUnlock()

	// 只在写入时加锁
	l.mu.Lock()
	if l.config.Output == Console {
		color := colors[level]
		fmt.Fprintln(l.writer, color+formatted+"\033[0m")
	} else {
		fmt.Fprintln(l.writer, formatted)
	}
	l.mu.Unlock()

	// 如果日志级别为 Fatal，不退出程序，错误集中处理
	/*
		if level == Fatal {
			os.Exit(1)
		}
	*/
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(Debug, 0, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(Info, 0, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(Warn, 0, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(Error, 0, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(Fatal, 0, format, args...)
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	l.config.Level = level
	l.mu.Unlock()
}

// SetFormatter changes the log formatter
func (l *Logger) SetFormatter(formatter string) {
	l.mu.Lock()
	l.config.Formatter = formatter
	l.mu.Unlock()
}

// Close releases resources held by the logger
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.config.Output == File {
		if closer, ok := l.writer.(*lumberjack.Logger); ok {
			return closer.Close()
		}
	}
	return nil
}
