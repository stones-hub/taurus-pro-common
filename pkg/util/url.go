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
	"strconv"
	"strings"
)

// URI中的数字替换
func ReplaceURI(uri string) string {
	// 按/切割
	parts := strings.Split(strings.Trim(uri, "/"), "/")

	// 数字替换为{num}
	replaced := make([]string, 0)
	for i := 0; i < len(parts); i++ {
		if _, err := strconv.Atoi(parts[i]); err == nil { // 数字替换为{num}
			replaced = append(replaced, "{num}")
		} else {
			replaced = append(replaced, parts[i])
		}
	}
	// 用/连接起来
	return strings.Join(replaced, "/")
}
