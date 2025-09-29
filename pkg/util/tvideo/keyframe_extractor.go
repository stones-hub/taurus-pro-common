package tvideo

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// KeyframeExtractionMode 关键帧提取模式
type KeyframeExtractionMode string

const (
	ModeSmart    KeyframeExtractionMode = "smart"    // 智能模式
	ModeUniform  KeyframeExtractionMode = "uniform"  // 均匀分布模式
	ModeInterval KeyframeExtractionMode = "interval" // 时间间隔模式

)

// KeyframeExtractionOptions 关键帧提取选项
type KeyframeExtractionOptions struct {
	MaxFrames        int                              `json:"max_frames"`    // 最大关键帧数量
	Mode             KeyframeExtractionMode           `json:"mode"`          // 提取模式
	TimeInterval     *float64                         `json:"time_interval"` // 时间间隔（秒）
	OutputDir        string                           `json:"output_dir"`    // 输出目录
	EnableDebug      bool                             `json:"enable_debug"`  // 是否启用调试模式
	ProgressCallback func(progress *KeyframeProgress) `json:"-"`             // 进度回调
}

// KeyframeProgress 关键帧提取进度信息
type KeyframeProgress struct {
	Coverage          float64 `json:"coverage"`            // 覆盖进度 (0-1)
	ElapsedSeconds    float64 `json:"elapsed_seconds"`     // 已处理时间
	DurationSeconds   float64 `json:"duration_seconds"`    // 总时长
	SavedFrames       int     `json:"saved_frames"`        // 已保存帧数
	MaxFrames         int     `json:"max_frames"`          // 最大帧数
	NewFramePath      string  `json:"new_frame_path"`      // 新帧路径
	NewFrameTimestamp float64 `json:"new_frame_timestamp"` // 新帧时间戳
	ChangeScore       float64 `json:"change_score"`        // 变化得分
	QualityScore      float64 `json:"quality_score"`       // 质量得分
	Width             int     `json:"width"`               // 帧宽度
	Height            int     `json:"height"`              // 帧高度
	FileSize          int64   `json:"file_size"`           // 文件大小
	LogMessage        string  `json:"log_message"`         // 日志消息
}

// VideoInfo 视频信息
type VideoInfo struct {
	TotalFrames int     `json:"total_frames"` // 总帧数
	FPS         float64 `json:"fps"`          // 帧率
	Width       int     `json:"width"`        // 宽度
	Height      int     `json:"height"`       // 高度
	Duration    float64 `json:"duration"`     // 时长（秒）
}

// KeyframeExtractionResult 关键帧提取结果
type KeyframeExtractionResult struct {
	KeyframePaths  []string   `json:"keyframe_paths"`  // 关键帧文件路径列表
	VideoInfo      *VideoInfo `json:"video_info"`      // 视频信息
	ProcessingTime float64    `json:"processing_time"` // 处理耗时（秒）
	ExtractionMode string     `json:"extraction_mode"` // 提取模式
	TotalFrames    int        `json:"total_frames"`    // 总帧数
	Success        bool       `json:"success"`         // 是否成功
	ErrorMessage   string     `json:"error_message"`   // 错误信息
}

// DebugInfo 调试信息
type DebugInfo struct {
	VideoInfo        *VideoInfo             `json:"video_info"`        // 视频信息
	ExtractionParams map[string]interface{} `json:"extraction_params"` // 提取参数
	Keyframes        []string               `json:"keyframes"`         // 关键帧路径
	Performance      map[string]interface{} `json:"performance"`       // 性能信息
	SceneAnalysis    map[string]interface{} `json:"scene_analysis"`    // 场景分析信息
	QualityMetrics   map[string]interface{} `json:"quality_metrics"`   // 质量指标信息
	ExtractionStats  map[string]interface{} `json:"extraction_stats"`  // 提取统计信息
	Error            string                 `json:"error,omitempty"`   // 错误信息
}

// KeyframeExtractor 关键帧提取器
type KeyframeExtractor struct {
	PythonPath  string // Python可执行文件路径, 即python地址
	scriptPath  string // 临时Python脚本路径, 即脚本地址
	tempDirPath string // 临时文件目录路径, 即临时文件存储目录
}

// KeyframeExtractorOption 配置选项函数类型
type KeyframeExtractorOption func(*KeyframeExtractor)

// WithPythonPath 设置Python路径
func WithPythonPath(path string) KeyframeExtractorOption {
	return func(k *KeyframeExtractor) {
		k.PythonPath = path
	}
}

// checkPythonAvailable 内部方法，检查Python环境
func (k *KeyframeExtractor) checkPythonAvailable() error {
	// 检查Python
	if err := exec.Command(k.PythonPath, "--version").Run(); err != nil {
		return fmt.Errorf("python不可用: %v", err)
	}

	// 检查必要的Python包
	cmd := exec.Command(k.PythonPath, "-c", "import cv2, numpy")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("缺少必要的Python包(opencv-python, numpy): %v", err)
	}

	return nil
}

// NewKeyframeExtractor 创建关键帧提取器
func NewKeyframeExtractor(opts ...KeyframeExtractorOption) (*KeyframeExtractor, error) {
	// 设置默认值
	k := &KeyframeExtractor{
		PythonPath: "python3", // 默认使用python3
	}

	// 应用选项
	for _, opt := range opts {
		opt(k)
	}

	// 检查Python环境
	if err := k.checkPythonAvailable(); err != nil {
		return nil, fmt.Errorf("python环境检查失败: %v", err)
	}

	// 创建临时脚本
	scriptPath, err := k.createTempScript()
	if err != nil {
		return nil, fmt.Errorf("创建临时脚本失败: %v", err)
	}
	k.scriptPath = scriptPath
	fmt.Printf("临时Python脚本路径: %s\n", scriptPath)

	return k, nil
}

// Close 实现io.Closer接口，用于资源清理,
// 注意: 需要用户主动清理
func (k *KeyframeExtractor) Close() error {
	if k.tempDirPath != "" {
		// 清理整个临时目录
		if err := os.RemoveAll(k.tempDirPath); err != nil {
			return fmt.Errorf("清理临时目录失败: %v", err)
		}
		k.tempDirPath = ""
		k.scriptPath = ""
	}
	return nil
}

// ExtractKeyframes 提取关键帧（统一的主方法）
func (k *KeyframeExtractor) ExtractKeyframes(ctx context.Context, videoPath string, options *KeyframeExtractionOptions) (*KeyframeExtractionResult, *DebugInfo, error) {
	startTime := time.Now()

	// 验证输入文件
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("视频文件不存在: %s", videoPath)
	}

	// 设置默认选项
	if options == nil {
		options = k.CreateDefaultOptions()
		options.OutputDir = filepath.Join(filepath.Dir(videoPath), "keyframes")
	}

	// 创建输出目录
	if err := os.MkdirAll(options.OutputDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 构建Python命令
	args := k.buildPythonArgs(k.scriptPath, videoPath, options)

	// 执行Python脚本
	cmd := exec.CommandContext(ctx, k.PythonPath, args...)

	var debugInfo *DebugInfo
	if options.EnableDebug {
		// 如果启用了调试模式，准备接收调试信息
		debugInfo = &DebugInfo{}
	}

	// 处理输出和进度
	if options.ProgressCallback != nil {
		// 创建管道来捕获输出
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, debugInfo, fmt.Errorf("创建输出管道失败: %v", err)
		}

		// 启动命令
		if err := cmd.Start(); err != nil {
			return nil, debugInfo, fmt.Errorf("启动Python脚本失败: %v", err)
		}

		// 在goroutine中处理输出
		go k.handlePythonOutput(stdout, options.ProgressCallback)

		// 等待命令完成
		if err := cmd.Wait(); err != nil {
			return nil, debugInfo, fmt.Errorf("python脚本执行失败: %v", err)
		}
	} else {
		// 没有进度回调时，直接执行
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return nil, debugInfo, fmt.Errorf("python脚本执行失败: %v", err)
		}
	}

	// 获取关键帧文件列表
	keyframePaths, err := k.getKeyframePaths(options.OutputDir)
	if err != nil {
		return nil, debugInfo, fmt.Errorf("获取关键帧文件失败: %v", err)
	}

	// 获取视频信息
	videoInfo, _ := k.GetVideoInfo(ctx, videoPath)

	// 如果启用了调试模式，加载调试信息
	if options.EnableDebug {
		debugInfo, _ = k.loadDebugInfo(options.OutputDir)
	}

	result := &KeyframeExtractionResult{
		KeyframePaths:  keyframePaths,
		VideoInfo:      videoInfo,
		ProcessingTime: time.Since(startTime).Seconds(),
		ExtractionMode: string(options.Mode),
		TotalFrames:    len(keyframePaths),
		Success:        true,
	}

	return result, debugInfo, nil
}

// GetVideoInfo 获取视频信息
func (k *KeyframeExtractor) GetVideoInfo(ctx context.Context, videoPath string) (*VideoInfo, error) {
	// 构建Python命令获取视频信息
	args := []string{
		k.scriptPath,
		"--get-info",
		videoPath,
	}

	cmd := exec.CommandContext(ctx, k.PythonPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取视频信息失败: %v", err)
	}

	// 解析JSON输出
	var videoInfo VideoInfo
	if err := json.Unmarshal(output, &videoInfo); err != nil {
		return nil, fmt.Errorf("解析视频信息失败: %v", err)
	}

	return &videoInfo, nil
}

// createTempScript 创建临时Python脚本文件
func (k *KeyframeExtractor) createTempScript() (string, error) {
	// 创建唯一的临时目录（使用进程ID和时间戳确保唯一性）
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("keyframe_extractor_%d_%d", os.Getpid(), time.Now().UnixNano()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %v", err)
	}
	k.tempDirPath = tempDir

	// 创建临时文件
	scriptPath := filepath.Join(tempDir, "extractor.py")
	file, err := os.OpenFile(scriptPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		os.RemoveAll(tempDir) // 清理临时目录
		return "", fmt.Errorf("创建临时脚本文件失败: %v", err)
	}
	defer file.Close()

	// 写入脚本内容
	if _, err := file.WriteString(pythonScript); err != nil {
		os.RemoveAll(tempDir) // 清理临时目录
		return "", fmt.Errorf("写入脚本内容失败: %v", err)
	}

	return scriptPath, nil
}

// buildPythonArgs 构建Python命令行参数
func (k *KeyframeExtractor) buildPythonArgs(scriptPath string, videoPath string, options *KeyframeExtractionOptions) []string {
	args := []string{
		scriptPath,
		videoPath,
		"--out", options.OutputDir,
		"--max-frames", strconv.Itoa(options.MaxFrames),
		"--mode", string(options.Mode),
	}

	if options.TimeInterval != nil {
		args = append(args, "--time-interval", strconv.FormatFloat(*options.TimeInterval, 'f', -1, 64))
	}

	// 如果启用调试模式，添加--debug参数
	if options.EnableDebug {
		args = append(args, "--debug")
	}

	return args
}

// getKeyframePaths 获取关键帧文件路径列表
func (k *KeyframeExtractor) getKeyframePaths(outputDir string) ([]string, error) {
	// 读取输出目录中的关键帧文件
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("读取输出目录失败: %v", err)
	}

	var keyframePaths []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "keyframe_") && strings.HasSuffix(entry.Name(), ".jpg") {
			keyframePaths = append(keyframePaths, filepath.Join(outputDir, entry.Name()))
		}
	}

	return keyframePaths, nil
}

// loadDebugInfo 加载调试信息
func (k *KeyframeExtractor) loadDebugInfo(outputDir string) (*DebugInfo, error) {
	// 查找调试文件
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("读取输出目录失败: %v", err)
	}

	var debugFile string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "debug_keyframes_") && strings.HasSuffix(entry.Name(), ".json") {
			debugFile = filepath.Join(outputDir, entry.Name())
			break
		}
	}

	if debugFile == "" {
		return nil, fmt.Errorf("未找到调试文件")
	}

	// 读取调试文件
	data, err := os.ReadFile(debugFile)
	if err != nil {
		return nil, fmt.Errorf("读取调试文件失败: %v", err)
	}

	// 解析JSON
	var debugInfo DebugInfo
	if err := json.Unmarshal(data, &debugInfo); err != nil {
		return nil, fmt.Errorf("解析调试文件失败: %v", err)
	}

	return &debugInfo, nil
}

// handlePythonOutput 处理Python脚本的输出，解析进度信息
func (k *KeyframeExtractor) handlePythonOutput(stdout io.ReadCloser, progressCallback func(progress *KeyframeProgress)) {
	defer stdout.Close()

	// 使用更大的buffer来读取输出
	reader := bufio.NewReaderSize(stdout, 64*1024)

	for {
		// 读取一行，包括分隔符
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				progressCallback(&KeyframeProgress{
					LogMessage: fmt.Sprintf("读取Python输出时出错: %v", err),
				})
			}
			break
		}

		// 处理这一行
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 尝试解析JSON格式的进度信息
		var progress KeyframeProgress
		if err := json.Unmarshal([]byte(line), &progress); err == nil {
			// 成功解析为进度信息，调用回调
			progressCallback(&progress)
		} else {
			// 不是JSON格式，可能是普通日志，也通过回调传递
			progressCallback(&KeyframeProgress{
				LogMessage: line,
			})
		}
	}
}

// GetSupportedModes 获取支持的提取模式
func (k *KeyframeExtractor) GetSupportedModes() []KeyframeExtractionMode {
	return []KeyframeExtractionMode{
		ModeSmart,
		ModeUniform,
		ModeInterval,
	}
}

// CreateDefaultOptions 创建默认选项
func (k *KeyframeExtractor) CreateDefaultOptions() *KeyframeExtractionOptions {
	return &KeyframeExtractionOptions{
		MaxFrames:   300,
		Mode:        ModeSmart,
		EnableDebug: false,
	}
}

// CreatePresetOptions 创建预设选项
func (k *KeyframeExtractor) CreatePresetOptions(preset string) *KeyframeExtractionOptions {
	switch preset {
	case "fast":
		// 快速提取：较少帧数，均匀分布
		return &KeyframeExtractionOptions{
			MaxFrames:   100,
			Mode:        ModeUniform,
			EnableDebug: false,
		}
	case "quality":
		// 质量优先：智能提取，较多帧数
		return &KeyframeExtractionOptions{
			MaxFrames:   500,
			Mode:        ModeSmart,
			EnableDebug: true,
		}
	case "interval":
		// 间隔提取：固定时间间隔
		interval := 5.0 // 5秒间隔
		return &KeyframeExtractionOptions{
			MaxFrames:    300,
			Mode:         ModeInterval,
			TimeInterval: &interval,
			EnableDebug:  false,
		}
	default:
		// 默认配置
		return k.CreateDefaultOptions()
	}
}

const (
	// 嵌入的Python脚本内容
	pythonScript = `"""
关键帧提取核心实现
python keyframe_extractor_core.py --out ./keyframes --max-frames 300 --mode smart --time-interval 1.0 demo.mp4
"""

import os
import cv2
import time
import json
import logging
import numpy as np

logger = logging.getLogger(__name__)


# 智能提取配置常量（与历史实现保持一致）
SMART_EXTRACTION_CONFIG = {
    'SCENE_CHANGE_THRESHOLD': 35.0,
    'MIN_INTERVAL': 0.5,
    'MAX_INTERVAL': 3.0,
    'HIST_WEIGHT': 0.3,
    'SSIM_WEIGHT': 0.3,
    'EDGE_WEIGHT': 0.3,
    'MOTION_WEIGHT': 0.1,
    'QUALITY_WEIGHT': 0.4,
    'CHANGE_WEIGHT': 0.6,
    'QUALITY_THRESHOLD': 20.0,
}


def _monitor_performance(func):
    def wrapper(*args, **kwargs):
        start_time = time.time()
        video_path = args[0] if args else kwargs.get('video_path')
        try:
            result = func(*args, **kwargs)
            duration = time.time() - start_time
            keyframe_count = len(result) if result else 0
            logger.info(f"关键帧提取成功: {video_path}, 耗时: {duration:.2f}s, 提取数量: {keyframe_count}")
            return result
        except Exception as e:
            duration = time.time() - start_time
            logger.error(f"❌ 关键帧提取失败: {video_path}, 耗时: {duration:.2f}s, 错误: {e}")
            raise
    return wrapper


def extract_keyframes_with_fallback(video_path, output_dir, max_frames=200, method='smart', time_interval=None, progress_callback=None):
    try:
        return extract_keyframes(video_path, output_dir, max_frames, method, time_interval=time_interval, progress_callback=progress_callback)
    except Exception as e:
        logger.warning(f"智能提取失败: {e}, 降级到间隔提取")
        try:
            return extract_keyframes(video_path, output_dir, max_frames, 'interval', time_interval=time_interval, progress_callback=progress_callback)
        except Exception as e2:
            logger.error(f"间隔提取也失败: {e2}, 使用最小提取")
            return _extract_minimal_keyframes(video_path, output_dir, max_frames)


def _extract_minimal_keyframes(video_path, output_dir, max_frames):
    try:
        os.makedirs(output_dir, exist_ok=True)
        cap = cv2.VideoCapture(video_path)
        if not cap.isOpened():
            return []
        total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
        fps = cap.get(cv2.CAP_PROP_FPS)
        keyframe_paths = []
        frame_indices = []
        if max_frames >= 1:
            frame_indices.append(0)
        if max_frames >= 2:
            frame_indices.append(total_frames - 1)
        if max_frames >= 3:
            frame_indices.append(total_frames // 2)
        if max_frames > 3:
            interval = total_frames // max_frames
            for i in range(1, max_frames - 2):
                idx = i * interval
                if idx not in frame_indices and idx < total_frames:
                    frame_indices.append(idx)
        frame_indices = sorted(set(frame_indices))[:max_frames]
        for i, frame_idx in enumerate(frame_indices):
            cap.set(cv2.CAP_PROP_POS_FRAMES, frame_idx)
            ret, frame = cap.read()
            if ret:
                timestamp = frame_idx / fps if fps > 0 else frame_idx
                filename = f"keyframe_{i:03d}_{timestamp:.2f}s.jpg"
                filepath = os.path.join(output_dir, filename)
                resized_frame = _resize_to_480p(frame)
                cv2.imwrite(filepath, resized_frame, [int(cv2.IMWRITE_JPEG_QUALITY), 85])
                keyframe_paths.append(filepath)
        cap.release()
        logger.info(f"最小提取完成: {len(keyframe_paths)}帧")
        return keyframe_paths
    except Exception as e:
        logger.error(f"最小提取失败: {e}")
        return []


@_monitor_performance
def extract_keyframes(video_path, output_dir, max_frames=200, method='smart', time_interval=None, progress_callback=None):
    try:
        os.makedirs(output_dir, exist_ok=True)
        cap = cv2.VideoCapture(video_path)
        if not cap.isOpened():
            raise Exception(f"无法打开视频文件: {video_path}")
        total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
        fps = cap.get(cv2.CAP_PROP_FPS)
        duration = total_frames / fps
        
        # 直接打印进度信息到标准输出
        def print_progress(progress_data):
            import json
            import sys
            print(json.dumps(progress_data), flush=True)
            sys.stdout.flush()
            
        def send_log(message):
            print_progress({'log_message': message})
            logger.info(message)
            
        send_log(
            f"视频文件分析: 总帧数={total_frames}, 帧率={fps:.2f}FPS, 时长={duration:.2f}秒({duration//60:.0f}分{duration%60:.0f}秒), 理论密度={total_frames/duration:.1f}帧/秒"
            if duration > 0 else f"视频文件分析: 总帧数={total_frames}, 帧率={fps:.2f}FPS"
        )
        if method == 'smart':
            send_log("选择智能提取模式")
            keyframe_paths = _extract_smart_keyframes(cap, output_dir, max_frames, time_interval=time_interval, progress_callback=progress_callback)
        elif method == 'uniform':
            send_log("选择均匀分布提取模式")
            send_log("提取策略: 等间隔时间分布")
            keyframe_paths = _extract_uniform_keyframes(cap, output_dir, max_frames, total_frames)
        else:
            send_log("选择时间间隔提取模式")
            seconds_interval = None
            if duration > 300:
                computed_interval = duration / 300.0
                if time_interval is not None:
                    seconds_interval = max(float(time_interval), computed_interval)
                else:
                    seconds_interval = computed_interval
            else:
                if time_interval is not None:
                    seconds_interval = float(time_interval)
                else:
                    seconds_interval = duration / max_frames if max_frames > 0 else 1.0
            if fps and fps > 0:
                interval_frames = max(1, int(round(seconds_interval * fps)))
            else:
                interval_frames = max(1, total_frames // max_frames) if max_frames > 0 else 1
            keyframe_paths = _extract_interval_keyframes(cap, output_dir, interval_frames, max_frames)
        cap.release()
        send_log("关键帧提取任务完成!")
        send_log(f"   - 成功提取: {len(keyframe_paths)}帧")
        send_log(f"   - 文件列表: {[os.path.basename(p) for p in keyframe_paths[:5]]}" + ("..." if len(keyframe_paths) > 5 else ""))
        return keyframe_paths
    except Exception as e:
        logger.error(f"关键帧提取失败: {e}")
        return []


def _calculate_scene_change_score(frame1, frame2):
    try:
        gray1 = cv2.cvtColor(frame1, cv2.COLOR_BGR2GRAY)
        gray2 = cv2.cvtColor(frame2, cv2.COLOR_BGR2GRAY)
        hist1 = cv2.calcHist([frame1], [0, 1, 2], None, [32, 32, 32], [0, 256, 0, 256, 0, 256])
        hist2 = cv2.calcHist([frame2], [0, 1, 2], None, [32, 32, 32], [0, 256, 0, 256, 0, 256])
        hist_corr = cv2.compareHist(hist1, hist2, cv2.HISTCMP_CORREL)
        hist_score = (1 - max(0, hist_corr)) * 100
        mu1 = cv2.GaussianBlur(gray1.astype(np.float32), (5, 5), 1.0)
        mu2 = cv2.GaussianBlur(gray2.astype(np.float32), (5, 5), 1.0)
        mu1_sq = mu1 * mu1
        mu2_sq = mu2 * mu2
        mu1_mu2 = mu1 * mu2
        sigma1_sq = cv2.GaussianBlur(gray1.astype(np.float32) * gray1, (5, 5), 1.0) - mu1_sq
        sigma2_sq = cv2.GaussianBlur(gray2.astype(np.float32) * gray2, (5, 5), 1.0) - mu2_sq
        sigma12 = cv2.GaussianBlur(gray1.astype(np.float32) * gray2, (5, 5), 1.0) - mu1_mu2
        c1, c2 = (0.01 * 255) ** 2, (0.03 * 255) ** 2
        ssim_map = ((2 * mu1_mu2 + c1) * (2 * sigma12 + c2)) / ((mu1_sq + mu2_sq + c1) * (sigma1_sq + sigma2_sq + c2))
        ssim_score = (1 - np.mean(ssim_map)) * 100
        edges1 = cv2.Canny(gray1, 50, 150)
        edges2 = cv2.Canny(gray2, 50, 150)
        edge_diff = cv2.absdiff(edges1, edges2)
        edge_change_score = np.mean(edge_diff) * 2
        frame_diff = cv2.absdiff(gray1, gray2)
        motion_score = np.mean(frame_diff)
        config = SMART_EXTRACTION_CONFIG
        change_score = (
            hist_score * config['HIST_WEIGHT'] +
            ssim_score * config['SSIM_WEIGHT'] +
            edge_change_score * config['EDGE_WEIGHT'] +
            motion_score * config['MOTION_WEIGHT']
        )
        return min(100.0, max(0.0, change_score))
    except Exception:
        try:
            gray1 = cv2.cvtColor(frame1, cv2.COLOR_BGR2GRAY)
            gray2 = cv2.cvtColor(frame2, cv2.COLOR_BGR2GRAY)
            diff = cv2.absdiff(gray1, gray2)
            return np.mean(diff)
        except Exception:
            return 0.0


def _calculate_frame_quality(frame):
    try:
        gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
        laplacian_var = cv2.Laplacian(gray, cv2.CV_64F).var()
        edges = cv2.Canny(gray, 50, 150)
        edge_density = np.sum(edges > 0) / edges.size
        quality_score = laplacian_var * 0.7 + edge_density * 1000 * 0.3
        return quality_score
    except Exception:
        return 50.0


def _calculate_comprehensive_score(frame, prev_frame=None):
    config = SMART_EXTRACTION_CONFIG
    quality_score = _calculate_frame_quality(frame)
    normalized_quality = min(100.0, quality_score / 5.0)
    change_score = 0
    if prev_frame is not None:
        change_score = _calculate_scene_change_score(prev_frame, frame)
    total_score = (
        normalized_quality * config['QUALITY_WEIGHT'] +
        change_score * config['CHANGE_WEIGHT']
    )
    return total_score, normalized_quality, change_score


def _is_dark_frame(frame, threshold=35):
    try:
        gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
        mean_brightness = np.mean(gray)
        dark_ratio = np.sum(gray < threshold) / gray.size
        brightness_std = np.std(gray)
        hist = cv2.calcHist([gray], [0], None, [256], [0, 256])
        hist = hist / hist.sum()
        entropy = -np.sum(hist * np.log2(hist + 1e-7))
        is_dark = (
            mean_brightness < threshold or
            dark_ratio > 0.95 or
            (brightness_std < 10 and entropy < 3.0)
        )
        return is_dark
    except Exception:
        return False


def _check_frame_similarity(frame1_gray, frame2_gray, threshold=0.75):
    try:
        hist1 = cv2.calcHist([frame1_gray], [0], None, [32], [0, 256])
        hist2 = cv2.calcHist([frame2_gray], [0], None, [32], [0, 256])
        hist_similarity = cv2.compareHist(hist1, hist2, cv2.HISTCMP_CORREL)
        small1 = cv2.resize(frame1_gray, (32, 32))
        small2 = cv2.resize(frame2_gray, (32, 32))
        template_similarity = cv2.matchTemplate(small1, small2, cv2.TM_CCOEFF_NORMED)[0][0]
        similarity = 0.5 * hist_similarity + 0.5 * template_similarity
        return similarity > threshold
    except Exception:
        return False


def _resize_to_480p(frame):
    try:
        h, w = frame.shape[:2]
        if h == 0 or w == 0:
            return frame
        target_h = 720
        if h == target_h:
            return frame
        scale = target_h / float(h)
        new_w = max(1, int(round(w * scale)))
        resized = cv2.resize(frame, (new_w, target_h), interpolation=cv2.INTER_AREA)
        return resized
    except Exception:
        return frame


def _extract_content_driven_keyframes(cap, output_dir, max_frames, progress_callback=None):
    import json
    import sys
    import os
    
    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    fps = cap.get(cv2.CAP_PROP_FPS)
    duration = total_frames / fps if fps > 0 else 0
    keyframe_paths = []  # 初始化关键帧路径列表
    
    def send_progress(current_time, message=None, frame_info=None):
        if not progress_callback:
            return
            
        # 只在保存关键帧时发送进度信息
        if frame_info:
            progress = {
                'coverage': min(1.0, current_time / duration) if duration > 0 else 0.0,
                'elapsed_seconds': float(current_time),
                'duration_seconds': float(duration),
                'saved_frames': len(keyframe_paths),
                'max_frames': int(max_frames),
                'new_frame_path': frame_info.get('new_frame_path', ''),
                'new_frame_timestamp': float(frame_info.get('new_frame_timestamp', 0)),
                'change_score': float(frame_info.get('change_score', 0)),
                'quality_score': float(frame_info.get('quality_score', 0)),
                'width': int(frame_info.get('width', 0)),
                'height': int(frame_info.get('height', 0)),
                'file_size': int(frame_info.get('file_size', 0)),
                'log_message': frame_info.get('log_message', '')
            }
            print(json.dumps(progress), flush=True)
            sys.stdout.flush()
        elif message:  # 只发送日志消息
            progress = {
                'coverage': 0.0,
                'elapsed_seconds': 0.0,
                'duration_seconds': float(duration),
                'saved_frames': len(keyframe_paths),
                'max_frames': int(max_frames),
                'new_frame_path': '',
                'new_frame_timestamp': 0.0,
                'change_score': 0.0,
                'quality_score': 0.0,
                'width': 0,
                'height': 0,
                'file_size': 0,
                'log_message': message
            }
            print(json.dumps(progress), flush=True)
            sys.stdout.flush()
        
    def send_log(message):
        send_progress(0, message=message)
    send_log("=== 内容驱动关键帧提取模式 ===")
    send_log(
        f"视频信息: 总帧数={total_frames}, 帧率={fps:.2f}FPS, 时长={duration:.2f}秒({duration//60:.0f}分{duration%60:.0f}秒), 理论密度={total_frames/duration:.1f}帧/秒"
        if duration > 0 else f"视频信息: 总帧数={total_frames}, 帧率={fps:.2f}FPS"
    )
    config = SMART_EXTRACTION_CONFIG
    min_interval = config['MIN_INTERVAL']
    max_interval = config['MAX_INTERVAL']
    scene_change_threshold = config['SCENE_CHANGE_THRESHOLD']
    send_log(f"提取参数: 最小间隔={min_interval}秒, 最大间隔={max_interval}秒, 场景变化阈值={scene_change_threshold}, 最大关键帧数={max_frames}")
    send_log("优化策略: 自适应步长 + 场景变化检测 + 质量评估")
    keyframe_paths = []
    last_saved_frame = None
    last_saved_timestamp = -1
    current_time = 0.0
    adaptive_step = 0.5
    skipped_dark = 0
    skipped_similar = 0
    last_report_time = 0
    used_filenames = set()
    send_log("开始逐帧分析...")
    active_second = 0
    has_active_second = False
    best_second_frame = None
    best_second_timestamp = None
    best_second_score = -1.0
    best_second_quality = 0.0
    best_second_change = 0.0
    def commit_best_of_second():
        nonlocal best_second_frame, best_second_timestamp, best_second_score
        nonlocal best_second_quality, best_second_change
        nonlocal last_saved_frame, last_saved_timestamp
        nonlocal keyframe_paths, used_filenames, adaptive_step
        if best_second_frame is None or best_second_timestamp is None:
            return
            
        # 发送处理进度
        send_progress(best_second_timestamp)
        
        if last_saved_frame is not None:
            try:
                current_gray = cv2.cvtColor(best_second_frame, cv2.COLOR_BGR2GRAY)
                last_gray = cv2.cvtColor(last_saved_frame, cv2.COLOR_BGR2GRAY)
                if _check_frame_similarity(last_gray, current_gray, threshold=0.8):
                    return
            except Exception:
                pass
        base_filename = f"keyframe_{len(keyframe_paths):03d}_{best_second_timestamp:.2f}s.jpg"
        filename = base_filename
        counter = 1
        while filename in used_filenames:
            name_part = base_filename.replace('.jpg', '')
            filename = f"{name_part}_v{counter}.jpg"
            counter += 1
        used_filenames.add(filename)
        filepath = os.path.join(output_dir, filename)
        resized_frame = _resize_to_480p(best_second_frame)
        cv2.imwrite(filepath, resized_frame, [cv2.IMWRITE_JPEG_QUALITY, 85])
        keyframe_paths.append(filepath)
        last_saved_frame = best_second_frame.copy()
        last_saved_timestamp = best_second_timestamp
        
        # 获取帧信息
        try:
            file_size = os.path.getsize(filepath) if os.path.exists(filepath) else 0
            height, width = best_second_frame.shape[:2]
            frame_info = {
                'saved_frames': int(len(keyframe_paths)),
                'new_frame_path': filepath,
                'new_frame_timestamp': float(best_second_timestamp),
                'change_score': float(best_second_change),
                'quality_score': float(best_second_quality),
                'width': int(width),
                'height': int(height),
                'file_size': int(file_size)
            }
            
            # 发送进度和帧信息
            send_progress(best_second_timestamp, 
                message=f"保存关键帧: {filename} | 综合得分={best_second_score:.1f} | 图像质量={best_second_quality:.1f}",
                frame_info=frame_info)
                
        except Exception as e:
            send_log(f"处理帧信息时出错: {e}")
        if best_second_change > scene_change_threshold:
            adaptive_step = min_interval
        else:
            adaptive_step = min(max_interval, adaptive_step * 1.5)
        best_second_frame = None
        best_second_timestamp = None
        best_second_score = -1.0
        best_second_quality = 0.0
        best_second_change = 0.0
    while current_time < duration and len(keyframe_paths) < max_frames:
        # 每秒发送一次进度更新
        if current_time - last_report_time >= 1:
            last_report_time = current_time
            progress = (current_time / duration) * 100 if duration > 0 else 0
            send_progress(current_time, 
                message=f"进度 {progress:.1f}% | 已提取 {len(keyframe_paths)} 帧 | 当前时间 {current_time:.1f}s")
            
            # 每10秒做一次垃圾回收
            if current_time % 10 == 0:
                import gc
                gc.collect()
        cur_sec = int(current_time)
        if not has_active_second:
            active_second = cur_sec
            has_active_second = True
        elif cur_sec != active_second:
            if len(keyframe_paths) < max_frames:
                commit_best_of_second()
            active_second = cur_sec
        frame_idx = int(current_time * fps)
        cap.set(cv2.CAP_PROP_POS_FRAMES, frame_idx)
        ret, frame = cap.read()
        if not ret:
            # 发送进度更新（即使读取失败）
            send_progress(current_time, message=f"跳过帧 {frame_idx} (读取失败)")
            current_time += adaptive_step
            continue
            
        # 不再每秒发送进度更新，只在保存关键帧时发送
        if _is_dark_frame(frame):
            skipped_dark += 1
            if skipped_dark % 10 == 1:
                send_log(f"跳过黑屏帧: 时间{current_time:.2f}s (亮度过低)")
            current_time += adaptive_step
            continue
        total_score, quality_score, change_score = _calculate_comprehensive_score(frame, last_saved_frame)
        if total_score > best_second_score:
            best_second_frame = frame.copy()
            best_second_timestamp = current_time
            best_second_score = total_score
            best_second_quality = quality_score
            best_second_change = change_score
        if change_score > scene_change_threshold:
            adaptive_step = min_interval
        else:
            adaptive_step = min(max_interval, adaptive_step * 1.2)
        current_time += adaptive_step
    if len(keyframe_paths) < max_frames:
        commit_best_of_second()
    send_log(f"内容驱动提取完成: 最终保存={len(keyframe_paths)}帧, 跳过黑屏帧={skipped_dark}帧, 跳过相似帧={skipped_similar}帧, 覆盖时长={min(current_time, duration):.2f}秒" + (f", 平均间隔={duration/len(keyframe_paths):.2f}秒/帧" if len(keyframe_paths) > 0 else ""))
    return keyframe_paths


def _extract_smart_keyframes(cap, output_dir, max_frames, time_interval=None, progress_callback=None):
    start_time = time.time()
    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    fps = cap.get(cv2.CAP_PROP_FPS)
    duration = total_frames / fps if fps > 0 else 0
    
    # 直接打印进度信息到标准输出
    def print_progress(progress_data):
        import json
        import sys
        print(json.dumps(progress_data), flush=True)
        sys.stdout.flush()
    
    def send_log(message):
        print_progress({'log_message': message})
        logger.info(message)
    
    send_log("=== 优化智能关键帧提取 ===")
    send_log(f"视频信息: FPS={fps:.2f}, 总帧数={total_frames}, 时长={duration:.2f}秒")
    send_log(f"目标帧数: {max_frames}")
    
    if duration <= 120:
        send_log("策略选择: 短视频内容驱动模式 (时长≤2分钟)")
    else:
        send_log("策略选择: 长视频内容驱动模式 (时长>2分钟)")
    
    send_log("提取策略: 仅内容驱动，按秒聚合取最佳帧（每秒一帧）")
    
    # 使用修改后的print_progress作为回调
    keyframe_paths = _extract_content_driven_keyframes(cap, output_dir, max_frames, print_progress)
    
    processing_time = time.time() - start_time
    send_log("\n智能提取总结:")
    send_log(f"   - 最终帧数: {len(keyframe_paths)}帧")
    send_log(f"   - 处理耗时: {processing_time:.2f}秒")
    send_log(f"   - 提取效率: {len(keyframe_paths)/processing_time:.1f}帧/秒" if processing_time > 0 else "   - 提取效率: N/A")
    send_log(f"   - 覆盖密度: {len(keyframe_paths)/duration:.2f}帧/秒" if duration > 0 else "   - 覆盖密度: N/A")
    
    return keyframe_paths


def _extract_uniform_keyframes(cap, output_dir, max_frames, total_frames, include_cover=True, start_index=0, used_filenames=None):
    keyframe_paths = []
    interval = max(1, total_frames // max_frames)
    if used_filenames is None:
        used_filenames = set()
    for i in range(max_frames):
        frame_idx = i * interval
        if not include_cover and frame_idx == 0:
            frame_idx += max(1, interval)
        if frame_idx >= total_frames:
            break
        cap.set(cv2.CAP_PROP_POS_FRAMES, frame_idx)
        ret, frame = cap.read()
        if ret:
            timestamp = frame_idx / cap.get(cv2.CAP_PROP_FPS)
            base_filename = f"keyframe_{start_index + i:03d}_{timestamp:.2f}s.jpg"
            filename = base_filename
            counter = 1
            while filename in used_filenames:
                name_part = base_filename.replace('.jpg', '')
                filename = f"{name_part}_v{counter}.jpg"
                counter += 1
            used_filenames.add(filename)
            filepath = os.path.join(output_dir, filename)
            out = _resize_to_480p(frame)
            cv2.imwrite(filepath, out, [int(cv2.IMWRITE_JPEG_QUALITY), 85])
            keyframe_paths.append(filepath)
    return keyframe_paths


def _extract_interval_keyframes(cap, output_dir, interval, max_frames):
    keyframe_paths = []
    frame_idx = 0
    saved_count = 0
    while saved_count < max_frames:
        cap.set(cv2.CAP_PROP_POS_FRAMES, frame_idx)
        ret, frame = cap.read()
        if not ret:
            break
        timestamp = frame_idx / cap.get(cv2.CAP_PROP_FPS)
        filename = f"keyframe_{saved_count:03d}_{timestamp:.2f}s.jpg"
        filepath = os.path.join(output_dir, filename)
        out = _resize_to_480p(frame)
        cv2.imwrite(filepath, out, [int(cv2.IMWRITE_JPEG_QUALITY), 85])
        keyframe_paths.append(filepath)
        frame_idx += interval
        saved_count += 1
    return keyframe_paths


def debug_keyframe_extraction(video_path, output_dir, max_frames=200, method='smart'):
    debug_info = {
        'video_info': {},
        'extraction_params': {},
        'keyframes': [],
        'performance': {},
        'scene_analysis': {
            'change_scores': {
                'min': 999999.0,  # 使用一个足够大的数字代替infinity
                'max': 0.0,
                'avg': 0.0,
                'std': 0.0,
                'scores': []  # 存储所有得分用于计算
            },
            'scene_changes': [],
            'frame_selection_reasons': []
        },
        'quality_metrics': {
            'clarity_scores': {
                'min': 999999.0,
                'max': 0.0,
                'avg': 0.0,
                'scores': []
            },
            'brightness_stats': {
                'min': 999999.0,
                'max': 0.0,
                'avg': 0.0,
                'values': []
            },
            'blur_detection': [],
            'frame_quality_scores': []
        },
        'extraction_stats': {
            'adaptive_steps': [],
            'skipped_frames': {
                'dark_frames': 0,
                'similar_frames': 0,
                'low_quality_frames': 0
            },
            'frame_intervals': [],
            'processing_stages': []
        }
    }
    
    # 添加处理阶段记录函数
    def add_processing_stage(stage, details):
        debug_info['extraction_stats']['processing_stages'].append({
            'stage': stage,
            'time': time.time(),
            'details': details
        })
    
    # 更新场景分析数据
    def update_scene_analysis(frame_idx, change_score, reason):
        scores = debug_info['scene_analysis']['change_scores']
        scores['min'] = min(scores['min'], change_score)
        scores['max'] = max(scores['max'], change_score)
        scores['scores'].append(change_score)
        
        if change_score > SMART_EXTRACTION_CONFIG['SCENE_CHANGE_THRESHOLD']:
            debug_info['scene_analysis']['scene_changes'].append({
                'frame': frame_idx,
                'score': change_score,
                'time': frame_idx / cap.get(cv2.CAP_PROP_FPS)
            })
        
        debug_info['scene_analysis']['frame_selection_reasons'].append({
            'frame': frame_idx,
            'reason': reason,
            'change_score': change_score
        })
    
    # 更新质量指标数据
    def update_quality_metrics(frame):
        # 计算亮度
        gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
        brightness = np.mean(gray)
        debug_info['quality_metrics']['brightness_stats']['values'].append(brightness)
        debug_info['quality_metrics']['brightness_stats']['min'] = min(
            debug_info['quality_metrics']['brightness_stats']['min'], 
            brightness
        )
        debug_info['quality_metrics']['brightness_stats']['max'] = max(
            debug_info['quality_metrics']['brightness_stats']['max'], 
            brightness
        )
        
        # 计算清晰度
        clarity = cv2.Laplacian(gray, cv2.CV_64F).var()
        debug_info['quality_metrics']['clarity_scores']['scores'].append(clarity)
        debug_info['quality_metrics']['clarity_scores']['min'] = min(
            debug_info['quality_metrics']['clarity_scores']['min'], 
            clarity
        )
        debug_info['quality_metrics']['clarity_scores']['max'] = max(
            debug_info['quality_metrics']['clarity_scores']['max'], 
            clarity
        )
        
        # 模糊检测
        is_blurry = clarity < 100  # 设置一个阈值
        debug_info['quality_metrics']['blur_detection'].append({
            'frame': len(debug_info['quality_metrics']['blur_detection']),
            'is_blurry': is_blurry,
            'clarity_score': clarity
        })
    
    # 在处理结束时计算平均值和标准差
    def finalize_debug_info():
        # 计算场景变化得分的统计信息
        scores = debug_info['scene_analysis']['change_scores']['scores']
        if scores:
            debug_info['scene_analysis']['change_scores']['avg'] = np.mean(scores)
            debug_info['scene_analysis']['change_scores']['std'] = np.std(scores)
        
        # 计算清晰度得分的平均值
        clarity_scores = debug_info['quality_metrics']['clarity_scores']['scores']
        if clarity_scores:
            debug_info['quality_metrics']['clarity_scores']['avg'] = np.mean(clarity_scores)
        
        # 计算亮度的平均值
        brightness_values = debug_info['quality_metrics']['brightness_stats']['values']
        if brightness_values:
            debug_info['quality_metrics']['brightness_stats']['avg'] = np.mean(brightness_values)
        
        # 清理中间数据
        debug_info['scene_analysis']['change_scores'].pop('scores', None)
        debug_info['quality_metrics']['clarity_scores'].pop('scores', None)
        debug_info['quality_metrics']['brightness_stats'].pop('values', None)
        
        # 如果min值仍然是初始值，设置为0
        if debug_info['scene_analysis']['change_scores']['min'] == 999999.0:
            debug_info['scene_analysis']['change_scores']['min'] = 0.0
        if debug_info['quality_metrics']['clarity_scores']['min'] == 999999.0:
            debug_info['quality_metrics']['clarity_scores']['min'] = 0.0
        if debug_info['quality_metrics']['brightness_stats']['min'] == 999999.0:
            debug_info['quality_metrics']['brightness_stats']['min'] = 0.0
    start_time = time.time()
    try:
        cap = cv2.VideoCapture(video_path)
        if cap.isOpened():
            debug_info['video_info'] = {
                'fps': cap.get(cv2.CAP_PROP_FPS),
                'total_frames': int(cap.get(cv2.CAP_PROP_FRAME_COUNT)),
                'width': int(cap.get(cv2.CAP_PROP_FRAME_WIDTH)),
                'height': int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT)),
                'duration': cap.get(cv2.CAP_PROP_FRAME_COUNT) / cap.get(cv2.CAP_PROP_FPS)
            }
            cap.release()
        debug_info['extraction_params'] = {
            'method': method,
            'max_frames': max_frames,
            'change_threshold': SMART_EXTRACTION_CONFIG['SCENE_CHANGE_THRESHOLD'],
            'quality_threshold': SMART_EXTRACTION_CONFIG['QUALITY_THRESHOLD'],
        }
        keyframes = extract_keyframes(video_path, output_dir, max_frames, method)
        processing_time = time.time() - start_time
        debug_info['performance'] = {
            'processing_time': processing_time,
            'total_keyframes': len(keyframes),
            'fps_performance': len(keyframes) / processing_time if processing_time > 0 else 0
        }
        debug_filename = f'debug_keyframes_{int(time.time())}.json'
        debug_filepath = os.path.join(output_dir, debug_filename)
        with open(debug_filepath, 'w', encoding='utf-8') as f:
            json.dump(debug_info, f, indent=2, ensure_ascii=False)
        logger.info(f"🔍 调试信息已保存: {debug_filepath}")
        return keyframes, debug_info
    except Exception as e:
        debug_info['error'] = str(e)
        logger.error(f"调试模式提取失败: {e}")
        return [], debug_info


def get_video_info(video_path):
    try:
        cap = cv2.VideoCapture(video_path)
        if not cap.isOpened():
            return None
        info = {
            'total_frames': int(cap.get(cv2.CAP_PROP_FRAME_COUNT)),
            'fps': cap.get(cv2.CAP_PROP_FPS),
            'width': int(cap.get(cv2.CAP_PROP_FRAME_WIDTH)),
            'height': int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT)),
        }
        info['duration'] = info['total_frames'] / info['fps'] if info['fps'] > 0 else 0
        cap.release()
        return info
    except Exception as e:
        logger.error(f"获取视频信息失败: {e}")
        return None


if __name__ == '__main__':
    # 轻量命令行入口：支持直接传入本地mp4文件进行关键帧提取
    try:
        import argparse
        import sys
        import os

        parser = argparse.ArgumentParser(description='本地mp4关键帧提取（独立，无需Django）')
        parser.add_argument('video', nargs='?', help='本地视频文件路径，如 /path/to/video.mp4')
        parser.add_argument('--out', help='输出目录（默认: 与视频同目录的 {stem}_keyframes）')
        parser.add_argument('--max-frames', type=int, default=300, help='最大关键帧数量，默认300')
        parser.add_argument('--mode', choices=['smart', 'uniform', 'interval'], default='smart', help='提取模式，默认smart')
        parser.add_argument('--time-interval', type=float, default=None, help='interval模式的秒级间隔；智能模式仅作参考')
        parser.add_argument('--get-info', action='store_true', help='仅获取视频信息，不提取关键帧')
        parser.add_argument('--debug', action='store_true', help='启用调试模式，生成调试信息文件')
        args = parser.parse_args()

        # 如果只是获取视频信息
        if args.get_info:
            if not args.video:
                print("[ERROR] 需要指定视频文件路径", file=sys.stderr)
                sys.exit(2)
            
            video_path = os.path.abspath(args.video)
            if not os.path.exists(video_path) or not os.path.isfile(video_path):
                print(f"[ERROR] 无效的视频文件: {video_path}", file=sys.stderr)
                sys.exit(2)
            
            video_info = get_video_info(video_path)
            if video_info:
                import json
                print(json.dumps(video_info, indent=2))
                sys.exit(0)
            else:
                print("[ERROR] 无法获取视频信息", file=sys.stderr)
                sys.exit(1)

        video_path = os.path.abspath(args.video)
        if not os.path.exists(video_path) or not os.path.isfile(video_path):
            print(f"[ERROR] 无效的视频文件: {video_path}", file=sys.stderr)
            sys.exit(2)

        # 计算默认输出目录：<video_dir>/<video_stem>_keyframes
        if args.out:
            output_dir = os.path.abspath(args.out)
        else:
            video_dir = os.path.dirname(video_path)
            video_stem = os.path.splitext(os.path.basename(video_path))[0]
            output_dir = os.path.join(video_dir, f"{video_stem}_keyframes")

        os.makedirs(output_dir, exist_ok=True)

        # 根据是否启用调试模式选择不同的提取方法
        if args.debug:
            paths, debug_info = debug_keyframe_extraction(
                video_path=video_path,
                output_dir=output_dir,
                max_frames=args.max_frames,
                method=args.mode
            )
        else:
            paths = extract_keyframes(
                video_path=video_path,
                output_dir=output_dir,
                max_frames=args.max_frames,
                method=args.mode,
                time_interval=args.time_interval,
                progress_callback=None,
            )

        if not paths:
            print("[ERROR] 未生成关键帧", file=sys.stderr)
            sys.exit(1)

        print(f"生成 {len(paths)} 张关键帧 → {output_dir}")
        # 打印前5个文件名作为简要输出
        try:
            import os as _os
            preview = [ _os.path.basename(p) for p in paths[:5] ]
            suffix = '...' if len(paths) > 5 else ''
            print(f"示例: {preview}{suffix}")
        except Exception:
            pass
        sys.exit(0)
    except SystemExit:
        raise
    except Exception as _e:
        import traceback as _tb, sys as _sys
        _tb.print_exc()
        _sys.exit(1)


`
)
