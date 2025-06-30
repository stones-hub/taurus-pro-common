package secure

import (
	"testing"
)

func TestBase642Byte(t *testing.T) {
	var (
		s     = "中华人民共和国"
		bas64 = "5Lit5Y2O5Lq65rCR5YWx5ZKM5Zu9"
		err   error
		b     []byte
	)

	if b, err = Base642Byte(bas64); err != nil {
		t.Error(err)
	} else {
		if string(b) != s {
			t.Error("Base64解码失败")
		}
	}
}

func TestByte2Base64(t *testing.T) {
	var (
		s     = "中华人民共和国"
		bas64 = "5Lit5Y2O5Lq65rCR5YWx5ZKM5Zu9"
	)

	if Byte2Base64([]byte(s)) != bas64 {
		t.Error("Base64编码失败")
	}

}
