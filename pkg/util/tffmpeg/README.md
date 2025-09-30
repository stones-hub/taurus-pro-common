# FFmpeg音频处理工具

这是一个基于FFmpeg的Go语言音频处理工具包，提供了丰富的音频处理功能，包括音频提取、格式转换、批量处理等。

## 功能特点

- 从视频中提取音频
- 音频格式转换
- 批量音频处理
- 实时进度监控
- 音频质量控制
- 预设配置支持
- 并发处理能力
- 完善的错误处理

## 支持的音频格式

- MP3 (最常用的音频格式)
- WAV (无损音频格式)
- AAC (高效压缩格式)
- FLAC (无损压缩格式)
- M4A (苹果音频格式)
- OGG (开源音频格式)

## 安装要求

### 系统要求
- Go 1.16+
- FFmpeg 4.0+
- FFprobe (通常随FFmpeg一起安装)

### FFmpeg安装指南

#### macOS
使用Homebrew安装：
```bash
brew install ffmpeg
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install ffmpeg
```

#### Linux (CentOS/RHEL)
```bash
sudo yum install epel-release
sudo yum install ffmpeg ffmpeg-devel
```

#### Windows
1. 访问 [FFmpeg官方下载页面](https://ffmpeg.org/download.html)
2. 下载Windows版本
3. 解压到指定目录
4. 将FFmpeg的bin目录添加到系统PATH环境变量

### 验证安装
安装完成后，在终端运行以下命令验证：
```bash
ffmpeg -version
ffprobe -version
```

### Go包安装
```bash
go get github.com/stones-hub/taurus-pro-common/pkg/util/tffmpeg
```

### 环境配置
默认情况下，工具会在以下路径查找FFmpeg和FFprobe：
- macOS/Linux: `/usr/local/bin/ffmpeg` 和 `/usr/local/bin/ffprobe`
- Windows: 系统PATH中的`ffmpeg.exe`和`ffprobe.exe`

如果您的FFmpeg安装在其他位置，可以在创建实例时指定路径：
```go
ffmpeg := tffmpeg.NewFFmpegAudio(
    tffmpeg.WithFFmpegPath("/path/to/ffmpeg"),
    tffmpeg.WithFFprobePath("/path/to/ffprobe"),
)
```

## 快速开始

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "github.com/stones-hub/taurus-pro-common/pkg/util/tffmpeg"
)

func main() {
    // 创建FFmpeg音频处理工具实例
    ffmpeg := tffmpeg.NewFFmpegAudio()

    // 检查FFmpeg是否可用
    if err := ffmpeg.CheckFFmpegAvailable(); err != nil {
        panic(err)
    }

    // 基本音频提取
    options := &tffmpeg.AudioExtractionOptions{
        Format:     tffmpeg.AudioFormatMP3,     // 输出MP3格式
        Quality:    tffmpeg.AudioQualityMedium, // 中等质量
        Bitrate:    "128k",                     // 128k比特率
        SampleRate: 44100,                      // 44.1kHz采样率
        Channels:   2,                          // 立体声
    }

    result, err := ffmpeg.ExtractAudioFromVideo(context.Background(), "input.mp4", "./output", options)
    if err != nil {
        panic(err)
    }

    fmt.Printf("音频提取成功: %s\n", result.OutputPath)
}
```

### 带进度监控的音频提取

```go
options := &tffmpeg.AudioExtractionOptions{
    Format:         tffmpeg.AudioFormatMP3,
    Quality:        tffmpeg.AudioQualityHigh,
    EnableProgress: true,
    ProgressCallback: func(progress *tffmpeg.FFmpegProgress) {
        fmt.Printf("处理进度: %.1f%%\n", progress.Progress)
    },
}

result, err := ffmpeg.ExtractAudioFromVideo(context.Background(), "input.mp4", "./output", options)
```

### 使用预设配置

```go
// 使用语音优化预设
speechConfig := ffmpeg.CreatePresetConfig("speech")
result, err := ffmpeg.ExtractAudioFromVideo(context.Background(), "input.mp4", "./output", speechConfig)

// 使用音乐优化预设
musicConfig := ffmpeg.CreatePresetConfig("music")
result, err := ffmpeg.ExtractAudioFromVideo(context.Background(), "input.mp4", "./output", musicConfig)
```

### 批量处理（用户自定义实现）

由于批量处理的需求和策略因人而异，我们建议用户根据自己的需求自行实现批量处理逻辑。以下是两种常见的实现方式：

#### 串行处理
```go
func batchExtractAudioSerial(ffmpeg *tffmpeg.FFmpegAudio, ctx context.Context, videoPaths []string, outputDir string) []*tffmpeg.AudioExtractionResult {
    results := make([]*tffmpeg.AudioExtractionResult, 0, len(videoPaths))
    
    options := &tffmpeg.AudioExtractionOptions{
        Format:  tffmpeg.AudioFormatMP3,
        Quality: tffmpeg.AudioQualityMedium,
    }
    
    for i, videoPath := range videoPaths {
        videoName := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
        videoOutputDir := filepath.Join(outputDir, videoName)
        
        result, err := ffmpeg.ExtractAudioFromVideo(ctx, videoPath, videoOutputDir, options)
        if err != nil {
            log.Printf("处理视频 %s 失败: %v", videoPath, err)
            continue
        }
        
        results = append(results, result)
    }
    
    return results
}
```

#### 并行处理
```go
func batchExtractAudioParallel(ffmpeg *tffmpeg.FFmpegAudio, ctx context.Context, videoPaths []string, outputDir string) []*tffmpeg.AudioExtractionResult {
    results := make([]*tffmpeg.AudioExtractionResult, 0, len(videoPaths))
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    // 限制并发数量
    maxConcurrency := 3
    semaphore := make(chan struct{}, maxConcurrency)
    
    for i, videoPath := range videoPaths {
        wg.Add(1)
        go func(index int, path string) {
            defer wg.Done()
            
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            videoName := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
            videoOutputDir := filepath.Join(outputDir, videoName)
            
            options := &tffmpeg.AudioExtractionOptions{
                Format:  tffmpeg.AudioFormatMP3,
                Quality: tffmpeg.AudioQualityMedium,
            }
            
            result, err := ffmpeg.ExtractAudioFromVideo(ctx, path, videoOutputDir, options)
            if err != nil {
                log.Printf("处理视频 %s 失败: %v", path, err)
                return
            }
            
            mu.Lock()
            results = append(results, result)
            mu.Unlock()
        }(i, videoPath)
    }
    
    wg.Wait()
    return results
}
```

## 高级功能

### 音频格式转换

```go
options := &tffmpeg.AudioExtractionOptions{
    Quality:    tffmpeg.AudioQualityHigh,
    SampleRate: 48000,
    Channels:   2,
}

result, err := ffmpeg.ConvertAudioFormat(
    context.Background(),
    "input.mp3",
    "./output",
    tffmpeg.AudioFormatAAC,
    options,
)
```

### 自定义音频处理

```go
options := &tffmpeg.AudioExtractionOptions{
    Format:        tffmpeg.AudioFormatMP3,
    Quality:       tffmpeg.AudioQualityHigh,
    Normalize:     true,         // 启用音量标准化
    RemoveSilence: true,         // 移除静音段
    Volume:        "1.5",        // 音量调整
    StartTime:     "00:00:10",   // 从10秒开始
    Duration:      "00:00:30",   // 提取30秒
}

result, err := ffmpeg.ExtractAudioFromVideo(context.Background(), "input.mp4", "./output", options)
```

## 预设配置说明

工具提供了四种预设配置：

1. `speech` - 语音优化
   - 单声道
   - 16kHz采样率
   - 64k比特率
   - 启用音量标准化和静音移除

2. `music` - 音乐优化
   - 立体声
   - 44.1kHz采样率
   - 192k比特率
   - 保持原始动态范围

3. `podcast` - 播客优化
   - 立体声
   - 44.1kHz采样率
   - 128k比特率
   - 启用音量标准化

4. `archive` - 归档配置
   - 立体声
   - 48kHz采样率
   - 320k比特率
   - FLAC格式
   - 保持原始质量

## 错误处理

工具提供了完善的错误处理机制：

- 输入文件检查
- 输出目录验证
- FFmpeg可用性检查
- 配置参数验证
- 上下文取消支持
- 并发安全保证

## 性能优化

- 支持并发处理
- 内存使用优化
- 进度监控开销最小化
- 支持不同质量级别的性能调优

## 注意事项

1. 确保系统已正确安装FFmpeg和FFprobe
2. 注意文件路径的权限问题
3. 大文件处理时建议启用进度监控
4. 并发处理时注意系统资源使用

## 贡献

欢迎提交Issue和Pull Request来帮助改进这个工具。

## 许可证

MIT License
