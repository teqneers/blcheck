package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
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
	rootCmd.PersistentFlags().CountP("quiet", "q", "quiet option to disable all output, overwrites verbose if set")
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))

	rootCmd.PersistentFlags().CountP("verbose", "v", "verbose output, supports levels: -v, -vv, -vvv")
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
}
