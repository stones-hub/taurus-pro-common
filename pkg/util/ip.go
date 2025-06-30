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
	"log"
	"net"
	"net/http"
	"strings"
)

func GetLocalIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, addr := range addrs {
		// 检查是否为 IP net.Addr 类型
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		// 获取 IP 地址
		ip := ipNet.IP

		// 排除回环地址
		if ip.IsLoopback() {
			continue
		}

		// 添加 IP 地址
		if ip.To4() != nil || ip.To16() != nil {
			ips = append(ips, ip.String())
		}
	}

	return ips, nil
}

/*
- 192.0.0.0/8 192.168.0.0/16 192.168.1.0/24
- /24：子网掩码为255.255.255.0，表示前24位是网络部分，后8位是主机部分。
- /16：子网掩码为255.255.0.0，表示前16位是网络部分，后16位是主机部分。
- /8：子网掩码为255.0.0.0，表示前8位是网络部分，后24位是主机部分。
*/
// isIPAllowed 检查IP是否在允许的网段中
func IsIPAllowed(ip string, allowedHosts []string) bool {
	for _, host := range allowedHosts {
		// 检查是否为CIDR格式
		if strings.Contains(host, "/") {
			_, ipNet, err := net.ParseCIDR(host)
			if err != nil {
				continue
			}
			if ipNet.Contains(net.ParseIP(ip)) {
				return true
			}
		} else {
			// 直接比较IP
			if ip == host {
				return true
			}
		}
	}
	return false
}

// GetRemoteIP 获取远程IP
func GetRemoteIP(r *http.Request) []string {
	var ips []string
	// 提取X-Forwarded-For
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips = strings.Split(xForwardedFor, ", ")
	}

	// 提取X-Real-Ip
	xRealIP := r.Header.Get("X-Real-Ip")
	if xRealIP != "" {
		ips = append(ips, xRealIP)
	}

	// 提取RemoteAddr
	remoteAddr := r.RemoteAddr
	parts := strings.Split(remoteAddr, ":")
	if len(parts) > 0 {
		ips = append(ips, parts[0])
	} else {
		log.Printf("RemoteAddr 格式错误 : %v\n", remoteAddr)
	}
	return ips
}
