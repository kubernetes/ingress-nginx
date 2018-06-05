package lib

import (
	"encoding/binary"
	"os"
	"unsafe"
)

type DataPos struct {
	OriginalOffset uint32
	Text           string
}

func ReadDataPos(inputFile *os.File, tableSize int, position uint32) (DataPos, error) {
	buffer := make([]byte, 4)
	_, err := inputFile.ReadAt(buffer, int64(int64(tableSize)+int64(position)*4))
	if err != nil {
		return DataPos{}, err
	}

	return DataPos{
		OriginalOffset: binary.BigEndian.Uint32(buffer),
	}, nil
}

func WriteDataPos(outputFile *os.File, tableSize int, position uint32, datapos DataPos) {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, datapos.OriginalOffset)
	outputFile.WriteAt(buffer, int64(tableSize)+int64(unsafe.Sizeof(position))*int64(position))
}

func WriteDataPosSlice(outputFile *os.File, tableSize int, dataposSlice []DataPos) {
	for i := range dataposSlice {
		WriteDataPos(outputFile, tableSize, uint32(i), dataposSlice[i])
	}
}

func LessThanDataPos(i DataPos, j DataPos) bool {
	return i.Text[0:1] < j.Text[0:1]
}
