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
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

// GenCSV writes data to a CSV file, data does not include headers
// Parameters:
//   - filename: CSV file path
//   - data: Data to be written
//   - headers: CSV header data, for scenarios requiring fixed header positions
func GenCSV(filename string, data []map[string]string, headers []string) error {
	var (
		fd     *os.File
		err    error
		writer *csv.Writer
	)

	if fd, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm); err != nil {
		log.Printf("Failed to generate CSV file: %s \n", err.Error())
		return err
	}

	defer fd.Close()

	writer = csv.NewWriter(fd)

	defer writer.Flush()

	if len(data) == 0 {
		return fmt.Errorf("data is empty")
	}

	if len(headers) == 0 {
		// Create a slice with header size
		headers = make([]string, 0, len(data[0]))
		for header := range data[0] {
			headers = append(headers, header)
		}
	}

	// Write headers
	if err = writer.Write(headers); err != nil {
		log.Printf("Failed to write headers: %s \n", err.Error())
		return err
	}

	// Write data
	for _, row := range data {
		record := make([]string, 0, len(headers))

		for _, header := range headers {
			record = append(record, row[header])
		}

		// log.Printf("Writing CSV data: %v \n", record)

		writer.Write(record)
	}

	return nil
}

/*
type Test struct {
	Name string `csv:"name" json:"name"`
	Age  int    `csv:"age" json:"age"`
}
*/
// ReadCSV reads data from a CSV file
// Parameters:
//   - filename: CSV file path
//   - result: Pointer to a slice of structs, e.g., &[]Test{}, cannot be a pointer to struct
func ReadCSV(fileanme string, result interface{}) error {
	var (
		fd               *os.File
		err              error
		reader           *csv.Reader
		invisibleHeaders []string
		headers          []string
		resultValue      reflect.Value
	)

	if fd, err = os.OpenFile(fileanme, os.O_RDONLY|os.O_APPEND, os.ModePerm); err != nil {
		return err
	}

	defer fd.Close()
	reader = csv.NewReader(fd)

	// Read the first line of CSV file as headers
	if invisibleHeaders, err = reader.Read(); err != nil {
		return err
	}

	// Filter out invisible characters from headers due to potential encoding issues
	for _, v := range invisibleHeaders {
		resRunes := []rune{}
		for _, r := range v {
			// ASCII codes less than or equal to 32 or greater than or equal to 127 are invisible characters
			if r > 32 && r < 127 {
				resRunes = append(resRunes, r)
			}
		}
		// Print ASCII encoding
		// fmt.Println(v, resRunes)
		headers = append(headers, string(resRunes))
	}

	// Get the reflection type of result
	resultValue = reflect.ValueOf(result)

	// Result must be a pointer to a slice
	if resultValue.Kind() != reflect.Ptr || resultValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result must be a pointer to a slice")
	}

	// Get the type of the slice that result points to: sliceType = []Test
	sliceType := resultValue.Elem().Type()

	// Get the type of elements in the slice: elementType = Test
	elementType := sliceType.Elem()

	// Create a new slice based on the type: []Test
	slice := reflect.MakeSlice(sliceType, 0, 0)

	// Read data rows
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Failed to read data row: %v\n", err)
			continue
		}

		// Create new struct instance
		element := reflect.New(elementType).Elem()

		// Iterate through struct fields
		for i := 0; i < element.NumField(); i++ {
			field := element.Type().Field(i)
			tag := field.Tag.Get("csv")
			if tag == "" {
				continue
			}

			// Find corresponding CSV column index
			colIndex := -1
			for j, header := range headers {
				if header == tag {
					colIndex = j
					break
				}
			}

			if colIndex == -1 || colIndex >= len(record) {
				continue
			}

			// Set field value
			fieldValue := element.Field(i)
			value := record[colIndex]

			switch fieldValue.Kind() {
			case reflect.String:
				fieldValue.SetString(value)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if v, err := strconv.ParseInt(value, 10, 64); err == nil {
					fieldValue.SetInt(v)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if v, err := strconv.ParseUint(value, 10, 64); err == nil {
					fieldValue.SetUint(v)
				}
			case reflect.Float32, reflect.Float64:
				if v, err := strconv.ParseFloat(value, 64); err == nil {
					fieldValue.SetFloat(v)
				}
			case reflect.Bool:
				if v, err := strconv.ParseBool(value); err == nil {
					fieldValue.SetBool(v)
				}
			case reflect.Slice:
				// Handle []byte type
				if fieldValue.Type().Elem().Kind() == reflect.Uint8 {
					fieldValue.SetBytes([]byte(value))
				}
			default:
				// Handle time.Time type
				if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
					if v, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
						fieldValue.Set(reflect.ValueOf(v))
					}
				} else if fieldValue.Type().Kind() == reflect.Interface {
					// Handle interface{} type
					fieldValue.Set(reflect.ValueOf(value))
				}
			}
		}

		// Add struct to slice
		slice = reflect.Append(slice, element)
	}

	// Set result
	resultValue.Elem().Set(slice)
	return nil
}
