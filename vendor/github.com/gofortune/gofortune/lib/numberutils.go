package lib

func Min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func Max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}
