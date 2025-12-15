package errs_test

import (
	stderrors "errors"
	"fmt"

	"github.com/stones-hub/taurus-pro-common/pkg/errs"
)

// 基本创建与提取
func ExampleNew() {
	err := errs.New(10001, "invalid params")
	fmt.Printf("Code: %d, Message: %s\n", errs.Code(err), errs.Message(err))
	// Output: Code: 10001, Message: invalid params
}

// 格式化创建
func ExampleErrorf() {
	err := errs.Errorf(10002, "invalid params: %s", "name")
	fmt.Printf("Code: %d, Message: %s\n", err.Code(), err.Msg())
	// Output: Code: 10002, Message: invalid params: name
}

// 包装错误
func ExampleWrap() {
	orig := fmt.Errorf("orig")
	err := errs.Wrap(orig, 20001, "wrap failed")
	fmt.Printf("Code: %d, Message: %s\n", errs.Code(err), errs.Message(err))
	// Output: Code: 20001, Message: wrap failed
}

// 普通错误比较
func ExampleIs_normal() {
	err1 := errs.New(30001, "a")
	err2 := errs.New(30001, "b")
	err3 := errs.New(30002, "c")
	fmt.Println(stderrors.Is(err1, err2))
	fmt.Println(stderrors.Is(err1, err3))
	// Output:
	// true
	// false
}

// 位掩码错误：组合、判断与比对
func ExampleIs_bitmask() {
	const (
		ErrInvalidParam uint64 = 1 << 0
		ErrNotFound     uint64 = 1 << 1
		ErrTimeout      uint64 = 1 << 2
	)

	combined := errs.NewBitmask(ErrInvalidParam|ErrNotFound, "combined")

	fmt.Println(stderrors.Is(combined, errs.NewBitmask(ErrInvalidParam, ""))) // 包含
	fmt.Println(stderrors.Is(combined, errs.NewBitmask(ErrTimeout, "")))      // 不包含

	code := errs.Code(combined)
	fmt.Println(code&ErrInvalidParam != 0)                  // 位检查
	fmt.Println(code&(ErrInvalidParam|ErrNotFound) == code) // 同时包含
	fmt.Println(code&(ErrTimeout) == 0)                     // 不包含
	// Output:
	// true
	// false
	// true
	// true
	// true
}

// 位掩码场景：表单验证（组合错误判断）
func Example_bitmaskFormValidation() {
	const (
		ErrUserEmpty uint64 = 1 << 0
		ErrPwdShort  uint64 = 1 << 1
		ErrEmailBad  uint64 = 1 << 2
	)

	validate := func(user, pwd, email string) error {
		var code uint64
		if user == "" {
			code |= ErrUserEmpty
		}
		if len(pwd) < 6 {
			code |= ErrPwdShort
		}
		if email == "" || !contains(email, "@") {
			code |= ErrEmailBad
		}
		if code != 0 {
			return errs.NewBitmask(code, "validation failed")
		}
		return nil
	}

	err := validate("", "123", "bad")
	if err != nil {
		c := errs.Code(err)
		fmt.Printf("code=%d\n", c)
		fmt.Println(stderrors.Is(err, errs.NewBitmask(ErrUserEmpty, "")))
		fmt.Println(stderrors.Is(err, errs.NewBitmask(ErrPwdShort, "")))
		fmt.Println(stderrors.Is(err, errs.NewBitmask(ErrEmailBad, "")))
		fmt.Println(c&(ErrUserEmpty|ErrPwdShort|ErrEmailBad) == c)
	}
	// Output:
	// code=7
	// true
	// true
	// true
	// true
}

// 混合使用：普通错误与位掩码错误
func Example_mixedModes() {
	normal := errs.New(40001, "normal")
	bitmask := errs.NewBitmask(1<<0|1<<1, "bm")

	fmt.Println(stderrors.Is(normal, errs.New(40001, "")))        // 精确匹配
	fmt.Println(stderrors.Is(bitmask, errs.NewBitmask(1<<0, ""))) // 位检查
	fmt.Println(stderrors.Is(bitmask, normal))                    // 模式不同，不匹配
	// Output:
	// true
	// true
	// false
}

// 辅助函数：简单包含
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
