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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	secondMutex    sync.Mutex
	nowSecondCache int64
)

// GetUnixMilliSeconds gets the current unix timestamp in milliseconds
func GetUnixMilliSeconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// GetUnixSeconds gets the current unix timestamp in seconds
func GetUnixSeconds() int64 {
	secondMutex.Lock()
	defer secondMutex.Unlock()

	// 1 second diff
	diff := time.Since(time.Unix(nowSecondCache, 0))
	if diff.Seconds() < 1 {
		return nowSecondCache
	}
	nowSecondCache = time.Now().Unix()
	return nowSecondCache
}

// Seconds2date converts seconds to a date string
func Seconds2date(timestamp int64, format string) string {
	t := time.Unix(timestamp, 0)
	formattedTime := t.Format(format)
	return formattedTime
}

// Milliseconds2date converts milliseconds to a date string
func Milliseconds2date(milliseconds int64, format string) string {
	t := time.UnixMilli(milliseconds)
	return t.Format(format)
}

func TimeFormatter(time time.Time) string {

	// 格式化时间
	return time.Format("2006-01-02 15:04:05")
}

func ParseDuration(d string) (time.Duration, error) {
	d = strings.TrimSpace(d)
	dr, err := time.ParseDuration(d)
	if err == nil {
		return dr, nil
	}
	if strings.Contains(d, "d") {
		index := strings.Index(d, "d")

		hour, _ := strconv.Atoi(d[:index])
		dr = time.Hour * 24 * time.Duration(hour)
		ndr, err := time.ParseDuration(d[index+1:])
		if err != nil {
			return dr, nil
		}
		return dr + ndr, nil
	}

	dv, err := strconv.ParseInt(d, 10, 64)
	return time.Duration(dv), err
}

// 将 ISO 8601 格式的时间字符串转换为更易读的格式，用于日志记录或显示。
// 例如："2024-01-01T00:00:00.000Z" 转换为 "2024-01-01 00:00:00.000"
func ParseTime(input []byte) string {
	var re = regexp.MustCompile(`"((\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2})(?:\.(\d{3}))\d*)(Z|[\+-]\d{2}:\d{2})"`)
	var substitution = "\"$2 $3.$4\""

	for re.Match(input) {
		input = re.ReplaceAll(input, []byte(substitution))
	}
	return string(input)
}
