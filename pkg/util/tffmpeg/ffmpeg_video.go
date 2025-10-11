package tffmpeg

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// VideoFormat 视频格式枚举
type VideoFormat string

const (
	VideoFormatMP4  VideoFormat = "mp4"  // MP4格式，最常用的视频格式
	VideoFormatMOV  VideoFormat = "mov"  // MOV格式，苹果设备常用格式
	VideoFormatAVI  VideoFormat = "avi"  // AVI格式，Windows常用格式
	VideoFormatMKV  VideoFormat = "mkv"  // MKV格式，支持多种编码和字幕
	VideoFormatWEBM VideoFormat = "webm" // WEBM格式，网页视频常用格式
	VideoFormatFLV  VideoFormat = "flv"  // FLV格式，Flash视频格式
)

// VideoQuality 视频质量枚举
type VideoQuality string

const (
	VideoQualityLow    VideoQuality = "low"    // 低质量：文件小，画质一般
	VideoQualityMedium VideoQuality = "medium" // 中等质量：平衡文件大小和画质
	VideoQualityHigh   VideoQuality = "high"   // 高质量：画质好，文件较大
	VideoQualityBest   VideoQuality = "best"   // 最佳质量：最高画质，文件最大
)

// VideoExtractionOptions 视频提取选项
type VideoExtractionOptions struct {
	Format           VideoFormat                    `json:"format"`          // 输出格式
	Quality          VideoQuality                   `json:"quality"`         // 质量等级
	Resolution       string                         `json:"resolution"`      // 分辨率，如"1920x1080"
	Bitrate          string                         `json:"bitrate"`         // 视频比特率，如"2000k"
	Fps              int                            `json:"fps"`             // 帧率，如30、60
	StartTime        string                         `json:"start_time"`      // 开始时间，格式如"00:00:10"
	Duration         string                         `json:"duration"`        // 持续时间，格式如"00:00:30"
	AudioEnabled     bool                           `json:"audio_enabled"`   // 是否包含音频
	AudioBitrate     string                         `json:"audio_bitrate"`   // 音频比特率
	EnableProgress   bool                           `json:"enable_progress"` // 是否启用进度监控
	ProgressCallback func(progress *FFmpegProgress) `json:"-"`               // 进度回调函数
}

// VideoExtractionResult 视频提取结果
type VideoExtractionResult struct {
	OutputPath     string    `json:"output_path"`     // 输出文件路径
	Duration       float64   `json:"duration"`        // 视频时长（秒）
	FileSize       int64     `json:"file_size"`       // 文件大小（字节）
	Format         string    `json:"format"`          // 视频格式
	Resolution     string    `json:"resolution"`      // 分辨率
	Bitrate        string    `json:"bitrate"`         // 比特率
	Fps            float64   `json:"fps"`             // 帧率
	HasAudio       bool      `json:"has_audio"`       // 是否包含音频
	ExtractedAt    time.Time `json:"extracted_at"`    // 提取时间
	ProcessingTime float64   `json:"processing_time"` // 处理耗时（秒）
}

// VideoInfo 视频信息结构
type VideoInfo struct {
	Duration   float64 `json:"duration"`   // 视频时长（秒）
	Bitrate    string  `json:"bitrate"`    // 比特率
	Resolution string  `json:"resolution"` // 分辨率
	Fps        float64 `json:"fps"`        // 帧率
	HasAudio   bool    `json:"has_audio"`  // 是否包含音频
}

// FFmpegVideo FFmpeg视频处理工具
type FFmpegVideo struct {
	ffmpegPath     string  // FFmpeg可执行文件路径
	ffprobePath    string  // FFprobe可执行文件路径
	tempDir        string  // 临时目录
	targetDuration float64 // 目标处理时长
}

// FFmpegVideoOption 配置选项函数类型
type FFmpegVideoOption func(*FFmpegVideo)

// WithVideoFFmpegPath 设置FFmpeg路径
func WithVideoFFmpegPath(path string) FFmpegVideoOption {
	return func(f *FFmpegVideo) {
		f.ffmpegPath = path
	}
}

// WithVideoFFprobePath 设置FFprobe路径
func WithVideoFFprobePath(path string) FFmpegVideoOption {
	return func(f *FFmpegVideo) {
		f.ffprobePath = path
	}
}

// WithVideoTempDir 设置临时目录
func WithVideoTempDir(dir string) FFmpegVideoOption {
	return func(f *FFmpegVideo) {
		f.tempDir = dir
	}
}

// NewFFmpegVideo 创建FFmpeg视频处理工具
func NewFFmpegVideo(opts ...FFmpegVideoOption) *FFmpegVideo {
	// 设置默认值
	tempDir, err := os.MkdirTemp("", "ffmpeg_video")
	if err != nil {
		tempDir = "/tmp/ffmpeg_video"
	}
	// 检查 tempDir 是否存在，不存在创建
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		os.MkdirAll(tempDir, 0755)
	}

	f := &FFmpegVideo{
		ffmpegPath:  "/usr/local/bin/ffmpeg",  // 默认使用系统PATH中的ffmpeg
		ffprobePath: "/usr/local/bin/ffprobe", // 默认使用系统PATH中的ffprobe
		tempDir:     tempDir,                  // 默认临时目录
	}

	// 应用选项
	for _, opt := range opts {
		opt(f)
	}

	return f
}

// ExtractVideo 从视频中提取片段（主要入口方法）
// 这是最常用的方法，支持从视频文件中提取指定时间段的片段并保存为指定格式
func (f *FFmpegVideo) ExtractVideo(ctx context.Context, videoPath string, outputDir string, options *VideoExtractionOptions) (*VideoExtractionResult, error) {
	// 验证FFmpeg可用性
	if err := f.CheckFFmpegAvailable(); err != nil {
		return nil, fmt.Errorf("FFmpeg检查失败: %v", err)
	}

	// 创建默认配置
	config := f.createDefaultConfig(options)

	// 验证配置
	if err := f.validateConfig(config, videoPath); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	// 执行视频提取
	if options.EnableProgress {
		return f.extractVideoWithProgress(ctx, videoPath, outputDir, config, options.ProgressCallback)
	} else {
		return f.extractVideo(ctx, videoPath, outputDir, config)
	}
}

// extractVideo 基本视频提取
func (f *FFmpegVideo) extractVideo(ctx context.Context, videoPath string, outputDir string, config *VideoExtractionConfig) (*VideoExtractionResult, error) {
	startTime := time.Now()

	// 验证输入文件
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("视频文件不存在: %s", videoPath)
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 生成输出文件名
	inputName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	outputFileName := fmt.Sprintf("%s_clip.%s", inputName, config.Format)
	outputPath := filepath.Join(outputDir, outputFileName)

	// 构建FFmpeg命令
	args := f.buildFFmpegArgs(videoPath, outputPath, config)

	// 执行FFmpeg命令
	cmd := exec.CommandContext(ctx, f.ffmpegPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("FFmpeg执行失败: %v", err)
	}

	// 获取输出文件信息
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("获取输出文件信息失败: %v", err)
	}

	// 获取视频信息
	videoInfo, err := f.getVideoInfo(ctx, outputPath)
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %v", err)
	}

	processingTime := time.Since(startTime).Seconds()

	return &VideoExtractionResult{
		OutputPath:     outputPath,
		Duration:       videoInfo.Duration,
		FileSize:       fileInfo.Size(),
		Format:         string(config.Format),
		Resolution:     videoInfo.Resolution,
		Bitrate:        videoInfo.Bitrate,
		Fps:            videoInfo.Fps,
		HasAudio:       videoInfo.HasAudio,
		ExtractedAt:    time.Now(),
		ProcessingTime: processingTime,
	}, nil
}

// extractVideoWithProgress 带进度监控的视频提取
func (f *FFmpegVideo) extractVideoWithProgress(ctx context.Context, videoPath string, outputDir string, config *VideoExtractionConfig, progressCallback func(progress *FFmpegProgress)) (*VideoExtractionResult, error) {
	startTime := time.Now()

	// 验证输入文件
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("视频文件不存在: %s", videoPath)
	}

	// 获取视频信息并初始化时长信息
	videoInfo, err := f.getVideoInfo(ctx, videoPath)
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %v", err)
	}

	// 设置目标时长
	if config.Duration != "" {
		// 如果指定了Duration，使用指定的时长
		f.targetDuration = f.parseTimeToSeconds(config.Duration)
	} else {
		// 否则使用视频总时长
		f.targetDuration = videoInfo.Duration
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 生成输出文件名
	inputName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	outputFileName := fmt.Sprintf("%s_clip.%s", inputName, config.Format)
	outputPath := filepath.Join(outputDir, outputFileName)

	// 构建FFmpeg命令
	args := f.buildFFmpegArgs(videoPath, outputPath, config)

	// 执行FFmpeg命令并监控进度
	cmd := exec.CommandContext(ctx, f.ffmpegPath, args...)

	// 创建管道来捕获stderr输出
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("创建stderr管道失败: %v", err)
	}

	// 设置stdout也输出到管道，因为某些FFmpeg版本会将进度信息输出到stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("创建stdout管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动FFmpeg失败: %v", err)
	}

	// 创建通道用于同步进度监控
	doneStderr := make(chan struct{})
	doneStdout := make(chan struct{})

	// 监控stderr进度
	go func() {
		defer close(doneStderr)
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)
		f.monitorProgress(scanner, progressCallback)
	}()

	// 监控stdout进度
	go func() {
		defer close(doneStdout)
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		f.monitorProgress(scanner, progressCallback)
	}()

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("FFmpeg执行失败: %v", err)
	}

	// 等待进度监控完成
	<-doneStderr
	<-doneStdout

	// 获取输出文件信息
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("获取输出文件信息失败: %v", err)
	}

	// 获取生成后的视频信息
	videoInfo, err = f.getVideoInfo(ctx, outputPath)
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %v", err)
	}

	processingTime := time.Since(startTime).Seconds()

	return &VideoExtractionResult{
		OutputPath:     outputPath,
		Duration:       videoInfo.Duration,
		FileSize:       fileInfo.Size(),
		Format:         string(config.Format),
		Resolution:     videoInfo.Resolution,
		Bitrate:        videoInfo.Bitrate,
		Fps:            videoInfo.Fps,
		HasAudio:       videoInfo.HasAudio,
		ExtractedAt:    time.Now(),
		ProcessingTime: processingTime,
	}, nil
}

// VideoExtractionConfig 内部使用的视频提取配置
type VideoExtractionConfig struct {
	Format       VideoFormat `json:"format"`        // 输出格式
	Resolution   string      `json:"resolution"`    // 分辨率
	Bitrate      string      `json:"bitrate"`       // 视频比特率
	Fps          int         `json:"fps"`           // 帧率
	Quality      int         `json:"quality"`       // 质量等级（0-51，0为最高质量）
	StartTime    string      `json:"start_time"`    // 开始时间
	Duration     string      `json:"duration"`      // 持续时间
	AudioEnabled bool        `json:"audio_enabled"` // 是否包含音频
	AudioBitrate string      `json:"audio_bitrate"` // 音频比特率
}

// createDefaultConfig 创建默认配置
func (f *FFmpegVideo) createDefaultConfig(options *VideoExtractionOptions) *VideoExtractionConfig {
	config := &VideoExtractionConfig{
		Format:       VideoFormatMP4, // 默认MP4格式
		Resolution:   "1280x720",     // 默认720p
		Bitrate:      "2000k",        // 默认视频比特率
		Fps:          30,             // 默认帧率
		Quality:      23,             // 默认质量等级
		AudioEnabled: true,           // 默认包含音频
		AudioBitrate: "128k",         // 默认音频比特率
	}

	// 应用用户选项
	if options != nil {
		if options.Format != "" {
			config.Format = options.Format
		}

		if options.Resolution != "" {
			config.Resolution = options.Resolution
		}

		if options.Bitrate != "" {
			config.Bitrate = options.Bitrate
		}

		if options.Fps > 0 {
			config.Fps = options.Fps
		}

		if options.StartTime != "" {
			config.StartTime = options.StartTime
		}

		if options.Duration != "" {
			config.Duration = options.Duration
		}

		config.AudioEnabled = options.AudioEnabled
		if options.AudioBitrate != "" {
			config.AudioBitrate = options.AudioBitrate
		}

		// 根据质量等级设置参数
		config.Quality = f.getQualityLevel(options.Quality)
		if options.Bitrate == "" {
			config.Bitrate = f.getQualityBitrate(options.Quality)
		}
	}

	return config
}

// getQualityLevel 获取质量等级对应的数值
func (f *FFmpegVideo) getQualityLevel(quality VideoQuality) int {
	switch quality {
	case VideoQualityLow:
		return 28
	case VideoQualityMedium:
		return 23
	case VideoQualityHigh:
		return 18
	case VideoQualityBest:
		return 15
	default:
		return 23
	}
}

// getQualityBitrate 获取质量等级对应的比特率
func (f *FFmpegVideo) getQualityBitrate(quality VideoQuality) string {
	switch quality {
	case VideoQualityLow:
		return "1000k"
	case VideoQualityMedium:
		return "2000k"
	case VideoQualityHigh:
		return "4000k"
	case VideoQualityBest:
		return "8000k"
	default:
		return "2000k"
	}
}

// buildFFmpegArgs 构建FFmpeg命令行参数
func (f *FFmpegVideo) buildFFmpegArgs(inputPath, outputPath string, config *VideoExtractionConfig) []string {
	args := []string{
		"-i", inputPath, // 输入文件
	}

	// 设置开始时间
	if config.StartTime != "" {
		args = append(args, "-ss", config.StartTime)
	}

	// 设置持续时间
	if config.Duration != "" {
		args = append(args, "-t", config.Duration)
	}

	// 设置视频编码器和参数
	switch config.Format {
	case VideoFormatMP4, VideoFormatMOV:
		args = append(args,
			"-c:v", "libx264", // 使用H.264编码
			"-preset", "medium", // 编码速度预设
			"-profile:v", "high", // 使用高规格
			"-level", "4.0", // 兼容性级别
			"-pix_fmt", "yuv420p", // 使用广泛支持的像素格式
		)
	case VideoFormatWEBM:
		args = append(args,
			"-c:v", "libvpx-vp9", // 使用VP9编码
			"-deadline", "good", // 平衡质量和速度
			"-cpu-used", "2", // CPU使用级别
			"-row-mt", "1", // 启用行级多线程
		)
	case VideoFormatFLV:
		args = append(args,
			"-c:v", "libx264", // 使用H.264编码
			"-preset", "medium", // 编码速度预设
			"-pix_fmt", "yuv420p", // 使用广泛支持的像素格式
		)
	default:
		args = append(args,
			"-c:v", "libx264", // 默认使用H.264编码
			"-preset", "medium", // 编码速度预设
			"-pix_fmt", "yuv420p", // 使用广泛支持的像素格式
		)
	}

	// 设置视频参数
	if config.Resolution != "" {
		// 解析分辨率
		parts := strings.Split(config.Resolution, "x")
		if len(parts) == 2 {
			width := parts[0]
			height := parts[1]
			// 使用更安全的缩放和填充参数
			args = append(args,
				"-vf", fmt.Sprintf("scale=w=%s:h=%s:force_original_aspect_ratio=decrease,pad=%s:%s:(ow-iw)/2:(oh-ih)/2:color=black",
					width, height, width, height),
			)
		}
	}

	// 设置比特率和质量
	if config.Bitrate != "" {
		args = append(args,
			"-b:v", config.Bitrate,
			"-maxrate", config.Bitrate,
			"-bufsize", config.Bitrate,
		)
	}
	if config.Fps > 0 {
		args = append(args,
			"-r", fmt.Sprintf("%d", config.Fps),
			"-fps_mode", "cfr", // 使用恒定帧率模式
		)
	}
	if config.Quality >= 0 {
		args = append(args, "-crf", fmt.Sprintf("%d", config.Quality))
	}

	// 音频设置
	if config.AudioEnabled {
		// 复制原始音频流
		args = append(args, "-map", "0:v:0") // 选择第一个视频流
		args = append(args, "-map", "0:a:0") // 选择第一个音频流

		switch config.Format {
		case VideoFormatMP4, VideoFormatMOV:
			args = append(args,
				"-c:a", "aac", // 使用AAC音频编码
				"-strict", "-2", // 允许实验性编码器
				"-ar", "44100", // 标准采样率
				"-ac", "2", // 双声道
			)
		case VideoFormatWEBM:
			args = append(args,
				"-c:a", "libopus", // 使用Opus音频编码
				"-ar", "48000", // WebM推荐采样率
				"-ac", "2", // 双声道
			)
		default:
			args = append(args,
				"-c:a", "aac", // 默认使用AAC音频编码
				"-ar", "44100", // 标准采样率
				"-ac", "2", // 双声道
			)
		}

		if config.AudioBitrate != "" {
			args = append(args, "-b:a", config.AudioBitrate)
		}
	} else {
		args = append(args, "-map", "0:v:0") // 只选择视频流
		args = append(args, "-an")           // 禁用音频
	}

	// 其他通用参数
	args = append(args,
		"-movflags", "+faststart", // 优化网络播放
		"-threads", "0", // 自动选择线程数
		"-max_muxing_queue_size", "1024", // 增加复用队列大小
		"-progress", "pipe:1", // 强制进度输出到stdout
		"-stats",   // 输出详细统计信息
		"-y",       // 覆盖输出文件
		outputPath, // 输出文件路径
	)

	return args
}

// monitorProgress 监控FFmpeg进度
func (f *FFmpegVideo) monitorProgress(scanner *bufio.Scanner, progressCallback func(progress *FFmpegProgress)) {
	for scanner.Scan() {
		line := scanner.Text()

		// 解析FFmpeg进度信息
		progress := f.parseProgressLine(line)
		if progress != nil && progressCallback != nil {
			progressCallback(progress)
		}
	}
}

// parseProgressLine 解析FFmpeg进度行
func (f *FFmpegVideo) parseProgressLine(line string) *FFmpegProgress {
	// 匹配时间信息
	timeRegex := regexp.MustCompile(`time=\s*(\d{2}:\d{2}:\d{2}(?:\.\d{2})?)`)
	timeMatch := timeRegex.FindStringSubmatch(line)

	// 匹配速度
	speedRegex := regexp.MustCompile(`speed=\s*([\d.]+)x`)
	speedMatch := speedRegex.FindStringSubmatch(line)

	// 如果没有找到时间信息，返回nil
	if len(timeMatch) < 2 {
		return nil
	}

	// 计算进度百分比
	var progress float64
	currentTime := f.parseTimeToSeconds(timeMatch[1])
	if f.targetDuration > 0 {
		progress = (currentTime / f.targetDuration) * 100
		if progress > 100 {
			progress = 100
		}
	}

	speed := "1.0x"
	if len(speedMatch) >= 2 {
		speed = speedMatch[1] + "x"
	}

	// 格式化总时长
	totalTime := ""
	if f.targetDuration > 0 {
		h := int(f.targetDuration) / 3600
		m := int(f.targetDuration) % 3600 / 60
		s := int(f.targetDuration) % 60
		ms := int((f.targetDuration - float64(int(f.targetDuration))) * 100)
		totalTime = fmt.Sprintf("%02d:%02d:%02d.%02d", h, m, s, ms)
	}

	return &FFmpegProgress{
		Progress:    progress,
		CurrentTime: timeMatch[1],
		TotalTime:   totalTime,
		Speed:       speed,
		Message:     line,
	}
}

// parseTimeToSeconds 将时间字符串转换为秒数
func (f *FFmpegVideo) parseTimeToSeconds(timeStr string) float64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return float64(hours*3600) + float64(minutes*60) + seconds
}

// getVideoInfo 获取视频文件信息
func (f *FFmpegVideo) getVideoInfo(ctx context.Context, videoPath string) (*VideoInfo, error) {
	// 使用ffprobe获取视频信息
	cmd := exec.CommandContext(ctx, f.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		videoPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe执行失败: %v", err)
	}

	// 解析JSON输出
	var probeResult struct {
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
		Streams []struct {
			CodecType  string `json:"codec_type"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &probeResult); err != nil {
		return nil, fmt.Errorf("解析ffprobe输出失败: %v", err)
	}

	// 查找视频流
	var videoStream *struct {
		CodecType  string `json:"codec_type"`
		Width      int    `json:"width"`
		Height     int    `json:"height"`
		RFrameRate string `json:"r_frame_rate"`
	}

	hasAudio := false
	for _, stream := range probeResult.Streams {
		if stream.CodecType == "video" && videoStream == nil {
			videoStream = &stream
		} else if stream.CodecType == "audio" {
			hasAudio = true
		}
	}

	if videoStream == nil {
		return nil, fmt.Errorf("未找到视频流")
	}

	// 解析时长
	duration, _ := strconv.ParseFloat(probeResult.Format.Duration, 64)

	// 解析帧率
	fps := 0.0
	if videoStream.RFrameRate != "" {
		parts := strings.Split(videoStream.RFrameRate, "/")
		if len(parts) == 2 {
			num, _ := strconv.ParseFloat(parts[0], 64)
			den, _ := strconv.ParseFloat(parts[1], 64)
			if den > 0 {
				fps = num / den
			}
		}
	}

	// 构建分辨率字符串
	resolution := fmt.Sprintf("%dx%d", videoStream.Width, videoStream.Height)

	return &VideoInfo{
		Duration:   duration,
		Bitrate:    probeResult.Format.BitRate + "bps",
		Resolution: resolution,
		Fps:        fps,
		HasAudio:   hasAudio,
	}, nil
}

// CheckFFmpegAvailable 检查FFmpeg是否可用
func (f *FFmpegVideo) CheckFFmpegAvailable() error {
	// 检查ffmpeg
	if err := exec.Command(f.ffmpegPath, "-version").Run(); err != nil {
		return fmt.Errorf("FFmpeg不可用: %v", err)
	}

	// 检查ffprobe
	if err := exec.Command(f.ffprobePath, "-version").Run(); err != nil {
		return fmt.Errorf("FFprobe不可用: %v", err)
	}

	return nil
}

// validateConfig 验证配置
func (f *FFmpegVideo) validateConfig(config *VideoExtractionConfig, videoPath string) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证格式
	validFormats := []VideoFormat{
		VideoFormatMP4, VideoFormatMOV, VideoFormatAVI,
		VideoFormatMKV, VideoFormatWEBM, VideoFormatFLV,
	}
	valid := false
	for _, format := range validFormats {
		if config.Format == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("不支持的视频格式: %s", config.Format)
	}

	// 验证分辨率格式
	if config.Resolution != "" {
		parts := strings.Split(config.Resolution, "x")
		if len(parts) != 2 {
			return fmt.Errorf("分辨率格式无效，应为 widthxheight 格式: %s", config.Resolution)
		}
		width, err := strconv.Atoi(parts[0])
		if err != nil || width <= 0 {
			return fmt.Errorf("分辨率宽度无效: %s", parts[0])
		}
		height, err := strconv.Atoi(parts[1])
		if err != nil || height <= 0 {
			return fmt.Errorf("分辨率高度无效: %s", parts[1])
		}
	}

	// 验证帧率
	if config.Fps > 0 && (config.Fps < 1 || config.Fps > 240) {
		return fmt.Errorf("帧率必须在1-240之间: %d", config.Fps)
	}

	// 验证质量等级
	if config.Quality >= 0 && (config.Quality < 0 || config.Quality > 51) {
		return fmt.Errorf("质量等级必须在0-51之间: %d", config.Quality)
	}

	// 验证时间格式
	timeRegex := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}(\.\d{2})?$`)

	// 检查StartTime和Duration是否同时存在或同时不存在
	if (config.StartTime == "") != (config.Duration == "") {
		return fmt.Errorf("StartTime和Duration必须同时设置或同时不设置")
	}

	// 如果设置了时间，验证格式和范围
	if config.StartTime != "" {
		if !timeRegex.MatchString(config.StartTime) {
			return fmt.Errorf("开始时间格式无效，应为 HH:MM:SS 或 HH:MM:SS.ms: %s", config.StartTime)
		}
		if !timeRegex.MatchString(config.Duration) {
			return fmt.Errorf("持续时间格式无效，应为 HH:MM:SS 或 HH:MM:SS.ms: %s", config.Duration)
		}

		// 获取视频总时长并验证时间范围
		videoInfo, err := f.getVideoInfo(context.Background(), videoPath)
		if err != nil {
			return fmt.Errorf("获取视频信息失败: %v", err)
		}

		startSeconds := f.parseTimeToSeconds(config.StartTime)
		durationSeconds := f.parseTimeToSeconds(config.Duration)
		totalSeconds := videoInfo.Duration

		// 验证开始时间不能超过视频总长度
		if startSeconds >= totalSeconds {
			return fmt.Errorf("开始时间 %s (%.2f秒) 超出视频总长度 %.2f秒", config.StartTime, startSeconds, totalSeconds)
		}

		// 验证开始时间+持续时间不能超过视频总长度
		if startSeconds+durationSeconds > totalSeconds {
			return fmt.Errorf("截取片段超出视频范围：开始时间 %s (%.2f秒) + 持续时间 %s (%.2f秒) = %.2f秒，超出视频总长度 %.2f秒",
				config.StartTime, startSeconds, config.Duration, durationSeconds, startSeconds+durationSeconds, totalSeconds)
		}
	}

	return nil
}
