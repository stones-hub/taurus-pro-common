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

// Package logx provides formatting capabilities for the logging system.
package logx

import (
	"fmt"
	"sync"
	"time"
)

// Formatter defines the interface for log message formatting.
type Formatter interface {
	// Format formats a log message with the given parameters.
	// level: the logging level
	// file: the source file where the log was called
	// line: the line number in the source file
	// message: the log message to format
	Format(level Level, file string, line int, message string) string
}

var (
	// formatterRegistry stores registered formatters
	formatterRegistry = make(map[string]Formatter)
	// registryMutex protects concurrent access to formatterRegistry
	registryMutex sync.RWMutex
)

// RegisterFormatter registers a new formatter with the given name.
// If a formatter with the same name already exists, it will be overwritten.
func RegisterFormatter(name string, formatter Formatter) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if formatter == nil {
		panic("formatter cannot be nil")
	}
	formatterRegistry[name] = formatter
}

// GetFormatter returns the formatter with the given name.
// If no formatter is found, returns the default formatter.
func GetFormatter(name string) Formatter {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	if formatter, exists := formatterRegistry[name]; exists {
		return formatter
	}
	return defaultFormatter{}
}

// defaultFormatter provides the default log formatting implementation.
type defaultFormatter struct{}

// Format implements the Formatter interface for defaultFormatter.
func (f defaultFormatter) Format(level Level, file string, line int, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	return fmt.Sprintf("[%s] [%s:%d] [%s] %s",
		timestamp,
		file,
		line,
		level.String(),
		message,
	)
}

// JSONFormatter formats log messages as JSON.
type JSONFormatter struct{}

// Format implements the Formatter interface for JSONFormatter.
func (f JSONFormatter) Format(level Level, file string, line int, message string) string {
	timestamp := time.Now().Format(time.RFC3339)
	return fmt.Sprintf(`{"time":"%s","level":"%s","file":"%s","line":%d,"message":"%s"}`,
		timestamp,
		level.String(),
		file,
		line,
		message,
	)
}

func init() {
	// Register built-in formatters
	RegisterFormatter("default", defaultFormatter{})
	RegisterFormatter("json", JSONFormatter{})
}
