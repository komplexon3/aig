package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aig",
	Short: "AIG is a CLI tool to launch customized docker containers",
	Long: `AIG (AI-Generated CLI) allows you to build and run docker containers 
by specifying base images and additional layers defined in Go code.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Root flags can be added here
}
