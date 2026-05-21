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
	if !strings.Contains(out, "v1.0") {
		t.Errorf("version output missing v1.0:\n%s", out)
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
