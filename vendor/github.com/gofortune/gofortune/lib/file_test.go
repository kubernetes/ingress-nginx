package lib

import (
	"io/ioutil"
	"path/filepath"
	"syscall"
	"testing"
)

func TestRemoveFileExtension(t *testing.T) {
	got := RemoveFileExtension("hello.goodbye")
	expected := "hello"

	if got != expected {
		t.Error("Expected " + expected + " got " + got)
	}
}

func TestFileExists(t *testing.T) {
	existentFile, err := ioutil.TempFile("", "gofortune")
	if err != nil {
		panic(err)
	}
	defer existentFile.Close()

	if FileExists(existentFile.Name()) != true {
		t.Error("Expedted file exists")
	}
}

func TestFileDoestExists(t *testing.T) {
	emptyDirectory, err := ioutil.TempDir("", "gofortune")
	if err != nil {
		panic(err)
	}
	defer syscall.Unlink(emptyDirectory)

	nonExistentFile := filepath.Join(emptyDirectory, "nonExistentFile.name")

	if FileExists(nonExistentFile) != false {
		t.Error("Expedted file doesnt exists")
	}
}
