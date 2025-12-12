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

package errs_test

import (
	stderrors "errors"
	"fmt"
	"log"

	"github.com/stones-hub/taurus-pro-common/pkg/errs"
)

// 示例：基本错误创建和使用
func ExampleNew() {
	err := errs.New(10001, "invalid params")
	fmt.Printf("Code: %d, Message: %s\n", err.Code(), err.Msg())
	// Output: Code: 10001, Message: invalid params
}

// 示例：使用格式化字符串创建错误
func ExampleErrorf() {
	err := errs.Errorf(10002, "invalid params: %s", "username")
	fmt.Printf("Code: %d, Message: %s\n", err.Code(), err.Msg())
	// Output: Code: 10002, Message: invalid params: username
}

// 示例：从任意 error 中提取错误码和消息
func ExampleCode() {
	err := errs.New(10003, "test error")
	code := errs.Code(err)
	msg := errs.Message(err)
	fmt.Printf("Code: %d, Message: %s\n", code, msg)
	// Output: Code: 10003, Message: test error
}

// 示例：错误包装
func ExampleWrap() {
	original := fmt.Errorf("original error")
	wrapped := errs.Wrap(original, 20001, "work failed")

	code := errs.Code(wrapped)
	msg := errs.Message(wrapped)
	fmt.Printf("Code: %d, Message: %s\n", code, msg)
	// Output: Code: 20001, Message: work failed
}

// 示例：实际业务场景
func Example_usage() {
	// 模拟业务函数
	doSomething := func(a, b int) (int, error) {
		if a < 0 {
			return 0, errs.New(10001, "invalid params")
		}
		return a + b, nil
	}

	// 包装错误后继续向上返回
	work := func() error {
		v, err := doSomething(-1, 2)
		if err != nil {
			errcode := errs.Code(err)
			log.Printf("doSomething failed: result = %d", errcode)
			// 包装后继续向上返回
			return errs.Wrap(err, 20001, "work failed")
		}
		_ = v
		return nil
	}

	err := work()
	if err != nil {
		log.Printf("work failed: %v", err)
		log.Printf("work failed: errcode = %d", errs.Code(err))
		log.Printf("work failed: errmsg = %s", errs.Message(err))
	}
}

// 示例：错误比较
func ExampleIs() {
	err1 := errs.New(10001, "error 1")
	err2 := errs.New(10001, "error 2") // 相同错误码
	err3 := errs.New(10002, "error 3") // 不同错误码

	fmt.Printf("err1 == err2 (same code): %v\n", stderrors.Is(err1, err2))
	fmt.Printf("err1 == err3 (different code): %v\n", stderrors.Is(err1, err3))
	// Output:
	// err1 == err2 (same code): true
	// err1 == err3 (different code): false
}

// 示例：错误类型断言
func ExampleAs() {
	err := errs.New(10001, "test error")

	var e *errs.Error
	if stderrors.As(err, &e) {
		fmt.Printf("Code: %d, Message: %s\n", e.Code(), e.Msg())
	}
	// Output: Code: 10001, Message: test error
}

// 示例：映射到外部协议
func Example_externalProtocol() {
	// 模拟 API 处理器
	handleRequest := func() {
		// 模拟业务逻辑
		err := errs.New(10001, "invalid params")

		if err != nil {
			// 提取错误码，映射到外部协议
			result := errs.Code(err)
			if result == 0 {
				result = 999999 // UnclassifiedError
			}

			// 在实际场景中，这里会将 result 设置到响应中
			fmt.Printf("Response Result: %d\n", result)
		}
	}

	handleRequest()
	// Output: Response Result: 10001
}
