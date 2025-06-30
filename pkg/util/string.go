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
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// 反转string字符串
func ReverseString(s string) string {
	l := len(s)
	buf := make([]byte, l)
	for i := 0; i < len(s); i++ {
		buf[l-1] = s[i]
		l--
	}
	return string(buf)
}

// 随机字符串
func RandString(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

// Normalize will remove any extra spaces, remove newlines, and trim leading and trailing spaces
func Normalize(s string) string {
	nlRe := regexp.MustCompile(`\r?\r`)
	re := regexp.MustCompile(`\s+`)

	s = nlRe.ReplaceAllString(s, " ")
	s = re.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}

// ChunkSplit input字符串每隔chunkSize个字符添加delimiter,并返回最终的string
func ChunkSplit(input string, chunkSize int, delimiter string) string {
	var chunks []string
	for i := 0; i < len(input); i += chunkSize {
		end := i + chunkSize
		if end > len(input) {
			end = len(input)
		}
		chunks = append(chunks, input[i:end])
	}

	// 如果最后一个块长度小于 chunkSize，则在其后面添加 delimiter
	if len(chunks) > 0 && len(chunks[len(chunks)-1]) < chunkSize {
		chunks[len(chunks)-1] += delimiter
	}

	return strings.Join(chunks, delimiter)
}

// GBK2UTF8 gbkBytes := []byte{0xb0, 0xc0} // "中"字的GBK编码
// 将GBK的 byte转成 UTF-8字符串
func GBK2UTF8(b []byte) (string, error) {
	// 创建一个GBK解码器
	decoder := simplifiedchinese.GBK.NewDecoder()
	// 使用解码器将gbkBytes转换为UTF-8编码的字符串
	utf8String, _, err := transform.String(decoder, string(b))
	if err != nil {
		return "", err
	}
	return utf8String, nil
}

func ParseStringSliceToUint64(s []string) []uint64 {
	iv := make([]uint64, len(s))

	for i, v := range s {
		// 以10进制的方式解析v， 最后保存为64 uint
		iv[i], _ = strconv.ParseUint(v, 10, 64)
	}

	/**
	// 将s 字符串用 base 进制转成 bitSize位的int类型
	strconv.ParseInt(s string, base int, bitSize int)
	// 将s 字符串用 base 进制转成 bitSize位的uint类型
	strconv.ParseUint(s string, base int , bitSize int)
	*/
	return iv
}
