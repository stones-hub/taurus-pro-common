package turl

import "testing"

func TestReplaceURI(t *testing.T) {
	tests := []struct {
		name     string
		uri      string
		expected string
	}{
		{
			name:     "基本路径替换",
			uri:      "/users/123/posts/456",
			expected: "users/{num}/posts/{num}",
		},
		{
			name:     "带前后斜杠",
			uri:      "/api/users/789/",
			expected: "api/users/{num}",
		},
		{
			name:     "不替换非独立数字",
			uri:      "/v1/user123/456",
			expected: "v1/user123/{num}",
		},
		{
			name:     "多级路径",
			uri:      "/api/v1/users/123/posts/456/comments/789",
			expected: "api/v1/users/{num}/posts/{num}/comments/{num}",
		},
		{
			name:     "无数字路径",
			uri:      "/api/users/profile",
			expected: "api/users/profile",
		},
		{
			name:     "空路径",
			uri:      "",
			expected: "",
		},
		{
			name:     "只有斜杠",
			uri:      "/",
			expected: "",
		},
		{
			name:     "只有数字",
			uri:      "/123",
			expected: "{num}",
		},
		{
			name:     "混合路径",
			uri:      "/api/v2/users/abc123/456/def",
			expected: "api/v2/users/abc123/{num}/def",
		},
		{
			name:     "连续斜杠",
			uri:      "//api//users//123//",
			expected: "api/users/{num}",
		},
		{
			name:     "带查询参数的数字",
			uri:      "/users/123?page=1",
			expected: "users/{num}?page=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceURI(tt.uri)
			if result != tt.expected {
				t.Errorf("ReplaceURI() = %v, want %v", result, tt.expected)
			}
		})
	}
}
