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
	"io"
	"log"
	"os/exec"
	"strings"
)

// CmdExec (COMMAND_NAME, workPath+SCRIPT_PATH, "-h", c.Host, "-p", strconv.Itoa(c.Port), "-t", strconv.Itoa(c.Timeout), "-e", c.Encode, "-c", "get_users")
func CmdExec(name string, args ...string) (string, error) {
	var (
		err    error
		stdout io.ReadCloser
		output []byte
	)
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %v\n", err)
		}
	}()
	cmd := exec.Command(name, args...)
	// 设置标准输出
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create stdout pipe: %v\n", err)
		return "", err
	}
	// 启动命令
	if err = cmd.Start(); err != nil {
		log.Printf("Failed to start command: %v\n", err)
		return "", err
	}
	// 读取标准输出
	output, err = io.ReadAll(stdout)
	if err != nil {
		log.Printf("Failed to read from stdout: %v\n", err)
		return "", err
	}

	// 关闭标准输出
	if err = stdout.Close(); err != nil {
		log.Printf("Failed to close stdout: %v\n", err)
		return "", err
	}

	// 等待命令完成
	if err = cmd.Wait(); err != nil {
		log.Printf("Failed to wait for command: %v\n", err)
		return "", err
	}

	// 解析输出结果
	return strings.TrimSpace(string(output)), nil
}
