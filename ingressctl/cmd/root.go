package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ingressctl",
	Short: "CLI",
	Long:  `CLI for managing ingress-nginx repo`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(newCmdCompletion(os.Stdout))
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}

const completionExample = `
# Installing bash completion on macOS using homebrew
## If running Bash 3.2 included with macOS
	brew install bash-completion
## or, if running Bash 4.1+
	brew install bash-completion@2
`

func newCmdCompletion(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "completion [shell]",
		Short:   "Output shell completion code",
		Long:    ``,
		Example: completionExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompletion(out, cmd, args)
		},
		ValidArgs: []string{"bash", "zsh", "fish"},
	}

	return cmd
}

func runCompletion(out io.Writer, cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments; expected only the shell type: %s", args)
	}

	if len(args) == 0 || args[0] == "bash" {
		return cmd.Root().GenBashCompletion(out)
	} else if args[0] == "zsh" {
		return cmd.Root().GenZshCompletion(out)
	} else if args[0] == "fish" {
		return cmd.Root().GenFishCompletion(out, true)
	}
	return fmt.Errorf("unsupported shell: %s", args[0])
}
