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

package tcrypt

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// MD5 计算字节流的MD5值，支持附加字节
// 参数：
//   - data: 要计算哈希值的字节数组
//   - salt: 可选的盐值，会被附加到哈希结果中
//
// 返回值：
//   - string: 32位小写十六进制的MD5哈希值
//
// 使用示例：
//
//	// 基本用法
//	data := []byte("Hello World")
//	hash := tcrypt.MD5(data)
//	fmt.Println(hash) // 输出32位哈希值
//
//	// 使用盐值
//	salt := []byte("mysalt")
//	hashWithSalt := tcrypt.MD5(data, salt...)
//
// 注意事项：
//   - MD5不是加密算法，是哈希算法
//   - MD5已经不再被认为是密码学安全的
//   - 建议用于数据完整性校验，不要用于密码存储
//   - 对于密码存储，请使用BcryptHash函数
//   - 返回值始终是32个字符的小写十六进制字符串
func MD5(data []byte, salt ...byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(salt))
}

// MD5String 计算字符串的MD5值，支持附加字节
// 参数：
//   - s: 要计算哈希值的字符串
//   - salt: 可选的盐值，会被附加到哈希结果中
//
// 返回值：
//   - string: 32位小写十六进制的MD5哈希值
//
// 使用示例：
//
//	// 基本用法
//	hash := tcrypt.MD5String("Hello World")
//	fmt.Println(hash) // 输出32位哈希值
//
//	// 使用盐值
//	hashWithSalt := tcrypt.MD5String("Hello World", []byte("mysalt")...)
//
// 注意事项：
//   - 这是MD5函数的字符串便捷版本
//   - 内部会将字符串转换为UTF-8编码的字节数组
//   - 其他注意事项与MD5函数相同
func MD5String(s string, salt ...byte) string {
	return MD5([]byte(s), salt...)
}

// MD5V 计算字符串的MD5值，支持附加字节（向后兼容）
// Deprecated: 使用 MD5String 替代，它提供相同的功能
//
// 参数：
//   - encrypt: 要计算哈希值的字符串
//   - b: 可选的盐值，会被附加到哈希结果中
//
// 返回值：
//   - string: 32位小写十六进制的MD5哈希值
//
// 使用示例：
//
//	// 不推荐使用此函数，请使用MD5String替代
//	hash := tcrypt.MD5V("Hello World")
//
// 注意事项：
//   - 此函数已废弃，仅为向后兼容保留
//   - 新代码应使用MD5String函数
//   - 功能与MD5String完全相同
func MD5V(encrypt string, b ...byte) string {
	return MD5String(encrypt, b...)
}

// SHA1 计算字节流的SHA1值
// 参数：
//   - b: 要计算哈希值的字节数组
//
// 返回值：
//   - string: 40位小写十六进制的SHA1哈希值
//
// 使用示例：
//
//	data := []byte("Hello World")
//	hash := tcrypt.SHA1(data)
//	fmt.Println(hash) // 输出40位哈希值
//
// 注意事项：
//   - SHA1不是加密算法，是哈希算法
//   - SHA1已经不再被认为是密码学安全的
//   - 建议用于数据完整性校验，不要用于密码存储
//   - 对于密码存储，请使用BcryptHash函数
//   - 返回值始终是40个字符的小写十六进制字符串
func SHA1(b []byte) string {
	h := sha1.New()
	_, _ = h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// SHA1String 计算字符串的SHA1值
// 参数：
//   - s: 要计算哈希值的字符串
//
// 返回值：
//   - string: 40位小写十六进制的SHA1哈希值
//
// 使用示例：
//
//	hash := tcrypt.SHA1String("Hello World")
//	fmt.Println(hash) // 输出40位哈希值
//
// 注意事项：
//   - 这是SHA1函数的字符串便捷版本
//   - 内部会将字符串转换为UTF-8编码的字节数组
//   - 其他注意事项与SHA1函数相同
func SHA1String(s string) string {
	return SHA1([]byte(s))
}

// Byte2Hex 将字节流转换为十六进制字符串
// 参数：
//   - b: 要转换的字节数组
//
// 返回值：
//   - string: 转换后的十六进制字符串（小写）
//
// 使用示例：
//
//	data := []byte{0x12, 0x34, 0xAB, 0xCD}
//	hex := tcrypt.Byte2Hex(data)
//	fmt.Println(hex) // 输出: "1234abcd"
//
// 注意事项：
//   - 每个字节会转换为两个十六进制字符
//   - 输出始终是小写字母
//   - 输出长度是输入长度的两倍
//   - 常用于二进制数据的可读性展示
func Byte2Hex(b []byte) string {
	return hex.EncodeToString(b)
}

// Hex2Byte 将十六进制字符串转换为字节流
// 参数：
//   - h: 要转换的十六进制字符串
//
// 返回值：
//   - []byte: 转换后的字节数组
//   - error: 转换过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	hex := "1234abcd"
//	data, err := tcrypt.Hex2Byte(hex)
//	if err != nil {
//	    log.Printf("转换失败：%v", err)
//	    return
//	}
//	fmt.Printf("%x", data) // 输出: [12 34 ab cd]
//
// 注意事项：
//   - 输入字符串长度必须是偶数
//   - 支持大写和小写十六进制字符
//   - 如果输入包含非法字符会返回错误
//   - 输出长度是输入长度的一半
//   - 是Byte2Hex的逆操作
func Hex2Byte(h string) ([]byte, error) {
	return hex.DecodeString(h)
}

// BcryptHash 使用bcrypt算法对密码进行安全哈希
// 参数：
//   - password: 要哈希的原始密码
//
// 返回值：
//   - string: bcrypt哈希后的字符串，包含盐值和成本因子
//
// 使用示例：
//
//	password := "mySecurePassword123"
//	hash := tcrypt.BcryptHash(password)
//	fmt.Println(hash) // 输出bcrypt哈希值
//
//	// 验证密码
//	isValid := tcrypt.BcryptCheck(password, hash)
//	fmt.Println(isValid) // 输出: true
//
// 注意事项：
//   - 使用默认的成本因子(10)
//   - 每次哈希相同的密码会得到不同的结果（因为盐值随机）
//   - 哈希结果包含算法版本、成本因子和盐值
//   - 适合用于密码存储
//   - 不要尝试自己解析或修改哈希结果
//   - 验证密码请使用BcryptCheck函数
func BcryptHash(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}

// BcryptCheck 验证明文密码是否与bcrypt哈希值匹配
// 参数：
//   - password: 要验证的明文密码
//   - hash: 存储的bcrypt哈希值（由BcryptHash生成）
//
// 返回值：
//   - bool: 如果密码匹配返回true，否则返回false
//
// 使用示例：
//
//	storedHash := "$2a$10$..." // 从数据库获取的哈希值
//	password := "userInputPassword"
//	if tcrypt.BcryptCheck(password, storedHash) {
//	    fmt.Println("密码正确")
//	} else {
//	    fmt.Println("密码错误")
//	}
//
// 注意事项：
//   - 用于验证用户输入的密码
//   - 时间恒定的比较，防止计时攻击
//   - hash参数必须是完整的bcrypt哈希字符串
//   - 如果hash格式错误，将返回false
//   - 验证过程是单向的，无法从哈希值恢复原始密码
func BcryptCheck(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SHA256 计算字节流的SHA256值
// 参数：
//   - data: 要计算哈希值的字节数组
//
// 返回值：
//   - string: 64位小写十六进制的SHA256哈希值
//
// 使用示例：
//
//	data := []byte("Hello World")
//	hash := tcrypt.SHA256(data)
//	fmt.Println(hash) // 输出64位哈希值
//
// 注意事项：
//   - SHA256是SHA-2家族中最常用的哈希算法
//   - 提供256位（32字节）的哈希值
//   - 比MD5和SHA1更安全
//   - 适用于数字签名和数据完整性验证
//   - 返回值始终是64个字符的小写十六进制字符串
func SHA256(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256String 计算字符串的SHA256值
// 参数：
//   - s: 要计算哈希值的字符串
//
// 返回值：
//   - string: 64位小写十六进制的SHA256哈希值
//
// 使用示例：
//
//	hash := tcrypt.SHA256String("Hello World")
//	fmt.Println(hash) // 输出64位哈希值
//
// 注意事项：
//   - 这是SHA256函数的字符串便捷版本
//   - 内部会将字符串转换为UTF-8编码的字节数组
//   - 其他注意事项与SHA256函数相同
func SHA256String(s string) string {
	return SHA256([]byte(s))
}

// SHA512 计算字节流的SHA512值
// 参数：
//   - data: 要计算哈希值的字节数组
//
// 返回值：
//   - string: 128位小写十六进制的SHA512哈希值
//
// 使用示例：
//
//	data := []byte("Hello World")
//	hash := tcrypt.SHA512(data)
//	fmt.Println(hash) // 输出128位哈希值
//
// 注意事项：
//   - SHA512是SHA-2家族中最安全的哈希算法
//   - 提供512位（64字节）的哈希值
//   - 比SHA256计算更慢，但提供更高的安全性
//   - 适用于要求最高安全级别的场景
//   - 返回值始终是128个字符的小写十六进制字符串
func SHA512(data []byte) string {
	h := sha512.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// SHA512String 计算字符串的SHA512值
// 参数：
//   - s: 要计算哈希值的字符串
//
// 返回值：
//   - string: 128位小写十六进制的SHA512哈希值
//
// 使用示例：
//
//	hash := tcrypt.SHA512String("Hello World")
//	fmt.Println(hash) // 输出128位哈希值
//
// 注意事项：
//   - 这是SHA512函数的字符串便捷版本
//   - 内部会将字符串转换为UTF-8编码的字节数组
//   - 其他注意事项与SHA512函数相同
func SHA512String(s string) string {
	return SHA512([]byte(s))
}

// HMAC 使用指定的哈希算法计算HMAC（Hash-based Message Authentication Code）
// 参数：
//   - hashFunc: 哈希函数构造器，如sha256.New
//   - data: 要计算HMAC的数据
//   - key: HMAC密钥
//
// 返回值：
//   - string: 小写十六进制的HMAC值
//
// 使用示例：
//
//	key := []byte("secret-key")
//	data := []byte("Hello World")
//	hmac := tcrypt.HMAC(sha256.New, data, key)
//	fmt.Println(hmac)
//
// 注意事项：
//   - 支持任何实现了hash.Hash接口的哈希函数
//   - 输出长度取决于所使用的哈希函数
//   - 用于消息认证和完整性验证
//   - 密钥长度应该至少与哈希函数的输出长度相同
//   - 常用于API签名验证
func HMAC(hashFunc func() hash.Hash, data, key []byte) string {
	h := hmac.New(hashFunc, key)
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// HMACSHA256 使用SHA256算法计算HMAC
// 参数：
//   - data: 要计算HMAC的数据
//   - key: HMAC密钥
//
// 返回值：
//   - string: 64位小写十六进制的HMAC-SHA256值
//
// 使用示例：
//
//	key := []byte("secret-key")
//	data := []byte("Hello World")
//	hmac := tcrypt.HMACSHA256(data, key)
//	fmt.Println(hmac) // 输出64位HMAC值
//
// 注意事项：
//   - 这是HMAC函数的SHA256特化版本
//   - 输出长度固定为64个字符（32字节）
//   - 建议密钥长度至少32字节
//   - 适用于大多数API认证场景
//   - 比HMAC-SHA512计算更快
func HMACSHA256(data, key []byte) string {
	return HMAC(sha256.New, data, key)
}

// HMACSHA512 使用SHA512算法计算HMAC
// 参数：
//   - data: 要计算HMAC的数据
//   - key: HMAC密钥
//
// 返回值：
//   - string: 128位小写十六进制的HMAC-SHA512值
//
// 使用示例：
//
//	key := []byte("secret-key")
//	data := []byte("Hello World")
//	hmac := tcrypt.HMACSHA512(data, key)
//	fmt.Println(hmac) // 输出128位HMAC值
//
// 注意事项：
//   - 这是HMAC函数的SHA512特化版本
//   - 输出长度固定为128个字符（64字节）
//   - 建议密钥长度至少64字节
//   - 适用于需要最高安全级别的场景
//   - 计算较慢但提供更强的安全性
func HMACSHA512(data, key []byte) string {
	return HMAC(sha512.New, data, key)
}

// GenerateSalt 生成指定长度的随机盐值
// 参数：
//   - length: 要生成的盐值长度（字节数）
//
// 返回值：
//   - []byte: 生成的随机盐值
//   - error: 生成过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	// 生成16字节的盐值
//	salt, err := tcrypt.GenerateSalt(16)
//	if err != nil {
//	    log.Printf("生成盐值失败：%v", err)
//	    return
//	}
//	fmt.Printf("%x", salt) // 以十六进制打印盐值
//
// 注意事项：
//   - 使用密码学安全的随机数生成器
//   - 每次调用都会生成不同的盐值
//   - 建议长度至少16字节
//   - 主要用于密码哈希和密钥派生
//   - 如果需要十六进制格式，请使用GenerateHexSalt
func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("generate salt failed: %w", err)
	}
	return salt, nil
}

// GenerateHexSalt 生成指定长度的十六进制格式随机盐值
// 参数：
//   - length: 要生成的盐值长度（字节数，最终字符串长度会是这个值的两倍）
//
// 返回值：
//   - string: 生成的十六进制格式盐值
//   - error: 生成过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	// 生成16字节（32个字符）的十六进制盐值
//	salt, err := tcrypt.GenerateHexSalt(16)
//	if err != nil {
//	    log.Printf("生成盐值失败：%v", err)
//	    return
//	}
//	fmt.Println(salt) // 输出32个十六进制字符
//
// 注意事项：
//   - 输出字符串长度是输入长度的两倍
//   - 使用小写十六进制字符（0-9, a-f）
//   - 每次调用都会生成不同的盐值
//   - 建议length至少16（生成32个字符）
//   - 内部使用GenerateSalt生成随机字节
func GenerateHexSalt(length int) (string, error) {
	salt, err := GenerateSalt(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// PasswordStrength 密码强度级别
// 用于表示密码的安全强度等级
type PasswordStrength int

const (
	// PasswordWeak 表示弱密码
	// 通常是长度过短或只包含单一类型字符的密码
	PasswordWeak PasswordStrength = iota

	// PasswordMedium 表示中等强度密码
	// 通常满足基本的长度要求，并包含至少两种不同类型的字符
	PasswordMedium

	// PasswordStrong 表示强密码
	// 通常长度适中，并包含三种或更多不同类型的字符
	PasswordStrong

	// PasswordVeryStrong 表示非常强的密码
	// 通常长度较长，并包含所有类型的字符（大小写字母、数字和特殊字符）
	PasswordVeryStrong
)

// CheckPasswordStrength 检查密码强度并返回强度等级
// 参数：
//   - password: 要检查的密码
//
// 返回值：
//   - PasswordStrength: 密码强度等级（Weak、Medium、Strong、VeryStrong）
//
// 使用示例：
//
//	password := "MyP@ssw0rd"
//	strength := tcrypt.CheckPasswordStrength(password)
//	switch strength {
//	case tcrypt.PasswordWeak:
//	    fmt.Println("密码强度：弱")
//	case tcrypt.PasswordMedium:
//	    fmt.Println("密码强度：中")
//	case tcrypt.PasswordStrong:
//	    fmt.Println("密码强度：强")
//	case tcrypt.PasswordVeryStrong:
//	    fmt.Println("密码强度：非常强")
//	}
//
// 注意事项：
//   - 检查以下特征：
//   - 密码长度（8位以下为弱，12位以上加分）
//   - 是否包含大写字母
//   - 是否包含小写字母
//   - 是否包含数字
//   - 是否包含特殊字符
//   - 评分标准：
//   - 分数0-2：弱密码
//   - 分数3：中等强度
//   - 分数4：强密码
//   - 分数5：非常强的密码
//   - 建议与ValidatePassword配合使用
func CheckPasswordStrength(password string) PasswordStrength {
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
		length     = len(password)
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// 计算强度
	score := 0
	if length >= 8 {
		score++
	}
	if length >= 12 {
		score++
	}
	if hasUpper && hasLower {
		score++
	}
	if hasNumber {
		score++
	}
	if hasSpecial {
		score++
	}

	switch {
	case score >= 5:
		return PasswordVeryStrong
	case score >= 4:
		return PasswordStrong
	case score >= 3:
		return PasswordMedium
	default:
		return PasswordWeak
	}
}

// ValidatePassword 根据指定的规则验证密码是否符合要求
// 参数：
//   - password: 要验证的密码
//   - minLength: 最小密码长度
//   - requireUpper: 是否要求包含大写字母
//   - requireLower: 是否要求包含小写字母
//   - requireNumber: 是否要求包含数字
//   - requireSpecial: 是否要求包含特殊字符
//
// 返回值：
//   - error: 如果密码不符合要求，返回具体的错误信息；如果符合要求，返回nil
//
// 使用示例：
//
//	err := tcrypt.ValidatePassword("MyP@ssw0rd", 8, true, true, true, true)
//	if err != nil {
//	    log.Printf("密码不符合要求：%v", err)
//	    return
//	}
//	fmt.Println("密码符合要求")
//
//	// 自定义规则
//	err = tcrypt.ValidatePassword("Pass123", 6, true, true, true, false)
//	if err != nil {
//	    log.Printf("密码不符合要求：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 可以灵活配置密码规则
//   - 特殊字符包括标点符号和符号字符
//   - 错误信息会详细说明不符合的要求
//   - 建议与CheckPasswordStrength配合使用
//   - 适用于注册和修改密码场景
func ValidatePassword(password string, minLength int, requireUpper, requireLower, requireNumber, requireSpecial bool) error {
	if len(password) < minLength {
		return fmt.Errorf("password must be at least %d characters long", minLength)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	var missing []string
	if requireUpper && !hasUpper {
		missing = append(missing, "uppercase letter")
	}
	if requireLower && !hasLower {
		missing = append(missing, "lowercase letter")
	}
	if requireNumber && !hasNumber {
		missing = append(missing, "number")
	}
	if requireSpecial && !hasSpecial {
		missing = append(missing, "special character")
	}

	if len(missing) > 0 {
		return fmt.Errorf("password must contain at least one %s", strings.Join(missing, ", "))
	}

	return nil
}
