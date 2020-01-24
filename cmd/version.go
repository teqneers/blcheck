package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of blkchecker",
	Long:  `All software has versions. This is blkchecker's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("blkchecker -- Blacklist Mailbox Checker -- v0.1 -- HEAD")
	},
}
