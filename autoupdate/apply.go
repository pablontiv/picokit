package autoupdate

import (
	"io"
	"os"
	"path/filepath"
	"sort"
)

// ExecFn is the signature for the platform-specific execution function.
// It replaces the current process with the binary at path.
type ExecFn func(path string) error

// ApplyStagedIfAvailable detects the newest staged binary, and if it is newer
// than CurrentVersion, atomically replaces the running binary and re-execs.
// All errors (permissions, filesystem) are swallowed silently — the function
// never interrupts the current command.
func (u *Updater) ApplyStagedIfAvailable() error {
	if u.CurrentVersion == "dev" {
		return nil
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil
	}
	stagedBase := filepath.Join(cacheDir, u.Binary, "staged")

	newestTag, newestBin, err := u.findNewest(stagedBase)
	if err != nil || newestTag == "" {
		return nil
	}

	if !isNewer(newestTag, u.CurrentVersion) {
		return nil
	}

	currentBin, err := os.Executable()
	if err != nil {
		return nil
	}

	if err := atomicReplace(currentBin, newestBin); err != nil {
		return nil
	}

	// Re-exec — on Unix this never returns on success; on Windows we exit after
	// launching the new process.
	execFn := u.execFn
	if execFn == nil {
		execFn = platformExec
	}
	_ = execFn(currentBin)
	return nil
}

// findNewest scans stagedBase for version directories and returns the newest
// tag and its binary path.
func (u *Updater) findNewest(stagedBase string) (tag string, binPath string, err error) {
	entries, err := os.ReadDir(stagedBase)
	if err != nil {
		return "", "", err
	}

	type candidate struct {
		tag string
		bin string
	}
	var candidates []candidate
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		bin := filepath.Join(stagedBase, e.Name(), u.binaryName())
		if fileExists(bin) {
			candidates = append(candidates, candidate{e.Name(), bin})
		}
	}
	if len(candidates) == 0 {
		return "", "", nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return isNewer(candidates[i].tag, candidates[j].tag)
	})
	best := candidates[0]
	return best.tag, best.bin, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	if _, err := copyIO(in, out); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

func copyIO(src io.Reader, dst io.Writer) (int64, error) {
	return io.Copy(dst, src)
}
