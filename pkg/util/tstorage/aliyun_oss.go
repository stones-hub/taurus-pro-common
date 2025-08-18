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
	"mime/multipart"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// AliyunOSS 实现了基于阿里云对象存储服务的存储功能
type AliyunOSS struct {
	// Endpoint OSS访问域名
	Endpoint string
	// AccessKeyId 访问密钥ID
	AccessKeyId string
	// AccessKeySecret 访问密钥密码
	AccessKeySecret string
	// BucketName 存储空间名称
	BucketName string
	// BasePath 存储根路径
	BasePath string
	// BucketUrl 存储空间访问URL
	BucketUrl string
}

// UploadFile 将文件上传到阿里云OSS
// 参数：
//   - file: 要上传的文件（multipart.FileHeader）
//
// 返回值：
//   - string: 文件的访问URL
//   - string: 文件在OSS中的路径
//   - error: 如果上传过程中出现错误则返回错误信息
//
// 使用示例：
//
//	oss := &AliyunOSS{
//	    Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
//	    AccessKeyId:     "your-access-key-id",
//	    AccessKeySecret: "your-access-key-secret",
//	    BucketName:      "your-bucket-name",
//	    BasePath:        "your-base-path",
//	    BucketUrl:       "https://your-bucket.oss-cn-hangzhou.aliyuncs.com",
//	}
//
//	// 上传文件
//	file := ctx.Request.MultipartForm.File["file"][0]
//	url, path, err := oss.UploadFile(file)
//	if err != nil {
//	    return err
//	}
//	fmt.Printf("文件已上传：%s，访问地址：%s\n", path, url)
//
// 注意事项：
//   - 需要正确配置阿里云OSS的访问信息
//   - 文件会按日期存储在 uploads/YYYY-MM-DD/ 目录下
//   - 不会自动重命名文件，可能会覆盖同名文件
//   - 建议在生产环境中添加文件名唯一性检查
//   - 注意处理大文件时的内存使用
func (aliyunOSS *AliyunOSS) UploadFile(file *multipart.FileHeader) (string, string, error) {
	bucket, err := NewBucket(aliyunOSS)
	if err != nil {
		return "", "", errors.New("function AliyunOSS.NewBucket() Failed, err:" + err.Error())
	}

	// 读取本地文件。
	f, openError := file.Open()
	if openError != nil {
		return "", "", errors.New("function file.Open() Failed, err:" + openError.Error())
	}
	defer f.Close() // 创建文件 defer 关闭
	// 上传阿里云路径 文件名格式 自己可以改 建议保证唯一性
	// yunFileTmpPath := filepath.Join("uploads", time.Now().Format("2006-01-02")) + "/" + file.Filename
	yunFileTmpPath := aliyunOSS.BasePath + "/" + "uploads" + "/" + time.Now().Format("2006-01-02") + "/" + file.Filename

	// 上传文件流。
	err = bucket.PutObject(yunFileTmpPath, f)
	if err != nil {
		return "", "", errors.New("function formUploader.Put() Failed, err:" + err.Error())
	}

	return aliyunOSS.BucketUrl + "/" + yunFileTmpPath, yunFileTmpPath, nil
}

// DeleteFile 从阿里云OSS删除指定的文件
// 参数：
//   - key: 要删除的文件路径（相对于Bucket的完整路径）
//
// 返回值：
//   - error: 如果删除过程中出现错误则返回错误信息
//
// 使用示例：
//
//	oss := &AliyunOSS{
//	    Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
//	    AccessKeyId:     "your-access-key-id",
//	    AccessKeySecret: "your-access-key-secret",
//	    BucketName:      "your-bucket-name",
//	}
//
//	// 删除文件
//	err := oss.DeleteFile("uploads/2023-06-15/example.jpg")
//	if err != nil {
//	    return fmt.Errorf("删除文件失败：%w", err)
//	}
//
// 注意事项：
//   - key 必须是文件的完整路径，包括所有目录
//   - 如果要删除目录，必须先删除目录下的所有文件
//   - 删除不存在的文件可能不会返回错误
//   - 删除操作不可恢复，请谨慎使用
//   - 建议在删除前先检查文件是否存在
func (aliyunOSS *AliyunOSS) DeleteFile(key string) error {
	bucket, err := NewBucket(aliyunOSS)
	if err != nil {
		return errors.New("function AliyunOSS.NewBucket() Failed, err:" + err.Error())
	}

	// 删除单个文件。objectName表示删除OSS文件时需要指定包含文件后缀在内的完整路径，例如abc/efg/123.jpg。
	// 如需删除文件夹，请将objectName设置为对应的文件夹名称。如果文件夹非空，则需要将文件夹下的所有object删除后才能删除该文件夹。
	err = bucket.DeleteObject(key)
	if err != nil {
		return errors.New("function bucketManager.Delete() failed, err:" + err.Error())
	}

	return nil
}

// NewBucket 创建阿里云OSS的Bucket实例
// 参数：
//   - aliyunOSS: OSS配置信息
//
// 返回值：
//   - *oss.Bucket: Bucket实例
//   - error: 如果创建过程中出现错误则返回错误信息
//
// 使用示例：
//
//	oss := &AliyunOSS{
//	    Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
//	    AccessKeyId:     "your-access-key-id",
//	    AccessKeySecret: "your-access-key-secret",
//	    BucketName:      "your-bucket-name",
//	}
//
//	bucket, err := NewBucket(oss)
//	if err != nil {
//	    return fmt.Errorf("创建Bucket失败：%w", err)
//	}
//
//	// 使用bucket进行操作
//	err = bucket.PutObject("example.txt", strings.NewReader("Hello, World!"))
//
// 注意事项：
//   - 需要正确配置Endpoint和访问凭证
//   - Bucket必须已经存在
//   - 建议复用Bucket实例而不是频繁创建
//   - 注意处理连接超时和重试
func NewBucket(aliyunOSS *AliyunOSS) (*oss.Bucket, error) {
	// 创建OSSClient实例。
	client, err := oss.New(aliyunOSS.Endpoint, aliyunOSS.AccessKeyId, aliyunOSS.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	// 获取存储空间。
	bucket, err := client.Bucket(aliyunOSS.BucketName)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}
