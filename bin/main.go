package main

import (
	"context"
	"fmt"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/ctx"
	"github.com/stones-hub/taurus-pro-common/pkg/util"
)

func main() {
	fmt.Println("=== 表格测试 ===")
	// 定义表头
	headers := []string{"编号", "项目名称", "状态", "进度", "负责人"}

	// 准备表格数据
	data := [][]interface{}{
		{1, "项目A", "进行中", "75%", "张三"},
		{2, "项目B", "已完成", "100%", "李四"},
		{3, "项目C", "待开始", "0%", "王五"},
		{4, "项目D", "已暂停", "45%", "赵六"},
		{5, "项目E", "规划中", "10%", "孙七"},
	}

	// 使用工具包渲染表格
	util.RenderTable(headers, data)

	fmt.Println("\n=== Context 测试 ===")
	contextExample()

	cronExample()
}

func contextExample() {
	// 创建基础 context
	baseCtx := context.Background()

	// 创建一个请求ID
	requestID := "req-123456"

	// 使用 WithTaurusContext 创建新的 context
	taurusCtx := ctx.WithTaurusContext(baseCtx, requestID)

	// 获取 TaurusContext
	tc := ctx.GetTaurusContext(taurusCtx)

	// 测试设置和获取数据
	tc.Set("user_id", "user-001")
	tc.Set("role", "admin")
	tc.Set("login_time", time.Now())

	// 打印基本信息
	fmt.Printf("请求ID: %s\n", tc.GetRequestID())
	fmt.Printf("创建时间: %v\n", tc.AtTime)

	// 获取并打印存储的数据
	fmt.Printf("\n存储的数据:\n")
	fmt.Printf("用户ID: %v\n", tc.Get("user_id"))
	fmt.Printf("角色: %v\n", tc.Get("role"))
	fmt.Printf("登录时间: %v\n", tc.Get("login_time"))

	// 测试不存在的键
	fmt.Printf("\n获取不存在的键:\n")
	fmt.Printf("不存在的键: %v\n", tc.Get("not_exist"))

	// 测试存储不同类型的数据
	tc.Set("int_value", 42)
	tc.Set("float_value", 3.14)
	tc.Set("bool_value", true)
	tc.Set("slice_value", []string{"a", "b", "c"})
	tc.Set("map_value", map[string]int{"one": 1, "two": 2})

	fmt.Printf("\n不同类型的数据:\n")
	fmt.Printf("整数: %v\n", tc.Get("int_value"))
	fmt.Printf("浮点数: %v\n", tc.Get("float_value"))
	fmt.Printf("布尔值: %v\n", tc.Get("bool_value"))
	fmt.Printf("切片: %v\n", tc.Get("slice_value"))
	fmt.Printf("映射: %v\n", tc.Get("map_value"))
}

func cronExample() {

}
