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

package ttable

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// TableData 定义表格数据类型，是一个二维切片，每行包含任意类型的数据
// 示例：
//
//	data := ttable.TableData{
//	    {"Name", "Age", "City"},           // 表头
//	    {"John", 30, "New York"},          // 数据行
//	    {"Jane", 25, "San Francisco"},     // 数据行
//	}
//
// 注意事项：
//   - 每行的列数应该相同
//   - 支持任意类型的数据，会自动转换为字符串
//   - 通常第一行用作表头
//   - 用于 PrintTable 函数的输入
type TableData [][]interface{}

// PrintTable 将表格数据以格式化的方式打印到标准输出
// 参数：
//   - data: 要打印的表格数据
//
// 使用示例：
//
//	data := ttable.TableData{
//	    {"Name", "Age", "City"},
//	    {"John", 30, "New York"},
//	    {"Jane", 25, "San Francisco"},
//	}
//	ttable.PrintTable(data)
//	// 输出：
//	// Name | Age | City
//	// ----------------------
//	// John | 30  | New York
//	// ----------------------
//	// Jane | 25  | San Francisco
//	// ----------------------
//
// 注意事项：
//   - 自动计算每列的最大宽度
//   - 使用 | 作为列分隔符
//   - 使用 - 作为行分隔符
//   - 所有数据类型会被转换为字符串
//   - 列会自动对齐（左对齐）
//   - 适用于终端输出
//   - 如果需要其他格式，请使用 RenderTable
func PrintTable(data TableData) {
	// 检查空数据
	if len(data) == 0 {
		return
	}

	// 检查第一行是否为空
	if len(data[0]) == 0 {
		return
	}

	// 计算每列的最大宽度
	colWidths := make([]int, len(data[0]))
	for _, row := range data {
		for i, val := range row {
			strVal := formatValue(val)
			if len(strVal) > colWidths[i] {
				colWidths[i] = len(strVal)
			}
		}
	}

	// 打印表格
	for _, row := range data {
		for i, val := range row {
			strVal := formatValue(val)
			// 使用空格填充到最大宽度
			paddedVal := fmt.Sprintf("%-"+fmt.Sprintf("%ds", colWidths[i])+"", strVal)
			fmt.Print(paddedVal)
			if i < len(row)-1 {
				fmt.Print(" | ")
			}
		}
		fmt.Println()
		if len(row) > 0 {
			fmt.Println(strings.Repeat("-", sum(colWidths)+len(row)*2+1))
		}
	}
}

// formatValue 将任意类型的数据转换为适合显示的字符串格式
// 参数：
//   - val: 要转换的值，可以是任意类型
//
// 返回值：
//   - string: 转换后的字符串
//
// 支持的类型：
//   - string: 直接返回
//   - int: 转换为十进制字符串
//   - []interface{}: 转换为字符串表示
//   - map[string]interface{}: 转换为字符串表示
//   - 其他类型: 使用 fmt.Sprintf("%v") 转换
//
// 使用示例：
//
//	str := formatValue("hello")               // 返回 "hello"
//	str = formatValue(42)                     // 返回 "42"
//	str = formatValue([]interface{}{1, 2, 3}) // 返回 "[1 2 3]"
//	str = formatValue(map[string]interface{}{
//	    "name": "John",
//	    "age":  30,
//	})                                        // 返回 "map[age:30 name:John]"
//
// 注意事项：
//   - 主要用于表格打印功能
//   - 返回的字符串适合在表格中显示
//   - 复杂类型会被转换为简单的字符串表示
//   - 不处理格式化或对齐
func formatValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	case []interface{}:
		return fmt.Sprintf("%v", v)
	case map[string]interface{}:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// sum 计算整数切片中所有数字的总和
// 参数：
//   - nums: 要计算的整数切片
//
// 返回值：
//   - int: 所有数字的总和
//
// 使用示例：
//
//	total := sum([]int{1, 2, 3, 4, 5})  // 返回 15
//	total = sum([]int{})                 // 返回 0
//	total = sum([]int{-1, 0, 1})         // 返回 0
//
// 注意事项：
//   - 如果切片为空，返回 0
//   - 支持负数
//   - 可能会发生整数溢出
//   - 主要用于计算表格列宽
func sum(nums []int) int {
	total := 0
	for _, num := range nums {
		total += num
	}
	return total
}

// RenderTable 渲染表格，支持终端、HTML、Markdown 三种格式
// 参数：
//   - headers: 表头字符串切片
//   - lines: 表格数据，每行是一个任意类型的切片
//   - format: 输出格式，可选值：
//   - "terminal": 终端友好的格式（默认）
//   - "html": HTML 表格格式
//   - "markdown": Markdown 表格格式
//
// 使用示例：
//
//	headers := []string{"Name", "Age", "City"}
//	data := [][]interface{}{
//	    {"John", 30, "New York"},
//	    {"Jane", 25, "San Francisco"},
//	}
//
//	// 终端格式
//	ttable.RenderTable(headers, data, "terminal")
//	// 输出：
//	// +------+-----+---------------+
//	// | Name | Age | City          |
//	// +------+-----+---------------+
//	// | John | 30  | New York      |
//	// | Jane | 25  | San Francisco |
//	// +------+-----+---------------+
//
//	// HTML 格式
//	ttable.RenderTable(headers, data, "html")
//	// 输出 HTML 表格代码
//
//	// Markdown 格式
//	ttable.RenderTable(headers, data, "markdown")
//	// 输出 Markdown 表格代码
//
// 注意事项：
//   - 使用 go-pretty 库实现表格渲染
//   - 终端格式支持边框和对齐
//   - HTML 格式包含基本的表格标签
//   - Markdown 格式符合标准语法
//   - 所有数据类型会被转换为字符串
//   - 输出到标准输出（os.Stdout）
//   - 如果格式无效，默认使用终端格式
func RenderTable(headers []string, lines [][]interface{}, format string) {
	// 创建表格
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	// 设置表头
	headerRow := table.Row{}
	for _, header := range headers {
		headerRow = append(headerRow, header)
	}
	t.AppendHeader(headerRow)

	// 添加数据行
	for _, line := range lines {
		row := table.Row{}
		for _, item := range line {
			row = append(row, item)
		}
		t.AppendRow(row)
	}

	// 根据格式渲染
	switch format {
	case "html":
		t.RenderHTML()
	case "markdown":
		t.RenderMarkdown()
	default: // terminal
		// 设置终端友好的样式
		t.SetStyle(table.StyleLight)
		t.Style().Options.SeparateRows = true
		t.Style().Options.SeparateColumns = true
		t.Style().Options.DrawBorder = true
		t.Style().Options.SeparateHeader = true
		t.Render()
	}
}

// RenderTableMarkdown 将表格数据渲染为 Markdown 格式的字符串
// 参数：
//   - headers: 表头字符串切片
//   - lines: 表格数据，每行是一个任意类型的切片
//
// 返回值：
//   - string: Markdown 格式的表格字符串
//
// 使用示例：
//
//	headers := []string{"Name", "Age", "City"}
//	data := [][]interface{}{
//	    {"John", 30, "New York"},
//	    {"Jane", 25, "San Francisco"},
//	}
//	md := ttable.RenderTableMarkdown(headers, data)
//	// 返回：
//	// | Name | Age | City          |
//	// | --- | --- | --- |
//	// | John | 30 | New York |
//	// | Jane | 25 | San Francisco |
//
// 注意事项：
//   - 生成标准的 Markdown 表格语法
//   - 使用 | 作为列分隔符
//   - 使用 --- 作为表头分隔符
//   - 所有数据类型会被转换为字符串
//   - 不处理单元格内的 Markdown 语法
//   - 适用于生成 Markdown 文档
//   - 返回字符串，不直接输出
func RenderTableMarkdown(headers []string, lines [][]interface{}) string {
	var sb strings.Builder

	// 处理空表头的情况
	if len(headers) == 0 {
		sb.WriteString("| |\n")
		sb.WriteString("| --- |\n")
		return sb.String()
	}

	// 表头
	sb.WriteString("| ")
	sb.WriteString(strings.Join(headers, " | "))
	sb.WriteString(" |\n")
	// 分隔符
	sb.WriteString("|" + strings.Repeat(" --- |", len(headers)))
	sb.WriteString("\n")
	// 数据行
	for _, line := range lines {
		strLine := make([]string, len(line))
		for i, item := range line {
			strLine[i] = fmt.Sprintf("%v", item)
		}
		sb.WriteString("| ")
		sb.WriteString(strings.Join(strLine, " | "))
		sb.WriteString(" |\n")
	}
	return sb.String()
}

// RenderTableHTML 将表格数据渲染为 HTML 格式的字符串
// 参数：
//   - headers: 表头字符串切片
//   - lines: 表格数据，每行是一个任意类型的切片
//
// 返回值：
//   - string: HTML 格式的表格字符串
//
// 使用示例：
//
//	headers := []string{"Name", "Age", "City"}
//	data := [][]interface{}{
//	    {"John", 30, "New York"},
//	    {"Jane", 25, "San Francisco"},
//	}
//	html := ttable.RenderTableHTML(headers, data)
//	// 返回：
//	// <table border="1" cellspacing="0" cellpadding="4">
//	//   <tr><th>Name</th><th>Age</th><th>City</th></tr>
//	//   <tr><td>John</td><td>30</td><td>New York</td></tr>
//	//   <tr><td>Jane</td><td>25</td><td>San Francisco</td></tr>
//	// </table>
//
// 注意事项：
//   - 生成基本的 HTML 表格标签
//   - 包含基本的表格样式（边框和内边距）
//   - 使用 th 标签表示表头
//   - 使用 td 标签表示数据单元格
//   - 所有数据类型会被转换为字符串
//   - 不处理 HTML 转义
//   - 适用于生成 HTML 文档
//   - 返回字符串，不直接输出
func RenderTableHTML(headers []string, lines [][]interface{}) string {
	var sb strings.Builder
	sb.WriteString("<table border=\"1\" cellspacing=\"0\" cellpadding=\"4\">\n")
	// 表头
	sb.WriteString("  <tr>")
	for _, h := range headers {
		sb.WriteString(fmt.Sprintf("<th>%v</th>", h))
	}
	sb.WriteString("</tr>\n")
	// 数据行
	for _, line := range lines {
		sb.WriteString("  <tr>")
		for _, item := range line {
			sb.WriteString(fmt.Sprintf("<td>%v</td>", item))
		}
		sb.WriteString("</tr>\n")
	}
	sb.WriteString("</table>\n")
	return sb.String()
}
