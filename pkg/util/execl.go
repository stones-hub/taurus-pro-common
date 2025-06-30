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
	"encoding/csv"
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"time"
)

// ExcelWriter provides functionality to write data to Excel files in CSV format.
// It supports writing structured data with automatic type conversion.
type ExcelWriter struct {
	file    *os.File
	writer  *csv.Writer
	headers []string
}

// InitExcelWriter creates a new ExcelWriter instance.
// It initializes the CSV file with the provided headers.
//
// Parameters:
//   - filename: The path to the output CSV file
//   - headers: The column headers for the CSV file
//
// Returns:
//   - *ExcelWriter: A new ExcelWriter instance
//   - error: Any error that occurred during initialization
func InitExcelWriter(filename string, headers []string) (*ExcelWriter, error) {
	var (
		err    error
		file   *os.File
		writer *csv.Writer
	)

	// 判断目录是否存在
	file, err = os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer = csv.NewWriter(file)
	//  初始化的时候，先将头写到文件
	if err = writer.Write(headers); err != nil {
		_ = file.Close()
		return nil, err
	}

	return &ExcelWriter{
		file:    file,
		writer:  writer,
		headers: headers,
	}, nil
}

// WriteBatch writes a batch of data to the CSV file.
// It automatically converts different data types to strings.
//
// Parameters:
//   - datas: A slice of interface{} containing the data to write
//
// Returns:
//   - error: Any error that occurred during writing
func (excelWriter *ExcelWriter) WriteBatch(datas []interface{}) error {

	for _, record := range datas {

		// 以excel列头为长度，创建能存储一行数据的slice
		row := make([]string, len(excelWriter.headers))
		// 反射一行数据的结构体对象 {"":}
		rowVal := reflect.ValueOf(record)

		//  Excel头(header)的值和行结构体数据的字段名(KEY)是设置的一样的，否则下面没有办法通过字段名那到数据值
		for i, header := range excelWriter.headers {
			// 根据字段名称，获取字段名称存储的值
			fieldVal := rowVal.FieldByName(header)

			if fieldVal.IsValid() {

				switch fieldVal.Kind() { // 判断字段对应的值的类型

				case reflect.String:
					row[i] = fieldVal.String()

				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					row[i] = strconv.FormatInt(fieldVal.Int(), 10)

				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					row[i] = strconv.FormatUint(fieldVal.Uint(), 10)

				case reflect.Float32, reflect.Float64:
					row[i] = strconv.FormatFloat(fieldVal.Float(), 'f', -1, 64)

				case reflect.Bool:
					row[i] = strconv.FormatBool(fieldVal.Bool())

				case reflect.Struct:
					if fieldVal.Type() == reflect.TypeOf(time.Now()) { // 字段类型是时间结构体time.Time
						// 这个写法很有意思， 先将fieldVal转成interface在转成time.Time
						row[i] = fieldVal.Interface().(time.Time).Format("2006-01-02 15:04:05")
					} else { // 如果不是time.Time结构体类型，统一json成字符串在写入到Excel
						jsonBytes, _ := json.Marshal(fieldVal.Interface())
						row[i] = string(jsonBytes)
					}
				default:
					jsonBytes, _ := json.Marshal(fieldVal.Interface())
					row[i] = string(jsonBytes)
				}
			} else {
				row[i] = ""
			}
		}

		if err := excelWriter.writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// Close flushes any buffered data and closes the CSV file.
//
// Returns:
//   - error: Any error that occurred during closing
func (excelWriter *ExcelWriter) Close() error {
	excelWriter.writer.Flush()
	if err := excelWriter.writer.Error(); err != nil {
		return err
	}
	return excelWriter.file.Close()
}
