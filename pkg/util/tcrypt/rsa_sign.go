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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

// ParsePriKEY 解析PEM格式的RSA私钥字符串
// 参数：
//   - priKeyStr: PEM格式的私钥字符串
//
// 返回值：
//   - *rsa.PrivateKey: 解析后的RSA私钥对象
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	pemStr := `-----BEGIN RSA PRIVATE KEY-----
//	MIIEpAIBAAKCAQEA...
//	-----END RSA PRIVATE KEY-----`
//	privateKey, err := tcrypt.ParsePriKEY(pemStr)
//	if err != nil {
//	    log.Printf("解析私钥失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 支持PKCS1和PKCS8格式的私钥
//   - 输入必须是完整的PEM格式字符串
//   - 包含BEGIN和END标记
//   - 私钥内容必须是Base64编码
//   - 用于数字签名和解密场景
func ParsePriKEY(priKeyStr string) (*rsa.PrivateKey, error) {

	block, _ := pem.Decode([]byte(priKeyStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}

	// 解析私钥内容
	if privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		if iprivKey, err := x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
			return nil, err
		} else {
			return iprivKey.(*rsa.PrivateKey), nil
		}
	} else {
		return privKey, nil
	}
}

// ParsePubKEY 解析PEM格式的RSA公钥字符串
// 参数：
//   - pubKeyStr: PEM格式的公钥字符串
//
// 返回值：
//   - *rsa.PublicKey: 解析后的RSA公钥对象
//   - error: 解析过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	pemStr := `-----BEGIN PUBLIC KEY-----
//	MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...
//	-----END PUBLIC KEY-----`
//	publicKey, err := tcrypt.ParsePubKEY(pemStr)
//	if err != nil {
//	    log.Printf("解析公钥失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 使用PKIX格式解析公钥
//   - 输入必须是完整的PEM格式字符串
//   - 包含BEGIN和END标记
//   - 公钥内容必须是Base64编码
//   - 用于加密和验证签名场景
func ParsePubKEY(pubKeyStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pubKeyStr))
	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return parsedKey.(*rsa.PublicKey), nil
}

// SignWithRSA 使用RSA私钥对数据进行签名
// 参数：
//   - message: 要签名的原始数据
//   - privateKey: RSA私钥对象
//   - hashAlg: 哈希算法（如crypto.SHA256、crypto.SHA512等）
//
// 返回值：
//   - []byte: 签名数据
//   - error: 签名过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	// 使用SHA256算法签名
//	message := "Hello World"
//	signature, err := tcrypt.SignWithRSA(message, privateKey, crypto.SHA256)
//	if err != nil {
//	    log.Printf("签名失败：%v", err)
//	    return
//	}
//	fmt.Printf("签名结果：%x\n", signature)
//
//	// 使用SHA512算法签名
//	signature, err = tcrypt.SignWithRSA(message, privateKey, crypto.SHA512)
//	if err != nil {
//	    log.Printf("签名失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 支持多种哈希算法：
//   - crypto.MD5（不推荐）
//   - crypto.SHA1（不推荐）
//   - crypto.SHA256（推荐）
//   - crypto.SHA512（更安全）
//   - 使用PKCS1v15填充方案
//   - 签名结果的长度与RSA密钥长度相同
//   - 建议使用SHA256或更强的哈希算法
//   - 确保选择的哈希算法已启用（hashAlg.Available()）
func SignWithRSA(message string, privateKey *rsa.PrivateKey, hashAlg crypto.Hash) ([]byte, error) {
	h := hashAlg.New()
	h.Write([]byte(message))
	hashed := h.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, hashAlg, hashed)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %v", err)
	}

	return signature, nil
}

// VerifyWithRSA 使用RSA公钥验证数字签名
// 参数：
//   - originalMessage: 原始消息
//   - publicKey: RSA公钥对象
//   - signature: 签名数据（由SignWithRSA生成）
//   - hashAlg: 哈希算法（必须与签名时使用的算法相同）
//
// 返回值：
//   - error: 验证结果，如果签名有效则为nil，否则返回错误
//
// 使用示例：
//
//	// 使用SHA256算法验证签名
//	message := "Hello World"
//	err := tcrypt.VerifyWithRSA(message, publicKey, signature, crypto.SHA256)
//	if err != nil {
//	    log.Printf("签名无效：%v", err)
//	    return
//	}
//	fmt.Println("签名验证成功")
//
//	// 使用SHA512算法验证签名
//	err = tcrypt.VerifyWithRSA(message, publicKey, signature, crypto.SHA512)
//	if err != nil {
//	    log.Printf("签名无效：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 必须使用与签名时相同的哈希算法
//   - 支持的哈希算法同SignWithRSA
//   - 验证失败不意味着发生错误，可能是签名确实无效
//   - 使用PKCS1v15填充方案
//   - 提供详细的错误信息便于调试
//   - 验证过程是时间恒定的，防止计时攻击
func VerifyWithRSA(originalMessage string, publicKey *rsa.PublicKey, signature []byte, hashAlg crypto.Hash) error {
	h := hashAlg.New()
	h.Write([]byte(originalMessage))
	hashed := h.Sum(nil)

	err := rsa.VerifyPKCS1v15(publicKey, hashAlg, hashed, signature)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %v", err)
	}

	return nil
}
