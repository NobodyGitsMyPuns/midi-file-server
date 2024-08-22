package main

import "testing"

func TestReturnOne(t *testing.T) {
	if ReturnOne() != 1 {
		t.Error("Expected 1")
	}
}
