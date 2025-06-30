package util

import (
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestInArray(t *testing.T) {

	var (
		k = "stones"
		a = []string{"stones", "st", "ondes", "ones"}
		b = []string{"st", "ondes", "ones"}
	)

	testCase := []struct {
		p1       string
		p2       []string
		expected bool
	}{
		{k, a, true},
		{k, b, false},
	}

	for _, s := range testCase {
		t.Run(fmt.Sprintf("InArray(%v, %v)", s.p1, s.p2), func(t *testing.T) {
			assert.Equal(t, InArray(s.p1, s.p2), s.expected)
		})
	}
}
