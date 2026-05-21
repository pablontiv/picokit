package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// specVersion is the coverage-spec version this build implements.
const specVersion = "v1.1"

// buildVersion is set at build time via -ldflags.
var buildVersion = "dev"

var outputFormat string

var rootCmd = &cobra.Command{
	Use:     "pkcov",
	Short:   "Coverage gate and report tool implementing coverage-spec " + specVersion,
	Version: buildVersion,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "output format (text|json)")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
