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

package util

import (
	"encoding/xml"
	"os"
)

/*
XML文件读取, filePath 文件地址, v 需要映射的结构体指针
*/

func ReadFromXml(filePath string, v interface{}) error {

	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()
	decodeXML := xml.NewDecoder(fd)

	if err := decodeXML.Decode(v); err != nil {
		return err
	}

	return nil
}

// XML文件生成， v 需要写入xml文件的结构体指针， filePath 文件地址

func WriteToXml(v interface{}, filePath string) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()
	return xml.NewEncoder(fd).Encode(v)
}
