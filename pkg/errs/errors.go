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
	code      uint64
	msg       string
	isBitmask bool // 标识是否为位掩码模式
}

// New 创建一个新的错误（普通错误码模式）
func New(code uint64, msg string) *Error {
	return &Error{code: code, msg: msg, isBitmask: false}
}

// Errorf 使用格式化字符串创建错误（普通错误码模式）
func Errorf(code uint64, format string, args ...any) *Error {
	return &Error{code: code, msg: fmt.Sprintf(format, args...), isBitmask: false}
}

// NewBitmask 创建位掩码错误
func NewBitmask(code uint64, msg string) *Error {
	return &Error{code: code, msg: msg, isBitmask: true}
}

// Bitmaskf 使用格式化字符串创建位掩码错误
func Bitmaskf(code uint64, format string, args ...any) *Error {
	return &Error{code: code, msg: fmt.Sprintf(format, args...), isBitmask: true}
}

// Error 实现标准库 error 接口
// 对应标准库: error.Error() string
func (e *Error) Error() string {
	if e == nil {
		return "nil"
	}
	return fmt.Sprintf("code: %d, msg: %s", e.code, e.msg)
}

// Code 返回错误码
func (e *Error) Code() uint64 {
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

// Unwrap 实现标准库错误解包接口
// 对应标准库: errors.Unwrap(err error) error
// 用于支持错误链（error chain）
func (e *Error) Unwrap() error {
	return nil
}

// Is 实现标准库错误比较接口
// 对应标准库: errors.Is(err, target error) bool
// 当调用 errors.Is(err, target) 时，会自动调用此方法
// 如果是位掩码模式，使用完全包含检查（e.code & t.code == t.code）
// 这样无论是单个位还是组合位，都能正确判断目标是否完全被包含
func (e *Error) Is(target error) bool {
	if e == nil {
		return target == nil
	}
	t, ok := target.(*Error)
	if !ok {
		return false
	}

	// 如果源错误和目标错误的模式不同，不匹配
	if e.isBitmask != t.isBitmask {
		return false
	}

	// 如果源错误和目标错误都是位掩码模式，使用完全包含检查
	// 原理：e.code & t.code == t.code 表示源错误完全包含目标错误的所有位
	// 这样无论是单个位还是组合位，都能正确判断
	if e.isBitmask && t.isBitmask {
		return (e.code & t.code) == t.code
	}

	// 普通模式使用精确匹配
	return e.code == t.code
}

// 注意：Error 类型不实现 As() 方法
// 因为 errors.As() 会自动使用类型断言来处理
// wrappedError.As() 中调用 errors.As(&e.errInfo, target) 也能正常工作

// 未知错误常量
var errUnknown = New(999999, "unknown error")

// Code 从 error 中提取错误码
// 如果 error 是 *Error 类型，返回其错误码
// 否则返回未知错误码 999999
func Code(err error) uint64 {
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
func Wrap(err error, code uint64, msg string) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		errInfo: *New(code, msg),
		err:     err,
	}
}

// Wrapf 使用格式化字符串包装错误
func Wrapf(err error, code uint64, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		errInfo: *Errorf(code, format, args...),
		err:     err,
	}
}

// WrapBitmask 包装错误为位掩码错误
func WrapBitmask(err error, code uint64, msg string) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		errInfo: *NewBitmask(code, msg),
		err:     err,
	}
}

// WrapBitmaskf 使用格式化字符串包装错误为位掩码错误
func WrapBitmaskf(err error, code uint64, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		errInfo: *Bitmaskf(code, format, args...),
		err:     err,
	}
}

// wrappedError 包装错误，支持错误链
type wrappedError struct {
	errInfo Error // 使用 errInfo 避免与方法名 Error 冲突
	err     error
}

// Unwrap 实现标准库错误解包接口
// 对应标准库: errors.Unwrap(err error) error
// 返回被包装的原始错误，用于错误链遍历
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
func (e *wrappedError) Code() uint64 {
	return e.errInfo.Code()
}

// Msg 返回错误消息
func (e *wrappedError) Msg() string {
	return e.errInfo.Msg()
}

// Is 实现标准库错误比较接口
// 对应标准库: errors.Is(err, target error) bool
// 当调用 errors.Is(err, target) 时，会自动调用此方法
// 先检查包装的错误信息，再检查原始错误链
func (e *wrappedError) Is(target error) bool {
	if e.errInfo.Is(target) {
		return true
	}
	return errors.Is(e.err, target)
}

// As 实现标准库错误类型断言接口
// 对应标准库: errors.As(err error, target any) bool
// 当调用 errors.As(err, target) 时，会自动调用此方法
// 先检查包装的错误信息，再检查原始错误链
func (e *wrappedError) As(target any) bool {
	if errors.As(&e.errInfo, target) {
		return true
	}
	return errors.As(e.err, target)
}
