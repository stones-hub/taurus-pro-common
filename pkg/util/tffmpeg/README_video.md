# FFmpeg 视频处理工具

这是一个基于 FFmpeg 的 Go 语言视频处理工具库，提供了丰富的视频处理功能。

## 主要功能

- 视频格式转换
- 视频质量调整
- 视频分辨率修改
- 视频帧率控制
- 视频片段提取
- 音频处理控制
- 进度监控支持

## 支持的视频格式

- MP4 (.mp4)
- MOV (.mov)
- AVI (.avi)
- MKV (.mkv)
- WebM (.webm)
- FLV (.flv)

## 视频质量等级

- 低质量 (Low)：文件小，画质一般
- 中等质量 (Medium)：平衡文件大小和画质
- 高质量 (High)：画质好，文件较大
- 最佳质量 (Best)：最高画质，文件最大

## 使用方法

### 1. 创建 FFmpeg 实例

```go
ffmpeg := tffmpeg.NewFFmpegVideo(
    tffmpeg.WithVideoFFmpegPath("/usr/local/bin/ffmpeg"),
    tffmpeg.WithVideoFFprobePath("/usr/local/bin/ffprobe"),
    tffmpeg.WithVideoTempDir("/tmp/ffmpeg_video"),
)
```

### 2. 配置视频处理选项

```go
options := &tffmpeg.VideoExtractionOptions{
    Format:       tffmpeg.VideoFormatMP4,
    Quality:      tffmpeg.VideoQualityHigh,
    Resolution:   "1920x1080",
    Bitrate:      "2000k",
    Fps:          30,
    StartTime:    "00:00:10",
    Duration:     "00:00:30",
    AudioEnabled: true,
    AudioBitrate: "128k",
}
```

### 3. 处理视频

```go
result, err := ffmpeg.ExtractVideo(context.Background(), inputVideo, outputDir, options)
if err != nil {
    log.Printf("视频处理失败: %v\n", err)
    return
}
```

## 核心特性

### 自动配置优化

- 根据质量等级自动设置合适的编码参数
- 智能选择编码器和编码参数
- 自动处理音频流

### 格式适配

- 针对不同输出格式优化编码参数
- 自动选择最适合的编码器
- 保证输出视频的兼容性

### 进度监控

- 实时监控处理进度
- 支持进度回调函数
- 提供详细的处理状态信息

### 错误处理

- 完善的输入验证
- 详细的错误信息
- 异常情况的优雅处理

## 注意事项

1. 确保系统已安装 FFmpeg 和 FFprobe
2. 视频处理可能需要较大的系统资源
3. 处理大文件时建议启用进度监控
4. 注意检查输出目录的写入权限

## 依赖要求

- FFmpeg >= 4.0
- FFprobe >= 4.0
- Go >= 1.16

## 错误处理建议

1. 总是检查返回的错误
2. 验证输入文件的存在性
3. 确保输出目录的权限
4. 监控系统资源使用情况

## 性能优化建议

1. 合理设置视频比特率
2. 根据需求选择适当的质量等级
3. 使用合适的编码预设
4. 注意临时文件的清理
