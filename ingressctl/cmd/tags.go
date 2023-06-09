package cmd

import (
	"os"
	"strings"
)

func getIngressNGINXVersion() (string, error) {

	dat, err := os.ReadFile("TAG")
	CheckIfError(err, "Could not read TAG file")
	datString := string(dat)
	//remove newline
	datString = strings.Replace(datString, "\n", "", -1)
	return datString, nil
}
