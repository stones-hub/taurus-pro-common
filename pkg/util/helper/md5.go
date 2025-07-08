package helper

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}
