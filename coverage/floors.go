package coverage

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
	"os"
)

// Floors holds the coverage policy loaded from .coverage-floors.toml.
type Floors struct {
	Default  int      `toml:"default"`
	Packages []string `toml:"packages"`
}

// LoadFloors parses a .coverage-floors.toml file and validates required fields.
func LoadFloors(tomlPath string) (*Floors, error) {
	data, err := os.ReadFile(tomlPath)
	if err != nil {
		return nil, fmt.Errorf("read floors: %w", err)
	}
	var f Floors
	if err := toml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse floors: %w", err)
	}
	if f.Default <= 0 {
		return nil, fmt.Errorf("floors: default must be a positive integer, got %d", f.Default)
	}
	if len(f.Packages) == 0 {
		return nil, fmt.Errorf("floors: packages list is empty")
	}
	return &f, nil
}

// Violation records a package that fell below the required threshold.
type Violation struct {
	Package string
	Got     float64
	Need    float64
}

// Result holds the output of a Check call.
type Result struct {
	Total           float64
	PerPackage      map[string]Bucket
	Violations      []Violation
	SkippedPackages []string
	MissingPackages []string
}

// Check evaluates a Profile against Floors and returns a Result.
// A package listed in floors but absent from the profile is reported in MissingPackages.
// A package with no source statements (Skipped==true) is reported in SkippedPackages.
func Check(p *Profile, f *Floors) Result {
	perPkg := p.PerPackage()
	r := Result{
		Total:      p.Total(),
		PerPackage: perPkg,
	}

	threshold := float64(f.Default)

	for _, pkg := range f.Packages {
		b, ok := perPkg[pkg]
		if !ok {
			// Package has no entries in the coverage profile: no source statements.
			// Matches bash reference behavior (awk total==0 → SKIP).
			r.SkippedPackages = append(r.SkippedPackages, pkg)
			continue
		}
		if b.Skipped {
			r.SkippedPackages = append(r.SkippedPackages, pkg)
			continue
		}
		pct := b.Percent()
		if pct < threshold {
			r.Violations = append(r.Violations, Violation{
				Package: pkg,
				Got:     pct,
				Need:    threshold,
			})
		}
	}

	return r
}
