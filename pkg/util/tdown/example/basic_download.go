package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/util/tdown"
)

func main() {
	fmt.Println("=== 基础下载器测试 ===")

	// 创建下载目录
	os.MkdirAll("./downloads", 0755)

	// 测试用例1：小文件下载
	fmt.Println("\n1. 测试小文件下载...")
	// testSmallFile()

	// 测试用例2：大文件下载和断点续传
	fmt.Println("\n2. 测试大文件下载和断点续传...")
	testLargeFileWithResume()

	// 测试用例3：并发下载多个文件
	fmt.Println("\n3. 测试并发下载多个文件...")
	// testConcurrentDownload()

	// 测试用例4：顺序同步下载多个文件
	fmt.Println("\n4. 测试顺序同步下载多个文件...")
	// testSequentialDownload()

	fmt.Println("\n=== 测试完成 ===")
}

// 测试小文件下载
func testSmallFile() {
	url := "https://cdn-tx-admin.3k.com/2023/06/25/e412337e23b25dcea2876205b4bd29dd.mp4" // 1KB 测试文件

	downloader := tdown.NewDownloader(
		"./downloads", // 基础保存路径
		tdown.WithMaxRetries(3),
		tdown.WithTimeout(time.Minute*2),
		tdown.WithCallback(func(progress tdown.DownloadProgress) {
			fmt.Printf("[小文件] %s - 进度: %.1f%% (%d/%d bytes)\n",
				progress.CurrentStep,
				progress.Progress,
				progress.Downloaded,
				progress.TotalSize,
			)
		}),
	)

	ctx := context.Background()

	// 现在只需要提供相对路径，会自动基于基础路径
	_, err := downloader.DownloadFile(ctx, url, "video_test_1.mp4")
	if err != nil {
		log.Printf("小文件下载失败: %v", err)
		return
	}

	if stat, err := os.Stat("./downloads/video_test_1.mp4"); err == nil {
		fmt.Printf("小文件下载成功! 文件大小: %d bytes\n", stat.Size())
	}
}

// 测试大文件下载和断点续传
func testLargeFileWithResume() {
	url := "https://cdn-tx-admin.3k.com/2023/06/25/e412337e23b25dcea2876205b4bd29dd.mp4" // 10MB 测试文件

	downloader := tdown.NewDownloader(
		"./downloads",
		tdown.WithMaxRetries(3),
		tdown.WithTimeout(time.Minute*5),
		tdown.WithCallback(func(progress tdown.DownloadProgress) {
			fmt.Printf("[大文件] %s - 进度: %.1f%% (%d/%d bytes) 速度: %d KB/s\n",
				progress.CurrentStep,
				progress.Progress,
				progress.Downloaded,
				progress.TotalSize,
				progress.Speed/1024,
			)
		}),
	)
	ctx := context.Background()

	// 第一次下载
	fmt.Println("开始下载大文件...")
	_, err := downloader.DownloadFile(ctx, url, "video_test_2.mp4")
	if err != nil {
		log.Printf("大文件下载失败: %v", err)
		return
	}

	if stat, err := os.Stat("./downloads/video_test_2.mp4"); err == nil {
		fmt.Printf("大文件下载成功! 文件大小: %d bytes\n", stat.Size())
	}

}

// 测试并发下载多个文件
func testConcurrentDownload() {
	// 定义要下载的文件信息
	downloadTasks := []struct {
		url      string
		filename string
		name     string
	}{
		{
			url:      "https://cdn-tx-admin.3k.com/2023/06/25/e412337e23b25dcea2876205b4bd29dd.mp4",
			filename: "concurrent_file_1.mp4",
			name:     "文件1",
		},
		{
			url:      "https://cdn-tx-admin.3k.com/2023/06/25/e412337e23b25dcea2876205b4bd29dd.mp4",
			filename: "concurrent_file_2.mp4",
			name:     "文件2",
		},
	}

	// 使用 WaitGroup 等待所有协程完成
	var wg sync.WaitGroup
	results := make(chan DownloadResult, len(downloadTasks))

	// 启动并发下载
	for i, task := range downloadTasks {
		wg.Add(1)
		go func(index int, task struct {
			url      string
			filename string
			name     string
		}) {
			defer wg.Done()

			fmt.Printf("开始下载 %s...\n", task.name)
			startTime := time.Now()

			// 为每个协程创建独立的下载器实例，包含进度回调
			taskDownloader := tdown.NewDownloader(
				"./downloads",
				tdown.WithMaxRetries(3),
				tdown.WithTimeout(time.Minute*5),
				tdown.WithCallback(func(progress tdown.DownloadProgress) {
					fmt.Printf("[%s] %s - 进度: %.1f%% (%d/%d bytes) 速度: %d KB/s\n",
						task.name,
						progress.CurrentStep,
						progress.Progress,
						progress.Downloaded,
						progress.TotalSize,
						progress.Speed/1024,
					)
				}),
			)

			_, err := taskDownloader.DownloadFile(context.Background(), task.url, task.filename)
			duration := time.Since(startTime)

			result := DownloadResult{
				TaskName: task.name,
				Filename: task.filename,
				Error:    err,
				Duration: duration,
			}

			// 检查文件大小
			if err == nil {
				if stat, err := os.Stat("./downloads/" + task.filename); err == nil {
					result.FileSize = stat.Size()
				}
			}

			results <- result
		}(i, task)
	}

	// 等待所有下载完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	fmt.Println("\n=== 下载结果 ===")
	successCount := 0
	totalSize := int64(0)
	totalDuration := time.Duration(0)

	for result := range results {
		if result.Error != nil {
			fmt.Printf("❌ %s 下载失败: %v (耗时: %v)\n", result.TaskName, result.Error, result.Duration)
		} else {
			fmt.Printf("✅ %s 下载成功! 文件: %s, 大小: %d bytes, 耗时: %v\n",
				result.TaskName, result.Filename, result.FileSize, result.Duration)
			successCount++
			totalSize += result.FileSize
		}
		totalDuration += result.Duration
	}

	// 显示统计信息
	fmt.Printf("\n=== 统计信息 ===\n")
	fmt.Printf("成功下载: %d/%d 个文件\n", successCount, len(downloadTasks))
	fmt.Printf("总下载大小: %d bytes (%.2f MB)\n", totalSize, float64(totalSize)/(1024*1024))
	fmt.Printf("总耗时: %v\n", totalDuration)
	if successCount > 0 {
		avgSpeed := float64(totalSize) / totalDuration.Seconds() / 1024 // KB/s
		fmt.Printf("平均速度: %.2f KB/s\n", avgSpeed)
	}
}

// 测试顺序同步下载多个文件
func testSequentialDownload() {
	// 定义要下载的文件信息
	downloadTasks := []struct {
		url      string
		filename string
		name     string
	}{
		{
			url:      "https://cdn-tx-admin.3k.com/2023/06/25/e412337e23b25dcea2876205b4bd29dd.mp4",
			filename: "sequential_file_1.mp4",
			name:     "文件1",
		},
		{
			url:      "https://cdn-tx-admin.3k.com/2023/06/25/e412337e23b25dcea2876205b4bd29dd.mp4",
			filename: "sequential_file_2.mp4",
			name:     "文件2",
		},
	}

	// 创建一个下载器实例，用于所有文件的下载
	downloader := tdown.NewDownloader(
		"./downloads",
		tdown.WithMaxRetries(3),
		tdown.WithTimeout(time.Minute*5),
		tdown.WithCallback(func(progress tdown.DownloadProgress) {
			// 简单的进度显示，不区分具体文件
			fmt.Printf("[顺序下载] %s - 进度: %.1f%% (%d/%d bytes) 速度: %d KB/s\n",
				progress.CurrentStep,
				progress.Progress,
				progress.Downloaded,
				progress.TotalSize,
				progress.Speed/1024,
			)
		}),
	)

	ctx := context.Background()
	var results []DownloadResult
	totalStartTime := time.Now()

	fmt.Println("开始顺序下载文件...")

	// 顺序下载每个文件
	for i, task := range downloadTasks {
		fmt.Printf("\n--- 开始下载第 %d 个文件: %s ---\n", i+1, task.name)
		startTime := time.Now()

		_, err := downloader.DownloadFile(ctx, task.url, task.filename)
		duration := time.Since(startTime)

		result := DownloadResult{
			TaskName: task.name,
			Filename: task.filename,
			Error:    err,
			Duration: duration,
		}

		// 检查文件大小
		if err == nil {
			if stat, err := os.Stat("./downloads/" + task.filename); err == nil {
				result.FileSize = stat.Size()
			}
		}

		results = append(results, result)

		// 显示当前文件下载结果
		if err != nil {
			fmt.Printf("❌ %s 下载失败: %v (耗时: %v)\n", task.name, err, duration)
		} else {
			fmt.Printf("✅ %s 下载成功! 文件: %s, 大小: %d bytes, 耗时: %v\n",
				task.name, task.filename, result.FileSize, duration)
		}

		// 如果不是最后一个文件，添加短暂延迟
		if i < len(downloadTasks)-1 {
			fmt.Println("等待 1 秒后开始下一个文件...")
			time.Sleep(time.Second)
		}
	}

	totalDuration := time.Since(totalStartTime)

	// 显示最终统计信息
	fmt.Printf("\n=== 顺序下载结果统计 ===\n")
	successCount := 0
	totalSize := int64(0)
	totalDownloadTime := time.Duration(0)

	for _, result := range results {
		if result.Error == nil {
			successCount++
			totalSize += result.FileSize
		}
		totalDownloadTime += result.Duration
	}

	fmt.Printf("成功下载: %d/%d 个文件\n", successCount, len(downloadTasks))
	fmt.Printf("总下载大小: %d bytes (%.2f MB)\n", totalSize, float64(totalSize)/(1024*1024))
	fmt.Printf("总下载时间: %v\n", totalDownloadTime)
	fmt.Printf("总耗时(含等待): %v\n", totalDuration)
	if successCount > 0 {
		avgSpeed := float64(totalSize) / totalDownloadTime.Seconds() / 1024 // KB/s
		fmt.Printf("平均下载速度: %.2f KB/s\n", avgSpeed)
	}

	// 显示每个文件的详细信息
	fmt.Printf("\n=== 详细结果 ===\n")
	for i, result := range results {
		status := "❌ 失败"
		if result.Error == nil {
			status = "✅ 成功"
		}
		fmt.Printf("%d. %s %s - %s (%.2f MB, %v)\n",
			i+1, status, result.TaskName, result.Filename,
			float64(result.FileSize)/(1024*1024), result.Duration)
	}
}

// 下载结果结构体
type DownloadResult struct {
	TaskName string
	Filename string
	Error    error
	FileSize int64
	Duration time.Duration
}
