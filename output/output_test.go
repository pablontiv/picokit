package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name      string
		format    Format
		maxTokens int
	}{
		{
			name:      "text format",
			format:    FormatText,
			maxTokens: 0,
		},
		{
			name:      "json format with token limit",
			format:    FormatJSON,
			maxTokens: 100,
		},
		{
			name:      "robot format",
			format:    FormatRobot,
			maxTokens: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(tt.format, tt.maxTokens)
			if f == nil {
				t.Errorf("NewFormatter returned nil")
			}
			if f.Format != tt.format {
				t.Errorf("Expected format %v, got %v", tt.format, f.Format)
			}
			if f.MaxTokens != tt.maxTokens {
				t.Errorf("Expected MaxTokens %d, got %d", tt.maxTokens, f.MaxTokens)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		wantValid bool
		wantError bool
	}{
		{
			name: "simple struct",
			input: struct {
				Name string
			}{"foo"},
			wantValid: true,
			wantError: false,
		},
		{
			name: "map",
			input: map[string]interface{}{
				"key":   "value",
				"count": 42,
			},
			wantValid: true,
			wantError: false,
		},
		{
			name:      "slice",
			input:     []string{"a", "b", "c"},
			wantValid: true,
			wantError: false,
		},
		{
			name: "nested struct",
			input: struct {
				Name    string
				Details struct {
					Age int
				}
			}{
				Name: "test",
				Details: struct {
					Age int
				}{Age: 30},
			},
			wantValid: true,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(FormatJSON, 0)
			var buf bytes.Buffer
			err := f.WriteJSON(&buf, tt.input)

			if (err != nil) != tt.wantError {
				t.Errorf("WriteJSON() error = %v, wantError %v", err, tt.wantError)
			}

			output := buf.String()
			if tt.wantValid && !json.Valid([]byte(output)) {
				t.Errorf("WriteJSON() produced invalid JSON: %s", output)
			}

			// Check for indentation
			if strings.Contains(output, "  ") || output == "" {
				// Either has indentation or is empty - both are acceptable for indented JSON
			}
		})
	}
}

func TestWriteJSON_Indentation(t *testing.T) {
	f := NewFormatter(FormatJSON, 0)
	var buf bytes.Buffer

	input := map[string]interface{}{
		"Name": "foo",
		"Data": map[string]interface{}{
			"Value": 42,
		},
	}

	err := f.WriteJSON(&buf, input)
	if err != nil {
		t.Errorf("WriteJSON() error = %v", err)
	}

	output := buf.String()
	// Check that output is indented (contains newlines and spaces)
	if !strings.Contains(output, "\n") {
		t.Errorf("WriteJSON() output is not indented: %s", output)
	}
}

func TestWriteLinesText(t *testing.T) {
	f := NewFormatter(FormatText, 0)
	lines := []string{"line1", "line2", "line3"}

	var buf bytes.Buffer
	err := f.WriteLines(&buf, lines)
	if err != nil {
		t.Errorf("WriteLines() error = %v", err)
	}

	expected := "line1\nline2\nline3"
	if buf.String() != expected {
		t.Errorf("WriteLines() = %q, want %q", buf.String(), expected)
	}
}

func TestWriteLinesJSON(t *testing.T) {
	f := NewFormatter(FormatJSON, 0)
	lines := []string{"line1", "line2", "line3"}

	var buf bytes.Buffer
	err := f.WriteLines(&buf, lines)
	if err != nil {
		t.Errorf("WriteLines() error = %v", err)
	}

	output := buf.String()
	if !json.Valid([]byte(output)) {
		t.Errorf("WriteLines() produced invalid JSON: %s", output)
	}

	// Unmarshal to verify it's a string array
	var result []string
	err = json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if len(result) != len(lines) {
		t.Errorf("Expected %d lines, got %d", len(lines), len(result))
	}

	for i, line := range result {
		if line != lines[i] {
			t.Errorf("Line %d: expected %q, got %q", i, lines[i], line)
		}
	}
}

func TestWriteLinesRobot(t *testing.T) {
	f := NewFormatter(FormatRobot, 0)
	lines := []string{"output1", "output2", "output3"}

	var buf bytes.Buffer
	err := f.WriteLines(&buf, lines)
	if err != nil {
		t.Errorf("WriteLines() error = %v", err)
	}

	output := buf.String()
	expectedLines := []string{
		"result_0=output1",
		"result_1=output2",
		"result_2=output3",
	}

	actualLines := strings.TrimSpace(output)
	for i, expected := range expectedLines {
		if !strings.Contains(actualLines, expected) {
			t.Errorf("Line %d: expected to contain %q in output:\n%s", i, expected, actualLines)
		}
	}
}

func TestTokenCount(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
		tolerance int
	}{
		{
			name:      "hello world",
			text:      "hello world",
			expected:  2,
			tolerance: 1,
		},
		{
			name:      "single word",
			text:      "hello",
			expected:  1,
			tolerance: 1,
		},
		{
			name:      "empty string",
			text:      "",
			expected:  0,
			tolerance: 1,
		},
		{
			name:      "three words",
			text:      "the quick brown",
			expected:  3,
			tolerance: 1,
		},
		{
			name:      "many words",
			text:      "this is a longer sentence with multiple words",
			expected:  8,
			tolerance: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TokenCount(tt.text)
			// Allow ±tolerance
			if result < tt.expected-tt.tolerance || result > tt.expected+tt.tolerance {
				t.Errorf("TokenCount(%q) = %d, want ≈%d (±%d)", tt.text, result, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestTokenCount_Precision(t *testing.T) {
	// Test that 1.3 multiplier is applied correctly
	// 10 words * 1.3 = 13 tokens
	text := "one two three four five six seven eight nine ten"
	result := TokenCount(text)
	expected := 13
	if result != expected {
		t.Errorf("TokenCount(%q) = %d, want %d (10 words * 1.3 = 13)", text, result, expected)
	}
}

func TestWriteLinesWithTokenLimit(t *testing.T) {
	// Create a formatter with a small token limit
	f := NewFormatter(FormatText, 3)
	lines := []string{
		"hello",           // 1 token
		"world test",      // 2 tokens (1*1.3 ≈ 1, 1*1.3 ≈ 1, round = 2)
		"third line here", // 3 tokens (3 words * 1.3 ≈ 3.9 ≈ 3)
	}

	var buf bytes.Buffer
	err := f.WriteLines(&buf, lines)
	if err != nil {
		t.Errorf("WriteLines() error = %v", err)
	}

	output := buf.String()
	// With limit of 3 tokens, first line (1 token) should fit
	if !strings.Contains(output, "hello") {
		t.Errorf("Expected 'hello' in output with token limit: %s", output)
	}
}

func TestWriteLinesWithZeroTokenLimit(t *testing.T) {
	// Zero means no limit
	f := NewFormatter(FormatText, 0)
	lines := []string{"line1", "line2", "line3"}

	var buf bytes.Buffer
	err := f.WriteLines(&buf, lines)
	if err != nil {
		t.Errorf("WriteLines() error = %v", err)
	}

	output := buf.String()
	expected := "line1\nline2\nline3"
	if output != expected {
		t.Errorf("WriteLines() with limit 0 = %q, want %q", output, expected)
	}
}

func TestWriteLinesEmpty(t *testing.T) {
	formats := []struct {
		name   string
		format Format
	}{
		{"text", FormatText},
		{"json", FormatJSON},
		{"robot", FormatRobot},
	}

	for _, fmt := range formats {
		t.Run(fmt.name, func(t *testing.T) {
			f := NewFormatter(fmt.format, 0)
			var buf bytes.Buffer
			err := f.WriteLines(&buf, []string{})
			if err != nil {
				t.Errorf("WriteLines() error = %v", err)
			}

			output := buf.String()
			// For JSON, empty array is []
			if fmt.format == FormatJSON && !strings.Contains(output, "[]") {
				t.Errorf("WriteLines() with empty lines should produce [], got: %s", output)
			}
		})
	}
}

func TestFormatConstants(t *testing.T) {
	if FormatText != 0 {
		t.Errorf("FormatText should be 0, got %d", FormatText)
	}
	if FormatJSON != 1 {
		t.Errorf("FormatJSON should be 1, got %d", FormatJSON)
	}
	if FormatRobot != 2 {
		t.Errorf("FormatRobot should be 2, got %d", FormatRobot)
	}
}

func TestWriteJSON_StructWithName(t *testing.T) {
	f := NewFormatter(FormatJSON, 0)
	var buf bytes.Buffer

	input := struct {
		Name string
	}{"foo"}

	err := f.WriteJSON(&buf, input)
	if err != nil {
		t.Errorf("WriteJSON() error = %v", err)
	}

	output := buf.String()
	if !json.Valid([]byte(output)) {
		t.Errorf("WriteJSON() produced invalid JSON: %s", output)
	}

	// Verify it contains the field
	if !strings.Contains(output, "Name") || !strings.Contains(output, "foo") {
		t.Errorf("WriteJSON() output should contain Name and foo: %s", output)
	}
}

func TestWriteLinesRobot_Format(t *testing.T) {
	f := NewFormatter(FormatRobot, 0)
	lines := []string{"first output", "second output"}

	var buf bytes.Buffer
	err := f.WriteLines(&buf, lines)
	if err != nil {
		t.Errorf("WriteLines() error = %v", err)
	}

	output := buf.String()
	expectedPairs := []string{
		"result_0=first output",
		"result_1=second output",
	}

	for _, expected := range expectedPairs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected %q in robot format output, got:\n%s", expected, output)
		}
	}
}
