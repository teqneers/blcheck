package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	verbose bool

	rootCmd = &cobra.Command{
		Use:   "blkchecker",
		Short: "blkchecker is a very fast mailbox blacklist checker",
		Long:  `A Fast and Flexible mailbox blacklist checker.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
