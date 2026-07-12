package main

import (
	"flag"
	"fmt"
	"io"
)

// specVersion is the coverage-spec version this build implements.
const specVersion = "v1.1"

// buildVersion is set at build time via -ldflags.
var buildVersion = "dev"

const usage = `Usage: pkcov <command> [flags]

Commands:
  check     Check coverage against floors; exits 1 on violations
  report    Print per-package coverage table
  version   Print pkcov version and implemented spec version

Common flags:
  -o, --output string   output format (text|json) (default "text")

Run "pkcov <command> --help" for command flags.
`

// run dispatches to a subcommand and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "check":
		return runCheckCmd(args[1:], stdout, stderr)
	case "report":
		return runReportCmd(args[1:], stdout, stderr)
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "pkcov %s (coverage-spec %s)\n", buildVersion, specVersion)
		return 0
	case "help", "--help", "-h":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "pkcov: unknown command %q\n\n%s", args[0], usage)
		return 2
	}
}

// addOutputFlag registers -o/--output on fs and returns the bound value.
func addOutputFlag(fs *flag.FlagSet) *string {
	out := fs.String("output", "text", "output format (text|json)")
	fs.StringVar(out, "o", "text", "output format (text|json)")
	return out
}
