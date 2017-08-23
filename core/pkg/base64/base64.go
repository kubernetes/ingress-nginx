package base64

import (
	"encoding/base64"
	"strings"
)

// Base64Encode
func Base64Encode(s string) string {
	str := base64.URLEncoding.EncodeToString([]byte(s))
	return strings.Replace(str, "=", "", -1)
}
