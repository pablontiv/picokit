package coverage

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DetectModulePrefix reads a go.mod file and returns the module path with a
// trailing slash suitable for use as a coverage profile prefix
// (e.g. "github.com/foo/bar/").
func DetectModulePrefix(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("open go.mod: %w", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "module ") {
			continue
		}
		mod := strings.TrimSpace(strings.TrimPrefix(line, "module"))
		if mod == "" {
			return "", fmt.Errorf("go.mod: empty module path")
		}
		return mod + "/", nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan go.mod: %w", err)
	}
	return "", fmt.Errorf("go.mod: no module directive found")
}
