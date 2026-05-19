//go:build windows

package autoupdate

import (
	"fmt"
	"os"
	"os/exec"
)

// atomicReplace replaces dest with src on Windows.
// The running binary can't be renamed over, so we move dest aside, copy src
// into place, then remove the old file.
func atomicReplace(dest, src string) error {
	tmp := dest + ".old"
	if err := os.Rename(dest, tmp); err != nil {
		return err
	}
	if err := copyFile(src, dest); err != nil {
		_ = os.Rename(tmp, dest)
		return err
	}
	_ = os.Remove(tmp)
	return nil
}

// platformExec spawns a new process with the updated binary on Windows and
// exits the current one.
func platformExec(path string) error {
	cmd := exec.Command(path, os.Args[1:]...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
