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

// Package templates provides a flexible and thread-safe template management system.
// It supports loading HTML templates from directories and dynamic template addition.
package templates

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Manager manages multiple template objects in a thread-safe manner.
type Manager struct {
	templates map[string]*template.Template
	mu        sync.RWMutex // 保护 templates map 的并发访问
}

// Config represents the configuration for a template.
type TemplateOptions struct {
	Name string `json:"name" yaml:"name" toml:"name"` // Template name
	Path string `json:"path" yaml:"path" toml:"path"` // Template directory path
}

// New creates a new template manager and loads templates from configurations.
func New(configs ...TemplateOptions) (*Manager, func(), error) {
	m := &Manager{
		templates: make(map[string]*template.Template),
	}

	// 初始化加载模板不需要加锁，因为这时还没有其他goroutine访问
	for _, config := range configs {
		if err := m.loadFromConfig(config); err != nil {
			return nil, nil, fmt.Errorf("failed to load template %s: %w", config.Name, err)
		}
	}

	return m, func() {
		if err := m.Close(); err != nil {
			log.Printf("failed to close template manager: %v\n", err)
		} else {
			log.Printf("template manager closed successfully")
		}
	}, nil
}

// LoadFromConfigs initializes templates from multiple configurations.
// It returns error if any template loading fails.
func (m *Manager) LoadFromConfigs(configs ...TemplateOptions) error {
	for _, config := range configs {
		if err := m.loadFromConfig(config); err != nil {
			return fmt.Errorf("failed to load template %s: %w", config.Name, err)
		}
	}
	return nil
}

// loadFromConfig loads a single template configuration.
func (m *Manager) loadFromConfig(config TemplateOptions) error {
	if config.Name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	// 检查模板是否已存在
	if _, exists := m.templates[config.Name]; exists {
		return fmt.Errorf("template %s already exists", config.Name)
	}

	// 解析模板目录路径
	absPath, err := resolvePath(config.Path)
	if err != nil {
		return err
	}

	// 检查目录权限
	if err := ensureWritableDir(absPath); err != nil {
		return err
	}

	return m.loadTemplatesFromDir(config.Name, absPath)
}

// loadTemplatesFromDir loads templates from the specified directory including subdirectories.
func (m *Manager) loadTemplatesFromDir(name, dir string) error {
	tmpl := template.New(name)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			_, err := tmpl.ParseFiles(path)
			if err != nil {
				return fmt.Errorf("failed to parse template file %s: %w", path, err)
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	m.templates[name] = tmpl
	return nil
}

// AddTemplate adds or updates a template dynamically.
// name: the template group name
// templateName: the specific template name within the group
// content: the template content
func (m *Manager) AddTemplate(name, templateName, content string) error {
	// 动态添加模板需要加写锁
	m.mu.Lock()
	defer m.mu.Unlock()

	tmpl, exists := m.templates[name]
	if !exists {
		tmpl = template.New(name)
		m.templates[name] = tmpl
	}

	_, err := tmpl.New(templateName).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template content: %w", err)
	}
	return nil
}

// Render executes the specified template with given data.
// name: the template group name
// templateName: the specific template name within the group
// data: the data to be passed to the template
func (m *Manager) Render(name, templateName string, data interface{}) (string, error) {
	// 渲染时只需要读锁
	m.mu.RLock()
	tmpl, exists := m.templates[name]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("template group %s does not exist", name)
	}

	var sb strings.Builder
	if err := tmpl.ExecuteTemplate(&sb, templateName, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return sb.String(), nil
}

// Close releases any resources held by the template manager
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清理模板映射
	m.templates = nil

	// 如果将来添加了文件监听器，在这里关闭
	// if m.watcher != nil {
	//     m.watcher.Close()
	// }

	return nil
}

// resolvePath resolves the template directory path to absolute path
func resolvePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("template path cannot be empty")
	}

	// 如果是绝对路径，直接返回
	if filepath.IsAbs(path) {
		return path, nil
	}

	// 如果是相对路径，基于当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// 使用 Clean 处理 . 和 .. 等特殊路径
	return filepath.Clean(filepath.Join(dir, path)), nil
}

// ensureWritableDir checks if the directory exists and is writable
func ensureWritableDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("template directory does not exist: %s", path)
		}
		return fmt.Errorf("failed to access template directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("template path must be a directory: %s", path)
	}

	// 尝试打开目录
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("insufficient permissions to access template directory: %s", path)
	}
	f.Close()

	return nil
}
