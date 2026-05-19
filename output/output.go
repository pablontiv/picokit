package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Format represents the output format type.
type Format int

const (
	FormatText Format = iota
	FormatJSON
	FormatRobot
)

// Formatter handles output formatting with optional token limiting.
type Formatter struct {
	Format    Format
	MaxTokens int // 0 = no limit
}

// NewFormatter creates a new Formatter with the given format and token limit.
func NewFormatter(format Format, maxTokens int) *Formatter {
	return &Formatter{
		Format:    format,
		MaxTokens: maxTokens,
	}
}

// WriteJSON writes arbitrary data as JSON to the writer with indentation.
func (f *Formatter) WriteJSON(w io.Writer, v any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// WriteLines writes lines to the writer in the configured format.
// Respects MaxTokens if > 0: stops when estimated token count would exceed limit.
func (f *Formatter) WriteLines(w io.Writer, lines []string) error {
	// Apply token limiting if specified
	lines = f.limitLines(lines)

	switch f.Format {
	case FormatJSON:
		return f.writeLinesJSON(w, lines)
	case FormatRobot:
		return f.writeLinesRobot(w, lines)
	case FormatText:
		fallthrough
	default:
		return f.writeLinesText(w, lines)
	}
}

// TokenCount estimates the token count for a string.
// Uses heuristic: token count = int(float64(wordCount) * 1.3)
func TokenCount(text string) int {
	wordCount := len(strings.Fields(text))
	return int(float64(wordCount) * 1.3)
}

// Private helper methods

func (f *Formatter) writeLinesText(w io.Writer, lines []string) error {
	_, err := io.WriteString(w, strings.Join(lines, "\n"))
	return err
}

func (f *Formatter) writeLinesJSON(w io.Writer, lines []string) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(lines)
}

func (f *Formatter) writeLinesRobot(w io.Writer, lines []string) error {
	for i, line := range lines {
		_, err := fmt.Fprintf(w, "result_%d=%s\n", i, line)
		if err != nil {
			return err
		}
	}
	return nil
}

// limitLines limits the lines based on the MaxTokens setting.
func (f *Formatter) limitLines(lines []string) []string {
	if f.MaxTokens <= 0 {
		return lines
	}

	var limited []string
	tokenCount := 0

	for _, line := range lines {
		lineTokens := TokenCount(line)
		if tokenCount+lineTokens > f.MaxTokens && len(limited) > 0 {
			// Adding this line would exceed limit, and we already have at least one
			break
		}
		limited = append(limited, line)
		tokenCount += lineTokens
	}

	return limited
}
