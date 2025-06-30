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

package secure

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

// 创建密钥文件对 pub 公钥文件地址， pri 私钥文件地址, bits 1024 或者 2048 公私钥生成密钥位数
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

// RSA 公钥加密 public 公钥文件地址, plaintext 明文
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

// RSA 私钥解密 private 私钥文件地址, ciphertext 密文
func DecryptByPrivateKey(private string, ciphertext []byte) ([]byte, error) {

	pemBlock, err := os.ReadFile(private)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(pemBlock)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}

	// 解析私钥内容
	if privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		if iprivKey, err := x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
			return nil, err
		} else {
			if plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, iprivKey.(*rsa.PrivateKey), ciphertext); err != nil {
				return nil, err
			} else {
				return plaintext, nil
			}
		}
	} else {
		plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, ciphertext)
		if err != nil {
			return nil, err
		}
		return plaintext, nil
	}
}
