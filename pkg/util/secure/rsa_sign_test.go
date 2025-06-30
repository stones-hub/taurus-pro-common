package secure

import (
	"crypto"
	"fmt"
	"strings"
	"testing"
)

var (
	PUBLIC_KEY  = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDZ1C2tQ42R/MVEGLZ1cNpaIjaQBJeH9NDvZsOta4p1Wv/AwgHzdCc0gEEqEiJaKiwQ8gui/gFm6V+eHuAhB/BUE6USKMYeL1SATyBB99c1pegL/7nSI/XK65OKFWxc7DJm45ybqpr+GN1tnm8dYGqGhiFbBDWXL97KO6e+fmyTSQIDAQAB"
	PRIVATE_KEY = "MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBANnULa1DjZH8xUQYtnVw2loiNpAEl4f00O9mw61rinVa/8DCAfN0JzSAQSoSIloqLBDyC6L+AWbpX54e4CEH8FQTpRIoxh4vVIBPIEH31zWl6Av/udIj9crrk4oVbFzsMmbjnJuqmv4Y3W2ebx1gaoaGIVsENZcv3so7p75+bJNJAgMBAAECgYAribAzgFwMgNRA3xug75SFDW+Qa4qJ/xG/t++Gewcqm6ygr2ZKbb3kTXo42XUKRoGWRXqz8kb/dcfJx+wOThLnrHxDLeHVmcpBEYbKceEc73v4QwjHxEO3kcC1NDui4l8Xbf38PPUU/HY9FZmJ7JHxKKcqO1GNGgY8kakVDgvLAQJBAP3ZokutwVmkbaPp2R+Mt84fHtnNbV8I8UZRTkD4MSibW970QfucbSmX8Tmn9AZVTR6VtLvtf6CJihHqVJkPe6ECQQDbrHKTzo6ZiSolEaEY8w7EAjuYd4cMMNhdtGNUE2iuTsFvdGE1fIFjD+gU1T381/XyxFPvrPyeI+5MJKrbIDapAkBKne5W0HxFHVAdHl/0JijhLcSjwP6lMLu7L6sQ7eOFTCV1I9dBXnm4ADGoAPZ55hkFJHw7wVQCnGs5WOgFFcgBAkAq0TQMB0jYOFoUm5kQ6d9I6T6Ae1vBTov9x7lMm/PdddBSTxbbfAckLeeIl//bFqUDyqypnMgocsxx3vvGdkLxAkEA35n1cawItP0eMKEcxDr0c2UeVheE4dLOBCHh+c9Tl43G9yYCHoHVbEAmp3sfPMAgKhUTgQONwJ0J6PzKiYuj+A=="
	message     = "apiCode=4a0e7744ebaabd0a43b63f3a832e8ab7&entCode=7720b9880852514ab1e272e7d49ec238"
)

/*
	{"MD5", crypto.MD5},
	{"SHA1", crypto.SHA1},
	{"SHA256", crypto.SHA256},
	{"SHA512", crypto.SHA512},
*/

func TestSignWithMD5(t *testing.T) {
	priKEY, err := ParsePriKEY(fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s-----END PRIVATE KEY-----", chunkSplit(PRIVATE_KEY, 64, "\n")))
	if err != nil {
		t.Errorf("%v", err)
	}
	sign, err := SignWithRSA(message, priKEY, crypto.MD5)
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Logf("原字符:%s 签名(base64):%v", message, Byte2Base64(sign))

	// --------------> 验签
	pubKEY, err := ParsePubKEY(fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s-----END PUBLIC KEY-----", chunkSplit(PUBLIC_KEY, 64, "\n")))
	if err != nil {
		t.Errorf("%v", err)
	}

	if err := VerifyWithRSA(message, pubKEY, sign, crypto.MD5); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Log("验签成功")
	}
}

func TestVerifyWithRSA(t *testing.T) {
	priKEY, err := ParsePriKEY(fmt.Sprintf("-----BEGIN PRIVATE KEY-----\n%s-----END PRIVATE KEY-----", chunkSplit(PRIVATE_KEY, 64, "\n")))
	if err != nil {
		t.Errorf("%v", err)
	}
	sign, err := SignWithRSA(message, priKEY, crypto.SHA256)
	if err != nil {
		t.Errorf("%v", err)
	}
	t.Logf("原字符:%s 签名(base64):%v", message, Byte2Base64(sign))

	// --------------> 验签
	pubKEY, err := ParsePubKEY(fmt.Sprintf("-----BEGIN PUBLIC KEY-----\n%s-----END PUBLIC KEY-----", chunkSplit(PUBLIC_KEY, 64, "\n")))
	if err != nil {
		t.Errorf("%v", err)
	}

	if err := VerifyWithRSA(message, pubKEY, sign, crypto.SHA256); err != nil {
		t.Errorf("%v", err)
	} else {
		t.Log("验签成功")
	}

}

// ChunkSplit input字符串每隔chunkSize个字符添加delimiter,并返回最终的string
func chunkSplit(input string, chunkSize int, delimiter string) string {
	var chunks []string
	for i := 0; i < len(input); i += chunkSize {
		end := i + chunkSize
		if end > len(input) {
			end = len(input)
		}
		chunks = append(chunks, input[i:end])
	}

	// 如果最后一个块长度小于 chunkSize，则在其后面添加 delimiter
	if len(chunks) > 0 && len(chunks[len(chunks)-1]) < chunkSize {
		chunks[len(chunks)-1] += delimiter
	}

	return strings.Join(chunks, delimiter)
}
