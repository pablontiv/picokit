package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/pablontiv/picokit/coverage"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Print per-package coverage table",
	RunE:  runReport,
}

var reportProfile string
var reportModule string

func init() {
	reportCmd.Flags().StringVar(&reportProfile, "profile", "coverage.out", "path to coverage profile")
	reportCmd.Flags().StringVar(&reportModule, "module", "", "module prefix (auto-detected from go.mod if empty)")
	rootCmd.AddCommand(reportCmd)
}

func runReport(cmd *cobra.Command, _ []string) error {
	prefix, err := resolveModule(reportModule)
	if err != nil {
		return err
	}

	p, err := coverage.ParseProfile(reportProfile, prefix)
	if err != nil {
		return fmt.Errorf("parse profile: %w", err)
	}

	perPkg := p.PerPackage()
	pkgs := sortedKeys(perPkg)

	if outputFormat == "json" {
		type pkgEntry struct {
			Covered int     `json:"covered"`
			Total   int     `json:"total"`
			Skipped bool    `json:"skipped"`
			Percent float64 `json:"percent"`
		}
		out := struct {
			Total      float64             `json:"total"`
			PerPackage map[string]pkgEntry `json:"per_package"`
		}{
			Total:      p.Total(),
			PerPackage: make(map[string]pkgEntry, len(perPkg)),
		}
		for k, b := range perPkg {
			out.PerPackage[k] = pkgEntry{
				Covered: b.Covered,
				Total:   b.Total,
				Skipped: b.Skipped,
				Percent: b.Percent(),
			}
		}
		return json.NewEncoder(os.Stdout).Encode(out)
	}

	out := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(out, "Coverage Report:")
	_, _ = fmt.Fprintln(out, "================")
	for _, pkg := range pkgs {
		b := perPkg[pkg]
		if b.Skipped {
			_, _ = fmt.Fprintf(out, "SKIP: %s\n", pkg)
		} else {
			_, _ = fmt.Fprintf(out, "PASS: %s = %.1f%%\n", pkg, b.Percent())
		}
	}
	_, _ = fmt.Fprintf(out, "\nTOTAL: %.1f%%\n", p.Total())
	return nil
}

func sortedKeys(m map[string]coverage.Bucket) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
