package util

import (
	"testing"
)

func TestIsIdCard(t *testing.T) {
	if !IsIdCard("123456789012345") {
		t.Error("Expected true for valid ID card")
	}

	if IsIdCard("12345678901234") {
		t.Error("Expected false for invalid ID card")
	}

	if !IsIdCard("123456789012345678") {
		t.Error("Expected true for valid ID card")
	}

	if IsIdCard("12345678901234567") {
		t.Error("Expected false for invalid ID card")
	}
}
