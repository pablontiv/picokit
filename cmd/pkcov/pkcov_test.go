package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

const (
	testProfile = "../../coverage/testdata/coverage.out"
	testFloors  = "../../coverage/testdata/floors.toml"
	testModule  = "github.com/pablontiv/rootline"
)

func resetFlags() {
	outputFormat = "text"
}

// TestReportText verifies that report produces PASS/SKIP/TOTAL lines.
func TestReportText(t *testing.T) {
	resetFlags()
	reportProfile = testProfile
	reportModule = testModule

	buf := &bytes.Buffer{}
	reportCmd.SetOut(buf)
	if err := reportCmd.RunE(reportCmd, nil); err != nil {
		t.Fatalf("report: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "PASS:") {
		t.Errorf("report output missing PASS lines:\n%s", out)
	}
	if !strings.Contains(out, "TOTAL:") {
		t.Errorf("report output missing TOTAL line:\n%s", out)
	}
}

// TestCheckPasses verifies that check exits cleanly on valid fixtures.
func TestCheckPasses(t *testing.T) {
	resetFlags()
	checkProfile = testProfile
	checkFloors = testFloors
	checkModule = testModule

	buf := &bytes.Buffer{}
	checkCmd.SetOut(buf)
	if err := checkCmd.RunE(checkCmd, nil); err != nil {
		t.Fatalf("check: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "PASS:") {
		t.Errorf("check output missing PASS lines:\n%s", out)
	}
	if strings.Contains(out, "FAIL:") {
		t.Errorf("check output contains unexpected FAIL:\n%s", out)
	}
}

// TestCheckJSON verifies --output json produces parseable output with required fields.
func TestCheckJSON(t *testing.T) {
	resetFlags()
	outputFormat = "text" // will be overridden below
	checkProfile = testProfile
	checkFloors = testFloors
	checkModule = testModule

	// Redirect os.Stdout temporarily for JSON output
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	outputFormat = "json"
	runErr := checkCmd.RunE(checkCmd, nil)
	_ = w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}

	if runErr != nil {
		t.Fatalf("check json: %v", runErr)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("json unmarshal: %v\noutput: %s", err, buf.String())
	}
	for _, field := range []string{"total", "per_package", "violations", "skipped", "excluded"} {
		if _, ok := result[field]; !ok {
			t.Errorf("JSON missing field %q", field)
		}
	}
}

// TestVersionCmd verifies version output contains spec version.
func TestVersionCmd(t *testing.T) {
	buf := &bytes.Buffer{}
	versionCmd.SetOut(buf)
	versionCmd.Run(versionCmd, nil)
	out := buf.String()
	if !strings.Contains(out, "coverage-spec") {
		t.Errorf("version output missing spec version:\n%s", out)
	}
	if !strings.Contains(out, "v1.1") {
		t.Errorf("version output missing v1.1:\n%s", out)
	}
}

// TestResolveModuleFlag verifies resolveModule appends trailing slash.
func TestResolveModuleFlag(t *testing.T) {
	prefix, err := resolveModule("github.com/foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	if prefix != "github.com/foo/bar/" {
		t.Errorf("got %q, want %q", prefix, "github.com/foo/bar/")
	}

	// Already has slash
	prefix2, err := resolveModule("github.com/foo/bar/")
	if err != nil {
		t.Fatal(err)
	}
	if prefix2 != "github.com/foo/bar/" {
		t.Errorf("got %q, want %q", prefix2, "github.com/foo/bar/")
	}
}

// TestCheckV11AutoDiscovery verifies that check with a minimal v1.1 config (no packages)
// outputs one line per profile package sorted alphabetically.
func TestCheckV11AutoDiscovery(t *testing.T) {
	resetFlags()

	dir := t.TempDir()
	floorsPath := dir + "/floors.toml"
	if err := os.WriteFile(floorsPath, []byte("default = 85\n"), 0600); err != nil {
		t.Fatal(err)
	}

	checkProfile = testProfile
	checkFloors = floorsPath
	checkModule = testModule

	buf := &bytes.Buffer{}
	checkCmd.SetOut(buf)
	if err := checkCmd.RunE(checkCmd, nil); err != nil {
		t.Fatalf("check: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "PASS:") {
		t.Errorf("expected PASS lines in output:\n%s", out)
	}
	if strings.Contains(out, "FAIL:") {
		t.Errorf("unexpected FAIL in output:\n%s", out)
	}
	if !strings.Contains(out, "TOTAL:") {
		t.Errorf("expected TOTAL line in output:\n%s", out)
	}

	// Output must be sorted: verify each PASS/SKIP line comes after the previous one.
	var prev string
	for _, line := range strings.Split(out, "\n") {
		if !strings.HasPrefix(line, "PASS:") && !strings.HasPrefix(line, "SKIP:") && !strings.HasPrefix(line, "FAIL:") {
			continue
		}
		pkg := strings.Fields(line)[1]
		if prev != "" && pkg < prev {
			t.Errorf("output not sorted: %q came after %q", pkg, prev)
		}
		prev = pkg
	}
}

// TestCheckV11LegacyFloors verifies that a v1.0 floors config (with packages field)
// loads without error and produces the same output as a minimal v1.1 config.
func TestCheckV11LegacyFloors(t *testing.T) {
	resetFlags()

	dir := t.TempDir()
	legacyPath := dir + "/legacy.toml"
	legacyContent := "default = 85\npackages = [\"cmd/rootline\", \"internal/derive\"]\n"
	if err := os.WriteFile(legacyPath, []byte(legacyContent), 0600); err != nil {
		t.Fatal(err)
	}
	minimalPath := dir + "/minimal.toml"
	if err := os.WriteFile(minimalPath, []byte("default = 85\n"), 0600); err != nil {
		t.Fatal(err)
	}

	runCheck := func(floors string) string {
		resetFlags()
		checkProfile = testProfile
		checkFloors = floors
		checkModule = testModule
		buf := &bytes.Buffer{}
		checkCmd.SetOut(buf)
		if err := checkCmd.RunE(checkCmd, nil); err != nil {
			t.Fatalf("check(%s): %v", floors, err)
		}
		return buf.String()
	}

	outLegacy := runCheck(legacyPath)
	outMinimal := runCheck(minimalPath)

	if outLegacy != outMinimal {
		t.Errorf("legacy and minimal configs produced different output\nlegacy:\n%s\nminimal:\n%s",
			outLegacy, outMinimal)
	}
}

// TestReportJSON verifies that report --output json produces parseable JSON with total and per_package.
func TestReportJSON(t *testing.T) {
	resetFlags()
	reportProfile = testProfile
	reportModule = testModule
	outputFormat = "json"

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	runErr := reportCmd.RunE(reportCmd, nil)
	_ = w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if runErr != nil {
		t.Fatalf("report json: %v", runErr)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("json unmarshal: %v\noutput: %s", err, buf.String())
	}
	for _, field := range []string{"total", "per_package"} {
		if _, ok := result[field]; !ok {
			t.Errorf("JSON missing field %q", field)
		}
	}
}

// TestReportTextSkipped verifies that a test-only package (Total==0 in profile) shows as SKIP.
func TestReportTextSkipped(t *testing.T) {
	resetFlags()
	reportProfile = "testdata/skipped.out"
	reportModule = "github.com/pablontiv/picokit"

	buf := &bytes.Buffer{}
	reportCmd.SetOut(buf)
	if err := reportCmd.RunE(reportCmd, nil); err != nil {
		t.Fatalf("report: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "SKIP: testpkg") {
		t.Errorf("expected SKIP line for testpkg:\n%s", out)
	}
	if !strings.Contains(out, "PASS: real") {
		t.Errorf("expected PASS line for real:\n%s", out)
	}
}

// TestReportErrorBadProfile verifies that a nonexistent profile returns an error.
func TestReportErrorBadProfile(t *testing.T) {
	resetFlags()
	reportProfile = "nonexistent_profile.out"
	reportModule = testModule

	if err := reportCmd.RunE(reportCmd, nil); err == nil {
		t.Error("expected error for nonexistent profile, got nil")
	}
}

// TestReportErrorBadModule verifies that empty module with no go.mod in CWD returns an error.
func TestReportErrorBadModule(t *testing.T) {
	resetFlags()
	reportProfile = testProfile
	reportModule = "" // triggers auto-detect; no go.mod in cmd/pkcov/

	if err := reportCmd.RunE(reportCmd, nil); err == nil {
		t.Error("expected error for auto-detect with no go.mod, got nil")
	}
}

// TestCheckErrorBadModule verifies that empty module with no go.mod in CWD returns an error.
func TestCheckErrorBadModule(t *testing.T) {
	resetFlags()
	checkProfile = testProfile
	checkFloors = testFloors
	checkModule = "" // triggers auto-detect; no go.mod in cmd/pkcov/

	if err := checkCmd.RunE(checkCmd, nil); err == nil {
		t.Error("expected error for auto-detect with no go.mod, got nil")
	}
}

// TestCheckErrorBadProfile verifies that a nonexistent profile path returns an error.
func TestCheckErrorBadProfile(t *testing.T) {
	resetFlags()
	checkProfile = "nonexistent_profile.out"
	checkFloors = testFloors
	checkModule = testModule

	if err := checkCmd.RunE(checkCmd, nil); err == nil {
		t.Error("expected error for nonexistent profile, got nil")
	}
}

// TestCheckErrorBadFloors verifies that an invalid floors file returns an error.
func TestCheckErrorBadFloors(t *testing.T) {
	resetFlags()
	checkProfile = testProfile
	checkFloors = "/nonexistent/.coverage-floors.toml"
	checkModule = testModule

	if err := checkCmd.RunE(checkCmd, nil); err == nil {
		t.Error("expected error for nonexistent floors, got nil")
	}
}

// TestCheckTextExcluded verifies that excluded packages do not appear in text output.
func TestCheckTextExcluded(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	floorsPath := dir + "/floors.toml"
	if err := os.WriteFile(floorsPath, []byte("default = 85\nexclude = [\"internal/fuzzy\"]\n"), 0600); err != nil {
		t.Fatal(err)
	}

	checkProfile = testProfile
	checkFloors = floorsPath
	checkModule = testModule

	buf := &bytes.Buffer{}
	checkCmd.SetOut(buf)
	if err := checkCmd.RunE(checkCmd, nil); err != nil {
		t.Fatalf("check: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "internal/fuzzy") {
		t.Errorf("excluded package internal/fuzzy should not appear in output:\n%s", out)
	}
}

// TestCheckJSONExcluded verifies that excluded packages are omitted from per_package in JSON output.
func TestCheckJSONExcluded(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	floorsPath := dir + "/floors.toml"
	if err := os.WriteFile(floorsPath, []byte("default = 85\nexclude = [\"internal/fuzzy\"]\n"), 0600); err != nil {
		t.Fatal(err)
	}

	checkProfile = testProfile
	checkFloors = floorsPath
	checkModule = testModule

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	outputFormat = "json"

	runErr := checkCmd.RunE(checkCmd, nil)
	_ = w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if runErr != nil {
		t.Fatalf("check json with exclude: %v", runErr)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("json unmarshal: %v\noutput: %s", err, buf.String())
	}
	perPkg, _ := result["per_package"].(map[string]any)
	if _, found := perPkg["internal/fuzzy"]; found {
		t.Error("excluded package internal/fuzzy must not appear in per_package")
	}
	excluded, _ := result["excluded"].([]any)
	found := false
	for _, e := range excluded {
		if e == "internal/fuzzy" {
			found = true
		}
	}
	if !found {
		t.Errorf("internal/fuzzy should appear in excluded list, got: %v", excluded)
	}
}

// TestCheckTextSkipped verifies that a test-only package (Total==0) shows as SKIP in check text output.
func TestCheckTextSkipped(t *testing.T) {
	resetFlags()
	dir := t.TempDir()
	floorsPath := dir + "/floors.toml"
	if err := os.WriteFile(floorsPath, []byte("default = 85\n"), 0600); err != nil {
		t.Fatal(err)
	}

	checkProfile = "testdata/skipped.out"
	checkFloors = floorsPath
	checkModule = "github.com/pablontiv/picokit"

	buf := &bytes.Buffer{}
	checkCmd.SetOut(buf)
	if err := checkCmd.RunE(checkCmd, nil); err != nil {
		t.Fatalf("check: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "SKIP: testpkg") {
		t.Errorf("expected SKIP line for testpkg:\n%s", out)
	}
	if !strings.Contains(out, "PASS: real") {
		t.Errorf("expected PASS line for real:\n%s", out)
	}
}
