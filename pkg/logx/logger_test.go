package logx

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "logger-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取工作目录失败: %v", err)
	}

	tests := []struct {
		name    string
		config  LoggerOptions
		wantErr bool
	}{
		{
			name: "控制台日志",
			config: LoggerOptions{
				Name:   "console",
				Output: Console,
				Level:  Info,
			},
			wantErr: false,
		},
		{
			name: "文件日志-绝对路径",
			config: LoggerOptions{
				Name:     "abs-path",
				Output:   File,
				FilePath: filepath.Join(tempDir, "test.log"),
				Level:    Info,
			},
			wantErr: false,
		},
		{
			name: "文件日志-相对路径",
			config: LoggerOptions{
				Name:     "rel-path",
				Output:   File,
				FilePath: "test.log",
				Level:    Info,
			},
			wantErr: false,
		},
		{
			name: "文件日志-目录不存在",
			config: LoggerOptions{
				Name:     "non-exist-dir",
				Output:   File,
				FilePath: filepath.Join(tempDir, "non-exist-dir", "test.log"),
				Level:    Info,
			},
			wantErr: false, // 目录不存在会自动创建
		},
		{
			name: "空路径",
			config: LoggerOptions{
				Name:   "empty-path",
				Output: File,
			},
			wantErr: true, // 文件路径为空应该返回错误
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. 测试创建 logger
			logger, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			defer logger.Close()

			// 2. 测试日志级别过滤
			logger.Debug("debug message") // 不应该出现
			logger.Info("info message")   // 应该出现
			logger.Warn("warn message")   // 应该出现
			logger.Error("error message") // 应该出现

			// 3. 如果是文件日志，验证文件内容
			if tt.config.Output == File {
				// 获取日志文件的绝对路径
				var logPath string
				if filepath.IsAbs(tt.config.FilePath) {
					logPath = tt.config.FilePath
				} else {
					logPath = filepath.Join(wd, tt.config.FilePath)
				}

				// 等待一小段时间确保日志写入
				time.Sleep(time.Millisecond * 100)

				content, err := os.ReadFile(logPath)
				if err != nil {
					t.Errorf("读取日志文件失败: %v", err)
					return
				}
				logContent := string(content)

				// 验证日志级别过滤
				if strings.Contains(logContent, "debug message") {
					t.Error("不应该包含debug级别的日志")
				}
				if !strings.Contains(logContent, "info message") {
					t.Error("应该包含info级别的日志")
				}
				if !strings.Contains(logContent, "warn message") {
					t.Error("应该包含warn级别的日志")
				}
				if !strings.Contains(logContent, "error message") {
					t.Error("应该包含error级别的日志")
				}

				// 4. 测试动态修改日志级别
				logger.SetLevel(Debug)
				logger.Debug("debug enabled")

				// 等待日志写入
				time.Sleep(time.Millisecond * 100)

				content, err = os.ReadFile(logPath)
				if err != nil {
					t.Errorf("读取日志文件失败: %v", err)
					return
				}
				logContent = string(content)

				if !strings.Contains(logContent, "debug enabled") {
					t.Error("修改日志级别后应该包含debug级别的日志")
				}

				// 清理测试生成的日志文件
				os.Remove(logPath)
			}
		})
	}
}

// TestLoggerConcurrent 测试并发写入日志
func TestLoggerConcurrent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "logger-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logPath := filepath.Join(tempDir, "concurrent.log")
	logger, err := New(LoggerOptions{
		Name:     "test",
		Level:    Debug,
		Output:   File,
		FilePath: logPath,
	})
	if err != nil {
		t.Fatalf("创建logger失败: %v", err)
	}
	defer logger.Close()

	// 并发写入日志
	const goroutines = 10
	const messagesPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Info("goroutine %d message %d", id, j)
			}
		}(i)
	}

	wg.Wait()

	// 等待日志写入
	time.Sleep(time.Millisecond * 100)

	// 验证日志文件
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Errorf("读取日志文件失败: %v", err)
		return
	}

	// 检查日志行数
	lines := strings.Split(string(content), "\n")
	expectedLines := goroutines * messagesPerGoroutine
	if len(lines)-1 != expectedLines { // -1 是因为最后一行是空行
		t.Errorf("日志行数不匹配，期望 %d 行，实际 %d 行", expectedLines, len(lines)-1)
	}
}

func TestManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "manager-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name     string
		configs  []LoggerOptions
		wantErr  bool
		testFunc func(*Manager) error
	}{
		{
			name: "正常配置",
			configs: []LoggerOptions{
				{
					Name:   "console",
					Output: Console,
				},
				{
					Name:     "file",
					Output:   File,
					FilePath: filepath.Join(tempDir, "test1.log"),
				},
			},
			wantErr: false,
			testFunc: func(m *Manager) error {
				m.LInfo("console", "test message")
				m.LDebug("file", "debug message")
				return nil
			},
		},
		{
			name: "重复名称",
			configs: []LoggerOptions{
				{
					Name:   "same",
					Output: Console,
				},
				{
					Name:   "same",
					Output: Console,
				},
			},
			wantErr:  true,
			testFunc: nil,
		},
		{
			name:    "空配置",
			configs: []LoggerOptions{},
			wantErr: false,
			testFunc: func(m *Manager) error {
				// 应该使用默认logger
				m.LInfo("non-existent", "using default logger")
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, cleanup, err := BuildManager(tt.configs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			defer cleanup()

			if tt.testFunc != nil {
				if err := tt.testFunc(manager); err != nil {
					t.Errorf("测试函数执行失败: %v", err)
				}
			}
		})
	}
}

func TestManagerConcurrent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "manager-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	logPath := filepath.Join(tempDir, "manager.log")
	manager, cleanup, err := BuildManager(
		LoggerOptions{
			Name:   "test1",
			Level:  Debug,
			Output: Console,
		},
		LoggerOptions{
			Name:     "test2",
			Level:    Info,
			Output:   File,
			FilePath: logPath,
		},
	)
	if err != nil {
		t.Fatalf("创建manager失败: %v", err)
	}
	defer cleanup()

	var wg sync.WaitGroup
	messageCount := 100
	loggerNames := []string{"test1", "test2", "non-existent"}

	// 并发使用不同的logger写入日志
	for i := 0; i < messageCount; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// 随机选择一个logger
			name := loggerNames[i%len(loggerNames)]
			manager.LInfo(name, "concurrent log %d from %s", i, name)
		}(i)
	}

	wg.Wait()
	time.Sleep(time.Millisecond * 100) // 等待文件写入完成

	// 验证文件日志
	if content, err := os.ReadFile(logPath); err == nil {
		logLines := strings.Split(string(content), "\n")
		// 计算实际写入的日志行数（去除空行）
		actualLines := 0
		for _, line := range logLines {
			if strings.TrimSpace(line) != "" {
				actualLines++
			}
		}
		expectedLines := messageCount / len(loggerNames) // test2的日志行数
		if actualLines != expectedLines {
			t.Errorf("文件日志行数不匹配，期望 %d，实际 %d", expectedLines, actualLines)
		}
	} else {
		t.Errorf("读取日志文件失败: %v", err)
	}
}
