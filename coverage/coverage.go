// Package coverage parses Go coverage profiles and checks per-package floors.
package coverage

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Bucket holds statement counts for one package.
type Bucket struct {
	Covered int
	Total   int
	Skipped bool // true when no source statements exist (test-only package)
}

// Percent returns statement coverage in [0,100]. Returns 0 for skipped buckets.
func (b Bucket) Percent() float64 {
	if b.Skipped || b.Total == 0 {
		return 0
	}
	return 100 * float64(b.Covered) / float64(b.Total)
}

// Profile holds per-package coverage buckets parsed from a coverage profile.
type Profile struct {
	Packages map[string]Bucket
}

// ParseProfile reads a Go coverage profile produced by `go test -coverprofile`.
// modulePrefix is the module path with a trailing slash (e.g. "github.com/foo/bar/").
// Each package key is the suffix after modulePrefix (e.g. "internal/core").
func ParseProfile(profilePath, modulePrefix string) (*Profile, error) {
	f, err := os.Open(profilePath)
	if err != nil {
		return nil, fmt.Errorf("open profile: %w", err)
	}
	defer func() { _ = f.Close() }()

	// key: package suffix (relative to modulePrefix), value: accumulated bucket
	raw := make(map[string]Bucket)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		// format: <file>:<start>,<end> <stmts> <count>
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}
		filePart := parts[0] // e.g. github.com/foo/bar/internal/core/foo.go:12.3,14.5
		stmts, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		count, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}

		// strip file path: split on ":" to get the import path portion
		colonIdx := strings.LastIndex(filePart, ":")
		if colonIdx < 0 {
			continue
		}
		importPath := filePart[:colonIdx] // e.g. github.com/foo/bar/internal/core/foo.go

		if !strings.HasPrefix(importPath, modulePrefix) {
			continue
		}
		// drop the module prefix to get a relative file path, then take the directory
		rel := importPath[len(modulePrefix):]
		slashIdx := strings.LastIndex(rel, "/")
		var pkg string
		if slashIdx < 0 {
			pkg = "." // top-level package
		} else {
			pkg = rel[:slashIdx]
		}

		b := raw[pkg]
		b.Total += stmts
		if count > 0 {
			b.Covered += stmts
		}
		raw[pkg] = b
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan profile: %w", err)
	}

	return &Profile{Packages: raw}, nil
}

// Total returns aggregate statement coverage across all packages in the profile.
func (p *Profile) Total() float64 {
	var covered, total int
	for _, b := range p.Packages {
		covered += b.Covered
		total += b.Total
	}
	if total == 0 {
		return 0
	}
	return 100 * float64(covered) / float64(total)
}

// PerPackage returns the per-package bucket map. Packages with Total==0 are
// marked as Skipped. The returned map is a shallow copy.
func (p *Profile) PerPackage() map[string]Bucket {
	out := make(map[string]Bucket, len(p.Packages))
	for k, b := range p.Packages {
		if b.Total == 0 {
			b.Skipped = true
		}
		out[k] = b
	}
	return out
}
