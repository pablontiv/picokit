package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/pablontiv/picokit/coverage"
)

type checkOptions struct {
	profile string
	floors  string
	module  string
	output  string
}

func runCheckCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := checkOptions{}
	fs.StringVar(&opts.profile, "profile", "coverage.out", "path to coverage profile")
	fs.StringVar(&opts.floors, "floors", ".coverage-floors.toml", "path to floors config")
	fs.StringVar(&opts.module, "module", "", "module prefix (auto-detected from go.mod if empty)")
	out := addOutputFlag(fs)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	opts.output = *out
	code, err := runCheck(opts, stdout, stderr)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return code
}

// runCheck evaluates the coverage gate. It returns the process exit code
// (0 pass, 1 violations) and a non-nil error only for operational failures.
func runCheck(opts checkOptions, stdout, stderr io.Writer) (int, error) {
	prefix, err := resolveModule(opts.module)
	if err != nil {
		return 1, err
	}

	p, err := coverage.ParseProfile(opts.profile, prefix)
	if err != nil {
		return 1, fmt.Errorf("parse profile: %w", err)
	}
	f, err := coverage.LoadFloors(opts.floors)
	if err != nil {
		return 1, fmt.Errorf("load floors: %w", err)
	}

	r := coverage.Check(p, f)

	excluded := make(map[string]bool, len(f.Exclude))
	for _, pkg := range f.Exclude {
		excluded[pkg] = true
	}

	if opts.output == "json" {
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
			Excluded   []string            `json:"excluded"`
		}{
			Total:      r.Total,
			PerPackage: make(map[string]pkgEntry, len(r.PerPackage)),
			Violations: make([]violation, 0, len(r.Violations)),
			Skipped:    r.SkippedPackages,
			Excluded:   f.Exclude,
		}
		if out.Skipped == nil {
			out.Skipped = []string{}
		}
		if out.Excluded == nil {
			out.Excluded = []string{}
		}
		for k, b := range r.PerPackage {
			if excluded[k] {
				continue
			}
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
		if err := json.NewEncoder(stdout).Encode(out); err != nil {
			return 1, err
		}
		if len(r.Violations) > 0 {
			return 1, nil
		}
		return 0, nil
	}

	// text output
	w := stdout
	_, _ = fmt.Fprintln(w, "Coverage Report:")
	_, _ = fmt.Fprintln(w, "================")
	violations := make(map[string]bool, len(r.Violations))
	for _, v := range r.Violations {
		violations[v.Package] = true
	}
	for _, pkg := range sortedKeys(r.PerPackage) {
		if excluded[pkg] {
			continue
		}
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
		fmt.Fprintln(stderr, "\nERROR: Coverage floors not met for:")
		for _, v := range r.Violations {
			fmt.Fprintf(stderr, "  - %s (%.1f%% < %.0f%%)\n", v.Package, v.Got, v.Need)
		}
		return 1, nil
	}
	return 0, nil
}

// resolveModule returns the module prefix: uses flag value if set, otherwise
// auto-detects from go.mod in the current directory.
func resolveModule(flagVal string) (string, error) {
	if flagVal != "" {
		if flagVal[len(flagVal)-1] != '/' {
			return flagVal + "/", nil
		}
		return flagVal, nil
	}
	return coverage.DetectModulePrefix("go.mod")
}
