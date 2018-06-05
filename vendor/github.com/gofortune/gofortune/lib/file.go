package lib

import (
	"os"
	"path/filepath"
)

func RemoveFileExtension(file string) string {
	var extension = filepath.Ext(file)
	return file[0 : len(file)-len(extension)]
}

func FileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
