package gocaptcha

import (
	"testing"
)

func TestRandom(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Log(Random(0, 1))
	}
}
