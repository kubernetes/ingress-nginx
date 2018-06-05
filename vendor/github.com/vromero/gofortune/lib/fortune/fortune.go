// Package fortune provides fortune cookie selection. This package will not output
// any data to the terminal.
package fortune

import (
	"math/rand"
	"os"

	"errors"
	"path/filepath"
	"time"

	"regexp"

	"fmt"

	"github.com/gofortune/gofortune/lib"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func GetRandomFortune(rootNode FileSystemNodeDescriptor) (file string, data string, err error) {
	randomNode := GetRandomLeafNode(rootNode)
	if randomNode.NumEntries == 0 {
		panic("File is empty")
	}
	randomEntry := rand.Intn(int(randomNode.NumEntries))

	indexFile, err := os.Open(randomNode.IndexPath)
	defer indexFile.Close()
	if err != nil {
		panic("Can't open index file")
	}

	dataPos, err := lib.ReadDataPos(indexFile, int(lib.DataTableSize), uint32(randomEntry))
	if err != nil {
		panic("Can't read index file")
	}
	indexFile.Close()

	fortuneFile, err := os.Open(randomNode.Path)
	if err != nil {
		panic("Can't read fortune file")
	}
	defer fortuneFile.Close()

	data, err = lib.ReadData(fortuneFile, int64(dataPos.OriginalOffset))
	return filepath.Base(randomNode.Path), data, err
}

func GetFilteredRandomFortune(rootNode FileSystemNodeDescriptor, filter func(string) bool) (file string, data string, err error) {
	if filter == nil {
		return "", "", errors.New("Filter can't be nil")
	}

	for {
		file, data, err = GetRandomFortune(rootNode)
		if filter(data) {
			break
		}
	}
	return
}

func MatchFortunes(fsDescriptor FileSystemNodeDescriptor, expression string, printer func(string)) {
	matchingExpression := regexp.MustCompile(expression)
	matchFortunesWithCompiledExpression(fsDescriptor, matchingExpression, printer)
}

func matchFortunesWithCompiledExpression(fsDescriptor FileSystemNodeDescriptor, expression *regexp.Regexp, printer func(string)) {
	if len(fsDescriptor.Children) > 0 {
		for i := range fsDescriptor.Children {
			matchFortunesWithCompiledExpression(fsDescriptor.Children[i], expression, printer)
		}
	} else {
		indexFile, err := os.Open(fsDescriptor.IndexPath)
		defer indexFile.Close()
		if err != nil {
			panic("Can't open index file")
		}

		fortuneFile, err := os.Open(fsDescriptor.Path)
		defer fortuneFile.Close()
		if err != nil {
			panic("Can't read fortune file")
		}

		for i := int64(0); i < int64(fsDescriptor.NumEntries); i++ {
			dataPos, err := lib.ReadDataPos(indexFile, int(lib.DataTableSize), uint32(i))
			if err != nil {
				panic(fmt.Sprintf("Can't read from index file, fortune number : %v", i))
			}

			data, err := lib.ReadData(fortuneFile, int64(dataPos.OriginalOffset))
			if expression.MatchString(data) {
				printer(data)
			}
		}
	}
}
