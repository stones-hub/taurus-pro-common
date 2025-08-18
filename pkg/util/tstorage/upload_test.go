package tstorage

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"testing"
)

// mockOSS 是一个用于测试的 OSS 接口实现
type mockOSS struct {
	uploadFunc func(*multipart.FileHeader) (string, string, error)
	deleteFunc func(string) error
}

func (m *mockOSS) UploadFile(file *multipart.FileHeader) (string, string, error) {
	if m.uploadFunc != nil {
		return m.uploadFunc(file)
	}
	return "", "", nil
}

func (m *mockOSS) DeleteFile(key string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(key)
	}
	return nil
}

func TestNewOss(t *testing.T) {
	tests := []struct {
		name     string
		ossType  string
		expected string
	}{
		{
			name:     "本地存储",
			ossType:  "local",
			expected: "*tstorage.Local",
		},
		{
			name:     "腾讯云COS",
			ossType:  "tencent-cos",
			expected: "*tstorage.TencentCOS",
		},
		{
			name:     "阿里云OSS",
			ossType:  "aliyun-oss",
			expected: "*tstorage.AliyunOSS",
		},
		{
			name:     "无效类型",
			ossType:  "invalid",
			expected: "*tstorage.Local", // 默认返回本地存储
		},
		{
			name:     "空类型",
			ossType:  "",
			expected: "*tstorage.Local", // 默认返回本地存储
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oss := NewOss(tt.ossType)
			actualType := GetType(oss)
			if actualType != tt.expected {
				t.Errorf("NewOss() type = %v, want %v", actualType, tt.expected)
			}
		})
	}
}

// GetType 返回接口实现的具体类型名称
func GetType(oss OSS) string {
	return fmt.Sprintf("%T", oss)
}

// TestOSSInterface 测试 OSS 接口的基本功能
func TestOSSInterface(t *testing.T) {
	// 创建一个模拟的文件头
	fileHeader := &multipart.FileHeader{
		Filename: "test.txt",
		Size:     100,
	}

	// 测试上传成功
	t.Run("Upload Success", func(t *testing.T) {
		mock := &mockOSS{
			uploadFunc: func(fh *multipart.FileHeader) (string, string, error) {
				return "http://example.com/test.txt", "test.txt", nil
			},
		}

		url, key, err := mock.UploadFile(fileHeader)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if url != "http://example.com/test.txt" {
			t.Errorf("Expected URL = %v, got %v", "http://example.com/test.txt", url)
		}
		if key != "test.txt" {
			t.Errorf("Expected key = %v, got %v", "test.txt", key)
		}
	})

	// 测试上传失败
	t.Run("Upload Failure", func(t *testing.T) {
		expectedError := errors.New("upload failed")
		mock := &mockOSS{
			uploadFunc: func(fh *multipart.FileHeader) (string, string, error) {
				return "", "", expectedError
			},
		}

		_, _, err := mock.UploadFile(fileHeader)
		if err != expectedError {
			t.Errorf("Expected error = %v, got %v", expectedError, err)
		}
	})

	// 测试删除成功
	t.Run("Delete Success", func(t *testing.T) {
		mock := &mockOSS{
			deleteFunc: func(key string) error {
				return nil
			},
		}

		err := mock.DeleteFile("test.txt")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	// 测试删除失败
	t.Run("Delete Failure", func(t *testing.T) {
		expectedError := errors.New("delete failed")
		mock := &mockOSS{
			deleteFunc: func(key string) error {
				return expectedError
			},
		}

		err := mock.DeleteFile("test.txt")
		if err != expectedError {
			t.Errorf("Expected error = %v, got %v", expectedError, err)
		}
	})
}

// 基准测试
func BenchmarkNewOss(b *testing.B) {
	ossTypes := []string{"local", "tencent-cos", "aliyun-oss", "invalid"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ossType := range ossTypes {
			NewOss(ossType)
		}
	}
}

func BenchmarkOSSInterface(b *testing.B) {
	// 创建一个模拟的文件头
	fileHeader := &multipart.FileHeader{
		Filename: "test.txt",
		Size:     100,
	}

	// 测试上传性能
	b.Run("Upload", func(b *testing.B) {
		mock := &mockOSS{
			uploadFunc: func(fh *multipart.FileHeader) (string, string, error) {
				return "http://example.com/test.txt", "test.txt", nil
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.UploadFile(fileHeader)
		}
	})

	// 测试删除性能
	b.Run("Delete", func(b *testing.B) {
		mock := &mockOSS{
			deleteFunc: func(key string) error {
				return nil
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mock.DeleteFile("test.txt")
		}
	})
}

func BenchmarkLocalStorage(b *testing.B) {
	// 测试本地存储创建
	b.Run("Create", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewOss("local")
		}
	})

	// 清理测试目录
	defer func() {
		os.RemoveAll("./test_uploads")
	}()
}

func BenchmarkTencentCOS(b *testing.B) {
	// 测试腾讯云COS创建
	b.Run("Create", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewOss("tencent-cos")
		}
	})
}

func BenchmarkAliyunOSS(b *testing.B) {
	// 测试阿里云OSS创建
	b.Run("Create", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			NewOss("aliyun-oss")
		}
	})
}
