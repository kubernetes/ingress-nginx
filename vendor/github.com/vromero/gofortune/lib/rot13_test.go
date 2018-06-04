package lib

import "testing"

func TestAlphabet(t *testing.T) {
	expected := "NOPQRSTUVWXYZABCDEFGHIJKLMnopqrstuvwxyzabcdefghijklm"
	got := Rot13("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	if Rot13("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz") != expected {
		t.Error("Expected " + expected + " got " + got)
	}
}
