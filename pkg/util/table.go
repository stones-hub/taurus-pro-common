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

	"github.com/olekukonko/tablewriter"
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

// RenderTable 渲染表格，带颜色
// headers 表头
// lines 表格数据
func RenderTable(headers []string, lines [][]interface{}) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)

	// 动态设置颜色
	headColors := make([]tablewriter.Colors, len(headers))
	colColors := make([]tablewriter.Colors, len(headers))

	for i := range headers {
		headColors[i] = tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor}
		colColors[i] = tablewriter.Colors{tablewriter.FgHiGreenColor + i%5} // 使用不同的颜色
	}

	table.SetHeaderColor(headColors...)
	table.SetColumnColor(colColors...)

	// 填充表格数据
	for _, line := range lines {
		strLine := make([]string, len(line))
		for i, item := range line {
			strLine[i] = fmt.Sprintf("%v", item) // 将每个数据项转换为字符串
		}
		table.Append(strLine)
	}

	// 渲染表格
	table.Render()
}
