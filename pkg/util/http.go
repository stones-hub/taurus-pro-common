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
	"net/url"
)

// doHttpRequest is a generic HTTP request function that handles different types of requests
// Parameters:
//   - method: HTTP method (GET, POST, etc.)
//   - url: Target URL
//   - payload: Request payload (can be map, string, or any JSON-serializable type)
//   - headers: HTTP headers to be set
//
// Returns:
//   - []byte: Response body
//   - error: Any error that occurred during the request
func doHttpRequest(method, url string, payload interface{}, headers map[string]string) ([]byte, error) {
	var (
		err         error
		request     *http.Request
		response    *http.Response
		body        []byte
		jsonPayload []byte
	)

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
	if response, err = http.DefaultClient.Do(request); err != nil {
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
//
// Returns:
//   - []byte: Response body
//   - error: Any error that occurred during the request
func HttpPost(url string, payload interface{}, headers map[string]string) ([]byte, error) {
	return doHttpRequest("POST", url, payload, headers)
}

// HttpGet is a wrapper for making GET requests
// Parameters:
//   - url: Target URL
//   - headers: HTTP headers to be set
//
// Returns:
//   - []byte: Response body
//   - error: Any error that occurred during the request
func HttpGet(url string, headers map[string]string) ([]byte, error) {
	return doHttpRequest("GET", url, nil, headers)
}

// HttpRequest is a comprehensive HTTP client wrapper that supports various request types
// Parameters:
//   - URL: Target URL
//   - method: HTTP method (GET, POST, etc.)
//   - headers: HTTP headers to be set
//   - params: URL query parameters
//   - data: Request body data (will be JSON encoded)
//
// Returns:
//   - *http.Response: HTTP response object
//   - error: Any error that occurred during the request
func HttpRequest(URL string, method string, headers map[string]string, params map[string]string, data any) (*http.Response, error) {
	var (
		err      error
		u        *url.URL
		query    url.Values
		body     = &bytes.Buffer{} // Set body data
		dataJson []byte
		req      *http.Request
		resp     *http.Response
	)
	// Create URL
	u, err = url.Parse(URL)
	if err != nil {
		return nil, err
	}

	// Add query parameters
	query = u.Query()
	for k, v := range params {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()

	// Encode data as JSON
	if data != nil {
		dataJson, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
		//  fmt.Println("http send data:", string(bodyData))
		body = bytes.NewBuffer(dataJson)
	}

	// Create request
	req, err = http.NewRequest(method, u.String(), body)

	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Send request
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// Return response for caller to handle
	return resp, nil
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
