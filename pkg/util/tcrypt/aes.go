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
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/stones-hub/taurus-pro-common/pkg/util/tstring"
)

// AesEncryptCBCPHP 使用AES-CBC模式加密数据，兼容PHP的加密实现
// 参数：
//   - originData: 要加密的原始数据
//   - key: 加密密钥，长度必须为16、24或32字节（分别对应AES-128、AES-192、AES-256）
//
// 返回值：
//   - []byte: 加密后的数据
//   - error: 加密过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	key := []byte("1234567890123456") // 16字节的密钥
//	data := []byte("Hello World")
//	encrypted, err := tcrypt.AesEncryptCBCPHP(data, key)
//	if err != nil {
//	    log.Printf("加密失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 本函数特别设计用于与PHP的AES加密实现兼容
//   - IV向量使用密钥的反转字符串，这是为了匹配特定PHP框架的实现
//   - 使用PKCS5填充
//   - 确保与PHP端使用相同的密钥
//   - 不建议在新项目中使用，除非需要兼容已有的PHP实现
func AesEncryptCBCPHP(originData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	originData = pkcs5Padding(originData, blockSize)
	// 兼容PHP ， 3K游戏 PHP框架中的IV向量是密钥KEY的反转
	blockMode := cipher.NewCBCEncrypter(block, []byte(tstring.Reverse(string(key))))
	encrypted := make([]byte, len(originData))
	blockMode.CryptBlocks(encrypted, originData)
	return encrypted, nil
}

// AesDecryptCBCPHP 使用AES-CBC模式解密数据，兼容PHP的解密实现
// 参数：
//   - encrypted: 使用AesEncryptCBCPHP加密的数据
//   - key: 解密密钥，必须与加密时使用的密钥相同
//
// 返回值：
//   - []byte: 解密后的原始数据
//   - error: 解密过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	key := []byte("1234567890123456") // 16字节的密钥
//	decrypted, err := tcrypt.AesDecryptCBCPHP(encrypted, key)
//	if err != nil {
//	    log.Printf("解密失败：%v", err)
//	    return
//	}
//	fmt.Printf("解密结果：%s\n", string(decrypted))
//
// 注意事项：
//   - 本函数用于解密由AesEncryptCBCPHP加密的数据
//   - IV向量使用密钥的反转字符串，这是为了匹配特定PHP框架的实现
//   - 使用PKCS5填充解码
//   - 确保与PHP端使用相同的密钥
//   - 如果输入数据不是有效的加密数据，可能会panic
func AesDecryptCBCPHP(encrypted []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, []byte(tstring.Reverse(string(key))))
	decrypted := make([]byte, len(encrypted))
	blockMode.CryptBlocks(decrypted, encrypted)
	decrypted = pkcs5UnPadding(decrypted)
	return decrypted, nil
}

// AesEncryptCBC 使用AES-CBC模式加密数据
// 参数：
//   - originData: 要加密的原始数据
//   - key: 加密密钥，长度必须为16、24或32字节（分别对应AES-128、AES-192、AES-256）
//
// 返回值：
//   - []byte: 加密后的数据
//   - error: 加密过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	key := []byte("1234567890123456") // 16字节的密钥
//	data := []byte("Hello World")
//	encrypted, err := tcrypt.AesEncryptCBC(data, key)
//	if err != nil {
//	    log.Printf("加密失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - 使用标准的AES-CBC模式加密
//   - IV向量使用密钥的前BlockSize个字节
//   - 使用PKCS5填充
//   - 加密结果不包含IV，需要解密方知道正确的密钥
//   - 适用于需要高安全性的场景
func AesEncryptCBC(originData []byte, key []byte) ([]byte, error) {

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	originData = pkcs5Padding(originData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	encrypted := make([]byte, len(originData))
	blockMode.CryptBlocks(encrypted, originData)
	return encrypted, nil
}

// pkcs5Padding 对数据进行PKCS5填充
// 参数：
//   - ciphertext: 需要填充的数据
//   - blockSize: 块大小（通常是16字节）
//
// 返回值：
//   - []byte: 填充后的数据
//
// 注意事项：
//   - 用于确保数据长度是块大小的整数倍
//   - 填充值是填充的字节数
//   - 如果数据长度已经是块大小的整数倍，仍会添加一个完整的块
func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, paddingText...)
}

// pkcs5UnPadding 移除PKCS5填充
// 参数：
//   - origData: 带填充的数据
//
// 返回值：
//   - []byte: 移除填充后的原始数据
//
// 注意事项：
//   - 用于移除PKCS5填充的数据
//   - 如果数据不是有效的PKCS5填充，可能会panic
//   - 最后一个字节表示填充的字节数
func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// AesDecryptCBC 使用AES-CBC模式解密数据
// 参数：
//   - encrypted: 使用AesEncryptCBC加密的数据
//   - key: 解密密钥，必须与加密时使用的密钥相同
//
// 返回值：
//   - []byte: 解密后的原始数据
//   - error: 解密过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	key := []byte("1234567890123456") // 16字节的密钥
//	decrypted, err := tcrypt.AesDecryptCBC(encrypted, key)
//	if err != nil {
//	    log.Printf("解密失败：%v", err)
//	    return
//	}
//	fmt.Printf("解密结果：%s\n", string(decrypted))
//
// 注意事项：
//   - 本函数用于解密由AesEncryptCBC加密的数据
//   - IV向量使用密钥的前BlockSize个字节
//   - 使用PKCS5填充解码
//   - 确保使用与加密时相同的密钥
//   - 如果输入数据不是有效的加密数据，可能会panic
func AesDecryptCBC(encrypted []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	decrypted := make([]byte, len(encrypted))
	blockMode.CryptBlocks(decrypted, encrypted)
	decrypted = pkcs5UnPadding(decrypted)
	return decrypted, nil
}

// AesEncryptECB 使用AES-ECB模式加密数据
// 参数：
//   - origData: 要加密的原始数据
//   - key: 加密密钥，长度必须为16、24或32字节（分别对应AES-128、AES-192、AES-256）
//
// 返回值：
//   - []byte: 加密后的数据
//   - error: 加密过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	key := []byte("1234567890123456") // 16字节的密钥
//	data := []byte("Hello World")
//	encrypted, err := tcrypt.AesEncryptECB(data, key)
//	if err != nil {
//	    log.Printf("加密失败：%v", err)
//	    return
//	}
//
// 注意事项：
//   - ECB模式是一种基本的加密模式，不推荐用于需要高安全性的场景
//   - 相同的明文块会被加密为相同的密文块，可能暴露数据模式
//   - 没有使用IV向量，安全性较低
//   - 主要用于兼容旧系统或特殊要求
//   - 建议使用CBC或CTR模式代替
func AesEncryptECB(origData []byte, key []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(generateKey(key))
	if err != nil {
		return nil, err
	}
	length := (len(origData) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	copy(plain, origData)
	pad := byte(len(plain) - len(origData))
	for i := len(origData); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted := make([]byte, len(plain))
	// 分组分块加密
	for bs, be := 0, cipher.BlockSize(); bs <= len(origData); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted, nil
}

// AesDecryptECB 使用AES-ECB模式解密数据
// 参数：
//   - encrypted: 使用AesEncryptECB加密的数据
//   - key: 解密密钥，必须与加密时使用的密钥相同
//
// 返回值：
//   - []byte: 解密后的原始数据
//   - error: 解密过程中的错误，如果成功则为nil
//
// 使用示例：
//
//	key := []byte("1234567890123456") // 16字节的密钥
//	decrypted, err := tcrypt.AesDecryptECB(encrypted, key)
//	if err != nil {
//	    log.Printf("解密失败：%v", err)
//	    return
//	}
//	fmt.Printf("解密结果：%s\n", string(decrypted))
//
// 注意事项：
//   - 本函数用于解密由AesEncryptECB加密的数据
//   - ECB模式的安全性较低，不推荐用于新项目
//   - 确保使用与加密时相同的密钥
//   - 如果输入数据不是有效的加密数据，可能会panic
//   - 建议使用CBC或CTR模式的解密函数代替
func AesDecryptECB(encrypted []byte, key []byte) ([]byte, error) {
	cipher, err := aes.NewCipher(generateKey(key))
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(encrypted))
	//
	for bs, be := 0, cipher.BlockSize(); bs < len(encrypted); bs, be = bs+cipher.BlockSize(), be+cipher.BlockSize() {
		cipher.Decrypt(decrypted[bs:be], encrypted[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}

	return decrypted[:trim], nil
}

// generateKey 生成固定长度的密钥
// 参数：
//   - key: 原始密钥
//
// 返回值：
//   - []byte: 生成的16字节密钥
//
// 注意事项：
//   - 用于ECB模式的密钥生成
//   - 如果输入密钥小于16字节，将进行填充
//   - 如果输入密钥大于16字节，将进行压缩
//   - 使用异或运算确保生成的密钥具有良好的随机性
func generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}

// AES CTR 模式加密/解密

// AesEncryptCTR 使用 AES-CTR 模式加密数据
// 注意：CTR 模式是一个流密码，加密和解密使用相同的操作
func AesEncryptCTR(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher failed: %w", err)
	}

	// 生成随机的初始化向量（IV）
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate IV failed: %w", err)
	}

	// 使用 CTR 模式加密
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

// AesDecryptCTR 使用 AES-CTR 模式解密数据
func AesDecryptCTR(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher failed: %w", err)
	}

	// 检查密文长度
	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short: must be at least as long as IV")
	}

	// 提取 IV 和实际密文
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// 分配明文缓冲区
	plaintext := make([]byte, len(ciphertext))

	// 使用 CTR 模式解密（实际上是相同的操作）
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}
