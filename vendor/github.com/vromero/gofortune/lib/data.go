package lib

import (
	"bufio"
	"bytes"
	"os"
)

// Reads a whole fortune from the fortune base file
func ReadData(inputFile *os.File, pos int64) (string, error) {
	buffer := make([]byte, 512)
	_, err := inputFile.ReadAt(buffer, pos)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(bytes.NewReader(buffer))
	scanner.Split(advanceAwareSplitter)
	scanner.Scan()

	return string(RemoveCRLF(scanner.Bytes())), nil
}

// advanceAwareSplitter splits byte array in lines returning also how many bytes were read.
// It differs from the default line splitter by the ability to return how many bytes were read.
func advanceAwareSplitter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte("\n%\n")); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
