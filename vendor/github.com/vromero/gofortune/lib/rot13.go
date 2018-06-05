package lib

import "strings"

func Rot13(input string) string {
	return strings.Map(Rot13Rune, input)
}

func Rot13Rune(r rune) rune {
	if (r >= 'A' && r < 'N') || (r >= 'a' && r < 'n') {
		return r + 13
	} else if (r > 'M' && r <= 'Z') || (r > 'm' && r <= 'z') {
		return r - 13
	}
	return r
}
