package util

import (
	"testing"
)

func TestIsIdCard(t *testing.T) {
	// 测试有效的18位身份证号码
	if !IsIdCard("421003198705072317") {
		t.Error("Expected true for valid ID card")
	}

	// 测试无效的身份证号码（长度不对）
	if IsIdCard("12345678901234") {
		t.Error("Expected false for invalid ID card")
	}
}
