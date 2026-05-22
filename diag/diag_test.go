package diag

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// Test Error layer
func TestErrorNew(t *testing.T) {
	err := New("ERR_FOO", "something failed")
	if err == nil {
		t.Fatal("New() returned nil")
	}
	if err.Code != "ERR_FOO" {
		t.Fatalf("Code = %q, want %q", err.Code, "ERR_FOO")
	}
	if err.Message != "something failed" {
		t.Fatalf("Message = %q, want %q", err.Message, "something failed")
	}
	if err.Cause != nil {
		t.Fatalf("Cause = %v, want nil", err.Cause)
	}
}

func TestErrorString(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "simple error without cause",
			err:  New("ERR_FOO", "algo falló"),
			want: "ERR_FOO: algo falló",
		},
		{
			name: "error with cause",
			err:  Wrap("ERR_BAR", "wrapped error", errors.New("root cause")),
			want: "ERR_BAR: wrapped error (root cause)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Fatalf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	rootErr := errors.New("root error")
	wrappedErr := Wrap("ERR_WRAPPED", "wrapper message", rootErr)

	if wrappedErr.Unwrap() != rootErr {
		t.Fatalf("Unwrap() did not return original error")
	}

	// New errors should have nil cause
	newErr := New("ERR_NEW", "new error")
	if newErr.Unwrap() != nil {
		t.Fatalf("Unwrap() of New() should return nil, got %v", newErr.Unwrap())
	}
}

func TestErrorImplementsError(t *testing.T) {
	var _ error = (*Error)(nil)
}

// Test Report layer
func TestNewReport(t *testing.T) {
	diagnostics := []Diagnostic{
		{ID: "TEST_001", Severity: SeverityInfo, Message: "info message"},
		{ID: "TEST_002", Severity: SeverityWarning, Message: "warning message"},
		{ID: "TEST_003", Severity: SeverityError, Message: "error message"},
	}

	report := NewReport("test/kind", "/root/path", diagnostics)

	if report.Version != 1 {
		t.Fatalf("Version = %d, want 1", report.Version)
	}
	if report.Kind != "test/kind" {
		t.Fatalf("Kind = %q, want %q", report.Kind, "test/kind")
	}
	if report.Root != "/root/path" {
		t.Fatalf("Root = %q, want %q", report.Root, "/root/path")
	}
	if len(report.Diagnostics) != 3 {
		t.Fatalf("Diagnostics len = %d, want 3", len(report.Diagnostics))
	}

	// Check summary
	if report.Summary.Status != SummaryStatusError {
		t.Fatalf("Summary.Status = %q, want %q", report.Summary.Status, SummaryStatusError)
	}
	if report.Summary.Errors != 1 {
		t.Fatalf("Summary.Errors = %d, want 1", report.Summary.Errors)
	}
	if report.Summary.Warnings != 1 {
		t.Fatalf("Summary.Warnings = %d, want 1", report.Summary.Warnings)
	}
	if report.Summary.Infos != 1 {
		t.Fatalf("Summary.Infos = %d, want 1", report.Summary.Infos)
	}
}

func TestRenderJSONProducesValidJSON(t *testing.T) {
	diagnostics := []Diagnostic{
		{ID: "TEST_001", Severity: SeverityError, Message: "error", Path: "docs/roadmap/test.md", Details: map[string]any{"key": "value"}},
	}
	report := NewReport("test/check", "/repo", diagnostics)

	var buf bytes.Buffer
	if err := RenderJSON(&buf, report); err != nil {
		t.Fatalf("RenderJSON error = %v", err)
	}

	// Verify JSON is valid
	if !json.Valid(buf.Bytes()) {
		t.Fatalf("RenderJSON produced invalid JSON:\n%s", buf.String())
	}

	// Verify content
	var decoded Report
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}
	if decoded.Version != 1 || decoded.Kind != "test/check" {
		t.Fatalf("decoded report = %#v", decoded)
	}
	if len(decoded.Diagnostics) != 1 || decoded.Diagnostics[0].ID != "TEST_001" {
		t.Fatalf("decoded diagnostics = %#v", decoded.Diagnostics)
	}
	if decoded.Diagnostics[0].Details == nil || decoded.Diagnostics[0].Details["key"] != "value" {
		t.Fatalf("decoded details = %#v", decoded.Diagnostics[0].Details)
	}
}

func TestRenderTextProducesReadableOutput(t *testing.T) {
	diagnostics := []Diagnostic{
		{ID: "TEST_INFO", Severity: SeverityInfo, Message: "info message"},
		{ID: "TEST_WARN", Severity: SeverityWarning, Message: "warning message", Path: "docs/roadmap/test.md"},
		{ID: "TEST_ERR", Severity: SeverityError, Message: "error message"},
	}
	report := NewReport("test/doctor", "/repo", diagnostics)

	var buf bytes.Buffer
	if err := RenderText(&buf, report); err != nil {
		t.Fatalf("RenderText error = %v", err)
	}

	text := buf.String()

	// Verify required content
	required := []string{
		"test/doctor",
		"status: error",
		"errors: 1",
		"warnings: 1",
		"infos: 1",
		"[info] TEST_INFO: info message",
		"[warning] TEST_WARN docs/roadmap/test.md: warning message",
		"[error] TEST_ERR: error message",
	}
	for _, want := range required {
		if !strings.Contains(text, want) {
			t.Fatalf("text missing %q:\n%s", want, text)
		}
	}
}

func TestExitCodeOK(t *testing.T) {
	report := NewReport("test/check", "/repo", nil)
	if got := ExitCode(report, false); got != ExitOK {
		t.Fatalf("ExitCode(empty, false) = %d, want %d", got, ExitOK)
	}
}

func TestExitCodeValidationError(t *testing.T) {
	diagnostics := []Diagnostic{
		{ID: "TEST_ERR", Severity: SeverityError, Message: "error"},
	}
	report := NewReport("test/check", "/repo", diagnostics)
	if got := ExitCode(report, false); got != ExitValidation {
		t.Fatalf("ExitCode(error, false) = %d, want %d", got, ExitValidation)
	}
}

func TestExitCodeWarningStrictMode(t *testing.T) {
	diagnostics := []Diagnostic{
		{ID: "TEST_WARN", Severity: SeverityWarning, Message: "warning"},
	}
	report := NewReport("test/lint", "/repo", diagnostics)

	// Non-strict should return OK
	if got := ExitCode(report, false); got != ExitOK {
		t.Fatalf("ExitCode(warning, false) = %d, want %d", got, ExitOK)
	}

	// Strict should return ExitValidation
	if got := ExitCode(report, true); got != ExitValidation {
		t.Fatalf("ExitCode(warning, true) = %d, want %d", got, ExitValidation)
	}
}

func TestExitCodeCustomExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		severity Severity
		strict   bool
		want     int
	}{
		{
			name:     "usage error",
			code:     ExitUsage,
			severity: SeverityError,
			strict:   false,
			want:     ExitUsage,
		},
		{
			name:     "environment error",
			code:     ExitEnvironment,
			severity: SeverityError,
			strict:   false,
			want:     ExitEnvironment,
		},
		{
			name:     "internal error",
			code:     ExitInternal,
			severity: SeverityError,
			strict:   false,
			want:     ExitInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnostics := []Diagnostic{
				{ID: "TEST", Severity: tt.severity, Message: "message", ExitCode: tt.code},
			}
			report := NewReport("test", "/repo", diagnostics)
			if got := ExitCode(report, tt.strict); got != tt.want {
				t.Fatalf("ExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestExitCodeHighestCodeWins(t *testing.T) {
	// When multiple errors with different codes, highest wins
	diagnostics := []Diagnostic{
		{ID: "TEST_VALIDATION", Severity: SeverityError, Message: "validation", ExitCode: ExitValidation},
		{ID: "TEST_INTERNAL", Severity: SeverityError, Message: "internal", ExitCode: ExitInternal},
	}
	report := NewReport("test", "/repo", diagnostics)
	if got := ExitCode(report, false); got != ExitInternal {
		t.Fatalf("ExitCode(multiple) = %d, want %d", got, ExitInternal)
	}
}

func TestPathNormalization(t *testing.T) {
	// Test that paths are normalized to forward slashes
	diagnostics := []Diagnostic{
		{ID: "TEST_001", Severity: SeverityInfo, Message: "test", Path: "docs/roadmap/test.md"},
	}
	report := NewReport("test", "/repo", diagnostics)
	// On Unix, ToSlash is a no-op, but on Windows it converts backslashes to forward slashes
	// This test verifies that paths are properly set
	if report.Diagnostics[0].Path != "docs/roadmap/test.md" {
		t.Fatalf("Path not normalized: %q", report.Diagnostics[0].Path)
	}
	if report.Root != "/repo" {
		t.Fatalf("Root not normalized: %q", report.Root)
	}
}

func TestSummarizeStatus(t *testing.T) {
	tests := []struct {
		name        string
		diagnostics []Diagnostic
		wantStatus  string
		wantErrors  int
		wantWarns   int
		wantInfos   int
	}{
		{
			name:        "empty",
			diagnostics: nil,
			wantStatus:  SummaryStatusOK,
		},
		{
			name: "info only",
			diagnostics: []Diagnostic{
				{Severity: SeverityInfo, Message: "info"},
			},
			wantStatus: SummaryStatusOK,
			wantInfos:  1,
		},
		{
			name: "warning only",
			diagnostics: []Diagnostic{
				{Severity: SeverityWarning, Message: "warn"},
			},
			wantStatus: SummaryStatusWarning,
			wantWarns:  1,
		},
		{
			name: "error only",
			diagnostics: []Diagnostic{
				{Severity: SeverityError, Message: "error"},
			},
			wantStatus: SummaryStatusError,
			wantErrors: 1,
		},
		{
			name: "mixed: error takes precedence",
			diagnostics: []Diagnostic{
				{Severity: SeverityInfo, Message: "info"},
				{Severity: SeverityWarning, Message: "warn"},
				{Severity: SeverityError, Message: "error"},
			},
			wantStatus: SummaryStatusError,
			wantErrors: 1,
			wantWarns:  1,
			wantInfos:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := NewReport("test", "/repo", tt.diagnostics)
			if report.Summary.Status != tt.wantStatus {
				t.Fatalf("Status = %q, want %q", report.Summary.Status, tt.wantStatus)
			}
			if report.Summary.Errors != tt.wantErrors {
				t.Fatalf("Errors = %d, want %d", report.Summary.Errors, tt.wantErrors)
			}
			if report.Summary.Warnings != tt.wantWarns {
				t.Fatalf("Warnings = %d, want %d", report.Summary.Warnings, tt.wantWarns)
			}
			if report.Summary.Infos != tt.wantInfos {
				t.Fatalf("Infos = %d, want %d", report.Summary.Infos, tt.wantInfos)
			}
		})
	}
}

func TestDiagnosticWithDetails(t *testing.T) {
	details := map[string]any{
		"expected": "TXXX format",
		"count":    42,
		"nested": map[string]any{
			"key": "value",
		},
	}
	diagnostics := []Diagnostic{
		{ID: "TEST_DETAILS", Severity: SeverityError, Message: "test", Details: details},
	}
	report := NewReport("test", "/repo", diagnostics)

	var buf bytes.Buffer
	if err := RenderJSON(&buf, report); err != nil {
		t.Fatalf("RenderJSON error = %v", err)
	}

	var decoded Report
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if decoded.Diagnostics[0].Details == nil {
		t.Fatal("Details is nil in decoded report")
	}
	if decoded.Diagnostics[0].Details["expected"] != "TXXX format" {
		t.Fatalf("Details['expected'] = %v, want 'TXXX format'", decoded.Diagnostics[0].Details["expected"])
	}
	if decoded.Diagnostics[0].Details["count"].(float64) != 42 {
		t.Fatalf("Details['count'] = %v, want 42", decoded.Diagnostics[0].Details["count"])
	}
}

func TestRenderTextWithPathlessAndPathedDiagnostics(t *testing.T) {
	diagnostics := []Diagnostic{
		{ID: "TEST_NO_PATH", Severity: SeverityInfo, Message: "info message"},
		{ID: "TEST_WITH_PATH", Severity: SeverityWarning, Message: "warn message", Path: "docs/roadmap/T001-task.md"},
	}
	report := NewReport("test/check", "/repo", diagnostics)

	var buf bytes.Buffer
	if err := RenderText(&buf, report); err != nil {
		t.Fatalf("RenderText error = %v", err)
	}

	text := buf.String()
	if !strings.Contains(text, "[info] TEST_NO_PATH: info message") {
		t.Fatalf("missing pathless diagnostic:\n%s", text)
	}
	if !strings.Contains(text, "[warning] TEST_WITH_PATH docs/roadmap/T001-task.md: warn message") {
		t.Fatalf("missing pathed diagnostic:\n%s", text)
	}
}

func TestRenderJSONHandlesEmptyDiagnostics(t *testing.T) {
	report := NewReport("test/check", "/repo", nil)

	var buf bytes.Buffer
	if err := RenderJSON(&buf, report); err != nil {
		t.Fatalf("RenderJSON error = %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Fatalf("RenderJSON produced invalid JSON:\n%s", buf.String())
	}

	var decoded Report
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}
	if decoded.Summary.Status != SummaryStatusOK {
		t.Fatalf("Expected ok status, got %q", decoded.Summary.Status)
	}
	if len(decoded.Diagnostics) != 0 {
		t.Fatalf("Expected empty diagnostics, got %d", len(decoded.Diagnostics))
	}
}

// FailingWriter is a writer that always fails
type FailingWriter struct{}

func (fw *FailingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed")
}

func TestRenderTextHandlesWriteError(t *testing.T) {
	report := NewReport("test/check", "/repo", nil)

	fw := &FailingWriter{}
	if err := RenderText(fw, report); err == nil {
		t.Fatal("RenderText should return error when write fails")
	}
}

func TestRenderJSONHandlesEncodingError(t *testing.T) {
	// Create a diagnostic that would fail to encode
	report := NewReport("test/check", "/repo", []Diagnostic{
		{ID: "TEST", Severity: SeverityError, Message: "test"},
	})

	fw := &FailingWriter{}
	if err := RenderJSON(fw, report); err == nil {
		t.Fatal("RenderJSON should return error when encoding fails")
	}
}
