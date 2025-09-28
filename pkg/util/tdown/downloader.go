package tdown

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 下载进度信息
type DownloadProgress struct {
	TotalSize    int64     `json:"total_size"`              // 总大小（字节）
	Downloaded   int64     `json:"downloaded"`              // 已下载大小（字节）
	Progress     float64   `json:"progress"`                // 进度百分比 (0-100)
	Speed        int64     `json:"speed"`                   // 下载速度（字节/秒）
	ETA          int64     `json:"eta"`                     // 预计剩余时间（秒）
	CurrentStep  string    `json:"current_step"`            // 当前步骤
	ErrorMessage string    `json:"error_message,omitempty"` // 错误信息
	Timestamp    time.Time `json:"timestamp"`               // 时间戳
}

// 下载进度回调函数类型
type ProgressCallback func(progress DownloadProgress)

// 下载器结构体
type Downloader struct {
	SavePath   string           // 保存路径
	Callback   ProgressCallback // 进度回调函数
	MaxRetries int              // 最大重试次数
	Timeout    time.Duration    // 超时时间
}

// 创建新的下载器
func NewDownloader(savePath string, options ...func(*Downloader)) *Downloader {
	d := &Downloader{
		SavePath:   savePath,
		MaxRetries: 3,
		Timeout:    time.Minute * 30, // 30分钟
	}
	for _, opt := range options {
		opt(d)
	}
	return d
}

// Option 函数类型
type Option func(*Downloader)

// 设置保存路径的 Option（覆盖构造函数中的基础路径）
func WithSavePath(savePath string) Option {
	return func(d *Downloader) {
		d.SavePath = savePath
	}
}

// 设置进度回调的 Option
func WithCallback(callback ProgressCallback) Option {
	return func(d *Downloader) {
		d.Callback = callback
	}
}

// 设置最大重试次数的 Option
func WithMaxRetries(maxRetries int) Option {
	return func(d *Downloader) {
		d.MaxRetries = maxRetries
	}
}

// 设置超时时间的 Option
func WithTimeout(timeout time.Duration) Option {
	return func(d *Downloader) {
		d.Timeout = timeout
	}
}

// 下载文件并支持进度回调
func (d *Downloader) DownloadFile(ctx context.Context, url, filePath string) error {
	// 1. 构建完整文件路径（基于 SavePath）
	var fullPath string
	if d.SavePath != "" {
		fullPath = filepath.Join(d.SavePath, filePath)
	} else {
		fullPath = filePath
	}

	// 2. 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("create download directory failed: %w", err)
	}

	// 发送开始下载进度
	d.sendProgress(DownloadProgress{
		CurrentStep: "start downloading",
		Progress:    0,
		Timestamp:   time.Now(),
	})

	// 2. 创建HTTP客户端
	client := &http.Client{Timeout: d.Timeout}

	// 3. 获取文件大小
	totalSize, err := d.fetchFileSize(ctx, client, url)
	if err != nil {
		return fmt.Errorf("fetch file size failed: %w", err)
	}

	// 发送文件信息进度
	d.sendProgress(DownloadProgress{
		CurrentStep: "fetch file info",
		Progress:    5,
		TotalSize:   totalSize,
		Timestamp:   time.Now(),
	})

	// 4. 执行下载
	return d.execute(ctx, client, url, fullPath, totalSize)
}

// 获取文件大小
func (d *Downloader) fetchFileSize(ctx context.Context, client *http.Client, url string) (int64, error) {
	// 1. 通过HEAD请求获取文件大小
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.ContentLength > 0 {
		return resp.ContentLength, nil
	}

	// 2. 如果HEAD请求无法获取大小，尝试Range请求
	req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Range", "bytes=0-0")

	resp, err = client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	contentRange := resp.Header.Get("Content-Range")
	if contentRange != "" {
		parts := strings.Split(contentRange, "/")
		if len(parts) == 2 {
			if size, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				return size, nil
			}
		}
	}
	return 0, fmt.Errorf("fetch file size failed")
}

// 执行实际下载
func (d *Downloader) execute(ctx context.Context, client *http.Client, url, filePath string, totalSize int64) error {
	var downloaded int64      // 已下载大小
	var lastTime = time.Now() // 最后下载时间
	var lastDownloaded int64  // 最后下载的字节数
	var supportsResume bool   // 是否支持断点续传

	// 检查文件是否已存在
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file failed: %w", err)
	}
	defer file.Close()

	// 获取已下载大小
	if stat, err := file.Stat(); err == nil {
		downloaded = stat.Size()    // 设置已下载大小
		lastDownloaded = downloaded // 设置最后下载的字节数
		lastTime = time.Now()       // 设置最后下载时间
		file.Seek(downloaded, 0)    // 设置文件已下载末端位置

		if downloaded >= totalSize && totalSize > 0 {
			// 文件之前已经下载完成，通知回调函数显示进度
			d.sendProgress(DownloadProgress{
				CurrentStep: "file already exists",
				Progress:    100,
				TotalSize:   totalSize,
				Downloaded:  totalSize,
				Timestamp:   time.Now(),
			})
			return nil
		}
	}

	// -------------程序运行到这里，不管之前有没有下载当前文件，此时文件都已经设置到它应该在的位置（downloaded, lastDownloaded, lastTime, seek(downloaded, 0)）-------------

	// 开始下载, 支持重试次数
	for retry := 0; retry <= d.MaxRetries; retry++ {
		if retry > 0 {
			time.Sleep(time.Duration(retry) * time.Second)
			// 重试时重新获取文件已下载末端位置
			if stat, err := file.Stat(); err == nil {
				downloaded = stat.Size()
				lastDownloaded = downloaded
				lastTime = time.Now()
				file.Seek(downloaded, 0)
			}
		}

		// 创建请求
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			if retry >= d.MaxRetries {
				return fmt.Errorf("create request failed: %w", err)
			}
			continue
		}

		// 不管服务端是否支持断点续传，都设置Range请求头
		if downloaded > 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", downloaded))
		}

		// 执行请求
		resp, err := client.Do(req)
		if err != nil {
			if retry >= d.MaxRetries {
				return fmt.Errorf("download failed: %w", err)
			}
			continue
		}
		defer resp.Body.Close()

		// 处理响应
		switch resp.StatusCode {
		case http.StatusPartialContent:
			// 服务器支持断点续传
			supportsResume = true
		case http.StatusOK:
			// 服务器不支持断点续传，重新开始, 无论之前下载了多少内容，都需要重新开始
			file.Seek(0, 0)
			// 清空文件
			file.Truncate(0)
			// 重置下载状态, 已下载大小和最后下载的字节数
			downloaded = 0
			lastDownloaded = 0
			lastTime = time.Now()
			// 服务器不支持断点续传
			supportsResume = false
		default:
			if retry >= d.MaxRetries {
				return fmt.Errorf("download failed, status code: %d", resp.StatusCode)
			}
			continue
		}

		// -------------程序运行到这里，不管之前有没有下载当前文件(含重试)，此时文件再次设置到它应该在的位置（downloaded, lastDownloaded, lastTime, seek(downloaded, 0)）-------------

		progressReader := &ProgressReader{
			Reader:     resp.Body,
			TotalSize:  totalSize,
			Downloaded: downloaded,
			Callback: func(historyDownloaded int64) {
				now := time.Now()
				elapsed := now.Sub(lastTime).Seconds()

				// 只有当时间间隔大于0时才计算速度，避免除零错误
				if elapsed > 0 {
					// 计算当前下载速度（字节/秒）
					speed := int64(float64(historyDownloaded-lastDownloaded) / elapsed)

					progress := float64(0)
					if totalSize > 0 {
						progress = float64(historyDownloaded) / float64(totalSize) * 100
					}

					// 计算预计剩余时间（秒）
					eta := int64(0)
					if speed > 0 && totalSize > historyDownloaded {
						eta = (totalSize - historyDownloaded) / speed
					}

					// 根据是否支持断点续传确定当前步骤描述
					step := "downloading"
					if supportsResume {
						step = "resuming"
					}

					d.sendProgress(DownloadProgress{
						CurrentStep: step,              // 当前步骤描述
						Progress:    progress,          // 进度百分比 (0-100)
						TotalSize:   totalSize,         // 总大小
						Downloaded:  historyDownloaded, // 当前文件已经下载的字节数
						Speed:       speed,             // 下载速度（字节/秒）
						ETA:         eta,               // 预计剩余时间（秒）
						Timestamp:   now,               // 时间戳
					})
				}

				lastTime = now
				lastDownloaded = historyDownloaded
			},
		}

		_, err = io.Copy(file, progressReader)
		if err != nil {
			if retry >= d.MaxRetries {
				return fmt.Errorf("write file failed: %w", err)
			}
			continue
		}

		// 下载成功
		break
	}

	// 发送完成进度
	d.sendProgress(DownloadProgress{
		CurrentStep: "download completed",
		Progress:    100,
		TotalSize:   totalSize,
		Downloaded:  totalSize,
		Timestamp:   time.Now(),
	})

	return nil
}

// 进度读取器
type ProgressReader struct {
	Reader     io.Reader   // 读取器 response.Body
	TotalSize  int64       // 总大小
	Downloaded int64       // 已下载大小
	Callback   func(int64) // 回调函数, 传入当前文件已经下载的字节数
}

// 从网络中读取数据到p[]byte中，并回调进度
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.Reader.Read(p)
	// 更新已下载大小
	pr.Downloaded += int64(n)
	// 回调进度
	if pr.Callback != nil {
		pr.Callback(pr.Downloaded)
	}

	return n, err
}

// 发送进度回调
func (d *Downloader) sendProgress(progress DownloadProgress) {
	if d.Callback != nil {
		go d.Callback(progress) // 异步发送，避免阻塞下载
	}
}
