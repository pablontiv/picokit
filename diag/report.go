package diag

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
)

const (
	ExitOK          = 0
	ExitValidation  = 1
	ExitUsage       = 2
	ExitEnvironment = 3
	ExitInternal    = 4
)

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

const (
	SummaryStatusOK      = "ok"
	SummaryStatusWarning = "warning"
	SummaryStatusError   = "error"
)

type Severity string

type Report struct {
	Version     int          `json:"version"`
	Kind        string       `json:"kind"`
	Summary     Summary      `json:"summary"`
	Root        string       `json:"root"`
	RoadmapRoot string       `json:"roadmap_root"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Summary struct {
	Status   string `json:"status"`
	Errors   int    `json:"errors"`
	Warnings int    `json:"warnings"`
	Infos    int    `json:"infos"`
}

type Diagnostic struct {
	ID       string         `json:"id"`
	Severity Severity       `json:"severity"`
	Message  string         `json:"message"`
	Path     string         `json:"path,omitempty"`
	Details  map[string]any `json:"details,omitempty"`
	ExitCode int            `json:"-"`
}

func NewReport(kind string, root string, diagnostics []Diagnostic) Report {
	copied := make([]Diagnostic, len(diagnostics))
	copy(copied, diagnostics)
	for i := range copied {
		copied[i].Path = filepath.ToSlash(copied[i].Path)
	}
	return Report{
		Version:     1,
		Kind:        kind,
		Summary:     summarize(copied),
		Root:        filepath.ToSlash(root),
		RoadmapRoot: "",
		Diagnostics: copied,
	}
}

func RenderJSON(w io.Writer, report Report) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(report)
}

func RenderText(w io.Writer, report Report) error {
	if _, err := fmt.Fprintf(w, "%s\nstatus: %s\nerrors: %d\nwarnings: %d\ninfos: %d\n", report.Kind, report.Summary.Status, report.Summary.Errors, report.Summary.Warnings, report.Summary.Infos); err != nil {
		return err
	}
	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Path == "" {
			if _, err := fmt.Fprintf(w, "[%s] %s: %s\n", diagnostic.Severity, diagnostic.ID, diagnostic.Message); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "[%s] %s %s: %s\n", diagnostic.Severity, diagnostic.ID, diagnostic.Path, diagnostic.Message); err != nil {
			return err
		}
	}
	return nil
}

func ExitCode(report Report, strict bool) int {
	code := ExitOK
	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Severity == SeverityWarning && strict && code < ExitValidation {
			code = ExitValidation
		}
		if diagnostic.Severity != SeverityError {
			continue
		}
		diagnosticCode := diagnostic.ExitCode
		if diagnosticCode == 0 {
			diagnosticCode = ExitValidation
		}
		if diagnosticCode > code {
			code = diagnosticCode
		}
	}
	return code
}

func summarize(diagnostics []Diagnostic) Summary {
	summary := Summary{Status: SummaryStatusOK}
	for _, diagnostic := range diagnostics {
		switch diagnostic.Severity {
		case SeverityError:
			summary.Errors++
		case SeverityWarning:
			summary.Warnings++
		case SeverityInfo:
			summary.Infos++
		}
	}
	if summary.Errors > 0 {
		summary.Status = SummaryStatusError
	} else if summary.Warnings > 0 {
		summary.Status = SummaryStatusWarning
	}
	return summary
}
