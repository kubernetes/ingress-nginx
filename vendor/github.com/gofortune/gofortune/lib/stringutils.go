package lib

func RemoveCRLF(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[0 : len(data)-1]
		if len(data) > 0 && data[len(data)-1] == '\r' {
			data = data[0 : len(data)-1]
		}
	}
	return data
}
