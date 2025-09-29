"""
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
        def send_log(message):
            if progress_callback:
                progress_callback({'log_message': message})
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
    total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
    fps = cap.get(cv2.CAP_PROP_FPS)
    duration = total_frames / fps if fps > 0 else 0
    def send_log(message):
        if progress_callback:
            progress_callback({'log_message': message})
        logger.info(message)
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
        send_log(f"保存关键帧: {filename} | 综合得分={best_second_score:.1f} | 图像质量={best_second_quality:.1f}")
        if progress_callback:
            try:
                file_size = os.path.getsize(filepath) if os.path.exists(filepath) else 0
                height, width = best_second_frame.shape[:2]
                progress_callback({
                    'coverage': min(1.0, best_second_timestamp / duration) if duration > 0 else 0.0,
                    'elapsed_seconds': float(best_second_timestamp),
                    'duration_seconds': float(duration),
                    'saved_frames': int(len(keyframe_paths)),
                    'max_frames': int(max_frames),
                    'new_frame_path': filepath,
                    'new_frame_timestamp': float(best_second_timestamp),
                    'change_score': float(best_second_change),
                    'quality_score': float(best_second_quality),
                    'width': int(width),
                    'height': int(height),
                    'file_size': int(file_size)
                })
            except Exception:
                pass
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
        if current_time - last_report_time >= 10:
            last_report_time = current_time
            progress = (current_time / duration) * 100 if duration > 0 else 0
            send_log(f"进度 {progress:.1f}% | 已提取 {len(keyframe_paths)} 帧 | 当前时间 {current_time:.1f}s")
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
            current_time += adaptive_step
            continue
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
    logger.info(f"\n=== 优化智能关键帧提取 ===")
    logger.info(f"视频信息: FPS={fps:.2f}, 总帧数={total_frames}, 时长={duration:.2f}秒")
    logger.info(f"目标帧数: {max_frames}")
    if duration <= 120:
        logger.info(f"策略选择: 短视频内容驱动模式 (时长≤2分钟)")
    else:
        logger.info(f"策略选择: 长视频内容驱动模式 (时长>2分钟)")
    logger.info(f"提取策略: 仅内容驱动，按秒聚合取最佳帧（每秒一帧）")
    keyframe_paths = _extract_content_driven_keyframes(cap, output_dir, max_frames, progress_callback)
    processing_time = time.time() - start_time
    logger.info(f"\n智能提取总结:")
    logger.info(f"   - 最终帧数: {len(keyframe_paths)}帧")
    logger.info(f"   - 处理耗时: {processing_time:.2f}秒")
    logger.info(f"   - 提取效率: {len(keyframe_paths)/processing_time:.1f}帧/秒" if processing_time > 0 else "   - 提取效率: N/A")
    logger.info(f"   - 覆盖密度: {len(keyframe_paths)/duration:.2f}帧/秒" if duration > 0 else "   - 覆盖密度: N/A")
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
        'performance': {}
    }
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
        parser.add_argument('video', help='本地视频文件路径，如 /path/to/video.mp4')
        parser.add_argument('--out', help='输出目录（默认: 与视频同目录的 {stem}_keyframes）')
        parser.add_argument('--max-frames', type=int, default=300, help='最大关键帧数量，默认300')
        parser.add_argument('--mode', choices=['smart', 'uniform', 'interval'], default='smart', help='提取模式，默认smart')
        parser.add_argument('--time-interval', type=float, default=None, help='interval模式的秒级间隔；智能模式仅作参考')
        args = parser.parse_args()

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


