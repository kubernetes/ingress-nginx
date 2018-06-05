package main

import (
	"github.com/gofortune/gofortune/cmd"
	"os"
	"path/filepath"
)

func main() {
	processAliases()
	cmd.Execute()
}

// To maintain compatibility with the classic tools: fortune, strfile. GoFortune supports to have its
// executable renamed or symlinked. If the appropriate names are found, the command line will
// be altered to honor the original syntax.
func processAliases() {
	switch getExecutableName() {
	case "fortune":
		os.Args = append([]string{"gofortune", "fortune"}, os.Args[1:]...)
	case "strfile":
		os.Args = append([]string{"gofortune", "strfile"}, os.Args[1:]...)
	}
}

// Finds the name of the executable (or symlink to executable) used to call this program
func getExecutableName() string {
	return filepath.Base(os.Args[0])
}
