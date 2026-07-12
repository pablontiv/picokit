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

func TestReportText(t *testing.T) {
	buf := &bytes.Buffer{}
	err := runReport(reportOptions{profile: testProfile, module: testModule, output: "text"}, buf)
	if err != nil {
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

func TestCheckPasses(t *testing.T) {
	buf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	code, err := runCheck(checkOptions{profile: testProfile, floors: testFloors, module: testModule, output: "text"}, buf, errBuf)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if code != 0 {
		t.Fatalf("check exit code = %d, want 0; stderr:\n%s", code, errBuf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "PASS:") {
		t.Errorf("check output missing PASS lines:\n%s", out)
	}
	if strings.Contains(out, "FAIL:") {
		t.Errorf("check output contains unexpected FAIL:\n%s", out)
	}
}

func TestCheckJSON(t *testing.T) {
	buf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	code, err := runCheck(checkOptions{profile: testProfile, floors: testFloors, module: testModule, output: "json"}, buf, errBuf)
	if err != nil {
		t.Fatalf("check json: %v", err)
	}
	if code != 0 {
		t.Fatalf("check exit code = %d, want 0", code)
	}
	var out struct {
		Total      float64                `json:"total"`
		PerPackage map[string]interface{} `json:"per_package"`
		Violations []interface{}          `json:"violations"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if out.Total <= 0 {
		t.Errorf("total = %v, want > 0", out.Total)
	}
}

func TestRunDispatch(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode int
		wantOut  string
	}{
		{"version subcommand", []string{"version"}, 0, "coverage-spec"},
		{"version flag", []string{"--version"}, 0, "coverage-spec"},
		{"help", []string{"--help"}, 0, "Usage: pkcov"},
		{"no args", []string{}, 2, ""},
		{"unknown command", []string{"bogus"}, 2, ""},
		{"check via dispatch", []string{"check", "--profile", testProfile, "--floors", testFloors, "--module", testModule}, 0, "PASS:"},
		{"check usage error", []string{"check", "--nope"}, 2, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
			code := run(tc.args, stdout, stderr)
			if code != tc.wantCode {
				t.Fatalf("run(%v) = %d, want %d; stderr:\n%s", tc.args, code, tc.wantCode, stderr.String())
			}
			if tc.wantOut != "" && !strings.Contains(stdout.String(), tc.wantOut) {
				t.Errorf("stdout missing %q:\n%s", tc.wantOut, stdout.String())
			}
		})
	}
}

// TestCheckV11AutoDiscovery verifies that check with a minimal v1.1 config (no packages)
// outputs one line per profile package sorted alphabetically.
func TestCheckV11AutoDiscovery(t *testing.T) {
	dir := t.TempDir()
	floorsPath := dir + "/floors.toml"
	if err := os.WriteFile(floorsPath, []byte("default = 85\n"), 0600); err != nil {
		t.Fatal(err)
	}

	buf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	code, err := runCheck(checkOptions{profile: testProfile, floors: floorsPath, module: testModule, output: "text"}, buf, errBuf)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if code != 0 {
		t.Fatalf("check exit code = %d, want 0; stderr:\n%s", code, errBuf.String())
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

	runCheckOpts := func(floors string) string {
		buf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
		code, err := runCheck(checkOptions{profile: testProfile, floors: floors, module: testModule, output: "text"}, buf, errBuf)
		if err != nil {
			t.Fatalf("check(%s): %v", floors, err)
		}
		if code != 0 {
			t.Fatalf("check(%s) exit code = %d, want 0; stderr:\n%s", floors, code, errBuf.String())
		}
		return buf.String()
	}

	outLegacy := runCheckOpts(legacyPath)
	outMinimal := runCheckOpts(minimalPath)

	if outLegacy != outMinimal {
		t.Errorf("legacy and minimal configs produced different output\nlegacy:\n%s\nminimal:\n%s",
			outLegacy, outMinimal)
	}
}

// TestReportJSON verifies that report --output json produces parseable JSON with total and per_package.
func TestReportJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	err := runReport(reportOptions{profile: testProfile, module: testModule, output: "json"}, buf)
	if err != nil {
		t.Fatalf("report json: %v", err)
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
	buf := &bytes.Buffer{}
	err := runReport(reportOptions{profile: "testdata/skipped.out", module: "github.com/pablontiv/picokit", output: "text"}, buf)
	if err != nil {
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
	buf := &bytes.Buffer{}
	err := runReport(reportOptions{profile: "nonexistent_profile.out", module: testModule, output: "text"}, buf)
	if err == nil {
		t.Error("expected error for nonexistent profile, got nil")
	}
}

// TestReportErrorBadModule verifies that empty module with no go.mod in CWD returns an error.
func TestReportErrorBadModule(t *testing.T) {
	buf := &bytes.Buffer{}
	err := runReport(reportOptions{profile: testProfile, module: "", output: "text"}, buf)
	if err == nil {
		t.Error("expected error for auto-detect with no go.mod, got nil")
	}
}

// TestCheckErrorBadModule verifies that empty module with no go.mod in CWD returns an error.
func TestCheckErrorBadModule(t *testing.T) {
	buf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	_, err := runCheck(checkOptions{profile: testProfile, floors: testFloors, module: "", output: "text"}, buf, errBuf)
	if err == nil {
		t.Error("expected error for auto-detect with no go.mod, got nil")
	}
}

// TestCheckErrorBadProfile verifies that a nonexistent profile path returns an error.
func TestCheckErrorBadProfile(t *testing.T) {
	buf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	_, err := runCheck(checkOptions{profile: "nonexistent_profile.out", floors: testFloors, module: testModule, output: "text"}, buf, errBuf)
	if err == nil {
		t.Error("expected error for nonexistent profile, got nil")
	}
}

// TestCheckErrorBadFloors verifies that an invalid floors file returns an error.
func TestCheckErrorBadFloors(t *testing.T) {
	buf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	_, err := runCheck(checkOptions{profile: testProfile, floors: "/nonexistent/.coverage-floors.toml", module: testModule, output: "text"}, buf, errBuf)
	if err == nil {
		t.Error("expected error for nonexistent floors, got nil")
	}
}

// TestCheckTextExcluded verifies that excluded packages do not appear in text output.
func TestCheckTextExcluded(t *testing.T) {
	dir := t.TempDir()
	floorsPath := dir + "/floors.toml"
	if err := os.WriteFile(floorsPath, []byte("default = 85\nexclude = [\"internal/fuzzy\"]\n"), 0600); err != nil {
		t.Fatal(err)
	}

	buf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	code, err := runCheck(checkOptions{profile: testProfile, floors: floorsPath, module: testModule, output: "text"}, buf, errBuf)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if code != 0 {
		t.Fatalf("check exit code = %d, want 0; stderr:\n%s", code, errBuf.String())
	}
	out := buf.String()
	if strings.Contains(out, "internal/fuzzy") {
		t.Errorf("excluded package internal/fuzzy should not appear in output:\n%s", out)
	}
}

// TestCheckJSONExcluded verifies that excluded packages are omitted from per_package in JSON output.
func TestCheckJSONExcluded(t *testing.T) {
	dir := t.TempDir()
	floorsPath := dir + "/floors.toml"
	if err := os.WriteFile(floorsPath, []byte("default = 85\nexclude = [\"internal/fuzzy\"]\n"), 0600); err != nil {
		t.Fatal(err)
	}

	buf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	code, err := runCheck(checkOptions{profile: testProfile, floors: floorsPath, module: testModule, output: "json"}, buf, errBuf)
	if err != nil {
		t.Fatalf("check json with exclude: %v", err)
	}
	if code != 0 {
		t.Fatalf("check exit code = %d, want 0", code)
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
	dir := t.TempDir()
	floorsPath := dir + "/floors.toml"
	if err := os.WriteFile(floorsPath, []byte("default = 85\n"), 0600); err != nil {
		t.Fatal(err)
	}

	buf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
	code, err := runCheck(checkOptions{profile: "testdata/skipped.out", floors: floorsPath, module: "github.com/pablontiv/picokit", output: "text"}, buf, errBuf)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if code != 0 {
		t.Fatalf("check exit code = %d, want 0; stderr:\n%s", code, errBuf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "SKIP: testpkg") {
		t.Errorf("expected SKIP line for testpkg:\n%s", out)
	}
	if !strings.Contains(out, "PASS: real") {
		t.Errorf("expected PASS line for real:\n%s", out)
	}
}
