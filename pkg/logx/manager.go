package logx

import (
	"fmt"
	"log"
	"sync"
)

// Manager manages multiple named loggers
type Manager struct {
	loggers map[string]*Logger
	mu      sync.RWMutex // 保护 loggers map 的并发访问
}

// Configure initializes loggers from configurations.
// If a logger with the same name already exists, it will be updated.
func BuildManager(configs ...LoggerOptions) (*Manager, func(), error) {
	m := &Manager{
		loggers: make(map[string]*Logger),
	}

	// 初始化时不需要加锁，因为还没有其他goroutine访问
	for _, cfg := range configs {
		if _, exists := m.loggers[cfg.Name]; exists {
			return nil, nil, fmt.Errorf("logger %s already exists", cfg.Name)
		}

		logger, err := New(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to configure logger %s: %w", cfg.Name, err)
		}

		m.loggers[cfg.Name] = logger
	}

	// 检查并创建默认logger
	if _, exists := m.loggers[defaultConfig.Name]; !exists {
		logger, err := New(defaultConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to configure default logger: %w", err)
		}
		m.loggers[defaultConfig.Name] = logger
	}

	return m, func() {
		// cleanup 时需要加锁，因为可能有并发访问
		m.mu.Lock()
		defer m.mu.Unlock()
		for _, logger := range m.loggers {
			if err := logger.Close(); err != nil {
				log.Printf("failed to close logger %s: %v\n", logger.config.Name, err)
			}
		}
	}, nil
}

// L returns a logger by name, returns default logger if not found
func (m *Manager) L(name string) *Logger {
	// 读取时需要加读锁
	m.mu.RLock()
	defer m.mu.RUnlock()

	logger, exists := m.loggers[name]
	if !exists {
		return m.loggers[defaultConfig.Name]
	}
	return logger
}

func (m *Manager) LInfo(name string, format string, args ...interface{}) {
	m.L(name).log(Info, 1, format, args...)
}

func (m *Manager) LDebug(name string, format string, args ...interface{}) {
	m.L(name).log(Debug, 1, format, args...)
}

func (m *Manager) LWarn(name string, format string, args ...interface{}) {
	m.L(name).log(Warn, 1, format, args...)
}

func (m *Manager) LError(name string, format string, args ...interface{}) {
	m.L(name).log(Error, 1, format, args...)
}

func (m *Manager) LFatal(name string, format string, args ...interface{}) {
	m.L(name).log(Fatal, 1, format, args...)
}

// LWithSkip logs a message with custom skip level
func (m *Manager) LWithSkip(name string, level Level, skip int, format string, args ...interface{}) {
	m.L(name).log(level, skip, format, args...)
}
