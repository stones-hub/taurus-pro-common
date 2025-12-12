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

// Package errs 提供统一的错误处理基础模块
// 支持错误码体系、错误包装/解包，遵循 Go 的 error 返回习惯
package errs

import (
	"errors"
	"fmt"
)

// Error 自定义错误类型，包含错误码和错误消息
type Error struct {
	code int32
	msg  string
}

// New 创建一个新的错误
func New(code int32, msg string) *Error {
	return &Error{code: code, msg: msg}
}

// Errorf 使用格式化字符串创建错误
func Errorf(code int32, format string, args ...any) *Error {
	return &Error{code: code, msg: fmt.Sprintf(format, args...)}
}

// Error 实现 error 接口
func (e *Error) Error() string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf("code: %d, msg: %s", e.code, e.msg)
}

// Code 返回错误码
func (e *Error) Code() int32 {
	if e == nil {
		return 0
	}
	return e.code
}

// Msg 返回错误消息
func (e *Error) Msg() string {
	if e == nil {
		return ""
	}
	return e.msg
}

// String 返回错误的字符串表示
func (e *Error) String() string {
	if e == nil {
		return "nil error"
	}
	return fmt.Sprintf("errcode: %d, errmsg: %s", e.code, e.msg)
}

// Unwrap 支持错误解包，用于 errors.Unwrap
func (e *Error) Unwrap() error {
	return nil
}

// Is 支持错误比较，用于 errors.Is
func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.code == t.code
}

// 未知错误常量
var errUnknown = New(999999, "unknown error")

// Code 从 error 中提取错误码
// 如果 error 是 *Error 类型，返回其错误码
// 否则返回未知错误码 999999
func Code(err error) int32 {
	if err == nil {
		return 0
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code()
	}
	return errUnknown.Code()
}

// Message 从 error 中提取错误消息
// 如果 error 是 *Error 类型，返回其错误消息
// 否则返回未知错误消息
func Message(err error) string {
	if err == nil {
		return ""
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Msg()
	}
	return err.Error()
}

// Wrap 包装错误，保留原始错误信息
// 返回一个新的 *Error，但保留原始错误链
func Wrap(err error, code int32, msg string) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		errInfo: *New(code, msg),
		err:     err,
	}
}

// Wrapf 使用格式化字符串包装错误
func Wrapf(err error, code int32, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		errInfo: *Errorf(code, format, args...),
		err:     err,
	}
}

// wrappedError 包装错误，支持错误链
type wrappedError struct {
	errInfo Error // 使用 errInfo 避免与方法名 Error 冲突
	err     error
}

// Unwrap 返回被包装的原始错误
func (e *wrappedError) Unwrap() error {
	return e.err
}

// Error 返回错误信息，包含被包装的错误
func (e *wrappedError) Error() string {
	if e.err == nil {
		return e.errInfo.Error()
	}
	return fmt.Sprintf("%s: %v", e.errInfo.Error(), e.err)
}

// Code 返回错误码
func (e *wrappedError) Code() int32 {
	return e.errInfo.Code()
}

// Msg 返回错误消息
func (e *wrappedError) Msg() string {
	return e.errInfo.Msg()
}

// Is 支持错误比较
func (e *wrappedError) Is(target error) bool {
	if e.errInfo.Is(target) {
		return true
	}
	return errors.Is(e.err, target)
}

// As 支持错误类型断言
func (e *wrappedError) As(target any) bool {
	if errors.As(&e.errInfo, target) {
		return true
	}
	return errors.As(e.err, target)
}
