package pathsec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ErrPathEscape = errors.New("path escapes root")
var ErrAbsolutePath = errors.New("absolute path is not allowed")

var windowsVolumePath = regexp.MustCompile(`^[A-Za-z]:/`)

// ResolveInside verifies that candidate path stays within root, resolving symlinks.
// Returns absolute path, relative path (forward-slash separated), and any error.
func ResolveInside(root string, candidate string) (string, string, error) {
	if strings.TrimSpace(candidate) == "" {
		return "", "", fmt.Errorf("path is empty")
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", "", fmt.Errorf("resolve root: %w", err)
	}
	absRoot = filepath.Clean(absRoot)

	normalized := normalizeSeparators(strings.TrimSpace(candidate))
	if windowsVolumePath.MatchString(normalized) || strings.HasPrefix(normalized, "//") {
		return "", "", fmt.Errorf("%w: %s", ErrAbsolutePath, candidate)
	}

	if strings.HasPrefix(normalized, "/") {
		return "", "", fmt.Errorf("%w: %s", ErrPathEscape, candidate)
	}
	target := filepath.Join(absRoot, filepath.FromSlash(normalized))

	rel, err := containedRel(absRoot, target)
	if err != nil {
		return "", "", err
	}

	if err := verifySymlinkContainment(absRoot, target); err != nil {
		return "", "", err
	}

	return target, filepath.ToSlash(rel), nil
}

func normalizeSeparators(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func containedRel(root string, target string) (string, error) {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return "", fmt.Errorf("resolve relative path: %w", err)
	}
	if rel == "." {
		return rel, nil
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("%w: %s", ErrPathEscape, target)
	}
	return rel, nil
}

func verifySymlinkContainment(root string, target string) error {
	evalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		evalRoot = root
	}
	evalRoot = filepath.Clean(evalRoot)

	evalTarget, err := evalExistingPrefix(target)
	if err != nil {
		return err
	}
	_, err = containedRel(evalRoot, evalTarget)
	return err
}

func evalExistingPrefix(target string) (string, error) {
	clean := filepath.Clean(target)
	probe := clean
	var missing []string

	for {
		if _, err := os.Lstat(probe); err == nil {
			eval, err := filepath.EvalSymlinks(probe)
			if err != nil {
				return "", fmt.Errorf("resolve symlink: %w", err)
			}
			for i := len(missing) - 1; i >= 0; i-- {
				eval = filepath.Join(eval, missing[i])
			}
			return filepath.Clean(eval), nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("inspect path: %w", err)
		}

		parent := filepath.Dir(probe)
		if parent == probe {
			return clean, nil
		}
		missing = append(missing, filepath.Base(probe))
		probe = parent
	}
}
