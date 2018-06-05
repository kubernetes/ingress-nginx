package lib

import (
	"bytes"
	"testing"
)

const TEST_STRING = "hello"

func TestRemoveN(t *testing.T) {
	validateRemoveX(t, "\n")
}

func TestRemoveRN(t *testing.T) {
	validateRemoveX(t, "\r\n")
}

func TestDontRemoveValid(t *testing.T) {
	validateRemoveX(t, "")
}

func validateRemoveX(t *testing.T, suffix string) {
	removed := RemoveCRLF([]byte(TEST_STRING + suffix))

	if bytes.Compare(removed, []byte(TEST_STRING)) != 0 {
		t.Error("\\n or \\r was not removed")
	}
}
