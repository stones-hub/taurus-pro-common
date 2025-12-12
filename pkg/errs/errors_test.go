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

package errs

import (
	"errors"
	"fmt"
	"testing"
)

func TestError_New(t *testing.T) {
	err := New(10001, "invalid params")
	if err == nil {
		t.Fatal("New should not return nil")
	}
	if err.Code() != 10001 {
		t.Errorf("expected code 10001, got %d", err.Code())
	}
	if err.Msg() != "invalid params" {
		t.Errorf("expected msg 'invalid params', got '%s'", err.Msg())
	}
}

func TestError_Errorf(t *testing.T) {
	err := Errorf(10002, "invalid params: %s", "username")
	if err == nil {
		t.Fatal("Errorf should not return nil")
	}
	if err.Code() != 10002 {
		t.Errorf("expected code 10002, got %d", err.Code())
	}
	expectedMsg := "invalid params: username"
	if err.Msg() != expectedMsg {
		t.Errorf("expected msg '%s', got '%s'", expectedMsg, err.Msg())
	}
}

func TestError_Error(t *testing.T) {
	err := New(10003, "test error")
	expected := "code: 10003, msg: test error"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestError_String(t *testing.T) {
	err := New(10004, "test error")
	expected := "errcode: 10004, errmsg: test error"
	if err.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.String())
	}
}

func TestCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int32
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: 0,
		},
		{
			name:     "Error type",
			err:      New(10005, "test"),
			expected: 10005,
		},
		{
			name:     "standard error",
			err:      fmt.Errorf("standard error"),
			expected: 999999, // unknown error code
		},
		{
			name:     "wrapped error",
			err:      Wrap(fmt.Errorf("original"), 10006, "wrapped"),
			expected: 10006,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := Code(tt.err)
			if code != tt.expected {
				t.Errorf("expected code %d, got %d", tt.expected, code)
			}
		})
	}
}

func TestMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "Error type",
			err:      New(10007, "test message"),
			expected: "test message",
		},
		{
			name:     "standard error",
			err:      fmt.Errorf("standard error"),
			expected: "standard error",
		},
		{
			name:     "wrapped error",
			err:      Wrap(fmt.Errorf("original"), 10008, "wrapped"),
			expected: "wrapped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := Message(tt.err)
			if msg != tt.expected {
				t.Errorf("expected msg '%s', got '%s'", tt.expected, msg)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	original := fmt.Errorf("original error")
	wrapped := Wrap(original, 10009, "wrapped error")

	if wrapped == nil {
		t.Fatal("Wrap should not return nil")
	}

	code := Code(wrapped)
	if code != 10009 {
		t.Errorf("expected code 10009, got %d", code)
	}

	msg := Message(wrapped)
	if msg != "wrapped error" {
		t.Errorf("expected msg 'wrapped error', got '%s'", msg)
	}

	// 测试错误解包
	unwrapped := errors.Unwrap(wrapped)
	if unwrapped != original {
		t.Errorf("expected unwrapped error to be original, got %v", unwrapped)
	}
}

func TestWrapf(t *testing.T) {
	original := fmt.Errorf("original error")
	wrapped := Wrapf(original, 10010, "wrapped error: %s", "details")

	if wrapped == nil {
		t.Fatal("Wrapf should not return nil")
	}

	code := Code(wrapped)
	if code != 10010 {
		t.Errorf("expected code 10010, got %d", code)
	}

	expectedMsg := "wrapped error: details"
	msg := Message(wrapped)
	if msg != expectedMsg {
		t.Errorf("expected msg '%s', got '%s'", expectedMsg, msg)
	}
}

func TestError_Is(t *testing.T) {
	err1 := New(10011, "error 1")
	err2 := New(10011, "error 2") // 相同错误码
	err3 := New(10012, "error 3") // 不同错误码

	if !errors.Is(err1, err2) {
		t.Error("errors with same code should be equal")
	}

	if errors.Is(err1, err3) {
		t.Error("errors with different codes should not be equal")
	}
}

func TestWrappedError_Is(t *testing.T) {
	original := New(10013, "original")
	wrapped := Wrap(original, 10014, "wrapped")

	// 应该匹配包装的错误
	if !errors.Is(wrapped, New(10014, "")) {
		t.Error("wrapped error should match its own code")
	}

	// 应该匹配原始错误
	if !errors.Is(wrapped, original) {
		t.Error("wrapped error should match original error")
	}
}

func TestWrappedError_As(t *testing.T) {
	original := fmt.Errorf("original")
	wrapped := Wrap(original, 10015, "wrapped")

	var e *Error
	if !errors.As(wrapped, &e) {
		t.Fatal("should be able to extract Error from wrapped error")
	}

	if e.Code() != 10015 {
		t.Errorf("expected code 10015, got %d", e.Code())
	}
}

func TestError_NilSafety(t *testing.T) {
	var err *Error

	// 测试 nil Error 的方法
	if err.Code() != 0 {
		t.Errorf("nil Error.Code() should return 0, got %d", err.Code())
	}

	if err.Msg() != "" {
		t.Errorf("nil Error.Msg() should return empty string, got '%s'", err.Msg())
	}

	if err.Error() != "nil" {
		t.Errorf("nil Error.Error() should return 'nil', got '%s'", err.Error())
	}

	if err.String() != "nil error" {
		t.Errorf("nil Error.String() should return 'nil error', got '%s'", err.String())
	}
}

func TestUsageExample(t *testing.T) {
	// 模拟实际使用场景
	doSomething := func(a, b int) (int, error) {
		if a < 0 {
			return 0, New(10001, "invalid params")
		}
		return a + b, nil
	}

	work := func() error {
		v, err := doSomething(-1, 2)
		if err != nil {
			errcode := Code(err)
			if errcode != 10001 {
				t.Errorf("expected errcode 10001, got %d", errcode)
			}
			// 包装后继续向上返回
			return Wrap(err, 20001, "work failed")
		}
		_ = v
		return nil
	}

	err := work()
	if err == nil {
		t.Fatal("work should return error")
	}

	errcode := Code(err)
	if errcode != 20001 {
		t.Errorf("expected errcode 20001, got %d", errcode)
	}

	errmsg := Message(err)
	if errmsg != "work failed" {
		t.Errorf("expected errmsg 'work failed', got '%s'", errmsg)
	}
}

