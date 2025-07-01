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

// Level defines logging priority.
type Level int

const (
	// Debug level for detailed troubleshooting
	Debug Level = iota
	// Info level for general operational entries
	Info
	// Warn level for non-critical issues
	Warn
	// Error level for errors that should be addressed
	Error
	// Fatal level for critical issues that require immediate attention
	Fatal
	// None level to disable logging
	None
)

// String returns the string representation of Level
func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	case None:
		return ""
	default:
		return fmt.Sprintf("LEVEL(%d)", l)
	}
}

// OutputType defines where logs should be written
type OutputType string

const (
	// Console writes logs to standard output
	Console OutputType = "console"
	// File writes logs to a file
	File OutputType = "file"
)

// colors defines ANSI color codes for console output
var colors = map[Level]string{
	Debug: "\033[36m", // Cyan
	Info:  "\033[32m", // Green
	Warn:  "\033[33m", // Yellow
	Error: "\033[31m", // Red
	Fatal: "\033[35m", // Purple
}

// Config defines logger configuration options
type Config struct {
	Name       string     // Logger name
	Prefix     string     // Log prefix (only used for file output)
	Level      Level      // Minimum logging level
	Output     OutputType // Output destination (console/file)
	FilePath   string     // Log file path (absolute or relative)
	MaxSize    int        // Maximum size of log file in MB
	MaxBackups int        // Maximum number of old log files to retain
	MaxAge     int        // Maximum number of days to retain old log files
	Compress   bool       // Whether to compress old log files
	Formatter  string     // Name of custom log formatter
}

// Manager manages multiple named loggers
type Manager struct {
	loggers map[string]*Logger
	mu      sync.RWMutex
}

// Logger represents a single logger instance
type Logger struct {
	config Config
	writer io.Writer
	mu     sync.Mutex
}

// NewManager creates a new logger manager
func NewManager() *Manager {
	return &Manager{
		loggers: make(map[string]*Logger),
	}
}

// Configure initializes loggers from configurations
func (m *Manager) Configure(configs []Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cfg := range configs {
		if err := m.configure(cfg); err != nil {
			return fmt.Errorf("failed to configure logger %s: %w", cfg.Name, err)
		}
	}
	return nil
}

// configure creates or updates a single logger
func (m *Manager) configure(cfg Config) error {
	logger, err := newLogger(cfg)
	if err != nil {
		return err
	}
	m.loggers[cfg.Name] = logger
	return nil
}

// GetLogger returns a named logger
func (m *Manager) GetLogger(name string) (*Logger, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	logger, exists := m.loggers[name]
	if !exists {
		return nil, fmt.Errorf("logger %s not found", name)
	}
	return logger, nil
}

// newLogger creates a new Logger instance
func newLogger(cfg Config) (*Logger, error) {
	var writer io.Writer

	if cfg.Output == File {
		logPath, err := resolveLogPath(cfg.FilePath)
		if err != nil {
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
	}, nil
}

// resolveLogPath resolves the absolute path for log file
func resolveLogPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	absPath := filepath.Join(dir, path)
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create log directory: %w", err)
	}

	return absPath, nil
}

// log formats and writes a log message
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.config.Level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown"
		line = 0
	}

	msg := fmt.Sprintf(format, args...)
	formatted := GetFormatter(l.config.Formatter).Format(level, file, line, msg)

	if l.config.Output == Console {
		color := colors[level]
		fmt.Fprintln(l.writer, color+formatted+"\033[0m")
	} else {
		fmt.Fprintln(l.writer, formatted)
	}

	if level == Fatal {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(Debug, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(Info, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(Warn, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(Error, format, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(Fatal, format, args...)
}

// SetLevel changes the logging level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Level = level
}

// SetFormatter changes the log formatter
func (l *Logger) SetFormatter(formatter string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Formatter = formatter
}
