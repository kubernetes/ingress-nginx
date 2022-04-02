package cmd

import (
	"fmt"
	"runtime"

	"github.com/jet/kube-webhook-certgen/core"
	"github.com/spf13/cobra"
)

var version = &cobra.Command{
	Use:   "version",
	Short: "Prints the CLI version information",
	Run:   versionCmdRun,
}

func versionCmdRun(cmd *cobra.Command, args []string) {
	fmt.Printf("%s\n", core.Version)
	fmt.Printf("build %s\n", core.BuildTime)
	fmt.Printf("%s\n", runtime.Version())
}

func init() {
	rootCmd.AddCommand(version)
}
