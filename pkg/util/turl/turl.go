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

package turl

import (
	"strconv"
	"strings"
)

// ReplaceURI 将URI中的数字替换为{num}，用于规范化URI路径
// 参数：
//   - uri: 原始URI字符串，可以包含前后的斜杠
//
// 返回值：
//   - string: 替换后的URI字符串，其中所有独立的数字部分都被替换为{num}
//
// 使用示例：
//
//	// 基本用法
//	// uri := turl.ReplaceURI("/users/123/posts/456")
//	// fmt.Println(uri)  // 输出: users/{num}/posts/{num}
//
//	// 处理前后斜杠
//	// uri = turl.ReplaceURI("api/users/789/")
//	// fmt.Println(uri)  // 输出: api/users/{num}
//
//	// 只替换独立的数字部分
//	// uri = turl.ReplaceURI("/v1/user123/456")
//	// fmt.Println(uri)  // 输出: v1/user123/{num}
//
// 注意事项：
//   - 会自动去除URI前后的斜杠
//   - 只替换路径中独立的数字部分
//   - 不会替换包含在其他字符串中的数字
//   - 主要用于API路径的规范化和分类
func ReplaceURI(uri string) string {
	// 拆分查询字符串
	path := uri
	query := ""
	if idx := strings.Index(uri, "?"); idx >= 0 {
		path = uri[:idx]
		query = uri[idx:]
	}

	// 规范化路径，移除首尾斜杠
	path = strings.Trim(path, "/")
	if path == "" {
		// 只有斜杠或空路径
		return ""
	}

	// 按 / 拆分，忽略空段，处理连续斜杠
	rawParts := strings.Split(path, "/")
	parts := make([]string, 0, len(rawParts))
	for _, seg := range rawParts {
		if seg == "" {
			continue
		}
		// 纯数字段替换为 {num}
		if _, err := strconv.Atoi(seg); err == nil {
			parts = append(parts, "{num}")
		} else {
			parts = append(parts, seg)
		}
	}

	resultPath := strings.Join(parts, "/")
	// 拼回查询字符串
	return resultPath + query
}
