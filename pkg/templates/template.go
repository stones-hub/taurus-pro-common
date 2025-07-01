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
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Manager handles multiple template objects in a thread-safe manner.
type Manager struct {
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// Config represents the configuration for a template.
type Config struct {
	Name string `json:"name" yaml:"name" toml:"name"` // Template name
	Path string `json:"path" yaml:"path" toml:"path"` // Template directory path
}

// New creates a new template manager instance.
func New() *Manager {
	return &Manager{
		templates: make(map[string]*template.Template),
	}
}

// LoadFromConfigs initializes templates from multiple configurations.
// It returns error if any template loading fails.
func (m *Manager) LoadFromConfigs(configs []Config) error {
	for _, config := range configs {
		if err := m.loadFromConfig(config); err != nil {
			return fmt.Errorf("failed to load template %s: %w", config.Name, err)
		}
	}
	return nil
}

// loadFromConfig loads a single template configuration.
func (m *Manager) loadFromConfig(config Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.templates[config.Name]; exists {
		return fmt.Errorf("template %s already exists", config.Name)
	}

	absPath, err := filepath.Abs(config.Path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("template directory %s does not exist", absPath)
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
