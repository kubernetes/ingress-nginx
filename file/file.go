package file

import (
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
)

// SHA1 returns the SHA1 of a file.
func SHA1(filename string) string {
	hasher := sha1.New()
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}

	hasher.Write(s)
	return hex.EncodeToString(hasher.Sum(nil))
}
