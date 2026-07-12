package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"sort"

	"github.com/pablontiv/picokit/coverage"
)

type reportOptions struct {
	profile string
	module  string
	output  string
}

func runReportCmd(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := reportOptions{}
	fs.StringVar(&opts.profile, "profile", "coverage.out", "path to coverage profile")
	fs.StringVar(&opts.module, "module", "", "module prefix (auto-detected from go.mod if empty)")
	out := addOutputFlag(fs)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	opts.output = *out
	if err := runReport(opts, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func runReport(opts reportOptions, stdout io.Writer) error {
	prefix, err := resolveModule(opts.module)
	if err != nil {
		return err
	}

	p, err := coverage.ParseProfile(opts.profile, prefix)
	if err != nil {
		return fmt.Errorf("parse profile: %w", err)
	}

	perPkg := p.PerPackage()
	pkgs := sortedKeys(perPkg)

	if opts.output == "json" {
		type pkgEntry struct {
			Covered int     `json:"covered"`
			Total   int     `json:"total"`
			Skipped bool    `json:"skipped"`
			Percent float64 `json:"percent"`
		}
		jsonOut := struct {
			Total      float64             `json:"total"`
			PerPackage map[string]pkgEntry `json:"per_package"`
		}{
			Total:      p.Total(),
			PerPackage: make(map[string]pkgEntry, len(perPkg)),
		}
		for k, b := range perPkg {
			jsonOut.PerPackage[k] = pkgEntry{
				Covered: b.Covered,
				Total:   b.Total,
				Skipped: b.Skipped,
				Percent: b.Percent(),
			}
		}
		return json.NewEncoder(stdout).Encode(jsonOut)
	}

	out := stdout
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
