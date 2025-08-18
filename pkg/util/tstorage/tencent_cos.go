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
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// TencentCOS 实现了基于腾讯云对象存储服务的存储功能
type TencentCOS struct {
	// Bucket 存储桶名称
	Bucket string
	// Region 地域信息
	Region string
	// SecretID 访问密钥ID
	SecretID string
	// SecretKey 访问密钥密码
	SecretKey string
	// PathPrefix 存储路径前缀
	PathPrefix string
	// BaseURL 访问基础URL
	BaseURL string
}

// UploadFile 将文件上传到腾讯云COS
// 参数：
//   - file: 要上传的文件（multipart.FileHeader）
//
// 返回值：
//   - string: 文件的访问URL
//   - string: 文件在COS中的路径
//   - error: 如果上传过程中出现错误则返回错误信息
//
// 使用示例：
//
//	cos := &TencentCOS{
//	    Bucket:     "your-bucket-1234567890",
//	    Region:     "ap-guangzhou",
//	    SecretID:   "your-secret-id",
//	    SecretKey:  "your-secret-key",
//	    PathPrefix: "uploads",
//	    BaseURL:    "https://your-bucket-1234567890.cos.ap-guangzhou.myqcloud.com",
//	}
//
//	// 上传文件
//	file := ctx.Request.MultipartForm.File["file"][0]
//	url, path, err := cos.UploadFile(file)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("文件已上传：%s，访问地址：%s\n", path, url)
//
// 注意事项：
//   - 需要正确配置腾讯云COS的访问信息
//   - 文件名会被加上时间戳前缀避免冲突
//   - 文件会存储在 PathPrefix 指定的目录下
//   - 注意处理大文件时的内存使用
//   - 建议在生产环境中添加更多的错误处理
func (tencentCOS *TencentCOS) UploadFile(file *multipart.FileHeader) (string, string, error) {
	client := NewClient(tencentCOS)
	f, openError := file.Open()
	if openError != nil {
		return "", "", errors.New("function file.Open() failed, err:" + openError.Error())
	}
	defer f.Close() // 创建文件 defer 关闭
	fileKey := fmt.Sprintf("%d%s", time.Now().Unix(), file.Filename)

	_, err := client.Object.Put(context.Background(), tencentCOS.PathPrefix+"/"+fileKey, f, nil)
	if err != nil {
		panic(err)
	}
	return tencentCOS.BaseURL + "/" + tencentCOS.PathPrefix + "/" + fileKey, fileKey, nil
}

// DeleteFile 从腾讯云COS删除指定的文件
// 参数：
//   - key: 要删除的文件名（不包含PathPrefix）
//
// 返回值：
//   - error: 如果删除过程中出现错误则返回错误信息
//
// 使用示例：
//
//	cos := &TencentCOS{
//	    Bucket:     "your-bucket-1234567890",
//	    Region:     "ap-guangzhou",
//	    SecretID:   "your-secret-id",
//	    SecretKey:  "your-secret-key",
//	    PathPrefix: "uploads",
//	}
//
//	// 删除文件
//	err := cos.DeleteFile("1623744000example.jpg")
//	if err != nil {
//	    return fmt.Errorf("删除文件失败：%w", err)
//	}
//
// 注意事项：
//   - key 是文件名，不需要包含 PathPrefix
//   - 实际删除的路径是 PathPrefix/key
//   - 删除不存在的文件可能会返回错误
//   - 删除操作不可恢复，请谨慎使用
//   - 建议在删除前先检查文件是否存在
func (tencentCOS *TencentCOS) DeleteFile(key string) error {
	client := NewClient(tencentCOS)
	name := tencentCOS.PathPrefix + "/" + key
	_, err := client.Object.Delete(context.Background(), name)
	if err != nil {
		return errors.New("function bucketManager.Delete() failed, err:" + err.Error())
	}
	return nil
}

// NewClient 创建腾讯云COS的客户端实例
// 参数：
//   - tencentCOS: COS配置信息
//
// 返回值：
//   - *cos.Client: COS客户端实例
//
// 使用示例：
//
//	cos := &TencentCOS{
//	    Bucket:     "your-bucket-1234567890",
//	    Region:     "ap-guangzhou",
//	    SecretID:   "your-secret-id",
//	    SecretKey:  "your-secret-key",
//	}
//
//	client := NewClient(cos)
//
//	// 使用客户端进行操作
//	_, err := client.Object.Put(context.Background(), "example.txt",
//	    strings.NewReader("Hello, World!"), nil)
//
// 注意事项：
//   - 需要正确配置访问凭证和地域信息
//   - 客户端会自动处理认证信息
//   - 建议复用客户端实例而不是频繁创建
//   - 注意处理连接超时和重试
//   - 生产环境建议添加自定义Transport配置
func NewClient(tencentCOS *TencentCOS) *cos.Client {
	urlStr, _ := url.Parse("https://" + tencentCOS.Bucket + ".cos." + tencentCOS.Region + ".myqcloud.com")
	baseURL := &cos.BaseURL{BucketURL: urlStr}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  tencentCOS.SecretID,
			SecretKey: tencentCOS.SecretKey,
		},
	})
	return client
}
