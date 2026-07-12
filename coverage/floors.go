package coverage

import (
	"fmt"
	"os"
)

// Floors holds the coverage policy loaded from .coverage-floors.toml.
type Floors struct {
	Default int `toml:"default"`
	// Deprecated: Packages is accepted for backward compatibility with v1.0 configs
	// but its content is ignored in v1.1. The gate applies to all packages discovered
	// in the coverage profile. Use Exclude to opt out specific packages.
	Packages []string `toml:"packages,omitempty"`
	Exclude  []string `toml:"exclude,omitempty"`
}

// LoadFloors parses a .coverage-floors.toml file and validates required fields.
func LoadFloors(tomlPath string) (*Floors, error) {
	data, err := os.ReadFile(tomlPath)
	if err != nil {
		return nil, fmt.Errorf("read floors: %w", err)
	}
	f, err := parseFloorsTOML(data)
	if err != nil {
		return nil, fmt.Errorf("parse floors: %w", err)
	}
	if f.Default <= 0 {
		return nil, fmt.Errorf("floors: default must be a positive integer, got %d", f.Default)
	}
	return f, nil
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
// The gate applies to every package discovered in the coverage profile, except
// those listed in f.Exclude. Packages with no source statements (Skipped==true)
// are reported in SkippedPackages.
func Check(p *Profile, f *Floors) Result {
	perPkg := p.PerPackage()
	r := Result{
		Total:      p.Total(),
		PerPackage: perPkg,
	}

	threshold := float64(f.Default)

	excluded := make(map[string]bool, len(f.Exclude))
	for _, pkg := range f.Exclude {
		excluded[pkg] = true
	}

	for pkg, b := range perPkg {
		if excluded[pkg] {
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
