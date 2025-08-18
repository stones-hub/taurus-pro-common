package tnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"mime/multipart"
)

// DefaultTimeout 默认超时时间
const DefaultTimeout = 30 * time.Second

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries  int           // 最大重试次数
	InitialWait time.Duration // 初始等待时间
	MaxWait     time.Duration // 最大等待时间
	Multiplier  float64       // 等待时间乘数
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxRetries:  3,
	InitialWait: time.Second,
	MaxWait:     10 * time.Second,
	Multiplier:  2.0,
}

// RateLimiter 请求限速器
type RateLimiter struct {
	tokens   chan struct{}
	interval time.Duration
	stop     chan struct{}
}

// NewRateLimiter 创建一个新的请求限速器
// 参数：
//   - rps: 每秒允许的请求数量，如果小于等于0则默认为1
//
// 返回值：
//   - *RateLimiter: 返回限速器实例，可以通过 Wait() 方法等待令牌
//
// 使用示例：
//
//	limiter := NewRateLimiter(100) // 限制每秒100个请求
//	defer limiter.Stop()           // 记得在不使用时停止限速器
//
//	// 在发送请求前等待令牌
//	limiter.Wait()
//	// 发送请求...
func NewRateLimiter(rps int) *RateLimiter {
	if rps <= 0 {
		rps = 1
	}
	interval := time.Second / time.Duration(rps)
	limiter := &RateLimiter{
		tokens:   make(chan struct{}, 1),
		interval: interval,
		stop:     make(chan struct{}),
	}
	go limiter.run()
	return limiter
}

// run 运行限速器
func (r *RateLimiter) run() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			select {
			case r.tokens <- struct{}{}:
			default:
			}
		case <-r.stop:
			return
		}
	}
}

// Wait 等待令牌
func (r *RateLimiter) Wait() {
	<-r.tokens
}

// Stop 停止限速器
func (r *RateLimiter) Stop() {
	close(r.stop)
}

// RequestTracer 请求跟踪器
type RequestTracer struct {
	StartTime time.Time
	Duration  time.Duration
	Request   *http.Request
	Response  *http.Response
	Error     error
}

// String 格式化跟踪信息
func (t *RequestTracer) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[%s] %s %s\n", t.StartTime.Format(time.RFC3339), t.Request.Method, t.Request.URL))
	sb.WriteString(fmt.Sprintf("Duration: %s\n", t.Duration))
	if t.Response != nil {
		sb.WriteString(fmt.Sprintf("Status: %d\n", t.Response.StatusCode))
		sb.WriteString("Headers:\n")
		for k, v := range t.Response.Header {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, strings.Join(v, ", ")))
		}
	}
	if t.Error != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", t.Error))
	}
	return sb.String()
}

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	Timeout     time.Duration
	RetryConfig *RetryConfig
	ProxyURL    string
	RateLimit   int // 每秒请求数
	Trace       bool
}

// HTTPClient HTTP客户端
type HTTPClient struct {
	client      *http.Client
	config      HTTPClientConfig
	rateLimiter *RateLimiter
}

// NewHTTPClient 创建一个新的HTTP客户端，支持重试、限速和请求跟踪等高级功能
// 参数：
//   - config: HTTP客户端配置，包含以下选项：
//   - Timeout: 请求超时时间，如果为0则使用默认值30秒
//   - RetryConfig: 重试配置，包含最大重试次数、等待时间等
//   - ProxyURL: 代理服务器URL，如果不为空则使用代理
//   - RateLimit: 每秒最大请求数，如果大于0则启用限速
//   - Trace: 是否启用请求跟踪，启用后会打印详细的请求信息
//
// 返回值：
//   - *HTTPClient: HTTP客户端实例
//   - error: 如果创建过程中出现错误则返回错误信息
//
// 使用示例：
//
//	config := HTTPClientConfig{
//	    Timeout: 30 * time.Second,
//	    RetryConfig: &RetryConfig{
//	        MaxRetries: 3,
//	        InitialWait: time.Second,
//	    },
//	    RateLimit: 100,  // 限制每秒100个请求
//	    Trace: true,     // 启用请求跟踪
//	}
//	client, err := NewHTTPClient(config)
//	if err != nil {
//	    return err
//	}
//	defer client.Close()
func NewHTTPClient(config HTTPClientConfig) (*HTTPClient, error) {
	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		MaxIdleConnsPerHost: 10,
	}

	// 设置代理
	if config.ProxyURL != "" {
		proxyURL, err := url.Parse(config.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("parse proxy URL failed: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	// 设置默认超时
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}

	// 创建客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	// 创建限速器
	var rateLimiter *RateLimiter
	if config.RateLimit > 0 {
		rateLimiter = NewRateLimiter(config.RateLimit)
	}

	return &HTTPClient{
		client:      client,
		config:      config,
		rateLimiter: rateLimiter,
	}, nil
}

// Do 执行HTTP请求
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	var tracer *RequestTracer
	if c.config.Trace {
		tracer = &RequestTracer{
			StartTime: time.Now(),
			Request:   req,
		}
		defer func() {
			tracer.Duration = time.Since(tracer.StartTime)
			fmt.Println(tracer.String())
		}()
	}

	// 应用限速
	if c.rateLimiter != nil {
		c.rateLimiter.Wait()
	}

	// 执行请求（带重试）
	var resp *http.Response
	var err error

	retryConfig := c.config.RetryConfig
	if retryConfig == nil {
		retryConfig = &DefaultRetryConfig
	}

	wait := retryConfig.InitialWait
	for i := 0; i <= retryConfig.MaxRetries; i++ {
		resp, err = c.client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			if tracer != nil {
				tracer.Response = resp
			}
			return resp, nil
		}

		if i == retryConfig.MaxRetries {
			break
		}

		// 准备重试
		if resp != nil {
			resp.Body.Close()
		}

		// 计算等待时间
		if i > 0 {
			wait = time.Duration(float64(wait) * retryConfig.Multiplier)
			if wait > retryConfig.MaxWait {
				wait = retryConfig.MaxWait
			}
		}

		time.Sleep(wait)

		// 创建新的请求（因为Body可能已经被读取）
		newReq := req.Clone(req.Context())
		if req.Body != nil {
			body, ok := req.Body.(io.ReadSeeker)
			if !ok {
				return nil, fmt.Errorf("request body must be seekable for retry")
			}
			_, err = body.Seek(0, 0)
			if err != nil {
				return nil, fmt.Errorf("seek request body failed: %w", err)
			}
			newReq.Body = req.Body
		}
		req = newReq
	}

	if tracer != nil {
		tracer.Response = resp
		tracer.Error = err
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", retryConfig.MaxRetries, err)
	}
	return resp, nil
}

// Close 关闭客户端
func (c *HTTPClient) Close() {
	if c.rateLimiter != nil {
		c.rateLimiter.Stop()
	}
	c.client.CloseIdleConnections()
}

// doHttpRequest 是一个通用的HTTP请求函数，处理不同类型的请求
func doHttpRequest(method, url string, payload interface{}, headers map[string]string, timeout time.Duration) ([]byte, error) {
	// 设置默认超时时间
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	// 创建带超时的HTTP客户端
	client := &http.Client{
		Timeout: timeout,
	}

	// 处理payload
	var reqBody io.Reader
	if payload != nil {
		var jsonPayload []byte
		var err error

		switch p := payload.(type) {
		case []byte:
			jsonPayload = p
		case string:
			jsonPayload = []byte(p)
		case io.Reader:
			reqBody = p
		default:
			if jsonPayload, err = json.Marshal(payload); err != nil {
				return nil, fmt.Errorf("marshal payload failed: %w", err)
			}
		}

		if jsonPayload != nil {
			reqBody = bytes.NewBuffer(jsonPayload)
		}
	}

	// 创建HTTP请求
	request, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	for k, v := range headers {
		request.Header.Set(k, v)
	}

	// 设置默认的Content-Type
	if request.Header.Get("Content-Type") == "" && reqBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	// 发送请求
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer response.Body.Close()

	// 检查状态码
	if response.StatusCode >= 400 {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", response.StatusCode, string(body))
	}

	// 读取响应体
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	return body, nil
}

// HttpPost 是POST请求的包装器
func HttpPost(url string, payload interface{}, headers map[string]string, timeout time.Duration) ([]byte, error) {
	return doHttpRequest("POST", url, payload, headers, timeout)
}

// HttpPostWithDefaultTimeout 是使用默认超时的POST请求的便利函数
func HttpPostWithDefaultTimeout(url string, payload interface{}, headers map[string]string) ([]byte, error) {
	return doHttpRequest("POST", url, payload, headers, DefaultTimeout)
}

// HttpGet 发送HTTP GET请求并返回响应内容
// 参数：
//   - url: 请求的目标URL
//   - headers: 自定义请求头，如果为nil则只使用默认头部
//   - timeout: 请求超时时间，如果为0则使用默认值30秒
//
// 返回值：
//   - []byte: 服务器响应的内容
//   - error: 如果请求过程中出现错误则返回错误信息
//
// 使用示例：
//
//	headers := map[string]string{
//	    "Authorization": "Bearer " + token,
//	    "Accept": "application/json",
//	}
//	resp, err := HttpGet("http://api.example.com/users", headers, 10*time.Second)
//	if err != nil {
//	    return err
//	}
//	// 处理响应...
//
// 注意事项：
//   - 响应状态码大于等于400时会返回错误
//   - 响应体会被完整读入内存，注意大响应的内存使用
//   - 如果需要更多控制，请使用 HTTPClient
func HttpGet(url string, headers map[string]string, timeout time.Duration) ([]byte, error) {
	return doHttpRequest("GET", url, nil, headers, timeout)
}

// HttpGetWithDefaultTimeout 是使用默认超时的GET请求的便利函数
func HttpGetWithDefaultTimeout(url string, headers map[string]string) ([]byte, error) {
	return doHttpRequest("GET", url, nil, headers, DefaultTimeout)
}

// ReadResponse 读取并返回HTTP响应的完整内容
// 参数：
//   - res: HTTP响应对象
//
// 返回值：
//   - []byte: 响应体的完整内容
//   - error: 如果读取过程中出现错误则返回错误信息
//
// 使用示例：
//
//	resp, err := http.Get("http://example.com")
//	if err != nil {
//	    return err
//	}
//	defer resp.Body.Close()
//
//	body, err := tnet.ReadResponse(resp)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("响应内容：%s\n", string(body))
//
// 注意事项：
//   - 会读取响应体的全部内容到内存中
//   - 调用者负责关闭响应体（resp.Body.Close()）
//   - 适用于需要完整读取响应内容的场景
//   - 对于大响应要注意内存使用
func ReadResponse(res *http.Response) ([]byte, error) {
	return io.ReadAll(res.Body)
}

// GetLocalIPs 获取本地所有非回环IP地址列表
// 返回值：
//   - []string: 本地IP地址列表，包含IPv4和IPv6地址
//   - error: 如果获取IP地址失败则返回错误信息
//
// 使用示例：
//
//	ips, err := tnet.GetLocalIPs()
//	if err != nil {
//	    return err
//	}
//	for _, ip := range ips {
//	    fmt.Printf("本地IP: %s\n", ip)
//	}
//
// 注意事项：
//   - 排除了回环地址（127.0.0.1和::1）
//   - 同时返回IPv4和IPv6地址
//   - 返回所有网络接口的IP
//   - 适用于需要获取服务器IP的场景
func GetLocalIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, addr := range addrs {
		// 检查是否为 IP net.Addr 类型
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		// 获取 IP 地址
		ip := ipNet.IP

		// 排除回环地址
		if ip.IsLoopback() {
			continue
		}

		// 添加 IP 地址
		if ip.To4() != nil || ip.To16() != nil {
			ips = append(ips, ip.String())
		}
	}

	return ips, nil
}

/*
- 192.0.0.0/8 192.168.0.0/16 192.168.1.0/24
- /24：子网掩码为255.255.255.0，表示前24位是网络部分，后8位是主机部分。
- /16：子网掩码为255.255.0.0，表示前16位是网络部分，后16位是主机部分。
- /8：子网掩码为255.0.0.0，表示前8位是网络部分，后24位是主机部分。
*/
// IsIPAllowed 检查IP是否在允许的网段中
func IsIPAllowed(ip string, allowedHosts []string) bool {
	for _, host := range allowedHosts {
		// 检查是否为CIDR格式
		if strings.Contains(host, "/") {
			_, ipNet, err := net.ParseCIDR(host)
			if err != nil {
				continue
			}
			if ipNet.Contains(net.ParseIP(ip)) {
				return true
			}
		} else {
			// 直接比较IP
			if ip == host {
				return true
			}
		}
	}
	return false
}

// GetRemoteIP 从HTTP请求中获取客户端的真实IP地址
// 按以下优先级顺序返回第一个有效的IP地址：
//  1. X-Real-IP 头部字段
//  2. X-Forwarded-For 头部字段的第一个IP（通常是最原始的客户端IP）
//  3. RemoteAddr（直接连接的客户端地址）
//
// 参数：
//   - r: HTTP请求对象
//
// 返回值：
//   - string: 返回清理后的IP地址字符串。如果是IPv6地址会返回压缩形式，
//     IPv4地址返回点分十进制形式。如果无法获取有效IP则返回空字符串。
//
// 使用示例：
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    clientIP := GetRemoteIP(r)
//	    if clientIP == "" {
//	        http.Error(w, "无法获取客户端IP", http.StatusBadRequest)
//	        return
//	    }
//	    // 使用clientIP...
//	}
//
// 注意事项：
//   - 返回的IP地址已经过清理和验证，保证是有效的IP格式
//   - 如果请求经过多层代理，建议配合 GetAllRemoteIPs 使用以获取完整的IP链路
//   - 在处理敏感操作时，建议配合 IsIPAllowed 使用以验证IP是否在允许范围内
func GetRemoteIP(r *http.Request) string {
	// 检查 X-Real-IP
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return sanitizeIP(ip)
	}

	// 检查 X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return sanitizeIP(ips[0]) // 返回链中的第一个IP
		}
	}

	// 使用 RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// 如果分割失败，可能是因为没有端口号
		return sanitizeIP(r.RemoteAddr)
	}
	return sanitizeIP(host)
}

// GetAllRemoteIPs 获取所有相关的远程IP地址
func GetAllRemoteIPs(r *http.Request) []string {
	var ips []string

	// 提取X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for _, ip := range strings.Split(xff, ",") {
			if cleanIP := sanitizeIP(ip); cleanIP != "" {
				ips = append(ips, cleanIP)
			}
		}
	}

	// 提取X-Real-IP
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		if cleanIP := sanitizeIP(xrip); cleanIP != "" {
			ips = append(ips, cleanIP)
		}
	}

	// 提取RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if cleanIP := sanitizeIP(host); cleanIP != "" {
			ips = append(ips, cleanIP)
		}
	} else if cleanIP := sanitizeIP(r.RemoteAddr); cleanIP != "" {
		ips = append(ips, cleanIP)
	}

	return ips
}

// sanitizeIP 清理和验证IP地址
func sanitizeIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}

	// 尝试解析IP地址
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return ""
	}

	// 如果是IPv6地址，返回压缩形式
	if parsedIP.To4() == nil {
		return parsedIP.String()
	}

	// 对于IPv4地址，返回点分十进制形式
	return parsedIP.String()
}

// FileUploader 文件上传器接口
type FileUploader interface {
	GetFileName() string
	GetFileSize() int64
	GetFileContent() (io.ReadCloser, error)
	GetContentType() string
}

// LocalFileUploader 本地文件上传器
type LocalFileUploader struct {
	filePath string
	fileInfo os.FileInfo
}

// NewLocalFileUploader 创建本地文件上传器
func NewLocalFileUploader(filePath string) (*LocalFileUploader, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("get file info failed: %w", err)
	}
	return &LocalFileUploader{
		filePath: filePath,
		fileInfo: fileInfo,
	}, nil
}

func (u *LocalFileUploader) GetFileName() string {
	return u.fileInfo.Name()
}

func (u *LocalFileUploader) GetFileSize() int64 {
	return u.fileInfo.Size()
}

func (u *LocalFileUploader) GetFileContent() (io.ReadCloser, error) {
	file, err := os.Open(u.filePath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	return file, nil
}

func (u *LocalFileUploader) GetContentType() string {
	return u.fileInfo.Mode().String()
}

// MultipartFileUploader multipart.FileHeader 上传器
type MultipartFileUploader struct {
	file *multipart.FileHeader
}

// NewMultipartFileUploader 创建 multipart.FileHeader 上传器
func NewMultipartFileUploader(file *multipart.FileHeader) *MultipartFileUploader {
	return &MultipartFileUploader{file: file}
}

func (u *MultipartFileUploader) GetFileName() string {
	return u.file.Filename
}

func (u *MultipartFileUploader) GetFileSize() int64 {
	return u.file.Size
}

func (u *MultipartFileUploader) GetFileContent() (io.ReadCloser, error) {
	file, err := u.file.Open()
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	return file, nil
}

func (u *MultipartFileUploader) GetContentType() string {
	return ""
}

// uploadFile 提供了一个通用的文件上传实现，支持多种上传源和自定义参数
// 参数：
//   - URL: 上传目标地址
//   - params: 上传时的额外参数，会被添加到URL的查询字符串中
//   - uploader: 文件上传器接口，支持本地文件和multipart.FileHeader
//   - fileFieldName: 文件字段的名称
//   - timeout: 上传超时时间，如果为0则使用默认值30秒
//
// 返回值：
//   - []byte: 服务器的响应内容
//   - error: 如果上传过程中出现错误则返回错误信息
//
// 使用示例：
//
//	// 使用本地文件上传器
//	uploader, err := NewLocalFileUploader("/path/to/file.jpg")
//	if err != nil {
//	    return nil, err
//	}
//	params := map[string]string{
//	    "type": "image",
//	    "tag": "profile",
//	}
//	resp, err := uploadFile("http://example.com/upload", params, uploader, "file", 60*time.Second)
//
// 注意事项：
//   - 上传的文件内容会被完整读入内存，注意大文件的内存使用
//   - 上传失败会自动关闭文件和清理资源
//   - 响应状态码大于等于400时会返回错误
func uploadFile(URL string, params map[string]string, uploader FileUploader, fileFieldName string, timeout time.Duration) ([]byte, error) {
	// 设置默认超时时间
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	// 解析上传地址
	u, err := url.Parse(URL)
	if err != nil {
		return nil, fmt.Errorf("parse URL failed: %w", err)
	}

	// 添加查询参数
	query := u.Query()
	for k, v := range params {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()

	// 获取文件内容
	fileContent, err := uploader.GetFileContent()
	if err != nil {
		return nil, err
	}
	defer fileContent.Close()

	// 创建表单
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	// 创建文件字段
	formFile, err := form.CreateFormFile(fileFieldName, uploader.GetFileName())
	if err != nil {
		return nil, fmt.Errorf("create form file failed: %w", err)
	}

	// 写入文件内容
	if _, err = io.Copy(formFile, fileContent); err != nil {
		return nil, fmt.Errorf("copy file content failed: %w", err)
	}

	// 添加额外的表单字段
	if err = form.WriteField("filename", uploader.GetFileName()); err != nil {
		return nil, fmt.Errorf("write filename field failed: %w", err)
	}

	if err = form.WriteField("filelength", strconv.FormatInt(uploader.GetFileSize(), 10)); err != nil {
		return nil, fmt.Errorf("write filelength field failed: %w", err)
	}

	if contentType := uploader.GetContentType(); contentType != "" {
		if err = form.WriteField("content-type", contentType); err != nil {
			return nil, fmt.Errorf("write content-type field failed: %w", err)
		}
	}

	// 关闭表单
	if err = form.Close(); err != nil {
		return nil, fmt.Errorf("close form failed: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", form.FormDataContentType())

	// 发送请求
	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	return respBody, nil
}

// UploadFile2Remote 将文件上传到远端
func UploadFile2Remote(URL string, params map[string]string, filePath string, fileFieldName string) ([]byte, error) {
	uploader, err := NewLocalFileUploader(filePath)
	if err != nil {
		return nil, err
	}
	return uploadFile(URL, params, uploader, fileFieldName, DefaultTimeout)
}

// UploadFile2RemoteWithTimeout 将文件上传到远端，支持自定义超时时间
func UploadFile2RemoteWithTimeout(URL string, params map[string]string, filePath string, fileFieldName string, timeout time.Duration) ([]byte, error) {
	uploader, err := NewLocalFileUploader(filePath)
	if err != nil {
		return nil, err
	}
	return uploadFile(URL, params, uploader, fileFieldName, timeout)
}

// UploadFile2RemoteWithFileHeader 将multipart.FileHeader上传到远端
func UploadFile2RemoteWithFileHeader(URL string, params map[string]string, file *multipart.FileHeader, fileFieldName string) ([]byte, error) {
	uploader := NewMultipartFileUploader(file)
	return uploadFile(URL, params, uploader, fileFieldName, DefaultTimeout)
}

// UploadFile2RemoteWithFileHeaderWithTimeout 将multipart.FileHeader上传到远端，支持自定义超时时间
func UploadFile2RemoteWithFileHeaderWithTimeout(URL string, params map[string]string, file *multipart.FileHeader, fileFieldName string, timeout time.Duration) ([]byte, error) {
	uploader := NewMultipartFileUploader(file)
	return uploadFile(URL, params, uploader, fileFieldName, timeout)
}
