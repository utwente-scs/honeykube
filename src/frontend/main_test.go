package main

import (
	"testing"
)

func test_randNum(t *testing.T) {
	randstr := randNum()

	if len(randstr) != 4 {
		t.Errorf("Incorrent random number string generated!")
	}
}
