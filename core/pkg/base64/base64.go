package base64

import (
	"encoding/base64"
	"strings"
)

// Encode encodes a string to base64 removing the equals character
func Encode(s string) string {
	str := base64.URLEncoding.EncodeToString([]byte(s))
	return strings.Replace(str, "=", "", -1)
}
