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

// Author: yelei
// Email: 61647649@qq.com
// Date: 2025-06-13

package util

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// WriteLine writes a single line to a file
// Parameters:
//   - filename: Path to the target file
//   - b: Byte slice containing the content to write
//
// Returns:
//   - error: Any error that occurred during the operation
func WriteLine(filename string, b []byte) error {
	var (
		err    error
		fd     *os.File
		writer *bufio.Writer
	)
	fd, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer fd.Close()
	writer = bufio.NewWriter(fd)

	_, err = writer.WriteString(string(b) + "\n")
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}

// ReadAll reads the entire content of a file into a string
// Parameters:
//   - filename: Path to the file to read
//
// Returns:
//   - string: The file content
//   - error: Any error that occurred during the operation
func ReadAll(filename string) (string, error) {
	var (
		err     error
		fd      *os.File
		scanner *bufio.Scanner
		content string
	)

	fd, err = os.OpenFile(filename, os.O_RDONLY|os.O_APPEND, os.ModePerm)

	if err != nil {
		return "", err
	}
	defer fd.Close()

	scanner = bufio.NewScanner(fd)

	for scanner.Scan() {
		line := scanner.Text()
		content += line
	}

	if err = scanner.Err(); err != nil {
		return "", err
	}

	if len(content) == 0 {
		return "", fmt.Errorf("file is empty")
	}

	return content, nil
}

// ReadLine reads a file line by line and returns the content as a slice of interfaces
// Parameters:
//   - filename: Path to the file to read
//
// Returns:
//   - []interface{}: Slice containing each line as an interface
//   - error: Any error that occurred during the operation
func ReadLine(filename string) ([]interface{}, error) {
	var (
		err     error
		fd      *os.File
		scanner *bufio.Scanner
		content []interface{}
	)

	fd, err = os.OpenFile(filename, os.O_RDONLY|os.O_APPEND, os.ModePerm)

	if err != nil {
		return nil, err
	}
	defer fd.Close()

	scanner = bufio.NewScanner(fd)

	for scanner.Scan() {
		line := scanner.Text()
		content = append(content, line)
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return content, nil
}

// FetchAllDir recursively traverses a directory and returns all file paths
// Parameters:
//   - path: Root directory path to traverse
//
// Returns:
//   - []string: Slice of file paths found
//   - error: Any error that occurred during the operation
func FetchAllDir(path string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(path, func(filepath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			files = append(files, filepath)
		}
		return nil
	})

	return files, err
}

// WalkDir recursively traverses a directory with a maximum depth limit
// Parameters:
//   - path: Root directory path to traverse
//   - maxDepth: Maximum directory depth to traverse
//
// Returns:
//   - []string: Slice of file paths found within the depth limit
//   - error: Any error that occurred during the operation
func WalkDir(path string, maxDepth int) ([]string, error) {
	var files []string

	// Calculate the depth of the root directory
	rootDepth := len(strings.Split(path, string(os.PathSeparator)))

	err := filepath.WalkDir(path, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate the current path's depth
		currentDepth := len(strings.Split(currentPath, string(os.PathSeparator))) - rootDepth

		// Check if depth exceeds maximum
		if currentDepth > maxDepth {
			return filepath.SkipDir // Skip this directory and its subdirectories
		}

		// If it's a file, add to result slice
		if !d.IsDir() {
			files = append(files, currentPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// CheckPath checks if a path is a file or directory
// Parameters:
//   - path: Path to check
//
// Returns:
//   - bool: true if directory, false if file
//   - error: Any error that occurred during the operation
func CheckPath(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if info.IsDir() {
		return true, nil
	}

	return false, nil
}

// GetCurrentPath gets the directory path of the current file
// Returns:
//   - string: Directory path of the current file
//   - error: Any error that occurred during the operation
func GetCurrentPath() (string, error) {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return "", fmt.Errorf("failed to get caller info: %v", ok)
	}
	dir := filepath.Dir(file)
	return dir, nil
}

// PathExists checks if a path exists and is a directory
// Parameters:
//   - path: Path to check
//
// Returns:
//   - bool: true if path exists and is a directory
//   - error: Any error that occurred during the operation
func PathExists(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return true, nil
		}
		return false, errors.New("file with same name exists")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateDir creates multiple directories
// Parameters:
//   - dirs: Variable number of directory paths to create
//
// Returns:
//   - error: Any error that occurred during the operation
func CreateDir(dirs ...string) (err error) {
	for _, v := range dirs {
		exist, err := PathExists(v)
		if err != nil {
			return err
		}
		if !exist {
			if err := os.MkdirAll(v, os.ModePerm); err != nil {
				return err
			}
		}
	}
	return err
}

// FileMove moves a file from source to destination
// Parameters:
//   - src: Source file path (absolute or relative)
//   - dst: Destination directory path (absolute or relative, must be a directory)
//
// Returns:
//   - error: Any error that occurred during the operation
func FileMove(src string, dst string) (err error) {
	if dst == "" {
		return nil
	}
	src, err = filepath.Abs(src)
	if err != nil {
		return err
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return err
	}
	revoke := false
	dir := filepath.Dir(dst)
Redirect:
	_, err = os.Stat(dir)
	if err != nil {
		err = os.MkdirAll(dir, 0o755)
		if err != nil {
			return err
		}
		if !revoke {
			revoke = true
			goto Redirect
		}
	}
	return os.Rename(src, dst)
}

// FileExist checks if a file exists
// Parameters:
//   - path: Path to check
//
// Returns:
//   - bool: true if file exists, false otherwise
func FileExist(path string) bool {
	fi, err := os.Lstat(path)
	if err == nil {
		return !fi.IsDir()
	}
	return !os.IsNotExist(err)
}
