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
	"mime/multipart"
)

// OSS 定义了对象存储服务的通用接口，支持多种存储实现
// 参考：
//   - [SliverHorn](https://github.com/SliverHorn)
//   - [ccfish86](https://github.com/ccfish86)
//
// 实现：
//   - Local: 本地文件系统存储
//   - TencentCOS: 腾讯云对象存储
//   - AliyunOSS: 阿里云对象存储
//
// 使用示例：
//
//	// 创建存储实例
//	storage := tstorage.NewOss("local")  // 或 "tencent-cos" 或 "aliyun-oss"
//
//	// 上传文件
//	file := ctx.Request.MultipartForm.File["file"][0]
//	url, key, err := storage.UploadFile(file)
//	if err != nil {
//	    return err
//	}
//
//	// 删除文件
//	err = storage.DeleteFile(key)
//	if err != nil {
//	    return err
//	}
//
// 注意事项：
//   - 不同实现可能需要不同的配置
//   - 返回的URL格式取决于具体实现
//   - 文件大小限制取决于具体实现
//   - 建议在初始化时进行配置验证
type OSS interface {
	UploadFile(file *multipart.FileHeader) (string, string, error)
	DeleteFile(key string) error
}

// NewOss 创建指定类型的对象存储服务实例
// 参数：
//   - ossType: 存储类型，可选值：
//   - "local": 本地文件系统存储（默认）
//   - "tencent-cos": 腾讯云对象存储
//   - "aliyun-oss": 阿里云对象存储
//
// 返回值：
//   - OSS: 对象存储服务实例
//
// 使用示例：
//
//	// 使用本地存储
//	storage := tstorage.NewOss("local")
//
//	// 使用腾讯云COS
//	storage := tstorage.NewOss("tencent-cos")
//
//	// 使用阿里云OSS
//	storage := tstorage.NewOss("aliyun-oss")
//
//	// 使用默认存储（本地）
//	storage := tstorage.NewOss("")
//
// 注意事项：
//   - 如果指定的类型无效，默认使用本地存储
//   - 不同存储类型需要不同的配置
//   - 建议使用常量定义存储类型
//   - 参考各存储实现的具体文档
//
// 参考：
//   - [SliverHorn](https://github.com/SliverHorn)
//   - [ccfish86](https://github.com/ccfish86)
func NewOss(ossType string) OSS {
	switch ossType {
	case "local":
		return &Local{}
	case "tencent-cos":
		return &TencentCOS{}
	case "aliyun-oss":
		return &AliyunOSS{}
	default:
		return &Local{}
	}
}
