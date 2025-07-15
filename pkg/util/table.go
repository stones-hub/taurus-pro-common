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
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// TableData 定义表格数据
type TableData [][]interface{}

// PrintTable 打印表格
func PrintTable(data TableData) {
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

// formatValue 将任意类型的数据转换为字符串
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

// sum 计算整数切片的总和
func sum(nums []int) int {
	total := 0
	for _, num := range nums {
		total += num
	}
	return total
}

// RenderTable 渲染表格，支持终端、HTML、Markdown 三种格式
// headers 表头
// lines 表格数据
// format 输出格式："terminal", "html", "markdown"
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

// RenderTableMarkdown 输出 Markdown 格式表格
func RenderTableMarkdown(headers []string, lines [][]interface{}) string {
	var sb strings.Builder
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

// RenderTableHTML 输出 HTML 格式表格
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
