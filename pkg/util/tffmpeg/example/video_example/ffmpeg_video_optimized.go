package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/util/tffmpeg"
)

// 示例配置
const (
	inputVideo = "input.mp4"
)

func main() {
	// 创建示例目录
	examples := []string{
		"basic",
		"high_quality",
		"progress",
		"format_convert",
		"clip",
		"compress",
		"no_audio",
	}

	for _, dir := range examples {
		outputDir := filepath.Join("output", dir)
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("创建目录失败 %s: %v", outputDir, err)
		}
	}

	// 创建FFmpeg视频处理工具实例
	ffmpeg := tffmpeg.NewFFmpegVideo(
		tffmpeg.WithVideoFFmpegPath("/usr/local/bin/ffmpeg"),
		tffmpeg.WithVideoFFprobePath("/usr/local/bin/ffprobe"),
	)

	// 运行所有示例
	fmt.Println("开始运行视频处理示例...")

	runBasicExample(ffmpeg)
	runHighQualityExample(ffmpeg)
	runProgressExample(ffmpeg)
	runFormatConvertExample(ffmpeg)
	runClipExample(ffmpeg)
	runCompressExample(ffmpeg)
	runNoAudioExample(ffmpeg)

	fmt.Println("\n所有示例执行完成！")

	// 清理输出目录
	os.RemoveAll("output")
}

// runBasicExample 基本视频处理示例
func runBasicExample(ffmpeg *tffmpeg.FFmpegVideo) {
	fmt.Println("\n=== 运行基本视频处理示例 ===")
	outputDir := filepath.Join("output", "basic")

	options := &tffmpeg.VideoExtractionOptions{
		Format:       tffmpeg.VideoFormatMP4,
		Quality:      tffmpeg.VideoQualityMedium,
		Resolution:   "1280x720",
		Fps:          30,
		AudioEnabled: true,
		StartTime:    "00:00:15.15",
		Duration:     "00:00:10.50",
	}

	result, err := ffmpeg.ExtractVideo(context.Background(), inputVideo, outputDir, "", options)
	if err != nil {
		log.Printf("基本视频处理失败: %v\n", err)
		return
	}

	fmt.Printf("基本视频处理完成:\n")
	fmt.Printf("- 输出文件: %s\n", result.OutputPath)
	fmt.Printf("- 分辨率: %s\n", result.Resolution)
	fmt.Printf("- 时长: %.2f秒\n", result.Duration)
}

// runHighQualityExample 高质量视频处理示例
func runHighQualityExample(ffmpeg *tffmpeg.FFmpegVideo) {
	fmt.Println("\n=== 运行高质量视频处理示例 ===")
	outputDir := filepath.Join("output", "high_quality")

	options := &tffmpeg.VideoExtractionOptions{
		Format:       tffmpeg.VideoFormatMP4,
		Quality:      tffmpeg.VideoQualityBest,
		Resolution:   "1920x1080",
		Bitrate:      "8000k",
		Fps:          60,
		AudioBitrate: "320k",
		AudioEnabled: true,
		StartTime:    "00:00:25",
		Duration:     "00:00:10",
	}

	result, err := ffmpeg.ExtractVideo(context.Background(), inputVideo, outputDir, "", options)
	if err != nil {
		log.Printf("高质量视频处理失败: %v\n", err)
		return
	}

	fmt.Printf("高质量视频处理完成:\n")
	fmt.Printf("- 输出文件: %s\n", result.OutputPath)
	fmt.Printf("- 分辨率: %s\n", result.Resolution)
	fmt.Printf("- 比特率: %s\n", result.Bitrate)
	fmt.Printf("- 帧率: %.2f\n", result.Fps)
}

// runProgressExample 带进度显示的视频处理示例
func runProgressExample(ffmpeg *tffmpeg.FFmpegVideo) {
	fmt.Println("\n=== 运行带进度显示的视频处理示例 ===")
	outputDir := filepath.Join("output", "progress")

	startTime := time.Now()
	options := &tffmpeg.VideoExtractionOptions{
		Format:         tffmpeg.VideoFormatMP4,
		Quality:        tffmpeg.VideoQualityHigh,
		Resolution:     "1920x1080",
		AudioEnabled:   true,   // 明确启用音频
		AudioBitrate:   "192k", // 设置音频比特率
		EnableProgress: true,
		StartTime:      "00:00:05",
		Duration:       "00:00:10",
		ProgressCallback: func(progress *tffmpeg.FFmpegProgress) {
			elapsed := time.Since(startTime)
			fmt.Printf("\r处理进度: %.2f%%, 当前时间: %s, 总时长: %s, 速度: %s, 已用时间: %v",
				progress.Progress,
				progress.CurrentTime,
				progress.TotalTime,
				progress.Speed,
				elapsed.Round(time.Second))
			log.Printf("处理进度: %.2f%%, 当前时间: %s, 总时长: %s, 速度: %s, 已用时间: %v", progress.Progress, progress.CurrentTime, progress.TotalTime, progress.Speed, elapsed.Round(time.Second))
		},
	}

	result, err := ffmpeg.ExtractVideo(context.Background(), inputVideo, outputDir, "", options)
	if err != nil {
		log.Printf("\n带进度显示的视频处理失败: %v\n", err)
		return
	}

	fmt.Printf("\n带进度显示的视频处理完成:\n")
	fmt.Printf("- 输出文件: %s\n", result.OutputPath)
	fmt.Printf("- 处理时间: %.2f秒\n", result.ProcessingTime)
}

// runFormatConvertExample 视频格式转换示例
func runFormatConvertExample(ffmpeg *tffmpeg.FFmpegVideo) {
	fmt.Println("\n=== 运行视频格式转换示例 ===")
	outputDir := filepath.Join("output", "format_convert")

	formats := []tffmpeg.VideoFormat{
		tffmpeg.VideoFormatWEBM,
		tffmpeg.VideoFormatMKV,
		tffmpeg.VideoFormatMOV,
	}

	for _, format := range formats {
		fmt.Printf("\n转换为 %s 格式...\n", format)
		options := &tffmpeg.VideoExtractionOptions{
			Format:       format,
			Quality:      tffmpeg.VideoQualityHigh,
			AudioEnabled: true,
			StartTime:    "00:00:05",
			Duration:     "00:00:10",
		}

		result, err := ffmpeg.ExtractVideo(context.Background(), inputVideo, outputDir, "", options)
		if err != nil {
			log.Printf("转换到%s格式失败: %v\n", format, err)
			continue
		}

		fmt.Printf("转换完成: %s\n", filepath.Base(result.OutputPath))
	}
}

// runClipExample 视频片段提取示例
func runClipExample(ffmpeg *tffmpeg.FFmpegVideo) {
	fmt.Println("\n=== 运行视频片段提取示例 ===")
	outputDir := filepath.Join("output", "clip")

	options := &tffmpeg.VideoExtractionOptions{
		Format:       tffmpeg.VideoFormatMP4,
		Quality:      tffmpeg.VideoQualityHigh,
		StartTime:    "00:00:05",
		Duration:     "00:00:10",
		Resolution:   "1280x720",
		AudioEnabled: true,
	}

	result, err := ffmpeg.ExtractVideo(context.Background(), inputVideo, outputDir, "", options)
	if err != nil {
		log.Printf("视频片段提取失败: %v\n", err)
		return
	}

	fmt.Printf("视频片段提取完成:\n")
	fmt.Printf("- 输出文件: %s\n", result.OutputPath)
	fmt.Printf("- 片段时长: %.2f秒\n", result.Duration)
}

// runCompressExample 视频压缩示例
func runCompressExample(ffmpeg *tffmpeg.FFmpegVideo) {
	fmt.Println("\n=== 运行视频压缩示例 ===")
	outputDir := filepath.Join("output", "compress")

	options := &tffmpeg.VideoExtractionOptions{
		Format:       tffmpeg.VideoFormatMP4,
		Quality:      tffmpeg.VideoQualityLow,
		Resolution:   "854x480",
		Bitrate:      "800k",
		Fps:          24,
		AudioBitrate: "64k",
		AudioEnabled: true,
		StartTime:    "00:00:05",
		Duration:     "00:00:10",
	}

	result, err := ffmpeg.ExtractVideo(context.Background(), inputVideo, outputDir, "", options)
	if err != nil {
		log.Printf("视频压缩失败: %v\n", err)
		return
	}

	// 获取原始文件大小
	originalInfo, _ := os.Stat(inputVideo)
	originalSize := originalInfo.Size()

	compressionRatio := float64(originalSize) / float64(result.FileSize)
	fmt.Printf("视频压缩完成:\n")
	fmt.Printf("- 输出文件: %s\n", result.OutputPath)
	fmt.Printf("- 原始大小: %.2f MB\n", float64(originalSize)/1024/1024)
	fmt.Printf("- 压缩后大小: %.2f MB\n", float64(result.FileSize)/1024/1024)
	fmt.Printf("- 压缩比: %.2f:1\n", compressionRatio)
}

// runNoAudioExample 无音频视频处理示例
func runNoAudioExample(ffmpeg *tffmpeg.FFmpegVideo) {
	fmt.Println("\n=== 运行无音频视频处理示例 ===")
	outputDir := filepath.Join("output", "no_audio")

	options := &tffmpeg.VideoExtractionOptions{
		Format:       tffmpeg.VideoFormatMP4,
		Quality:      tffmpeg.VideoQualityHigh,
		Resolution:   "1920x1080",
		AudioEnabled: false,
		StartTime:    "00:00:05",
		Duration:     "00:00:10",
	}

	result, err := ffmpeg.ExtractVideo(context.Background(), inputVideo, outputDir, "", options)
	if err != nil {
		log.Printf("无音频视频处理失败: %v\n", err)
		return
	}

	fmt.Printf("无音频视频处理完成:\n")
	fmt.Printf("- 输出文件: %s\n", result.OutputPath)
	fmt.Printf("- 是否包含音频: %v\n", result.HasAudio)
}
