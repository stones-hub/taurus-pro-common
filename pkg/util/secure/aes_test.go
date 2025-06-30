package secure

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
)

var origData = "啦x啦阿L_哈085哈$"

// 密钥 必须 16位
var k = "ABCDEFGHIJKLMNOP"

func TestAesCBCPHP(t *testing.T) {
	var (
		orig = "test"                     // 明文
		k    = "4a4Tm6s3PQLgXbpa"         // 密钥
		e    = "cXEamDlC/BAgFEEQ7dreEA==" // base64密文
	)

	if en, err := AesEncryptCBCPHP([]byte(orig), []byte(k)); err != nil {
		t.Errorf("aes加密失败,err:%v", err)
	} else {
		if e != base64.StdEncoding.EncodeToString(en) {
			t.Errorf("密文(hex) 校验失败, %s, %s !", e, base64.StdEncoding.EncodeToString(en))
		}

		t.Logf("密文(base64) 校验成功. (%s) Equal (%s) !", base64.StdEncoding.EncodeToString(en), e)

		if de, err := AesDecryptCBCPHP(en, []byte(k)); err != nil {
			t.Errorf("aes解密失败,err:%v", err)
		} else {
			if string(de) != orig {
				t.Errorf("明文校验失败, %s, %s !", orig, string(de))
			}

			t.Logf("明文校验成功. (%s) Equal (%s) !", orig, string(de))
		}
	}
}

func TestAesEncryptCBC(t *testing.T) {
	if encrypted, err := AesEncryptCBC([]byte(origData), []byte(k)); err != nil {
		t.Errorf("aes加密失败,err:%v", err)
	} else {
		if hex.EncodeToString(encrypted) != "8351089dedd5b303c5ff446638307b866cdb4362c57c78647295051c2428e648" {
			t.Error("密文(hex) 校验失败 !")
		}
		if base64.StdEncoding.EncodeToString(encrypted) != "g1EIne3VswPF/0RmODB7hmzbQ2LFfHhkcpUFHCQo5kg=" {
			t.Error("密文(base64) 校验失败 !")
		}
	}
}

func TestAesDecryptCBC(t *testing.T) {

	if encrypted, err := hex.DecodeString("8351089dedd5b303c5ff446638307b866cdb4362c57c78647295051c2428e648"); err != nil {
		t.Errorf("aes解密失败,err:%v", err)
	} else {
		if decrypted, err := AesDecryptCBC(encrypted, []byte(k)); err != nil {
			t.Errorf("aes解密失败,err:%v", err)
		} else {
			if string(decrypted) != origData {
				t.Error("明文校验失败 !")
			}
		}
	}
	if encrypted, err := base64.StdEncoding.DecodeString("g1EIne3VswPF/0RmODB7hmzbQ2LFfHhkcpUFHCQo5kg="); err != nil {
		t.Errorf("aes解密失败,err:%v", err)
	} else {
		if decrypted, err := AesDecryptCBC(encrypted, []byte(k)); err != nil {
			t.Errorf("aes解密失败,err:%v", err)
		} else {
			if string(decrypted) != origData {
				t.Error("明文校验失败 !")
			}
		}
	}
}

func TestAesEncryptECB(t *testing.T) {
	if encrypted, err := AesEncryptECB([]byte(origData), []byte(k)); err != nil {
		t.Errorf("aes加密失败,err:%v", err)
	} else {
		if hex.EncodeToString(encrypted) != "601bd5fa16df3c9f30f851982eb786586487f8c83fc0d9803838c213e63b27f9" {
			t.Error("密文(hex) 校验失败 !")
		}

		if base64.StdEncoding.EncodeToString(encrypted) != "YBvV+hbfPJ8w+FGYLreGWGSH+Mg/wNmAODjCE+Y7J/k=" {
			t.Error("密文(base64) 校验失败 !")
		}
	}
}

func TestAesDecryptECB(t *testing.T) {
	if encrypted, err := hex.DecodeString("601bd5fa16df3c9f30f851982eb786586487f8c83fc0d9803838c213e63b27f9"); err != nil {
		t.Errorf("aes解密失败,err:%v", err)
	} else {

		if decrypted, err := AesDecryptECB(encrypted, []byte(k)); err != nil {
			t.Errorf("aes解密失败,err:%v", err)
		} else {
			if string(decrypted) != origData {
				t.Error("明文校验失败 !")
			}
		}
	}
}

// 每次变化
func TestAesCTR(t *testing.T) {
	if encrypted, err := AesEncryptCTR([]byte(origData), []byte(k)); err != nil {
		t.Errorf("aes加密失败,err:%v", err)
	} else {

		if s, err := AesDecryptCTR(encrypted, []byte(k)); err != nil {
			t.Errorf("aes解密失败,err:%v", err)
		} else {
			// fmt.Printf("明文:%s, base(64)密文: %s, hex密文: %s ", origData, base64.StdEncoding.EncodeToString(encrypted), hex.EncodeToString(encrypted))
			if string(s) != origData {
				t.Errorf("明文校验失败 %s != %s !", origData, string(s))
			}
		}
	}
}
