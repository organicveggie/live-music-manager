package cmd

import (
	"github.com/organicveggie/livemusic/lm/cmd/scan"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "lm",
		Short: "Live Music manager",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(scan.Cmd)
}
