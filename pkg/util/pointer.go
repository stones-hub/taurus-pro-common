package util

import (
	"fmt"
	"reflect"
)

// ScanPointer 通用的指针赋值函数
// 参数说明：
//   - dest: 必须是指针类型（&变量），用于接收赋值结果
//   - src:  可以是指针或值类型，将被赋值给 dest 指向的变量
//
// 使用示例:
//
//	// 基本类型
//	var str string
//	helper.ScanPointer(&str, "hello")     // src 是值
//	helper.ScanPointer(&str, &"world")    // src 是指针
//
//	// 结构体
//	type User struct { Name string }
//	var user User
//	helper.ScanPointer(&user, User{Name: "张三"})    // src 是结构体
//	srcUser := &User{Name: "李四"}
//	helper.ScanPointer(&user, srcUser)              // src 是指针
//
//	// 错误示范
//	helper.ScanPointer(user, srcUser)    // ❌ dest 不能是值类型
//	var nilPtr *User
//	helper.ScanPointer(nilPtr, srcUser)  // ❌ dest 不能是 nil
func ScanPointer(dest interface{}, src interface{}) error {
	if dest == nil {
		return fmt.Errorf("destination pointer is nil")
	}

	// 获取目标的反射值
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer, got %v", dv.Kind())
	}
	if dv.IsNil() {
		return fmt.Errorf("destination pointer is nil")
	}

	// 获取指针指向的元素
	dv = dv.Elem()

	// 获取源数据的反射值
	sv := reflect.ValueOf(src)

	// 如果源数据是指针，获取其指向的值
	if sv.Kind() == reflect.Ptr {
		if sv.IsNil() {
			return fmt.Errorf("source value is nil")
		}
		sv = sv.Elem()
	}

	// 检查类型是否匹配
	if !sv.Type().ConvertibleTo(dv.Type()) {
		return fmt.Errorf("cannot convert source type %v to destination type %v", sv.Type(), dv.Type())
	}

	// 进行类型转换（如果需要）并赋值
	if sv.Type() != dv.Type() {
		sv = sv.Convert(dv.Type())
	}

	// 设置值
	dv.Set(sv)
	return nil
}

// MustScanPointer 是 ScanPointer 的简化版本，如果出错会 panic
// 在确定不会出错的场景下使用
func MustScanPointer(dest interface{}, src interface{}) {
	if err := ScanPointer(dest, src); err != nil {
		panic(err)
	}
}

// ScanPointersSlice 批量处理多个指针赋值
// 示例:
//
//	var (
//	    str string
//	    num int
//	)
//	err := ScanPointersSlice(
//	    []interface{}{&str, &num},
//	    []interface{}{"hello", 42},
//	)
func ScanPointersSlice(dests []interface{}, srcs []interface{}) error {
	if len(dests) != len(srcs) {
		return fmt.Errorf("destination and source slices must have same length")
	}

	for i := range dests {
		if err := ScanPointer(dests[i], srcs[i]); err != nil {
			return fmt.Errorf("error at index %d: %v", i, err)
		}
	}
	return nil
}
