package logx

import (
	"fmt"
	"os"
	"path/filepath"
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
type LoggerOptions struct {
	Name       string     // Logger name √
	Prefix     string     // Log prefix
	Level      Level      // Minimum logging level √
	Output     OutputType // Output destination (console/file) √
	FilePath   string     // Log file path (absolute or relative) √
	MaxSize    int        // Maximum size of log file in MB √
	MaxBackups int        // Maximum number of old log files to retain √
	MaxAge     int        // Maximum number of days to retain old log files √
	Compress   bool       // Whether to compress old log files √
	Formatter  string     // Name of custom log formatter √
}

// resolveLogPath resolves the absolute path for log file
func resolveLogPath(path string) (string, error) {
	// 如果path是绝对路径，则直接返回
	if filepath.IsAbs(path) {
		return path, nil
	}

	// 如果path是相对路径，则获取当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// 使用 Clean 处理 . 和 .. 等特殊路径
	return filepath.Clean(filepath.Join(dir, path)), nil
}

// ensureDirectoryExists checks if the directory exists and is accessible
func ensureDirectoryExists(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("failed to access directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dir)
	}

	// 尝试打开目录以验证权限
	f, err := os.OpenFile(dir, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("insufficient permissions to access directory: %s", err)
	}
	f.Close()

	return nil
}
