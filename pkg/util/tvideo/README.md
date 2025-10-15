# 视频关键帧提取工具

这是一个基于Go和Python的视频关键帧提取工具，支持多种提取模式，能够智能地从视频中提取代表性的关键帧。

## 功能特性

- 🎯 支持三种提取模式：
  - 智能模式（smart）：基于场景变化和图像质量的智能提取
  - 均匀分布模式（uniform）：按时间均匀提取
  - 时间间隔模式（interval）：按固定时间间隔提取

- 🛠 提供丰富的配置选项：
  - 最大关键帧数量控制
  - 时间间隔设置
  - 调试模式支持
  - 进度回调功能

- 📊 智能分析功能：
  - 场景变化检测
  - 图像质量评估
  - 黑帧检测
  - 相似帧过滤

## 环境要求

- Go 1.16+
- Python 3.7+
- OpenCV (cv2)
- NumPy

## 环境安装

推荐使用Conda创建独立的Python环境：

```bash
# 创建新的conda环境
conda create -n keyframe python=3.7
conda activate keyframe

# 安装所需的Python包
conda install -c conda-forge opencv
conda install numpy

# 验证安装
python -c "import cv2; import numpy; print('OpenCV version:', cv2.__version__)"
```

## 使用示例

### 基本使用

```go
package main

import (
    "context"
    "fmt"
    "github.com/stones-hub/taurus-pro-common/pkg/util/tvideo"
)

func main() {
    // 创建提取器实例
    extractor, err := tvideo.NewKeyframeExtractor()
    if err != nil {
        panic(err)
    }
    defer extractor.Close()

    // 创建提取选项
    options := &tvideo.KeyframeExtractionOptions{
        MaxFrames:   300,           // 最大提取300帧
        Mode:        tvideo.ModeSmart,  // 使用智能模式
        OutputDir:   "./keyframes", // 输出目录
        EnableDebug: true,         // 启用调试模式
    }

    // 执行提取
    result, debugInfo, err := extractor.ExtractKeyframes(
        context.Background(),
        "input.mp4",
        options,
    )
    if err != nil {
        panic(err)
    }

    fmt.Printf("成功提取 %d 个关键帧\n", len(result.KeyframePaths))
}
```

### 使用进度回调

```go
options := &tvideo.KeyframeExtractionOptions{
    MaxFrames: 300,
    Mode:     tvideo.ModeSmart,
    OutputDir: "./keyframes",
    ProgressCallback: func(progress *tvideo.KeyframeProgress) {
        fmt.Printf("进度: %.2f%%, 已处理: %.2f秒\n",
            progress.Coverage*100,
            progress.ElapsedSeconds,
        )
    },
}
```

### 使用预设配置

```go
// 使用快速提取预设
options := extractor.CreatePresetOptions("fast")

// 使用质量优先预设
options := extractor.CreatePresetOptions("quality")

// 使用间隔提取预设
options := extractor.CreatePresetOptions("interval")
```

## 提取模式说明

### 1. 智能模式 (ModeSmart)

- 基于场景变化和图像质量的智能提取
- 自动调整提取间隔
- 过滤低质量和重复帧
- 适合内容丰富的视频

### 2. 均匀分布模式 (ModeUniform)

- 按时间均匀分布提取关键帧
- 确保覆盖整个视频时长
- 适合变化平缓的视频

### 3. 时间间隔模式 (ModeInterval)

- 按固定时间间隔提取
- 可指定具体的时间间隔
- 适合需要精确控制的场景

## 调试功能

启用调试模式后，可以获得详细的提取过程信息：

- 场景变化分数
- 图像质量指标
- 提取决策依据
- 性能统计数据

```go
options := &tvideo.KeyframeExtractionOptions{
    EnableDebug: true,
}
```

## 注意事项

1. 确保Python环境正确配置，可以通过`python3 --version`和`python3 -c "import cv2"`验证

2. 对于大型视频文件，建议：
   - 适当增加MaxFrames值
   - 使用进度回调监控处理进度
   - 考虑使用均匀分布模式提高处理速度

3. 临时文件处理：
   - 工具会创建临时文件和目录
   - 使用完成后调用`Close()`清理
   - 确保输出目录有足够的存储空间

4. 错误处理：
   - 工具内置了多级容错机制
   - 智能模式失败会自动降级到其他模式
   - 建议实现错误监控和日志记录

## 性能优化建议

1. 预处理：
   - 对于超长视频，考虑先进行分段处理
   - 可以预先过滤掉不需要的视频段

2. 资源管理：
   - 及时清理临时文件
   - 控制并发处理的视频数量
   - 监控内存使用情况

3. 输出管理：
   - 定期清理旧的输出文件
   - 实现输出文件的归档策略

## 常见问题

1. Python环境问题：
   ```bash
   # 检查Python版本
   python3 --version

   # 检查OpenCV安装
   python3 -c "import cv2; print(cv2.__version__)"

   # 检查NumPy安装
   python3 -c "import numpy; print(numpy.__version__)"
   ```

2. 权限问题：
   - 确保有临时目录的写入权限
   - 确保有输出目录的写入权限

3. 内存问题：
   - 对于大型视频，注意监控内存使用
   - 考虑使用流式处理方式

## 许可证

Apache License 2.0
