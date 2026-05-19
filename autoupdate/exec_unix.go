//go:build !windows

package autoupdate

import (
	"os"
	"syscall"
)

// atomicReplace replaces dest with src atomically on Unix.
func atomicReplace(dest, src string) error {
	return os.Rename(src, dest)
}

// platformExec replaces the current process with the binary at path on Unix.
func platformExec(path string) error {
	return syscall.Exec(path, os.Args, os.Environ()) //nolint:gosec
}
