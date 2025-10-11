# FFmpeg 视频处理优化示例

这个示例展示了如何使用 FFmpeg 视频处理工具进行各种常见的视频处理操作。

## 示例目录结构

```
output/
├── basic/          # 基本视频处理示例
├── high_quality/   # 高质量视频处理示例
├── progress/       # 进度显示示例
├── format_convert/ # 格式转换示例
├── clip/           # 视频片段提取示例
├── compress/       # 视频压缩示例
└── no_audio/       # 无音频处理示例
```

## 示例详解

### 1. 基本视频处理 (Basic)

展示了最基本的视频处理操作：
- 转换为 MP4 格式
- 设置中等质量
- 调整为 720p 分辨率
- 设置 30fps 帧率
- 提取 10 秒片段

```go
options := &tffmpeg.VideoExtractionOptions{
    Format:       tffmpeg.VideoFormatMP4,
    Quality:      tffmpeg.VideoQualityMedium,
    Resolution:   "1280x720",
    Fps:          30,
    AudioEnabled: true,
    StartTime:    "00:00:15",
    Duration:     "00:00:10",
}
```

### 2. 高质量处理 (High Quality)

展示了高质量视频处理：
- 1080p 高清分辨率
- 60fps 高帧率
- 8Mbps 视频比特率
- 320kbps 音频比特率

### 3. 进度显示 (Progress)

展示了如何监控视频处理进度：
- 实时显示处理进度
- 显示当前处理时间
- 显示处理速度
- 计算总耗时

### 4. 格式转换 (Format Convert)

展示了多种格式转换：
- WebM 格式
- MKV 格式
- MOV 格式

### 5. 视频片段提取 (Clip)

展示了如何提取视频片段：
- 指定开始时间
- 设置片段时长
- 保持原始质量

### 6. 视频压缩 (Compress)

展示了视频压缩技术：
- 降低分辨率至 480p
- 降低视频比特率
- 降低音频质量
- 计算压缩比

### 7. 无音频处理 (No Audio)

展示了纯视频处理：
- 移除音频流
- 保持视频质量
- 高分辨率输出

## 使用方法

1. 准备输入视频文件：
```go
const inputVideo = "input.mp4"
```

2. 创建 FFmpeg 实例：
```go
ffmpeg := tffmpeg.NewFFmpegVideo(
    tffmpeg.WithVideoFFmpegPath("/usr/local/bin/ffmpeg"),
    tffmpeg.WithVideoFFprobePath("/usr/local/bin/ffprobe"),
)
```

3. 运行示例：
```go
runBasicExample(ffmpeg)
runHighQualityExample(ffmpeg)
runProgressExample(ffmpeg)
runFormatConvertExample(ffmpeg)
runClipExample(ffmpeg)
runCompressExample(ffmpeg)
runNoAudioExample(ffmpeg)
```

## 注意事项

1. 确保输入视频存在且可访问
2. 检查输出目录的写入权限
3. 监控系统资源使用情况
4. 注意处理大文件时的内存使用

## 最佳实践

1. 根据实际需求选择合适的示例
2. 适当调整参数以达到最佳效果
3. 在处理大文件时使用进度监控
4. 注意保存和清理临时文件

## 调试建议

1. 使用进度回调监控处理过程
2. 检查输出文件的属性
3. 验证处理结果的质量
4. 记录处理时间和资源使用
