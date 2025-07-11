package util

import (
	"fmt"
	"testing"
)

func TestGetLastPathSegments(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		count    int
		expected string
		hasError bool
	}{
		{
			name:     "正常情况-获取最后3个路径段",
			path:     "/Users/yelei/data/code/projects/go/taurus-pro/temp/kf_ai/app/controller/database_controller.go",
			count:    3,
			expected: "app/controller/database_controller.go",
			hasError: false,
		},
		{
			name:     "正常情况-获取最后2个路径段",
			path:     "/Users/yelei/data/code/projects/go/taurus-pro/temp/kf_ai/app/controller/database_controller.go",
			count:    2,
			expected: "controller/database_controller.go",
			hasError: false,
		},
		{
			name:     "正常情况-获取最后1个路径段",
			path:     "/Users/yelei/data/code/projects/go/taurus-pro/temp/kf_ai/app/controller/database_controller.go",
			count:    1,
			expected: "database_controller.go",
			hasError: false,
		},
		{
			name:     "路径开头和结尾有斜杠",
			path:     "///Users/yelei/data/code/projects/go/taurus-pro/temp/kf_ai/app/controller/database_controller.go///",
			count:    3,
			expected: "app/controller/database_controller.go",
			hasError: false,
		},
		{
			name:     "相对路径",
			path:     "temp/kf_ai/app/controller/database_controller.go",
			count:    3,
			expected: "app/controller/database_controller.go",
			hasError: false,
		},
		{
			name:     "路径段数量不足",
			path:     "app/controller",
			count:    3,
			expected: "app/controller",
			hasError: false,
		},
		{
			name:     "空路径",
			path:     "",
			count:    1,
			expected: "",
			hasError: true,
		},
		{
			name:     "只有斜杠的路径",
			path:     "///",
			count:    1,
			expected: "",
			hasError: true,
		},
		{
			name:     "count为0",
			path:     "/Users/yelei/data/code/projects/go/taurus-pro/temp/kf_ai/app/controller/database_controller.go",
			count:    0,
			expected: "",
			hasError: true,
		},
		{
			name:     "count为负数",
			path:     "/Users/yelei/data/code/projects/go/taurus-pro/temp/kf_ai/app/controller/database_controller.go",
			count:    -1,
			expected: "",
			hasError: true,
		},
		{
			name:     "没有路径",
			path:     "controller",
			count:    3,
			expected: "controller",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetLastPathSegments(tt.path, tt.count)

			fmt.Println(result, err)

			if tt.hasError {
				if err == nil {
					t.Errorf("期望有错误，但没有错误")
				}
			} else {
				if err != nil {
					t.Errorf("不期望有错误，但得到了错误: %v", err)
				}
				if result != tt.expected {
					t.Errorf("期望结果 %s，但得到 %s", tt.expected, result)
				}
			}
		})
	}
}
