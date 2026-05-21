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
	for _, field := range []string{"total", "per_package", "violations", "skipped"} {
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
