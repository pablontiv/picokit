package coverage

import (
	"fmt"
	"strconv"
	"strings"
)

// parseFloorsTOML parses the minimal TOML subset used by .coverage-floors.toml:
// comments, blank lines, "key = <integer>" assignments, and single-line
// double-quoted string arrays. Anything outside that subset is an error naming
// the offending line, so a config drifting beyond the subset fails loudly.
func parseFloorsTOML(data []byte) (*Floors, error) {
	var f Floors
	for i, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("line %d: expected key = value", i+1)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "default":
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("line %d: default must be an integer, got %q", i+1, val)
			}
			f.Default = n
		case "packages", "exclude":
			items, err := parseStringArray(val)
			if err != nil {
				return nil, fmt.Errorf("line %d: %s: %w", i+1, key, err)
			}
			if key == "packages" {
				f.Packages = items
			} else {
				f.Exclude = items
			}
		default:
			// Unknown keys are ignored for forward compatibility, matching
			// go-toml's lenient decoding into a fixed struct.
		}
	}
	return &f, nil
}

// stripComment removes a trailing # comment, ignoring # inside quoted strings.
func stripComment(line string) string {
	inStr := false
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case '\\':
			if inStr {
				i++
			}
		case '"':
			inStr = !inStr
		case '#':
			if !inStr {
				return line[:i]
			}
		}
	}
	return line
}

// parseStringArray parses a single-line TOML string array like ["a", "b"].
// Supported escapes inside strings: \" and \\. Trailing commas are allowed.
func parseStringArray(s string) ([]string, error) {
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return nil, fmt.Errorf("expected single-line string array")
	}
	inner := s[1 : len(s)-1]
	items := []string{}
	i := 0
	for {
		for i < len(inner) && (inner[i] == ' ' || inner[i] == '\t') {
			i++
		}
		if i >= len(inner) {
			return items, nil
		}
		if inner[i] != '"' {
			return nil, fmt.Errorf("expected quoted string")
		}
		i++
		var b strings.Builder
		closed := false
		for i < len(inner) {
			c := inner[i]
			if c == '\\' {
				if i+1 >= len(inner) || (inner[i+1] != '"' && inner[i+1] != '\\') {
					return nil, fmt.Errorf("unsupported escape sequence")
				}
				b.WriteByte(inner[i+1])
				i += 2
				continue
			}
			if c == '"' {
				closed = true
				i++
				break
			}
			b.WriteByte(c)
			i++
		}
		if !closed {
			return nil, fmt.Errorf("unterminated string")
		}
		items = append(items, b.String())
		for i < len(inner) && (inner[i] == ' ' || inner[i] == '\t') {
			i++
		}
		if i < len(inner) {
			if inner[i] != ',' {
				return nil, fmt.Errorf("expected ',' between array items")
			}
			i++
		}
	}
}
