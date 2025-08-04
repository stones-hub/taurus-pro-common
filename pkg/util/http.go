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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// DefaultTimeout 默认超时时间
const DefaultTimeout = 30 * time.Second

// doHttpRequest is a generic HTTP request function that handles different types of requests
// Parameters:
//   - method: HTTP method (GET, POST, etc.)
//   - url: Target URL
//   - payload: Request payload (can be map, string, or any JSON-serializable type)
//   - headers: HTTP headers to be set
//   - timeout: Request timeout (optional, uses DefaultTimeout if not specified)
//
// Returns:
//   - []byte: Response body
//   - error: Any error that occurred during the request
func doHttpRequest(method, url string, payload interface{}, headers map[string]string, timeout time.Duration) ([]byte, error) {
	var (
		err         error
		request     *http.Request
		response    *http.Response
		body        []byte
		jsonPayload []byte
	)

	// 设置默认超时时间
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	// 创建带超时的HTTP客户端
	client := &http.Client{
		Timeout: timeout,
	}

	// Process payload if it's a map or string type
	if payload != nil {
		switch p := payload.(type) {
		case []byte:
			jsonPayload = p
		case string:
			jsonPayload = []byte(p) // Convert string to byte array
		default:
			if jsonPayload, err = json.Marshal(payload); err != nil {
				return nil, err
			}
		}
	}

	// Create HTTP request
	if request, err = http.NewRequest(method, url, bytes.NewBuffer(jsonPayload)); err != nil {
		return nil, err
	}

	// Set request headers
	for k, v := range headers {
		request.Header.Set(k, v)
	}

	if request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", "application/json")
	}

	// Send request
	if response, err = client.Do(request); err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Read response body
	if body, err = io.ReadAll(response.Body); err != nil {
		return nil, err
	}

	return body, nil
}

// HttpPost is a wrapper for making POST requests
// Parameters:
//   - url: Target URL
//   - payload: Request payload
//   - headers: HTTP headers to be set
//   - timeout: Request timeout (optional, uses DefaultTimeout if not specified)
//
// Returns:
//   - []byte: Response body
//   - error: Any error that occurred during the request
func HttpPost(url string, payload interface{}, headers map[string]string, timeout time.Duration) ([]byte, error) {
	return doHttpRequest("POST", url, payload, headers, timeout)
}

// HttpPostWithDefaultTimeout is a convenience function for POST requests with default timeout
// Parameters:
//   - url: Target URL
//   - payload: Request payload
//   - headers: HTTP headers to be set
//
// Returns:
//   - []byte: Response body
//   - error: Any error that occurred during the request
func HttpPostWithDefaultTimeout(url string, payload interface{}, headers map[string]string) ([]byte, error) {
	return doHttpRequest("POST", url, payload, headers, DefaultTimeout)
}

// HttpGet is a wrapper for making GET requests
// Parameters:
//   - url: Target URL
//   - headers: HTTP headers to be set
//   - timeout: Request timeout (optional, uses DefaultTimeout if not specified)
//
// Returns:
//   - []byte: Response body
//   - error: Any error that occurred during the request
func HttpGet(url string, headers map[string]string, timeout time.Duration) ([]byte, error) {
	return doHttpRequest("GET", url, nil, headers, timeout)
}

// HttpGetWithDefaultTimeout is a convenience function for GET requests with default timeout
// Parameters:
//   - url: Target URL
//   - headers: HTTP headers to be set
//
// Returns:
//   - []byte: Response body
//   - error: Any error that occurred during the request
func HttpGetWithDefaultTimeout(url string, headers map[string]string) ([]byte, error) {
	return doHttpRequest("GET", url, nil, headers, DefaultTimeout)
}

// ReadResponse reads and returns the body of an HTTP response
// Parameters:
//   - res: HTTP response object
//
// Returns:
//   - []byte: Response body
//   - error: Any error that occurred while reading the response
func ReadResponse(res *http.Response) ([]byte, error) {
	return io.ReadAll(res.Body)
}
