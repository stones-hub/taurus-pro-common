// Package main FFmpeg音频处理工具测试程序
// 这是一个完整的测试程序，用于验证FFmpeg音频处理工具的各项功能
// 包括核心功能测试、高级功能测试、并发测试、错误处理测试和性能测试
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/util/tffmpeg"
)

// main 主函数
// 程序入口，执行完整的FFmpeg音频处理工具测试套件
func main() {
	fmt.Println("=== FFmpeg音频处理工具测试程序 ===")
	fmt.Printf("Go版本: %s, 操作系统: %s, 架构: %s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	// 创建FFmpeg音频处理工具
	ffmpeg := tffmpeg.NewFFmpegAudio()

	// 检查FFmpeg是否可用
	if err := ffmpeg.CheckFFmpegAvailable(); err != nil {
		log.Fatalf("FFmpeg不可用: %v", err)
	}

	fmt.Println("✅ FFmpeg可用，开始测试...")

	// 运行各种测试
	runCoreTests(ffmpeg)          // 核心功能测试
	runAdvancedTests(ffmpeg)      // 高级功能测试
	runConcurrencyTests(ffmpeg)   // 并发测试
	runErrorHandlingTests(ffmpeg) // 错误处理测试
	runPerformanceTests(ffmpeg)   // 性能测试

	fmt.Println("\n=== 所有测试完成 ===")
}

// runCoreTests 运行核心功能测试
// 测试FFmpeg音频处理工具的基本功能，包括音频提取、格式转换、批量处理等
// 参数:
//   - ffmpeg: FFmpeg音频处理工具实例
func runCoreTests(ffmpeg *tffmpeg.FFmpegAudio) {
	fmt.Println("\n=== 核心功能测试 ===")

	testVideoPath := findTestVideo()
	if testVideoPath == "" {
		fmt.Println("⚠️  未找到测试视频文件，跳过核心功能测试")
		return
	}

	fmt.Printf("找到测试视频: %s\n", testVideoPath)
	outputDir := "./test_output/core"
	os.MkdirAll(outputDir, 0755)
	ctx := context.Background()

	// 测试1: 基本音频提取
	// 测试从视频中提取音频的基本功能
	fmt.Println("\n测试1: 基本音频提取")
	basicOptions := &tffmpeg.AudioExtractionOptions{
		Format:     tffmpeg.AudioFormatMP3,     // 输出MP3格式
		Quality:    tffmpeg.AudioQualityMedium, // 中等质量
		Bitrate:    "128k",                     // 128k比特率
		SampleRate: 44100,                      // 44.1kHz采样率
		Channels:   2,                          // 立体声
	}

	result, err := ffmpeg.ExtractAudioFromVideo(ctx, testVideoPath, outputDir, basicOptions)
	if err != nil {
		fmt.Printf("❌ 基本音频提取失败: %v\n", err)
	} else {
		fmt.Printf("✅ 基本音频提取成功: %s (%.2f秒, %d bytes)\n",
			result.OutputPath, result.Duration, result.FileSize)
	}

	// 测试2: 带进度监控的提取
	// 测试实时进度监控功能
	fmt.Println("\n测试2: 带进度监控的提取")
	progressOptions := &tffmpeg.AudioExtractionOptions{
		Format:         tffmpeg.AudioFormatWAV,   // 输出WAV格式
		Quality:        tffmpeg.AudioQualityHigh, // 高质量
		Bitrate:        "192k",                   // 192k比特率
		SampleRate:     44100,                    // 44.1kHz采样率
		Channels:       2,                        // 立体声
		EnableProgress: true,                     // 启用进度监控
		ProgressCallback: func(progress *tffmpeg.FFmpegProgress) {
			// 进度回调函数，显示处理进度
			if progress.Progress > 0 {
				fmt.Printf("   进度: %.1f%%\n", progress.Progress)
			}
		},
	}

	_, err = ffmpeg.ExtractAudioFromVideo(ctx, testVideoPath, outputDir, progressOptions)
	if err != nil {
		fmt.Printf("❌ 带进度监控的音频提取失败: %v\n", err)
	} else {
		fmt.Printf("✅ 带进度监控的音频提取成功\n")
	}

	// 测试3: 格式转换
	// 测试音频格式转换功能
	fmt.Println("\n测试3: 格式转换")
	convertOptions := &tffmpeg.AudioExtractionOptions{
		Format:     tffmpeg.AudioFormatAAC,   // 转换为AAC格式
		Quality:    tffmpeg.AudioQualityHigh, // 高质量
		SampleRate: 48000,                    // 48kHz采样率
		Channels:   2,                        // 立体声
	}

	_, err = ffmpeg.ConvertAudioFormat(ctx, result.OutputPath, outputDir, tffmpeg.AudioFormatAAC, convertOptions)
	if err != nil {
		fmt.Printf("❌ 格式转换失败: %v\n", err)
	} else {
		fmt.Printf("✅ 格式转换成功\n")
	}

	// 测试4: 批量处理
	// 测试批量处理多个视频文件的功能
	fmt.Println("\n测试4: 批量处理")
	videoPaths := []string{testVideoPath, testVideoPath} // 使用同一个文件测试批量处理
	batchOptions := &tffmpeg.AudioExtractionOptions{
		Format:  tffmpeg.AudioFormatMP3,     // 输出MP3格式
		Quality: tffmpeg.AudioQualityMedium, // 中等质量
	}

	batchResults, err := ffmpeg.BatchExtractAudio(ctx, videoPaths, outputDir, batchOptions)
	if err != nil {
		fmt.Printf("❌ 批量处理失败: %v\n", err)
	} else {
		fmt.Printf("✅ 批量处理成功，共处理 %d 个文件\n", len(batchResults))
	}
}

// runAdvancedTests 运行高级功能测试
// 测试FFmpeg音频处理工具的高级功能，包括时间范围提取、音频处理功能组合、预设配置等
// 参数:
//   - ffmpeg: FFmpeg音频处理工具实例
func runAdvancedTests(ffmpeg *tffmpeg.FFmpegAudio) {
	fmt.Println("\n=== 高级功能测试 ===")

	testVideoPath := findTestVideo()
	if testVideoPath == "" {
		fmt.Println("⚠️  未找到测试视频文件，跳过高级功能测试")
		return
	}

	outputDir := "./test_output/advanced"
	os.MkdirAll(outputDir, 0755)
	ctx := context.Background()

	// 测试1: 时间范围提取
	// 测试从视频的特定时间范围提取音频
	fmt.Println("\n测试1: 时间范围提取")
	timeRangeOptions := &tffmpeg.AudioExtractionOptions{
		Format:     tffmpeg.AudioFormatMP3,     // 输出MP3格式
		Quality:    tffmpeg.AudioQualityMedium, // 中等质量
		StartTime:  "00:00:10",                 // 从10秒开始
		Duration:   "00:00:30",                 // 提取30秒
		SampleRate: 44100,                      // 44.1kHz采样率
		Channels:   2,                          // 立体声
	}

	result, err := ffmpeg.ExtractAudioFromVideo(ctx, testVideoPath, outputDir, timeRangeOptions)
	if err != nil {
		fmt.Printf("❌ 时间范围提取失败: %v\n", err)
	} else {
		fmt.Printf("✅ 时间范围提取成功: %.2f秒\n", result.Duration)
	}

	// 测试2: 音频处理功能组合
	// 测试多种音频处理功能的组合使用
	fmt.Println("\n测试2: 音频处理功能组合")
	combinedOptions := &tffmpeg.AudioExtractionOptions{
		Format:        tffmpeg.AudioFormatMP3,   // 输出MP3格式
		Quality:       tffmpeg.AudioQualityHigh, // 高质量
		Normalize:     true,                     // 启用音量标准化
		RemoveSilence: true,                     // 启用静音段移除
		Volume:        "1.5",                    // 音量调整为1.5倍
		SampleRate:    48000,                    // 48kHz采样率
		Channels:      2,                        // 立体声
	}

	_, err = ffmpeg.ExtractAudioFromVideo(ctx, testVideoPath, outputDir, combinedOptions)
	if err != nil {
		fmt.Printf("❌ 组合功能测试失败: %v\n", err)
	} else {
		fmt.Printf("✅ 组合功能测试成功\n")
	}

	// 测试3: 预设配置测试
	// 测试各种预设配置的效果
	fmt.Println("\n测试3: 预设配置测试")
	presets := []string{"speech", "music", "podcast", "archive"} // 测试四种预设配置
	for _, preset := range presets {
		config := ffmpeg.CreatePresetConfig(preset) // 获取预设配置
		_, err := ffmpeg.ExtractAudioFromVideo(ctx, testVideoPath, outputDir, config)
		if err != nil {
			fmt.Printf("❌ 预设 '%s' 失败: %v\n", preset, err)
		} else {
			fmt.Printf("✅ 预设 '%s' 成功\n", preset)
		}
	}
}

// runConcurrencyTests 运行并发测试
// 测试FFmpeg音频处理工具在并发环境下的稳定性和性能
// 参数:
//   - ffmpeg: FFmpeg音频处理工具实例
func runConcurrencyTests(ffmpeg *tffmpeg.FFmpegAudio) {
	fmt.Println("\n=== 并发测试 ===")

	testVideoPath := findTestVideo()
	if testVideoPath == "" {
		fmt.Println("⚠️  未找到测试视频文件，跳过并发测试")
		return
	}

	outputDir := "./test_output/concurrent"
	os.MkdirAll(outputDir, 0755)

	// 测试1: 并发音频提取
	// 测试多个goroutine同时进行音频提取
	fmt.Println("\n测试1: 并发音频提取")
	concurrentCount := 3 // 并发数量
	var wg sync.WaitGroup
	results := make(chan *tffmpeg.AudioExtractionResult, concurrentCount)
	errors := make(chan error, concurrentCount)

	startTime := time.Now()

	for i := 0; i < concurrentCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			ctx := context.Background()
			options := &tffmpeg.AudioExtractionOptions{
				Format:     tffmpeg.AudioFormatMP3,
				Quality:    tffmpeg.AudioQualityMedium,
				Bitrate:    "128k",
				SampleRate: 44100,
				Channels:   2,
			}

			taskOutputDir := filepath.Join(outputDir, fmt.Sprintf("task_%d", index))
			os.MkdirAll(taskOutputDir, 0755)

			result, err := ffmpeg.ExtractAudioFromVideo(ctx, testVideoPath, taskOutputDir, options)
			if err != nil {
				errors <- fmt.Errorf("任务 %d 失败: %v", index, err)
				return
			}

			results <- result
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	successCount := 0
	for result := range results {
		successCount++
		_ = result
	}

	for err := range errors {
		fmt.Printf("   ❌ %v\n", err)
	}

	elapsed := time.Since(startTime)
	fmt.Printf("✅ 并发测试完成: %d/%d 成功，耗时: %.2f秒\n", successCount, concurrentCount, elapsed.Seconds())

	// 测试2: 并发不同格式提取
	// 测试同时提取不同格式的音频
	fmt.Println("\n测试2: 并发不同格式提取")
	formats := []tffmpeg.AudioFormat{
		tffmpeg.AudioFormatMP3, tffmpeg.AudioFormatWAV, tffmpeg.AudioFormatAAC, // 测试三种不同格式
	}

	var wg2 sync.WaitGroup
	formatResults := make(chan *tffmpeg.AudioExtractionResult, len(formats))
	formatErrors := make(chan error, len(formats))

	startTime2 := time.Now()

	for i, format := range formats {
		wg2.Add(1)
		go func(index int, audioFormat tffmpeg.AudioFormat) {
			defer wg2.Done()

			ctx := context.Background()
			options := &tffmpeg.AudioExtractionOptions{
				Format:     audioFormat,
				Quality:    tffmpeg.AudioQualityMedium,
				SampleRate: 44100,
				Channels:   2,
			}

			formatOutputDir := filepath.Join(outputDir, fmt.Sprintf("format_%s", audioFormat))
			os.MkdirAll(formatOutputDir, 0755)

			result, err := ffmpeg.ExtractAudioFromVideo(ctx, testVideoPath, formatOutputDir, options)
			if err != nil {
				formatErrors <- fmt.Errorf("格式 %s 失败: %v", audioFormat, err)
				return
			}

			formatResults <- result
		}(i, format)
	}

	wg2.Wait()
	close(formatResults)
	close(formatErrors)

	formatSuccessCount := 0
	for result := range formatResults {
		formatSuccessCount++
		_ = result
	}

	for err := range formatErrors {
		fmt.Printf("   ❌ %v\n", err)
	}

	elapsed2 := time.Since(startTime2)
	fmt.Printf("✅ 并发格式测试完成: %d/%d 成功，耗时: %.2f秒\n", formatSuccessCount, len(formats), elapsed2.Seconds())
}

// runErrorHandlingTests 运行错误处理测试
// 测试FFmpeg音频处理工具的错误处理能力
// 参数:
//   - ffmpeg: FFmpeg音频处理工具实例
func runErrorHandlingTests(ffmpeg *tffmpeg.FFmpegAudio) {
	fmt.Println("\n=== 错误处理测试 ===")

	ctx := context.Background()
	outputDir := "./test_output/error_test"
	os.MkdirAll(outputDir, 0755)

	// 测试1: 不存在的输入文件
	// 测试处理不存在文件时的错误处理
	fmt.Println("\n测试1: 不存在的输入文件")
	_, err := ffmpeg.ExtractAudioFromVideo(ctx, "/nonexistent/video.mp4", outputDir, &tffmpeg.AudioExtractionOptions{})
	if err != nil {
		fmt.Printf("✅ 正确处理不存在的文件: %v\n", err)
	} else {
		fmt.Println("❌ 应该返回错误但没有")
	}

	// 测试2: 无效的输出目录
	// 测试处理无效输出目录时的错误处理
	fmt.Println("\n测试2: 无效的输出目录")
	_, err = ffmpeg.ExtractAudioFromVideo(ctx, "/dev/null", "/invalid/path/that/does/not/exist", &tffmpeg.AudioExtractionOptions{})
	if err != nil {
		fmt.Printf("✅ 正确处理无效输出目录: %v\n", err)
	} else {
		fmt.Println("❌ 应该返回错误但没有")
	}

	// 测试3: 上下文取消测试
	// 测试上下文取消时的错误处理
	fmt.Println("\n测试3: 上下文取消测试")
	testVideoPath := findTestVideo()
	if testVideoPath != "" {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消上下文

		_, err = ffmpeg.ExtractAudioFromVideo(ctx, testVideoPath, outputDir, &tffmpeg.AudioExtractionOptions{})
		if err != nil {
			fmt.Printf("✅ 正确处理上下文取消: %v\n", err)
		} else {
			fmt.Println("❌ 应该返回错误但没有")
		}
	} else {
		fmt.Println("⚠️  跳过上下文取消测试（需要测试视频文件）")
	}
}

// runPerformanceTests 运行性能测试
// 测试FFmpeg音频处理工具的性能表现，包括不同质量级别的性能对比和内存使用监控
// 参数:
//   - ffmpeg: FFmpeg音频处理工具实例
func runPerformanceTests(ffmpeg *tffmpeg.FFmpegAudio) {
	fmt.Println("\n=== 性能测试 ===")

	testVideoPath := findTestVideo()
	if testVideoPath == "" {
		fmt.Println("⚠️  未找到测试视频文件，跳过性能测试")
		return
	}

	outputDir := "./test_output/performance"
	os.MkdirAll(outputDir, 0755)

	// 测试1: 不同质量级别的性能对比
	// 测试不同质量级别对处理时间和文件大小的影响
	fmt.Println("\n测试1: 不同质量级别的性能对比")
	qualities := []tffmpeg.AudioQuality{
		tffmpeg.AudioQualityLow, tffmpeg.AudioQualityMedium, tffmpeg.AudioQualityHigh, // 测试三种质量级别
	}

	for _, quality := range qualities {
		startTime := time.Now()

		options := &tffmpeg.AudioExtractionOptions{
			Format:     tffmpeg.AudioFormatMP3,
			Quality:    quality,
			SampleRate: 44100,
			Channels:   2,
		}

		qualityOutputDir := filepath.Join(outputDir, string(quality))
		os.MkdirAll(qualityOutputDir, 0755)

		ctx := context.Background()
		result, err := ffmpeg.ExtractAudioFromVideo(ctx, testVideoPath, qualityOutputDir, options)

		elapsed := time.Since(startTime)

		if err != nil {
			fmt.Printf("   ❌ 质量 %s 失败: %v\n", quality, err)
		} else {
			fmt.Printf("   ✅ 质量 %s: 耗时 %.2f秒, 文件大小 %d bytes\n",
				quality, elapsed.Seconds(), result.FileSize)
		}
	}

	// 测试2: 内存使用监控
	// 监控音频处理过程中的内存使用情况
	fmt.Println("\n测试2: 内存使用监控")
	var m1, m2 runtime.MemStats
	runtime.GC()              // 强制垃圾回收
	runtime.ReadMemStats(&m1) // 记录处理前的内存状态

	options := &tffmpeg.AudioExtractionOptions{
		Format:     tffmpeg.AudioFormatMP3,
		Quality:    tffmpeg.AudioQualityMedium,
		SampleRate: 44100,
		Channels:   2,
	}

	ctx := context.Background()
	result, err := ffmpeg.ExtractAudioFromVideo(ctx, testVideoPath, outputDir, options)

	if err != nil {
		fmt.Printf("❌ 内存测试失败: %v\n", err)
	} else {
		runtime.GC()
		runtime.ReadMemStats(&m2)

		memUsed := m2.Alloc - m1.Alloc
		fmt.Printf("✅ 内存使用测试完成: %s\n", result.OutputPath)
		fmt.Printf("   内存使用: %d KB, 文件大小: %d bytes, 处理时间: %.2f秒\n",
			memUsed/1024, result.FileSize, result.ProcessingTime)
	}
}

// findTestVideo 查找测试视频文件
// 在多个可能的路径中查找测试视频文件
// 返回:
//   - string: 找到的测试视频文件路径，如果未找到返回空字符串
func findTestVideo() string {
	possiblePaths := []string{
		"./test_video.mp4", // 当前目录下的测试视频
		"./downloads/69123bca-1490-440e-9982-3aaf95a5aa8a/e412337e23b25dcea2876205b4bd29dd.mp4", // 下载目录中的视频
		"./downloads/test.mp4", // 下载目录中的测试视频
		"./test.mp4",           // 当前目录下的测试视频
		"../downloads/69123bca-1490-440e-9982-3aaf95a5aa8a/e412337e23b25dcea2876205b4bd29dd.mp4", // 上级目录的下载视频
		"../downloads/test.mp4", // 上级目录的测试视频
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
