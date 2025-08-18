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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// GenerateKeyPair 生成RSA公私钥对并保存到文件
// 参数：
//   - pub: 公钥文件保存路径
//   - pri: 私钥文件保存路径
//   - bits: 密钥长度（通常是1024或2048位）
//
// 返回值：
//   - error: 生成过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	err := tcrypt.GenerateKeyPair(
//	    "public.pem",  // 公钥文件路径
//	    "private.pem", // 私钥文件路径
//	    2048,          // 2048位密钥长度
//	)
//	if err != nil {
//	    log.Printf("生成密钥对失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 建议使用2048位或更长的密钥长度
//   - 1024位密钥在现代已经不够安全
//   - 公钥文件使用PKIX格式（PEM编码）
//   - 私钥文件使用PKCS1格式（PEM编码）
//   - 确保文件路径有写入权限
//   - 私钥文件应该有适当的访问权限限制
func GenerateKeyPair(pub string, pri string, bits int) error {
	priKey, err := rsa.GenerateKey(rand.Reader, bits) // 私钥
	if err != nil {
		return err
	}
	pubKey := &priKey.PublicKey // 公钥

	if createPub(pubKey, pub) != nil || createPri(priKey, pri) != nil {
		return errors.New("公私钥内容写入文件失败")
	}

	return nil
}

// createPub 将RSA公钥保存到PEM格式文件
// 参数：
//   - pubKey: RSA公钥
//   - filename: 保存文件的路径
//
// 返回值：
//   - error: 保存过程中的错误，如果成功则为nil
//
// 注意事项：
//   - 使用PKIX格式编码公钥
//   - 使用PEM格式封装
//   - PEM块类型为"PUBLIC KEY"
//   - 文件已存在会被覆盖
func createPub(pubKey *rsa.PublicKey, filename string) error {
	pubBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return err
	}
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil

}

// createPri 将RSA私钥保存到PEM格式文件
// 参数：
//   - priKey: RSA私钥
//   - filename: 保存文件的路径
//
// 返回值：
//   - error: 保存过程中的错误，如果成功则为nil
//
// 注意事项：
//   - 使用PKCS1格式编码私钥
//   - 使用PEM格式封装
//   - PEM块类型为"RSA PRIVATE KEY"
//   - 文件已存在会被覆盖
//   - 应该设置适当的文件权限
func createPri(priKey *rsa.PrivateKey, filename string) error {
	privDER := x509.MarshalPKCS1PrivateKey(priKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = pem.Encode(file, block)
	if err != nil {
		return err
	}
	return nil
}

// EncryptByPublicKey 使用RSA公钥加密数据
// 参数：
//   - public: 公钥文件路径（PEM格式）
//   - plaintext: 要加密的数据
//
// 返回值：
//   - []byte: 加密后的数据
//   - error: 加密过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	data := []byte("Hello World")
//	encrypted, err := tcrypt.EncryptByPublicKey("public.pem", data)
//	if err != nil {
//	    log.Printf("加密失败：%v", err)
//	    return
//	}
//	fmt.Printf("加密结果：%x\n", encrypted)
//
// 注意事项：
//   - 使用PKCS1v15填充方案
//   - 明文长度不能超过密钥长度减去11字节
//   - 支持PKIX格式的公钥文件
//   - 每次加密相同的数据得到的结果都不同
//   - 加密结果可以用对应的私钥解密
//   - 适用于加密小块数据（如密钥）
func EncryptByPublicKey(public string, plaintext []byte) ([]byte, error) {

	pemBlock, err := os.ReadFile(public)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemBlock)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)

	if err != nil {
		return nil, err
	}

	// 公钥加密
	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey.(*rsa.PublicKey), plaintext)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// DecryptByPrivateKey 使用RSA私钥解密数据
// 参数：
//   - private: 私钥文件路径（PEM格式）
//   - ciphertext: 要解密的数据（由EncryptByPublicKey加密）
//
// 返回值：
//   - []byte: 解密后的原始数据
//   - error: 解密过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	decrypted, err := tcrypt.DecryptByPrivateKey("private.pem", encrypted)
//	if err != nil {
//	    log.Printf("解密失败：%v", err)
//	    return
//	}
//	fmt.Printf("解密结果：%s\n", string(decrypted))
//
// 注意事项：
//   - 使用PKCS1v15填充方案
//   - 支持PKCS1和PKCS8格式的私钥文件
//   - 只能解密使用对应公钥加密的数据
//   - 解密失败可能意味着数据损坏或使用了错误的私钥
//   - 提供详细的错误信息便于调试
//   - 私钥文件需要妥善保管
func DecryptByPrivateKey(private string, ciphertext []byte) ([]byte, error) {
	// 读取私钥文件
	pemBlock, err := os.ReadFile(private)
	if err != nil {
		return nil, fmt.Errorf("read private key file failed: %w", err)
	}

	// 解码 PEM 格式
	block, _ := pem.Decode(pemBlock)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}

	// 尝试解析为 PKCS1 格式私钥
	var privateKey *rsa.PrivateKey
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		privateKey = key
	} else {
		// 如果 PKCS1 解析失败，尝试解析为 PKCS8 格式
		if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			if rsaKey, ok := key.(*rsa.PrivateKey); ok {
				privateKey = rsaKey
			} else {
				return nil, errors.New("private key is not RSA key")
			}
		} else {
			return nil, fmt.Errorf("parse private key failed: %w", err)
		}
	}

	// 使用私钥解密
	plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt failed: %w", err)
	}

	return plaintext, nil
}
