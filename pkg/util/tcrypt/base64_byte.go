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

package tcrypt

import "encoding/base64"

// Byte2Base64 将字节数组编码为base64字符串
// 参数：
//   - b: 要编码的字节数组
//
// 返回值：
//   - string: base64编码后的字符串
//
// 使用示例：
//
//	data := []byte("Hello, World!")
//	encoded := tcrypt.Byte2Base64(data)
//	fmt.Println(encoded) // 输出: "SGVsbG8sIFdvcmxkIQ=="
//
// 注意事项：
//   - 使用标准base64编码（RFC 4648）
//   - 编码结果可能包含字符：A-Z、a-z、0-9、+、/
//   - 如果输入长度不是3的倍数，输出会包含填充字符=
//   - 编码后的字符串长度会比原始数据增加约33%
func Byte2Base64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// Base642Byte 将base64字符串解码为字节数组
// 参数：
//   - h: 要解码的base64字符串
//
// 返回值：
//   - []byte: 解码后的字节数组
//   - error: 解码过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	encoded := "SGVsbG8sIFdvcmxkIQ=="
//	decoded, err := tcrypt.Base642Byte(encoded)
//	if err != nil {
//	    log.Printf("解码失败：%v", err)
//	    return
//	}
//	fmt.Println(string(decoded)) // 输出: "Hello, World!"
//
// 注意事项：
//   - 使用标准base64解码（RFC 4648）
//   - 输入必须是有效的base64编码字符串
//   - 如果输入包含非法字符会返回错误
//   - 输入长度必须是4的倍数（可能包含填充字符=）
//   - 解码后的数据长度约为输入字符串长度的3/4
func Base642Byte(h string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(h)
}
