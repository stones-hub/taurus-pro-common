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

// AudioFormat 音频格式枚举
// 支持多种常见的音频格式
type AudioFormat string

const (
	AudioFormatMP3  AudioFormat = "mp3"  // MP3格式，最常用的音频格式
	AudioFormatWAV  AudioFormat = "wav"  // WAV格式，无损音频格式
	AudioFormatAAC  AudioFormat = "aac"  // AAC格式，高效压缩格式
	AudioFormatFLAC AudioFormat = "flac" // FLAC格式，无损压缩格式
	AudioFormatM4A  AudioFormat = "m4a"  // M4A格式，苹果音频格式
	AudioFormatOGG  AudioFormat = "ogg"  // OGG格式，开源音频格式
)

// AudioQuality 音频质量枚举
// 定义了不同的音频质量等级，影响文件大小和音质
type AudioQuality string

const (
	AudioQualityLow    AudioQuality = "low"    // 低质量：文件小，音质一般
	AudioQualityMedium AudioQuality = "medium" // 中等质量：平衡文件大小和音质
	AudioQualityHigh   AudioQuality = "high"   // 高质量：音质好，文件较大
	AudioQualityBest   AudioQuality = "best"   // 最佳质量：最高音质，文件最大
)

// AudioExtractionOptions 音频提取选项
// 包含了音频提取过程中所有可配置的参数
type AudioExtractionOptions struct {
	Format           AudioFormat                    `json:"format"`          // 输出格式，如mp3、wav等
	Quality          AudioQuality                   `json:"quality"`         // 质量等级，影响文件大小和音质
	Bitrate          string                         `json:"bitrate"`         // 比特率，如"128k"、"192k"等
	SampleRate       int                            `json:"sample_rate"`     // 采样率，如44100、48000等
	Channels         int                            `json:"channels"`        // 声道数，1为单声道，2为立体声
	StartTime        string                         `json:"start_time"`      // 开始时间，格式如"00:00:10"
	Duration         string                         `json:"duration"`        // 持续时间，格式如"00:00:30"
	Volume           string                         `json:"volume"`          // 音量调整，如"1.5"表示1.5倍音量
	Normalize        bool                           `json:"normalize"`       // 是否标准化音量
	RemoveSilence    bool                           `json:"remove_silence"`  // 是否移除静音段
	EnableProgress   bool                           `json:"enable_progress"` // 是否启用进度监控
	ProgressCallback func(progress *FFmpegProgress) `json:"-"`               // 进度回调函数，用于实时监控处理进度
}

// AudioExtractionResult 音频提取结果
// 包含音频提取完成后的所有相关信息
type AudioExtractionResult struct {
	OutputPath     string    `json:"output_path"`     // 输出文件路径
	Duration       float64   `json:"duration"`        // 音频时长（秒）
	FileSize       int64     `json:"file_size"`       // 文件大小（字节）
	Format         string    `json:"format"`          // 音频格式
	Bitrate        string    `json:"bitrate"`         // 比特率
	SampleRate     int       `json:"sample_rate"`     // 采样率
	Channels       int       `json:"channels"`        // 声道数
	ExtractedAt    time.Time `json:"extracted_at"`    // 提取时间
	ProcessingTime float64   `json:"processing_time"` // 处理耗时（秒）
}

// FFmpegProgress 进度信息
// 用于实时监控FFmpeg处理进度
type FFmpegProgress struct {
	Progress    float64 `json:"progress"`     // 进度百分比 (0-100)
	CurrentTime string  `json:"current_time"` // 当前处理时间，格式如"00:00:05.00"
	TotalTime   string  `json:"total_time"`   // 总时长，格式如"00:01:30.00"
	Speed       string  `json:"speed"`        // 处理速度，如"1.01x"
	Message     string  `json:"message"`      // 状态消息，包含原始FFmpeg输出
}

// AudioInfo 音频信息结构
// 包含音频文件的基本信息
type AudioInfo struct {
	Duration   float64 `json:"duration"`    // 音频时长（秒）
	Bitrate    string  `json:"bitrate"`     // 比特率
	SampleRate int     `json:"sample_rate"` // 采样率
	Channels   int     `json:"channels"`    // 声道数
}

// FFmpegAudio FFmpeg音频处理工具
// 封装了FFmpeg和FFprobe的功能，提供便捷的音频处理接口
type FFmpegAudio struct {
	ffmpegPath  string // FFmpeg可执行文件路径
	ffprobePath string // FFprobe可执行文件路径
	tempDir     string // 临时目录
}

// FFmpegAudioOption 配置选项函数类型
// 用于灵活配置FFmpegAudio实例
type FFmpegAudioOption func(*FFmpegAudio)

// WithFFmpegPath 设置FFmpeg路径
// 如果FFmpeg不在系统PATH中，可以通过此选项指定路径
func WithFFmpegPath(path string) FFmpegAudioOption {
	return func(f *FFmpegAudio) {
		f.ffmpegPath = path
	}
}

// WithFFprobePath 设置FFprobe路径
// 如果FFprobe不在系统PATH中，可以通过此选项指定路径
func WithFFprobePath(path string) FFmpegAudioOption {
	return func(f *FFmpegAudio) {
		f.ffprobePath = path
	}
}

// WithTempDir 设置临时目录
// 指定临时文件存储目录
func WithTempDir(dir string) FFmpegAudioOption {
	return func(f *FFmpegAudio) {
		f.tempDir = dir
	}
}

// NewFFmpegAudio 创建FFmpeg音频处理工具
// 使用默认配置创建FFmpegAudio实例，可以通过选项函数进行自定义配置
// 参数:
//   - opts: 可选的配置函数，用于自定义FFmpeg和FFprobe路径等
//
// 返回:
//   - *FFmpegAudio: 配置好的FFmpeg音频处理工具实例
func NewFFmpegAudio(opts ...FFmpegAudioOption) *FFmpegAudio {
	// 设置默认值
	f := &FFmpegAudio{
		ffmpegPath:  "/usr/local/bin/ffmpeg",  // 默认使用系统PATH中的ffmpeg
		ffprobePath: "/usr/local/bin/ffprobe", // 默认使用系统PATH中的ffprobe
		tempDir:     "/tmp",                   // 默认临时目录
	}

	// 应用选项
	for _, opt := range opts {
		opt(f)
	}

	return f
}

// ExtractAudioFromVideo 从视频中提取音频（主要入口方法）
// 这是最常用的方法，支持从视频文件中提取音频并保存为指定格式
// 参数:
//   - ctx: 上下文，用于控制超时和取消操作
//   - videoPath: 输入视频文件路径
//   - outputDir: 输出目录路径
//   - options: 音频提取选项，包含格式、质量等配置
//
// 返回:
//   - *AudioExtractionResult: 提取结果，包含输出文件路径、时长等信息
//   - error: 错误信息
func (f *FFmpegAudio) ExtractAudioFromVideo(ctx context.Context, videoPath string, outputDir string, options *AudioExtractionOptions) (*AudioExtractionResult, error) {
	// 验证FFmpeg可用性
	if err := f.CheckFFmpegAvailable(); err != nil {
		return nil, fmt.Errorf("FFmpeg检查失败: %v", err)
	}

	// 创建默认配置
	config := f.createDefaultConfig(options)

	// 验证配置
	if err := f.validateConfig(config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %v", err)
	}

	// 执行音频提取
	if options.EnableProgress {
		return f.extractAudioWithProgress(ctx, videoPath, outputDir, config, options.ProgressCallback)
	} else {
		return f.extractAudio(ctx, videoPath, outputDir, config)
	}
}

// extractAudio 基本音频提取
// 执行基本的音频提取操作，不包含进度监控
// 参数:
//   - ctx: 上下文，用于控制超时和取消操作
//   - videoPath: 输入视频文件路径
//   - outputDir: 输出目录路径
//   - config: 内部音频提取配置
//
// 返回:
//   - *AudioExtractionResult: 提取结果
//   - error: 错误信息
func (f *FFmpegAudio) extractAudio(ctx context.Context, videoPath string, outputDir string, config *AudioExtractionConfig) (*AudioExtractionResult, error) {
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
	videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	fileExt := f.getFileExtension(config.Format)
	outputFileName := fmt.Sprintf("%s_audio.%s", videoName, fileExt)
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

	// 获取音频信息
	audioInfo, err := f.getAudioInfo(ctx, outputPath)
	if err != nil {
		// 如果获取音频信息失败，使用默认值
		audioInfo = &AudioInfo{
			Duration:   0,
			Bitrate:    config.Bitrate,
			SampleRate: config.SampleRate,
			Channels:   config.Channels,
		}
	}

	processingTime := time.Since(startTime).Seconds()

	return &AudioExtractionResult{
		OutputPath:     outputPath,
		Duration:       audioInfo.Duration,
		FileSize:       fileInfo.Size(),
		Format:         string(config.Format),
		Bitrate:        audioInfo.Bitrate,
		SampleRate:     audioInfo.SampleRate,
		Channels:       audioInfo.Channels,
		ExtractedAt:    time.Now(),
		ProcessingTime: processingTime,
	}, nil
}

// extractAudioWithProgress 带进度监控的音频提取
// 执行音频提取操作并提供实时进度监控
// 参数:
//   - ctx: 上下文，用于控制超时和取消操作
//   - videoPath: 输入视频文件路径
//   - outputDir: 输出目录路径
//   - config: 内部音频提取配置
//   - progressCallback: 进度回调函数，用于接收进度更新
//
// 返回:
//   - *AudioExtractionResult: 提取结果
//   - error: 错误信息
func (f *FFmpegAudio) extractAudioWithProgress(ctx context.Context, videoPath string, outputDir string, config *AudioExtractionConfig, progressCallback func(progress *FFmpegProgress)) (*AudioExtractionResult, error) {
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
	videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	fileExt := f.getFileExtension(config.Format)
	outputFileName := fmt.Sprintf("%s_audio.%s", videoName, fileExt)
	outputPath := filepath.Join(outputDir, outputFileName)

	// 构建FFmpeg命令
	args := f.buildFFmpegArgs(videoPath, outputPath, config)

	// 执行FFmpeg命令并监控进度
	cmd := exec.CommandContext(ctx, f.ffmpegPath, args...)

	// 创建管道来捕获stderr输出（FFmpeg的进度信息输出到stderr）
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("创建stderr管道失败: %v", err)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动FFmpeg失败: %v", err)
	}

	// 监控进度
	go f.monitorProgress(bufio.NewScanner(stderr), progressCallback)

	// 等待命令完成
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("FFmpeg执行失败: %v", err)
	}

	// 获取输出文件信息
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("获取输出文件信息失败: %v", err)
	}

	// 获取音频信息
	audioInfo, err := f.getAudioInfo(ctx, outputPath)
	if err != nil {
		// 如果获取音频信息失败，使用默认值
		audioInfo = &AudioInfo{
			Duration:   0,
			Bitrate:    config.Bitrate,
			SampleRate: config.SampleRate,
			Channels:   config.Channels,
		}
	}

	processingTime := time.Since(startTime).Seconds()

	return &AudioExtractionResult{
		OutputPath:     outputPath,
		Duration:       audioInfo.Duration,
		FileSize:       fileInfo.Size(),
		Format:         string(config.Format),
		Bitrate:        audioInfo.Bitrate,
		SampleRate:     audioInfo.SampleRate,
		Channels:       audioInfo.Channels,
		ExtractedAt:    time.Now(),
		ProcessingTime: processingTime,
	}, nil
}

// AudioExtractionConfig 内部使用的音频提取配置
// 这是内部使用的配置结构，与AudioExtractionOptions类似但更适合内部处理
type AudioExtractionConfig struct {
	Format        AudioFormat `json:"format"`         // 输出格式
	Bitrate       string      `json:"bitrate"`        // 比特率
	SampleRate    int         `json:"sample_rate"`    // 采样率
	Channels      int         `json:"channels"`       // 声道数
	Quality       int         `json:"quality"`        // 质量等级（0-9）
	StartTime     string      `json:"start_time"`     // 开始时间
	Duration      string      `json:"duration"`       // 持续时间
	Volume        string      `json:"volume"`         // 音量调整
	Normalize     bool        `json:"normalize"`      // 是否标准化
	RemoveSilence bool        `json:"remove_silence"` // 是否移除静音
}

// createDefaultConfig 创建默认配置
// 将用户选项转换为内部配置，并设置合理的默认值
// 参数:
//   - options: 用户提供的音频提取选项
//
// 返回:
//   - *AudioExtractionConfig: 内部使用的配置对象
func (f *FFmpegAudio) createDefaultConfig(options *AudioExtractionOptions) *AudioExtractionConfig {
	config := &AudioExtractionConfig{
		Format:        AudioFormatMP3, // 默认MP3格式
		Bitrate:       "128k",         // 默认比特率
		SampleRate:    44100,          // 默认采样率
		Channels:      2,              // 默认立体声
		Quality:       2,              // 默认质量等级
		Normalize:     false,          // 默认不标准化
		RemoveSilence: false,          // 默认不移除静音
	}

	// 应用用户选项
	if options != nil {
		if options.Format != "" {
			config.Format = options.Format
		}

		if options.Bitrate != "" {
			config.Bitrate = options.Bitrate
		}

		if options.SampleRate > 0 {
			config.SampleRate = options.SampleRate
		}

		if options.Channels > 0 {
			config.Channels = options.Channels
		}

		if options.StartTime != "" {
			config.StartTime = options.StartTime
		}

		if options.Duration != "" {
			config.Duration = options.Duration
		}

		if options.Volume != "" {
			config.Volume = options.Volume
		}

		config.Normalize = options.Normalize
		config.RemoveSilence = options.RemoveSilence

		// 根据质量等级设置参数
		config.Quality = f.getQualityLevel(options.Quality)
		if options.Bitrate == "" {
			config.Bitrate = f.getQualityBitrate(options.Quality)
		}
	}

	return config
}

// getQualityLevel 获取质量等级对应的数值
// 将用户友好的质量等级转换为FFmpeg内部使用的数值
// 参数:
//   - quality: 音频质量等级
//
// 返回:
//   - int: FFmpeg质量参数（0-9，0为最高质量）
func (f *FFmpegAudio) getQualityLevel(quality AudioQuality) int {
	switch quality {
	case AudioQualityLow:
		return 6
	case AudioQualityMedium:
		return 4
	case AudioQualityHigh:
		return 2
	case AudioQualityBest:
		return 0
	default:
		return 2
	}
}

// getQualityBitrate 获取质量等级对应的比特率
// 根据质量等级返回推荐的比特率设置
// 参数:
//   - quality: 音频质量等级
//
// 返回:
//   - string: 比特率字符串，如"128k"
func (f *FFmpegAudio) getQualityBitrate(quality AudioQuality) string {
	switch quality {
	case AudioQualityLow:
		return "96k"
	case AudioQualityMedium:
		return "128k"
	case AudioQualityHigh:
		return "192k"
	case AudioQualityBest:
		return "320k"
	default:
		return "128k"
	}
}

// getFileExtension 根据音频格式获取正确的文件扩展名
// 为不同的音频格式返回对应的文件扩展名
// 参数:
//   - format: 音频格式
//
// 返回:
//   - string: 文件扩展名，如"mp3"、"wav"等
func (f *FFmpegAudio) getFileExtension(format AudioFormat) string {
	switch format {
	case AudioFormatAAC:
		return "m4a" // AAC使用MP4容器，扩展名为m4a
	case AudioFormatM4A:
		return "m4a"
	default:
		return string(format)
	}
}

// buildFFmpegArgs 构建FFmpeg命令行参数
// 根据配置参数构建完整的FFmpeg命令行参数列表
// 参数:
//   - inputPath: 输入文件路径
//   - outputPath: 输出文件路径
//   - config: 音频提取配置
//
// 返回:
//   - []string: FFmpeg命令行参数列表
func (f *FFmpegAudio) buildFFmpegArgs(inputPath, outputPath string, config *AudioExtractionConfig) []string {
	args := []string{
		"-i", inputPath, // 输入文件
		"-vn", // 禁用视频流
	}

	// 根据格式选择编码器和容器格式
	switch config.Format {
	case AudioFormatMP3:
		args = append(args, "-acodec", "libmp3lame")
		args = append(args, "-f", "mp3")
	case AudioFormatAAC:
		args = append(args, "-acodec", "aac")
		args = append(args, "-f", "mp4") // AAC使用MP4容器
	case AudioFormatFLAC:
		args = append(args, "-acodec", "flac")
		args = append(args, "-f", "flac")
	case AudioFormatOGG:
		args = append(args, "-acodec", "libvorbis")
		args = append(args, "-f", "ogg")
	case AudioFormatWAV:
		args = append(args, "-acodec", "pcm_s16le")
		args = append(args, "-f", "wav")
	case AudioFormatM4A:
		args = append(args, "-acodec", "aac")
		args = append(args, "-f", "mp4")
	default:
		args = append(args, "-acodec", "libmp3lame") // 默认使用mp3编码器
		args = append(args, "-f", "mp3")
	}

	// 设置比特率
	if config.Bitrate != "" {
		args = append(args, "-b:a", config.Bitrate)
	}

	// 设置采样率
	if config.SampleRate > 0 {
		args = append(args, "-ar", fmt.Sprintf("%d", config.SampleRate))
	}

	// 设置声道数
	if config.Channels > 0 {
		args = append(args, "-ac", fmt.Sprintf("%d", config.Channels))
	}

	// 设置质量
	if config.Quality >= 0 && config.Quality <= 9 {
		args = append(args, "-q:a", fmt.Sprintf("%d", config.Quality))
	}

	// 设置开始时间
	if config.StartTime != "" {
		args = append(args, "-ss", config.StartTime)
	}

	// 设置持续时间
	if config.Duration != "" {
		args = append(args, "-t", config.Duration)
	}

	// 设置音量
	if config.Volume != "" {
		args = append(args, "-af", fmt.Sprintf("volume=%s", config.Volume))
	}

	// 音频标准化
	if config.Normalize {
		args = append(args, "-af", "loudnorm")
	}

	// 移除静音段
	if config.RemoveSilence {
		args = append(args, "-af", "silenceremove=start_periods=1:start_duration=1:start_threshold=-60dB:detection=peak")
	}

	// 覆盖输出文件
	args = append(args, "-y")

	// 输出文件路径
	args = append(args, outputPath)

	return args
}

// monitorProgress 监控FFmpeg进度
// 解析FFmpeg的stderr输出，提取进度信息并调用回调函数
// 参数:
//   - scanner: 用于读取FFmpeg输出的扫描器
//   - progressCallback: 进度回调函数
func (f *FFmpegAudio) monitorProgress(scanner *bufio.Scanner, progressCallback func(progress *FFmpegProgress)) {
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
// 从FFmpeg输出行中提取进度信息
// 参数:
//   - line: FFmpeg输出的一行文本
//
// 返回:
//   - *FFmpegProgress: 解析出的进度信息，如果解析失败返回nil
func (f *FFmpegAudio) parseProgressLine(line string) *FFmpegProgress {
	// FFmpeg进度信息格式示例：
	// frame=  123 fps= 25 q=28.0 size=    1024kB time=00:00:05.00 bitrate=1677.7kbits/s speed=1.01x

	// 匹配时间信息
	timeRegex := regexp.MustCompile(`time=(\d{2}:\d{2}:\d{2}\.\d{2})`)
	timeMatch := timeRegex.FindStringSubmatch(line)
	if len(timeMatch) < 2 {
		return nil
	}

	// 匹配总时长
	durationRegex := regexp.MustCompile(`Duration: (\d{2}:\d{2}:\d{2}\.\d{2})`)
	durationMatch := durationRegex.FindStringSubmatch(line)

	// 匹配速度
	speedRegex := regexp.MustCompile(`speed=([\d.]+)x`)
	speedMatch := speedRegex.FindStringSubmatch(line)

	// 计算进度百分比
	var progress float64
	if len(durationMatch) >= 2 {
		totalDuration := f.parseTimeToSeconds(durationMatch[1])
		currentTime := f.parseTimeToSeconds(timeMatch[1])
		if totalDuration > 0 {
			progress = (currentTime / totalDuration) * 100
		}
	}

	speed := "1.0x"
	if len(speedMatch) >= 2 {
		speed = speedMatch[1] + "x"
	}

	totalTime := ""
	if len(durationMatch) >= 2 {
		totalTime = durationMatch[1]
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
// 将"HH:MM:SS.ss"格式的时间字符串转换为秒数
// 参数:
//   - timeStr: 时间字符串，格式如"00:01:30.50"
//
// 返回:
//   - float64: 对应的秒数
func (f *FFmpegAudio) parseTimeToSeconds(timeStr string) float64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}

	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])
	seconds, _ := strconv.ParseFloat(parts[2], 64)

	return float64(hours*3600) + float64(minutes*60) + seconds
}

// getAudioInfo 获取音频文件信息
// 使用ffprobe获取音频文件的详细信息
// 参数:
//   - ctx: 上下文，用于控制超时和取消操作
//   - audioPath: 音频文件路径
//
// 返回:
//   - *AudioInfo: 音频信息，包含时长、比特率等
//   - error: 错误信息
func (f *FFmpegAudio) getAudioInfo(ctx context.Context, audioPath string) (*AudioInfo, error) {
	// 使用ffprobe获取音频信息
	cmd := exec.CommandContext(ctx, f.ffprobePath,
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		audioPath)

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
			SampleRate string `json:"sample_rate"`
			Channels   int    `json:"channels"`
			BitRate    string `json:"bit_rate"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(output, &probeResult); err != nil {
		return nil, fmt.Errorf("解析ffprobe输出失败: %v", err)
	}

	// 查找音频流
	var audioStream *struct {
		CodecType  string `json:"codec_type"`
		SampleRate string `json:"sample_rate"`
		Channels   int    `json:"channels"`
		BitRate    string `json:"bit_rate"`
	}

	for _, stream := range probeResult.Streams {
		if stream.CodecType == "audio" {
			audioStream = &stream
			break
		}
	}

	if audioStream == nil {
		return nil, fmt.Errorf("未找到音频流")
	}

	// 解析时长
	duration, _ := strconv.ParseFloat(probeResult.Format.Duration, 64)

	// 解析采样率
	sampleRate, _ := strconv.Atoi(audioStream.SampleRate)

	// 获取比特率
	bitrate := audioStream.BitRate
	if bitrate == "" {
		bitrate = probeResult.Format.BitRate
	}
	if bitrate != "" {
		bitrate = bitrate + "bps"
	}

	return &AudioInfo{
		Duration:   duration,
		Bitrate:    bitrate,
		SampleRate: sampleRate,
		Channels:   audioStream.Channels,
	}, nil
}

// BatchExtractAudio 批量提取音频
// 对多个视频文件进行批量音频提取
// 参数:
//   - ctx: 上下文，用于控制超时和取消操作
//   - videoPaths: 视频文件路径列表
//   - outputDir: 输出目录路径
//   - options: 音频提取选项
//
// 返回:
//   - []*AudioExtractionResult: 提取结果列表
//   - error: 错误信息
func (f *FFmpegAudio) BatchExtractAudio(ctx context.Context, videoPaths []string, outputDir string, options *AudioExtractionOptions) ([]*AudioExtractionResult, error) {
	results := make([]*AudioExtractionResult, 0, len(videoPaths))

	for i, videoPath := range videoPaths {
		// 为每个视频创建独立的输出目录
		videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
		videoOutputDir := filepath.Join(outputDir, videoName)

		// 提取音频
		result, err := f.ExtractAudioFromVideo(ctx, videoPath, videoOutputDir, options)
		if err != nil {
			// 记录错误但继续处理其他文件
			fmt.Printf("处理视频 %s 失败: %v\n", videoPath, err)
			continue
		}

		results = append(results, result)

		// 如果启用了进度监控，可以在这里报告批量处理进度
		if options != nil && options.ProgressCallback != nil {
			progress := &FFmpegProgress{
				Progress: float64(i+1) / float64(len(videoPaths)) * 100,
				Message:  fmt.Sprintf("批量处理进度: %d/%d", i+1, len(videoPaths)),
			}
			options.ProgressCallback(progress)
		}
	}

	return results, nil
}

// ConvertAudioFormat 转换音频格式
// 将音频文件从一种格式转换为另一种格式
// 参数:
//   - ctx: 上下文，用于控制超时和取消操作
//   - inputPath: 输入音频文件路径
//   - outputDir: 输出目录路径
//   - targetFormat: 目标音频格式
//   - options: 音频转换选项
//
// 返回:
//   - *AudioExtractionResult: 转换结果
//   - error: 错误信息
func (f *FFmpegAudio) ConvertAudioFormat(ctx context.Context, inputPath string, outputDir string, targetFormat AudioFormat, options *AudioExtractionOptions) (*AudioExtractionResult, error) {
	// 创建转换选项
	convertOptions := &AudioExtractionOptions{
		Format:           targetFormat,
		Quality:          options.Quality,
		Bitrate:          options.Bitrate,
		SampleRate:       options.SampleRate,
		Channels:         options.Channels,
		Normalize:        options.Normalize,
		RemoveSilence:    options.RemoveSilence,
		EnableProgress:   options.EnableProgress,
		ProgressCallback: options.ProgressCallback,
	}

	// 使用音频提取功能进行格式转换
	return f.ExtractAudioFromVideo(ctx, inputPath, outputDir, convertOptions)
}

// CreatePresetConfig 创建预设配置
// 根据预设名称创建预定义的音频提取配置
// 参数:
//   - preset: 预设名称，支持"speech"、"music"、"podcast"、"archive"
//
// 返回:
//   - *AudioExtractionOptions: 预设的音频提取选项
func (f *FFmpegAudio) CreatePresetConfig(preset string) *AudioExtractionOptions {
	switch preset {
	case "speech":
		// 语音优化配置
		return &AudioExtractionOptions{
			Format:        AudioFormatMP3,
			Quality:       AudioQualityMedium,
			Bitrate:       "64k",
			SampleRate:    16000, // 语音通常使用16kHz
			Channels:      1,     // 单声道
			Normalize:     true,  // 标准化音量
			RemoveSilence: true,  // 移除静音段
		}
	case "music":
		// 音乐优化配置
		return &AudioExtractionOptions{
			Format:        AudioFormatMP3,
			Quality:       AudioQualityHigh,
			Bitrate:       "192k",
			SampleRate:    44100, // CD质量
			Channels:      2,     // 立体声
			Normalize:     false, // 保持原始动态范围
			RemoveSilence: false, // 不移除静音
		}
	case "podcast":
		// 播客优化配置
		return &AudioExtractionOptions{
			Format:        AudioFormatMP3,
			Quality:       AudioQualityMedium,
			Bitrate:       "128k",
			SampleRate:    44100,
			Channels:      2,
			Normalize:     true, // 标准化音量
			RemoveSilence: true, // 移除静音段
		}
	case "archive":
		// 归档配置（高质量）
		return &AudioExtractionOptions{
			Format:        AudioFormatFLAC,
			Quality:       AudioQualityBest,
			Bitrate:       "320k",
			SampleRate:    48000, // 专业音频标准
			Channels:      2,
			Normalize:     false,
			RemoveSilence: false,
		}
	default:
		// 默认配置
		return &AudioExtractionOptions{
			Format:        AudioFormatMP3,
			Quality:       AudioQualityMedium,
			Bitrate:       "128k",
			SampleRate:    44100,
			Channels:      2,
			Normalize:     false,
			RemoveSilence: false,
		}
	}
}

// validateConfig 验证配置
// 验证音频提取配置的有效性
// 参数:
//   - config: 要验证的配置
//
// 返回:
//   - error: 验证失败时的错误信息
func (f *FFmpegAudio) validateConfig(config *AudioExtractionConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证格式
	validFormats := []AudioFormat{
		AudioFormatMP3, AudioFormatWAV, AudioFormatAAC,
		AudioFormatFLAC, AudioFormatM4A, AudioFormatOGG,
	}
	valid := false
	for _, format := range validFormats {
		if config.Format == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("不支持的音频格式: %s", config.Format)
	}

	// 验证采样率
	if config.SampleRate > 0 && (config.SampleRate < 8000 || config.SampleRate > 192000) {
		return fmt.Errorf("采样率必须在8000-192000之间: %d", config.SampleRate)
	}

	// 验证声道数
	if config.Channels > 0 && (config.Channels < 1 || config.Channels > 8) {
		return fmt.Errorf("声道数必须在1-8之间: %d", config.Channels)
	}

	// 验证质量等级
	if config.Quality > 0 && (config.Quality < 0 || config.Quality > 9) {
		return fmt.Errorf("质量等级必须在0-9之间: %d", config.Quality)
	}

	return nil
}

// CheckFFmpegAvailable 检查FFmpeg是否可用
// 验证FFmpeg和FFprobe是否已安装并可用
// 返回:
//   - error: 如果FFmpeg或FFprobe不可用则返回错误
func (f *FFmpegAudio) CheckFFmpegAvailable() error {
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
