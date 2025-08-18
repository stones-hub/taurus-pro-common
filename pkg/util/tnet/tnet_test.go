package tnet

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	// 创建限速器，每秒10个请求
	limiter := NewRateLimiter(10)
	defer limiter.Stop()

	// 测试限速效果
	start := time.Now()
	for i := 0; i < 20; i++ {
		limiter.Wait()
	}
	duration := time.Since(start)

	// 20个请求应该至少需要2秒
	if duration < 2*time.Second {
		t.Errorf("Rate limiter not working properly, took %v for 20 requests", duration)
	}
}

func TestHTTPClient(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	// 创建HTTP客户端
	config := HTTPClientConfig{
		Timeout: 5 * time.Second,
		RetryConfig: &RetryConfig{
			MaxRetries:  3,
			InitialWait: time.Millisecond * 100,
			MaxWait:     time.Second,
			Multiplier:  2.0,
		},
		RateLimit: 10,
		Trace:     true,
	}

	client, err := NewHTTPClient(config)
	if err != nil {
		t.Fatalf("Failed to create HTTP client: %v", err)
	}
	defer client.Close()

	// 创建请求
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// 验证响应
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "success" {
		t.Errorf("Expected response body 'success', got '%s'", string(body))
	}
}

func TestHTTPClientRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	config := HTTPClientConfig{
		Timeout: 5 * time.Second,
		RetryConfig: &RetryConfig{
			MaxRetries:  3,
			InitialWait: time.Millisecond * 100,
			MaxWait:     time.Second,
			Multiplier:  2.0,
		},
	}

	client, err := NewHTTPClient(config)
	if err != nil {
		t.Fatalf("Failed to create HTTP client: %v", err)
	}
	defer client.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
}

func TestGetLocalIPs(t *testing.T) {
	ips, err := GetLocalIPs()
	if err != nil {
		t.Fatalf("Failed to get local IPs: %v", err)
	}

	if len(ips) == 0 {
		t.Error("No local IPs found")
	}

	for _, ip := range ips {
		if ip == "127.0.0.1" || ip == "::1" {
			t.Errorf("Loopback address %s should not be included", ip)
		}
	}
}

func TestIsIPAllowed(t *testing.T) {
	tests := []struct {
		name         string
		ip           string
		allowedHosts []string
		want         bool
	}{
		{
			name:         "exact match",
			ip:           "192.168.1.1",
			allowedHosts: []string{"192.168.1.1"},
			want:         true,
		},
		{
			name:         "CIDR match",
			ip:           "192.168.1.100",
			allowedHosts: []string{"192.168.1.0/24"},
			want:         true,
		},
		{
			name:         "no match",
			ip:           "192.168.2.1",
			allowedHosts: []string{"192.168.1.0/24"},
			want:         false,
		},
		{
			name:         "invalid CIDR",
			ip:           "192.168.1.1",
			allowedHosts: []string{"invalid/cidr"},
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIPAllowed(tt.ip, tt.allowedHosts); got != tt.want {
				t.Errorf("IsIPAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRemoteIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remoteIP string
		want     string
	}{
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.1",
			},
			remoteIP: "10.0.0.1:1234",
			want:     "192.168.1.1",
		},
		{
			name: "X-Forwarded-For",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			remoteIP: "10.0.0.2:1234",
			want:     "192.168.1.1",
		},
		{
			name:     "RemoteAddr only",
			headers:  map[string]string{},
			remoteIP: "192.168.1.1:1234",
			want:     "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			req.RemoteAddr = tt.remoteIP
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			if got := GetRemoteIP(req); got != tt.want {
				t.Errorf("GetRemoteIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileUpload(t *testing.T) {
	// 创建临时文件
	content := []byte("test file content")
	tmpfile, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		if header.Filename != filepath.Base(tmpfile.Name()) {
			http.Error(w, "wrong filename", http.StatusBadRequest)
			return
		}

		uploadedContent, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !bytes.Equal(uploadedContent, content) {
			http.Error(w, "content mismatch", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "upload success")
	}))
	defer server.Close()

	// 测试本地文件上传
	params := map[string]string{
		"type": "test",
	}
	resp, err := UploadFile2Remote(server.URL, params, tmpfile.Name(), "file")
	if err != nil {
		t.Fatalf("UploadFile2Remote failed: %v", err)
	}

	if string(resp) != "upload success" {
		t.Errorf("Expected 'upload success', got '%s'", string(resp))
	}

	// 测试 multipart.FileHeader 上传
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(tmpfile.Name()))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	req := httptest.NewRequest("POST", server.URL, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	server.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rec.Code)
	}
}

func TestSanitizeIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		{
			name: "valid IPv4",
			ip:   "192.168.1.1",
			want: "192.168.1.1",
		},
		{
			name: "valid IPv6",
			ip:   "2001:db8::1",
			want: "2001:db8::1",
		},
		{
			name: "invalid IP",
			ip:   "invalid",
			want: "",
		},
		{
			name: "empty IP",
			ip:   "",
			want: "",
		},
		{
			name: "IP with spaces",
			ip:   "  192.168.1.1  ",
			want: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeIP(tt.ip); got != tt.want {
				t.Errorf("sanitizeIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetAllRemoteIPs(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remoteIP string
		want     []string
	}{
		{
			name: "all headers present",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
				"X-Real-IP":       "192.168.1.2",
			},
			remoteIP: "10.0.0.2:1234",
			want:     []string{"192.168.1.1", "10.0.0.1", "192.168.1.2", "10.0.0.2"},
		},
		{
			name:     "only remote addr",
			headers:  map[string]string{},
			remoteIP: "192.168.1.1:1234",
			want:     []string{"192.168.1.1"},
		},
		{
			name: "invalid IPs filtered",
			headers: map[string]string{
				"X-Forwarded-For": "invalid, 192.168.1.1",
				"X-Real-IP":       "invalid",
			},
			remoteIP: "192.168.1.2:1234",
			want:     []string{"192.168.1.1", "192.168.1.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com", nil)
			req.RemoteAddr = tt.remoteIP
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := GetAllRemoteIPs(req)
			if len(got) != len(tt.want) {
				t.Errorf("GetAllRemoteIPs() got %v IPs, want %v IPs", len(got), len(tt.want))
				return
			}

			for i, ip := range got {
				if ip != tt.want[i] {
					t.Errorf("GetAllRemoteIPs()[%d] = %v, want %v", i, ip, tt.want[i])
				}
			}
		})
	}
}

func TestDoHttpRequest(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 检查请求方法
		if r.Method != "POST" && r.Method != "GET" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 检查Content-Type
		if r.Method == "POST" && !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			http.Error(w, "invalid content type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	// 测试GET请求
	resp, err := HttpGet(server.URL, nil, 5*time.Second)
	if err != nil {
		t.Errorf("HttpGet failed: %v", err)
	}
	if string(resp) != "success" {
		t.Errorf("HttpGet response = %s, want success", string(resp))
	}

	// 测试POST请求
	payload := map[string]string{"key": "value"}
	resp, err = HttpPost(server.URL, payload, nil, 5*time.Second)
	if err != nil {
		t.Errorf("HttpPost failed: %v", err)
	}
	if string(resp) != "success" {
		t.Errorf("HttpPost response = %s, want success", string(resp))
	}

	// 测试带自定义头的请求
	headers := map[string]string{
		"X-Custom-Header": "test",
	}
	resp, err = HttpGet(server.URL, headers, 5*time.Second)
	if err != nil {
		t.Errorf("HttpGet with headers failed: %v", err)
	}
	if string(resp) != "success" {
		t.Errorf("HttpGet with headers response = %s, want success", string(resp))
	}
}

// 基准测试
func BenchmarkRateLimiter(b *testing.B) {
	limiter := NewRateLimiter(1000)
	defer limiter.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Wait()
	}
}

func BenchmarkGetLocalIPs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GetLocalIPs()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIsIPAllowed(b *testing.B) {
	allowedHosts := []string{"192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12"}
	testIP := "192.168.1.100"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsIPAllowed(testIP, allowedHosts)
	}
}

func BenchmarkGetRemoteIP(b *testing.B) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	req.Header.Set("X-Forwarded-For", "192.168.1.2, 10.0.0.1")
	req.Header.Set("X-Real-IP", "192.168.1.3")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetRemoteIP(req)
	}
}

func BenchmarkSanitizeIP(b *testing.B) {
	testIPs := []string{
		"192.168.1.1",
		"  192.168.1.2  ",
		"2001:db8::1",
		"invalid",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, ip := range testIPs {
			sanitizeIP(ip)
		}
	}
}

func BenchmarkHttpGet(b *testing.B) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := HttpGet(server.URL, nil, 5*time.Second)
		if err != nil {
			b.Fatal(err)
		}
		if string(resp) != "success" {
			b.Fatal("unexpected response")
		}
	}
}

func BenchmarkHttpPost(b *testing.B) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	payload := map[string]string{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := HttpPost(server.URL, payload, nil, 5*time.Second)
		if err != nil {
			b.Fatal(err)
		}
		if string(resp) != "success" {
			b.Fatal("unexpected response")
		}
	}
}
