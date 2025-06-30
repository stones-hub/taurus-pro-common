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
	"log"
	"net"
	"os"
	"runtime/debug"
	"strings"
)

// PanicRecover 有时候可能需要自己手动解决一些panic
func PanicRecover(stack bool) {
	if err := recover(); err != nil {
		// Check for a broken connection, as it is not really a
		// condition that warrants a panic stack trace.
		var brokenPipe bool
		if ne, ok := err.(*net.OpError); ok {
			if se, ok := ne.Err.(*os.SyscallError); ok {
				if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
					brokenPipe = true
				}
			}
		}

		if brokenPipe {
			log.Printf("[Recovery brokenPipe]: %+v \n\n", err)
			return
		}

		if stack {
			log.Printf("[Recovery from panic]: %+v \n %s \n", err, string(debug.Stack()))
		} else {
			log.Printf("[Recovery from panic]: %+v \n\n", err)
		}
	}
}
