package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/stones-hub/taurus-pro-common/pkg/util/tvideo"
)

// 进度回调处理函数
func createProgressCallback() func(progress *tvideo.KeyframeProgress) {
	var frameCount int
	var lastProgress float64

	return func(progress *tvideo.KeyframeProgress) {
		// 处理日志消息
		if progress.LogMessage != "" {
			fmt.Printf("[进度日志] %s\n", progress.LogMessage)
			return
		}

		// 更新进度
		currentProgress := progress.Coverage * 100
		if currentProgress > lastProgress {
			fmt.Printf("\r处理进度: %.1f%%...", currentProgress)
			os.Stdout.Sync()
			lastProgress = currentProgress
		}

		// 处理新关键帧信息
		if progress.NewFramePath != "" {
			frameCount++
			fmt.Printf("\n----------------------------------------\n")
			fmt.Printf("✓ 关键帧 #%d 已保存\n", frameCount)
			fmt.Printf("  文件名: %s\n", filepath.Base(progress.NewFramePath))
			fmt.Printf("  时间点: %.2f秒\n", progress.NewFrameTimestamp)
			fmt.Printf("  分辨率: %dx%d\n", progress.Width, progress.Height)
			fmt.Printf("  大小　: %.2fKB\n", float64(progress.FileSize)/1024)
			fmt.Printf("  变化度: %.2f\n", progress.ChangeScore)
			fmt.Printf("  质量度: %.2f\n", progress.QualityScore)
			fmt.Printf("----------------------------------------\n")
		}
	}
}

// 演示智能提取模式
func demonstrateSmartMode(extractor *tvideo.KeyframeExtractor, videoPath string) {
	fmt.Println("\n=== 智能提取模式演示 ===")

	// 创建输出目录
	outputDir := filepath.Join(filepath.Dir(videoPath), "smart_mode_output")

	// 配置选项
	options := &tvideo.KeyframeExtractionOptions{
		MaxFrames:        200,
		Mode:             tvideo.ModeSmart,
		OutputDir:        outputDir,
		EnableDebug:      true,
		ProgressCallback: createProgressCallback(),
	}

	// 执行提取
	startTime := time.Now()
	result, debugInfo, err := extractor.ExtractKeyframes(context.Background(), videoPath, options)
	if err != nil {
		log.Fatalf("智能提取失败: %v", err)
	}

	// 打印结果
	fmt.Printf("\n提取完成! 耗时: %.2f秒\n", time.Since(startTime).Seconds())
	fmt.Printf("总帧数: %d\n", result.TotalFrames)
	fmt.Printf("视频信息: %+v\n", result.VideoInfo)
	if debugInfo != nil {
		fmt.Printf("调试信息: %+v\n", debugInfo.Performance)
	}
}

// 演示均匀分布模式
func demonstrateUniformMode(extractor *tvideo.KeyframeExtractor, videoPath string) {
	fmt.Println("\n=== 均匀分布模式演示 ===")

	outputDir := filepath.Join(filepath.Dir(videoPath), "uniform_mode_output")

	options := &tvideo.KeyframeExtractionOptions{
		MaxFrames: 100,
		Mode:      tvideo.ModeUniform,
		OutputDir: outputDir,
	}

	result, _, err := extractor.ExtractKeyframes(context.Background(), videoPath, options)
	if err != nil {
		log.Fatalf("均匀提取失败: %v", err)
	}

	fmt.Printf("\n均匀提取完成!\n")
	fmt.Printf("提取帧数: %d\n", len(result.KeyframePaths))
}

// 演示时间间隔模式
func demonstrateIntervalMode(extractor *tvideo.KeyframeExtractor, videoPath string) {
	fmt.Println("\n=== 时间间隔模式演示 ===")

	outputDir := filepath.Join(filepath.Dir(videoPath), "interval_mode_output")

	interval := 2.0 // 2秒间隔
	options := &tvideo.KeyframeExtractionOptions{
		MaxFrames:    150,
		Mode:         tvideo.ModeInterval,
		TimeInterval: &interval,
		OutputDir:    outputDir,
	}

	result, _, err := extractor.ExtractKeyframes(context.Background(), videoPath, options)
	if err != nil {
		log.Fatalf("间隔提取失败: %v", err)
	}

	fmt.Printf("\n间隔提取完成!\n")
	fmt.Printf("提取帧数: %d\n", len(result.KeyframePaths))
}

// 演示预设配置
func demonstratePresets(extractor *tvideo.KeyframeExtractor, videoPath string) {
	fmt.Println("\n=== 预设配置演示 ===")

	presets := []string{"fast", "quality", "interval"}

	for _, preset := range presets {
		fmt.Printf("\n使用预设: %s\n", preset)

		outputDir := filepath.Join(filepath.Dir(videoPath), fmt.Sprintf("%s_preset_output", preset))
		options := extractor.CreatePresetOptions(preset)
		options.OutputDir = outputDir

		result, _, err := extractor.ExtractKeyframes(context.Background(), videoPath, options)
		if err != nil {
			fmt.Printf("预设 %s 提取失败: %v\n", preset, err)
			continue
		}

		fmt.Printf("预设 %s 提取完成: %d帧\n", preset, len(result.KeyframePaths))
	}
}

// go run comprehensive.go /path/to/video.mp4
func main() {
	// 检查命令行参数
	if len(os.Args) != 2 {
		fmt.Printf("用法: %s <视频文件路径>\n", os.Args[0])
		os.Exit(1)
	}

	videoPath := os.Args[1]

	// 创建提取器实例
	extractor, err := tvideo.NewKeyframeExtractor()
	if err != nil {
		log.Fatalf("创建提取器失败: %v", err)
	}
	defer extractor.Close()

	// 获取视频信息
	fmt.Println("=== 视频信息 ===")
	info, err := extractor.GetVideoInfo(context.Background(), videoPath)
	if err != nil {
		log.Fatalf("获取视频信息失败: %v", err)
	}
	fmt.Printf("分辨率: %dx%d\n", info.Width, info.Height)
	fmt.Printf("帧率: %.2f fps\n", info.FPS)
	fmt.Printf("时长: %.2f 秒\n", info.Duration)
	fmt.Printf("总帧数: %d\n", info.TotalFrames)

	// 演示不同模式
	demonstrateSmartMode(extractor, videoPath)
	demonstrateUniformMode(extractor, videoPath)
	demonstrateIntervalMode(extractor, videoPath)
	demonstratePresets(extractor, videoPath)
}
