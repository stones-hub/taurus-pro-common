package util

import "testing"

func TestPrintTable(t *testing.T) {
	// 示例数据
	data := TableData{
		{"Name", "Age", "Hobbies", "Details"},
		{"Alice", 25, []interface{}{"Reading", "Coding"}, map[string]interface{}{"City": "New York", "Job": "Engineer"}},
		{"Bob", 30, []interface{}{"Gaming", "Traveling"}, map[string]interface{}{"City": "Los Angeles", "Job": "Designer"}},
		{"Charlie", 22, []interface{}{"Cooking", "Painting"}, map[string]interface{}{"City": "San Francisco", "Job": "Artist"}},
	}
	PrintTable(data)
}
