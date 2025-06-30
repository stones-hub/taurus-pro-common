package secure

import "testing"

// 公钥加密
func TestEncryptByPublicKey(t *testing.T) {
	/*
		if en, err := EncryptByPublicKey("../../test/pub.pem", []byte("test")); err != nil {
			t.Errorf("EncryptByPublicKey error: %v", err)
		} else {
			t.Logf("EncryptByPublicKey: %s", Byte2Base64(en))
		}

	*/

}

// 私钥解密
func TestDecryptByPrivateKey(t *testing.T) {
	/*
		en := "gfIugjB7JYoOMl8lsXF0ZAJWyq4uffRjDBZEpGOlzolT64umdHejHoMREhELYJUCCmzLk00Igv6Ntk9YJDCcRLYO7otIIDGuSQGp/4KrDdwDOZDERrOt7U+nwd8fjP3d1DapWieuFn7BcUqDGQ1V50uuquipnh+ces1pQ4gAo9I="

		if b, err := Base642Byte(en); err != nil {
			t.Errorf("Base642Byte error: %v", err)
		} else {
			if de, err := DecryptByPrivateKey("../../test/pri.pem", b); err != nil {
				t.Errorf("DecryptByPrivateKey error: %v", err)
			} else {
				t.Logf("DecryptByPrivateKey: %s", de)
			}
		}

	*/
}

// 创建公私钥
func TestGenerateKeyPair(t *testing.T) {
	/*
		if err := GenerateKeyPair("../../test/pub.pem", "../../test/pri.pem", 1024); err != nil {
			t.Errorf("GenerateKeyPair error: %v", err)
		}

	*/
}
