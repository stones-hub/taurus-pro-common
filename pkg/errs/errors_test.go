package errs

import (
	"errors"
	"fmt"
	"testing"
)

// 辅助：构造普通和位掩码错误
var (
	ErrInvalidParam uint64 = 1 << 0
	ErrNotFound     uint64 = 1 << 1
	ErrTimeout      uint64 = 1 << 2
)

func TestNewAndErrorf(t *testing.T) {
	err := New(10001, "invalid params")
	if err.Code() != 10001 || err.Msg() != "invalid params" {
		t.Fatalf("unexpected err: %v", err)
	}

	errf := Errorf(10002, "invalid: %s", "name")
	if errf.Code() != 10002 || errf.Msg() != "invalid: name" {
		t.Fatalf("unexpected errf: %v", errf)
	}
}

func TestCodeAndMessage(t *testing.T) {
	std := fmt.Errorf("std")
	if Code(std) != errUnknown.Code() {
		t.Fatalf("expect unknown code for std error")
	}
	if Message(std) != "std" {
		t.Fatalf("expect std message")
	}

	e := New(20001, "biz")
	if Code(e) != 20001 || Message(e) != "biz" {
		t.Fatalf("unexpected code/msg")
	}
}

func TestWrap(t *testing.T) {
	orig := fmt.Errorf("orig")
	w := Wrap(orig, 30001, "wrap")
	if Code(w) != 30001 || Message(w) != "wrap" {
		t.Fatalf("wrap failed")
	}
	if errors.Unwrap(w) != orig {
		t.Fatalf("unwrap mismatch")
	}
}

func TestWrapBitmask(t *testing.T) {
	orig := fmt.Errorf("orig")
	w := WrapBitmask(orig, ErrInvalidParam|ErrNotFound, "bm wrap")
	if Code(w) != ErrInvalidParam|ErrNotFound {
		t.Fatalf("bitmask code mismatch")
	}
	if !errors.Is(w, NewBitmask(ErrInvalidParam, "")) {
		t.Fatalf("should contain ErrInvalidParam")
	}
	if errors.Is(w, NewBitmask(ErrTimeout, "")) {
		t.Fatalf("should not contain ErrTimeout")
	}
}

func TestIsNormal(t *testing.T) {
	err1 := New(40001, "a")
	err2 := New(40001, "b")
	err3 := New(40002, "c")
	if !errors.Is(err1, err2) {
		t.Fatalf("same code should match")
	}
	if errors.Is(err1, err3) {
		t.Fatalf("different code should not match")
	}
}

func TestIsBitmask(t *testing.T) {
	combined := NewBitmask(ErrInvalidParam|ErrNotFound, "combined")

	// 测试单个位：应该完全包含
	if !errors.Is(combined, NewBitmask(ErrInvalidParam, "")) {
		t.Fatalf("should match single bit")
	}
	if !errors.Is(combined, NewBitmask(ErrNotFound, "")) {
		t.Fatalf("should match single bit")
	}
	if errors.Is(combined, NewBitmask(ErrTimeout, "")) {
		t.Fatalf("should not match unrelated bit")
	}

	// 测试组合位：完全包含的情况
	// combined = ErrInvalidParam|ErrNotFound (0b0011)
	// 检查是否完全包含 ErrInvalidParam|ErrNotFound
	if !errors.Is(combined, NewBitmask(ErrInvalidParam|ErrNotFound, "")) {
		t.Fatalf("should match combined bits when fully contained")
	}

	// 测试组合位：部分包含的情况（应该返回 false，因为不完全包含）
	// combined = ErrInvalidParam|ErrNotFound (0b0011)
	// 检查是否完全包含 ErrInvalidParam|ErrTimeout (0b0101)
	// 结果：0b0011 & 0b0101 = 0b0001，不等于 0b0101，所以应该返回 false
	if errors.Is(combined, NewBitmask(ErrInvalidParam|ErrTimeout, "")) {
		t.Fatalf("should not match when only partially contained")
	}

	// 测试组合位：完全不包含的情况
	if errors.Is(combined, NewBitmask(ErrTimeout, "")) {
		t.Fatalf("should not match unrelated bit")
	}

	// 测试模式隔离：位掩码和普通错误不应该匹配
	normal := New(50001, "normal")
	if errors.Is(combined, normal) || errors.Is(normal, NewBitmask(ErrInvalidParam, "")) {
		t.Fatalf("bitmask and normal should not match")
	}
}

func TestWrappedError_Is_As(t *testing.T) {
	orig := New(60001, "orig")
	w := Wrap(orig, 60002, "wrap")
	if !errors.Is(w, orig) {
		t.Fatalf("wrapped should match original")
	}
	if !errors.Is(w, New(60002, "")) {
		t.Fatalf("wrapped should match its code")
	}
	var e *Error
	if !errors.As(w, &e) || e.Code() != 60002 {
		t.Fatalf("errors.As failed on wrapped")
	}
}

func TestNilSafety(t *testing.T) {
	var e *Error
	if e.Code() != 0 || e.Msg() != "" || e.Error() != "nil" || e.String() != "nil error" {
		t.Fatalf("nil safety failed")
	}
}
