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

package upload

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

// UploadFile2Remote 将文件上传到远端
// URL 远端地址
// params 请求参数
// filePath 要上传的文件路径
// fileType 文件类型
// fileFieldName 表单中存储文件的表单的字段名
func UploadFile2Remote(URL string, params map[string]string, filePath string, fileFieldName string) ([]byte, error) {
	var (
		err      error
		u        *url.URL
		query    url.Values
		body     = &bytes.Buffer{} // 空body
		form     *multipart.Writer
		file     *os.File
		fileInfo os.FileInfo
		formFile io.Writer // 在表单中用于存储文件的句柄
	)

	// -----------------> 解析上传地址 <-----------------
	u, err = url.Parse(URL)
	if err != nil {
		return nil, err
	}
	// 添加查询参数
	query = u.Query()
	for k, v := range params {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()
	// -----------------> 创建上传URL <-----------------

	// ------------->打开文件将要上传的文件 <-------------
	file, err = os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 获取文件属性比如大小
	fileInfo, err = file.Stat()
	if err != nil {
		return nil, err
	}
	// ------------->打开文件将要上传的文件 <-------------

	// 创建一个新的 multipart writer, 可以理解是一个空表单
	form = multipart.NewWriter(body)
	// 创建一个表单内的字段用于存储文件, fieldname就是表单内的字段名称， filename就是字段值
	formFile, err = form.CreateFormFile(fileFieldName, file.Name())
	if err != nil {
		return nil, err
	}
	// 将需要上传的文件流写入到表单字段中
	_, err = io.Copy(formFile, file)
	if err != nil {
		return nil, err
	}

	// 添加额外的表单字段
	err = form.WriteField("filename", file.Name())
	if err != nil {
		return nil, err
	}

	err = form.WriteField("filelength", strconv.FormatInt(fileInfo.Size(), 10))
	if err != nil {
		return nil, err
	}

	err = form.WriteField("content-type", fileInfo.Mode().String())
	if err != nil {
		return nil, err
	}

	// 关闭表单, 此时数据都写入了bodyData中
	err = form.Close()
	if err != nil {
		return nil, err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", u.String(), body) // 经过上面的写入，空body里面有了数据了
	if err != nil {
		return nil, err
	}

	// 设置 Content-Type 和 Content-Length
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.Header.Set("Content-Length", strconv.Itoa(len(body.Bytes())))

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 返回响应内容
	return respBody, nil
}

// Upload2Remote 将web段提交过来的文件，转存到URL地址下
// URL 远端地址
// params 请求参数
// file  的数据就是来源 _, header, err := c.Request.FormFile("file")
// fileFieldName 表单中存储文件的表单的字段名
func Upload2Remote(URL string, params map[string]string, file *multipart.FileHeader, fileFieldName string) ([]byte, error) {

	var (
		err      error
		u        *url.URL
		query    url.Values
		body     = &bytes.Buffer{}
		form     *multipart.Writer
		f        multipart.File
		formFile io.Writer // 在表单中用于存储文件的句柄
	)

	// -----------------> 解析上传地址 <-----------------
	u, err = url.Parse(URL)
	if err != nil {
		return nil, err
	}
	// 添加查询参数
	query = u.Query()
	for k, v := range params {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()
	// -----------------> 创建上传URL <-----------------

	// ------------->打开文件将要上传的文件 <-------------
	f, err = file.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// ------------->打开文件将要上传的文件 <-------------

	// 创建一个新的 multipart writer, 可以理解是一个表单
	form = multipart.NewWriter(body)
	// 创建一个表单内的字段用于存储文件, fieldname就是表单内的字段名称， filename就是字段值
	formFile, err = form.CreateFormFile(fileFieldName, file.Filename)
	if err != nil {
		return nil, err
	}
	// 将需要上传的文件流写入到表单字段中
	_, err = io.Copy(formFile, f)
	if err != nil {
		return nil, err
	}

	err = form.WriteField("filename", file.Filename)
	if err != nil {
		return nil, err
	}

	err = form.WriteField("filelength", strconv.FormatInt(file.Size, 10))
	if err != nil {
		return nil, err
	}

	err = form.WriteField("content-type", file.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	// 关闭表单, 此时数据都写入了bodyData中
	err = form.Close()
	if err != nil {
		return nil, err
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return nil, err
	}

	// 设置 Content-Type 和 Content-Length
	req.Header.Set("Content-Type", form.FormDataContentType())
	req.Header.Set("Content-Length", strconv.Itoa(len(body.Bytes())))

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应内容
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 返回响应内容
	return respBody, nil
}
