package cyclicwithtest

import "testing"

func TestUseB(t *testing.T) {
	if UseB() != 2 {
		t.Fatal("UseB")
	}
}
