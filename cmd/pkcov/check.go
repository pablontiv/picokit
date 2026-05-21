package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pablontiv/picokit/coverage"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check coverage against floors; exits 1 on violations",
	RunE:  runCheck,
}

var checkProfile string
var checkFloors string
var checkModule string

func init() {
	checkCmd.Flags().StringVar(&checkProfile, "profile", "coverage.out", "path to coverage profile")
	checkCmd.Flags().StringVar(&checkFloors, "floors", ".coverage-floors.toml", "path to floors config")
	checkCmd.Flags().StringVar(&checkModule, "module", "", "module prefix (auto-detected from go.mod if empty)")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, _ []string) error {
	prefix, err := resolveModule(checkModule)
	if err != nil {
		return err
	}

	p, err := coverage.ParseProfile(checkProfile, prefix)
	if err != nil {
		return fmt.Errorf("parse profile: %w", err)
	}
	f, err := coverage.LoadFloors(checkFloors)
	if err != nil {
		return fmt.Errorf("load floors: %w", err)
	}

	r := coverage.Check(p, f)

	if outputFormat == "json" {
		type violation struct {
			Package string  `json:"package"`
			Got     float64 `json:"got"`
			Need    float64 `json:"need"`
		}
		type pkgEntry struct {
			Covered int     `json:"covered"`
			Total   int     `json:"total"`
			Skipped bool    `json:"skipped"`
			Percent float64 `json:"percent"`
		}
		out := struct {
			Total      float64             `json:"total"`
			PerPackage map[string]pkgEntry `json:"per_package"`
			Violations []violation         `json:"violations"`
			Skipped    []string            `json:"skipped"`
		}{
			Total:      r.Total,
			PerPackage: make(map[string]pkgEntry, len(r.PerPackage)),
			Violations: make([]violation, 0, len(r.Violations)),
			Skipped:    r.SkippedPackages,
		}
		if out.Skipped == nil {
			out.Skipped = []string{}
		}
		for k, b := range r.PerPackage {
			out.PerPackage[k] = pkgEntry{
				Covered: b.Covered,
				Total:   b.Total,
				Skipped: b.Skipped,
				Percent: b.Percent(),
			}
		}
		for _, v := range r.Violations {
			out.Violations = append(out.Violations, violation{v.Package, v.Got, v.Need})
		}
		if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
			return err
		}
		if len(r.Violations) > 0 {
			os.Exit(1)
		}
		return nil
	}

	// text output
	w := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(w, "Coverage Report:")
	_, _ = fmt.Fprintln(w, "================")
	violations := make(map[string]bool, len(r.Violations))
	for _, v := range r.Violations {
		violations[v.Package] = true
	}
	for _, pkg := range sortedKeys(r.PerPackage) {
		b := r.PerPackage[pkg]
		switch {
		case b.Skipped:
			_, _ = fmt.Fprintf(w, "SKIP: %s\n", pkg)
		case violations[pkg]:
			_, _ = fmt.Fprintf(w, "FAIL: %s = %.1f%%\n", pkg, b.Percent())
		default:
			_, _ = fmt.Fprintf(w, "PASS: %s = %.1f%%\n", pkg, b.Percent())
		}
	}
	_, _ = fmt.Fprintf(w, "\nTOTAL: %.1f%%\n", r.Total)

	if len(r.Violations) > 0 {
		fmt.Fprintln(os.Stderr, "\nERROR: Coverage floors not met for:")
		for _, v := range r.Violations {
			fmt.Fprintf(os.Stderr, "  - %s (%.1f%% < %.0f%%)\n", v.Package, v.Got, v.Need)
		}
		os.Exit(1)
	}
	return nil
}

// resolveModule returns the module prefix: uses flag value if set, otherwise
// auto-detects from go.mod in the current directory.
func resolveModule(flag string) (string, error) {
	if flag != "" {
		if flag[len(flag)-1] != '/' {
			return flag + "/", nil
		}
		return flag, nil
	}
	return coverage.DetectModulePrefix("go.mod")
}
