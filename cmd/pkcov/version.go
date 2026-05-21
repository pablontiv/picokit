package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print pkcov version and implemented spec version",
	Run: func(cmd *cobra.Command, _ []string) {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "pkcov %s (coverage-spec %s)\n", buildVersion, specVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
