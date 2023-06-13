package cmd

import "runtime"

func getArch() string {
	return runtime.GOARCH
}
