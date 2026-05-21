package coverage

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	rootlineModulePrefix = "github.com/pablontiv/rootline/"
	fixtureProfile       = "testdata/coverage.out"
	fixtureFloors        = "testdata/floors.toml"
)

func approxEqual(a, b, tol float64) bool {
	return math.Abs(a-b) <= tol
}

// TestParseProfilePerPackage verifies that ParseProfile produces correct per-package
// statement counts matching the reference bash script output.
func TestParseProfilePerPackage(t *testing.T) {
	p, err := ParseProfile(fixtureProfile, rootlineModulePrefix)
	if err != nil {
		t.Fatalf("ParseProfile: %v", err)
	}

	perPkg := p.PerPackage()

	// internal/e2e is test-only: it never appears in the coverage profile.
	// PerPackage() only returns packages that have at least one line in the profile.
	if _, ok := perPkg["internal/e2e"]; ok {
		t.Error("internal/e2e should not appear in profile (test-only package)")
	}

	cases := []struct {
		pkg     string
		wantPct float64
	}{
		{"cmd/rootline", 85.3},
		{"internal/derive", 93.6},
		{"internal/extract", 95.2},
		{"internal/fix", 88.7},
		{"internal/fuzzy", 100.0},
		{"internal/graph", 88.1},
		{"internal/index", 88.8},
		{"internal/infer", 94.7},
		{"internal/migrate", 86.2},
		{"internal/proposal", 88.6},
		{"internal/query", 88.7},
		{"internal/rules", 88.8},
		{"internal/templates", 87.1},
	}

	for _, tc := range cases {
		b, ok := perPkg[tc.pkg]
		if !ok {
			t.Errorf("package %q missing from profile", tc.pkg)
			continue
		}
		if b.Skipped {
			t.Errorf("package %q: unexpected Skipped=true", tc.pkg)
			continue
		}
		got := b.Percent()
		if !approxEqual(got, tc.wantPct, 0.15) {
			t.Errorf("package %q: got %.1f%%, want %.1f%%", tc.pkg, got, tc.wantPct)
		}
	}
}

// TestProfileTotal verifies the aggregate coverage matches the reference output (88.9%).
func TestProfileTotal(t *testing.T) {
	p, err := ParseProfile(fixtureProfile, rootlineModulePrefix)
	if err != nil {
		t.Fatalf("ParseProfile: %v", err)
	}
	total := p.Total()
	if !approxEqual(total, 88.9, 0.5) {
		t.Errorf("Total: got %.1f%%, want ~88.9%%", total)
	}
}

// TestCheckPassesWithRootlineFloors verifies that Check produces no violations on the
// rootline fixtures.
func TestCheckPassesWithRootlineFloors(t *testing.T) {
	p, err := ParseProfile(fixtureProfile, rootlineModulePrefix)
	if err != nil {
		t.Fatalf("ParseProfile: %v", err)
	}
	f, err := LoadFloors(fixtureFloors)
	if err != nil {
		t.Fatalf("LoadFloors: %v", err)
	}

	r := Check(p, f)

	if len(r.Violations) != 0 {
		t.Errorf("expected no violations, got: %v", r.Violations)
	}
	// internal/e2e is test-only and must appear in SkippedPackages
	found := false
	for _, pkg := range r.SkippedPackages {
		if pkg == "internal/e2e" {
			found = true
		}
	}
	if !found {
		t.Errorf("internal/e2e not in SkippedPackages: %v", r.SkippedPackages)
	}
}

// TestCheckFailsWithArtificialViolation verifies that a threshold of 99 forces a failure.
func TestCheckFailsWithArtificialViolation(t *testing.T) {
	p, err := ParseProfile(fixtureProfile, rootlineModulePrefix)
	if err != nil {
		t.Fatalf("ParseProfile: %v", err)
	}
	// Use a real floors config but override threshold to force violations
	f := &Floors{
		Default:  99,
		Packages: []string{"cmd/rootline", "internal/derive", "internal/extract"},
	}

	r := Check(p, f)

	if len(r.Violations) == 0 {
		t.Error("expected violations with threshold=99, got none")
	}
	for _, v := range r.Violations {
		if v.Need != 99 {
			t.Errorf("violation threshold mismatch: got %.0f, want 99", v.Need)
		}
		if v.Got >= 99 {
			t.Errorf("violation %q: Got=%.1f%% which is not below 99%%", v.Package, v.Got)
		}
	}
}

// TestParseProfileErrors verifies error paths in ParseProfile.
func TestParseProfileErrors(t *testing.T) {
	_, err := ParseProfile("nonexistent.out", "github.com/x/y/")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// TestLoadFloorsErrors verifies validation and error paths in LoadFloors.
func TestLoadFloorsErrors(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		name    string
		content string
	}{
		{"nonexistent", ""},
		{"zero_default", "default = 0\npackages = [\"cmd/foo\"]\n"},
		{"empty_packages", "default = 85\npackages = []\n"},
		{"invalid_toml", "default = [\n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var path string
			if tc.name == "nonexistent" {
				path = filepath.Join(dir, "nope.toml")
			} else {
				path = filepath.Join(dir, tc.name+".toml")
				if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := LoadFloors(path)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

// TestBucketPercent verifies Bucket.Percent edge cases.
func TestBucketPercent(t *testing.T) {
	skipped := Bucket{Skipped: true}
	if skipped.Percent() != 0 {
		t.Errorf("skipped bucket: got %.1f, want 0", skipped.Percent())
	}
	zeroTotal := Bucket{Covered: 0, Total: 0}
	if zeroTotal.Percent() != 0 {
		t.Errorf("zero total bucket: got %.1f, want 0", zeroTotal.Percent())
	}
	full := Bucket{Covered: 10, Total: 10}
	if full.Percent() != 100 {
		t.Errorf("full bucket: got %.1f, want 100", full.Percent())
	}
}

// TestDetectModulePrefix verifies prefix extraction from a valid go.mod.
func TestDetectModulePrefix(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(gomod, []byte("module github.com/example/myrepo\n\ngo 1.21\n"), 0600); err != nil {
		t.Fatal(err)
	}

	prefix, err := DetectModulePrefix(gomod)
	if err != nil {
		t.Fatalf("DetectModulePrefix: %v", err)
	}
	if prefix != "github.com/example/myrepo/" {
		t.Errorf("got %q, want %q", prefix, "github.com/example/myrepo/")
	}
}

// TestDetectModulePrefixNotFound verifies error on missing file.
func TestDetectModulePrefixNotFound(t *testing.T) {
	_, err := DetectModulePrefix("/nonexistent/go.mod")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// TestDetectModulePrefixRejectsInvalid verifies that malformed go.mod files are rejected.
func TestDetectModulePrefixRejectsInvalid(t *testing.T) {
	dir := t.TempDir()

	cases := []struct{ name, content string }{
		{"no_module_directive", "go 1.21\n"},
		{"empty_module", "module \n"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gomod := filepath.Join(dir, tc.name+"_go.mod")
			if err := os.WriteFile(gomod, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			_, err := DetectModulePrefix(gomod)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

// TestEquivalence verifies that Check on the rootline fixtures produces the same
// PASS/FAIL/SKIP/TOTAL lines as the reference bash script.
func TestEquivalence(t *testing.T) {
	p, err := ParseProfile(fixtureProfile, rootlineModulePrefix)
	if err != nil {
		t.Fatalf("ParseProfile: %v", err)
	}
	f, err := LoadFloors(fixtureFloors)
	if err != nil {
		t.Fatalf("LoadFloors: %v", err)
	}

	r := Check(p, f)

	// Build a map of expected results from the reference bash output
	expected := map[string]string{
		"cmd/rootline":      "PASS",
		"internal/derive":   "PASS",
		"internal/e2e":      "SKIP",
		"internal/extract":  "PASS",
		"internal/fix":      "PASS",
		"internal/fuzzy":    "PASS",
		"internal/graph":    "PASS",
		"internal/index":    "PASS",
		"internal/infer":    "PASS",
		"internal/migrate":  "PASS",
		"internal/proposal": "PASS",
		"internal/query":    "PASS",
		"internal/rules":    "PASS",
		"internal/templates": "PASS",
	}

	// Build result from Go Check output
	violations := make(map[string]bool)
	for _, v := range r.Violations {
		violations[v.Package] = true
	}
	skipped := make(map[string]bool)
	for _, pkg := range r.SkippedPackages {
		skipped[pkg] = true
	}

	for pkg, wantStatus := range expected {
		var gotStatus string
		switch {
		case skipped[pkg]:
			gotStatus = "SKIP"
		case violations[pkg]:
			gotStatus = "FAIL"
		default:
			gotStatus = "PASS"
		}
		if gotStatus != wantStatus {
			t.Errorf("package %q: got %s, want %s", pkg, gotStatus, wantStatus)
		}
	}

	// Verify total coverage is close to bash reference (88.9%)
	total := p.Total()
	if !approxEqual(total, 88.9, 0.5) {
		t.Errorf("TOTAL: got %.1f%%, want ~88.9%%", total)
	}

	// Print equivalence lines for external diff (used by the AC grep comparison)
	for _, pkg := range f.Packages {
		b := r.PerPackage[pkg]
		switch {
		case b.Skipped:
			fmt.Printf("SKIP: %s\n", pkg)
		case violations[pkg]:
			fmt.Printf("FAIL: %s = %.1f%%\n", pkg, b.Percent())
		default:
			fmt.Printf("PASS: %s = %.1f%%\n", pkg, b.Percent())
		}
	}
	fmt.Printf("TOTAL: %.1f%%\n", total)

	// No violations expected
	if len(r.Violations) > 0 {
		var names []string
		for _, v := range r.Violations {
			names = append(names, v.Package)
		}
		t.Errorf("unexpected failures: %s", strings.Join(names, ", "))
	}
}
