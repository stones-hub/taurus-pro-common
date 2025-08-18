// Copyright (c) 2025 Taurus Team. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: yelei
// Email: 61647649@qq.com
// Date: 2025-06-13

package tstorage

import (
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var mu sync.Mutex

// Local 实现了基于本地文件系统的对象存储
// 参考：
//   - [piexlmax](https://github.com/piexlmax)
//   - [ccfish86](https://github.com/ccfish86)
//   - [SliverHorn](https://github.com/SliverHorn)
type Local struct {
	// StorePath 是文件存储的物理路径
	StorePath string
	// Path 是文件访问的URL路径
	Path string
}

// UploadFile 将文件上传到本地文件系统
// 参数：
//   - file: 要上传的文件（multipart.FileHeader）
//
// 返回值：
//   - string: 文件的访问URL路径
//   - string: 文件名
//   - error: 如果上传过程中出现错误则返回错误信息
//
// 使用示例：
//
//	local := &Local{
//	    StorePath: "./uploads",    // 物理存储路径
//	    Path:      "/static",      // URL访问路径
//	}
//
//	// 上传文件
//	file := ctx.Request.MultipartForm.File["file"][0]
//	url, filename, err := local.UploadFile(file)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("文件已上传：%s，访问路径：%s\n", filename, url)
//
// 注意事项：
//   - 自动创建存储目录
//   - 文件名会被重命名为：原名_时间戳.扩展名
//   - 同名文件不会覆盖（因为有时间戳）
//   - 确保存储路径有写入权限
//   - 大文件上传要考虑内存使用
//
// 参考：
//   - [piexlmax](https://github.com/piexlmax)
//   - [ccfish86](https://github.com/ccfish86)
//   - [SliverHorn](https://github.com/SliverHorn)
func (local *Local) UploadFile(file *multipart.FileHeader) (string, string, error) {
	// 读取文件后缀
	ext := filepath.Ext(file.Filename)
	// 读取文件名并加密
	name := strings.TrimSuffix(file.Filename, ext)
	// 拼接新文件名
	filename := name + "_" + time.Now().Format("20060102150405") + ext
	// 尝试创建此路径
	mkdirErr := os.MkdirAll(local.StorePath, os.ModePerm)
	if mkdirErr != nil {
		return "", "", errors.New("function os.MkdirAll() failed, err:" + mkdirErr.Error())
	}
	// 拼接路径和文件名
	p := local.StorePath + "/" + filename
	filepath := local.Path + "/" + filename

	f, openError := file.Open() // 读取文件
	if openError != nil {
		return "", "", errors.New("function file.Open() failed, err:" + openError.Error())
	}
	defer f.Close() // 创建文件 defer 关闭

	out, createErr := os.Create(p)
	if createErr != nil {
		return "", "", errors.New("function os.Create() failed, err:" + createErr.Error())
	}
	defer out.Close() // 创建文件 defer 关闭

	_, copyErr := io.Copy(out, f) // 传输（拷贝）文件
	if copyErr != nil {
		return "", "", errors.New("function io.Copy() failed, err:" + copyErr.Error())
	}
	return filepath, filename, nil
}

// DeleteFile 从本地文件系统删除指定的文件
// 参数：
//   - key: 要删除的文件名
//
// 返回值：
//   - error: 如果删除过程中出现错误则返回错误信息
//
// 使用示例：
//
//	local := &Local{
//	    StorePath: "./uploads",
//	}
//
//	// 删除文件
//	err := local.DeleteFile("example_20230615123456.jpg")
//	if err != nil {
//	    if os.IsNotExist(err) {
//	        fmt.Println("文件不存在")
//	    } else {
//	        fmt.Printf("删除失败：%v\n", err)
//	    }
//	    return err
//	}
//
// 注意事项：
//   - key 不能为空
//   - key 不能包含非法字符（如 ../ 或 \）
//   - 使用互斥锁保证并发安全
//   - 如果文件不存在会返回错误
//   - 删除操作不可恢复，请谨慎使用
//
// 参考：
//   - [piexlmax](https://github.com/piexlmax)
//   - [ccfish86](https://github.com/ccfish86)
//   - [SliverHorn](https://github.com/SliverHorn)
func (local *Local) DeleteFile(key string) error {
	// 检查 key 是否为空
	if key == "" {
		return errors.New("key不能为空")
	}

	// 验证 key 是否包含非法字符或尝试访问存储路径之外的文件
	if strings.Contains(key, "..") || strings.ContainsAny(key, `\/:*?"<>|`) {
		return errors.New("非法的key")
	}

	p := filepath.Join(local.StorePath, key)

	// 检查文件是否存在
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return errors.New("文件不存在")
	}

	// 使用文件锁防止并发删除
	mu.Lock()
	defer mu.Unlock()

	err := os.Remove(p)
	if err != nil {
		return errors.New("文件删除失败: " + err.Error())
	}

	return nil
}
