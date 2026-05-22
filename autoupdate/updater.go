package autoupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pablontiv/picokit/hashfile"
)

// Updater is a parameterized async binary updater. Each application creates
// an instance with its own repo, binary name, and no-update environment variable.
type Updater struct {
	Repo       string // e.g. "pablontiv/roadmapctl"
	Binary     string // e.g. "roadmapctl"
	EnvDisable string // env var name that disables updates when set to "1"; "" means no env opt-out

	// CurrentVersion is the running binary version. Set by caller before
	// invoking ApplyStagedIfAvailable.
	CurrentVersion string

	// execFn is the platform-specific execution function. Overridable in tests.
	execFn ExecFn

	// HTTP client for downloads; initialized to default in New, overridable in tests.
	httpClient *http.Client
	// GitHub API URL; defaults to github.com but can be overridden in tests.
	githubAPI string
	// GitHub download base URL; defaults to github.com but can be overridden in tests.
	githubDLBase string
}

// New creates a new Updater. envDisable is optional; if omitted or empty,
// no environment variable can disable the updater (only version=="dev" does).
func New(repo, binary string, envDisable ...string) *Updater {
	env := ""
	if len(envDisable) > 0 {
		env = envDisable[0]
	}
	return &Updater{
		Repo:         repo,
		Binary:       binary,
		EnvDisable:   env,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
		githubAPI:    "https://api.github.com/repos/" + repo + "/releases/latest",
		githubDLBase: "https://github.com/" + repo + "/releases/download/",
	}
}

// FetchAndStage downloads the latest release into a staging directory without
// interrupting the current command. It returns nil silently on network errors;
// a SHA256 mismatch is returned as an error because it signals a data integrity
// problem. No files are written on error.
func (u *Updater) FetchAndStage(currentVersion string) error {
	if currentVersion == "dev" || os.Getenv(u.EnvDisable) == "1" {
		return nil
	}

	tag, err := u.fetchLatestTag()
	if err != nil {
		return nil
	}

	if !isNewer(tag, currentVersion) {
		return nil
	}

	stageDir, err := u.stagingDir(tag)
	if err != nil {
		return nil
	}

	stagedBin := filepath.Join(stageDir, u.binaryName())
	if fileExists(stagedBin) {
		return nil
	}

	return u.stageRelease(tag, stageDir, stagedBin)
}

func (u *Updater) fetchLatestTag() (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u.githubAPI, nil) //nolint:gosec
	if err != nil {
		return "", err
	}
	resp, err := u.httpClient.Do(req) //nolint:gosec
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api: status %d", resp.StatusCode)
	}
	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("empty tag_name in github response")
	}
	return rel.TagName, nil
}

func (u *Updater) binaryName() string {
	if runtime.GOOS == "windows" {
		return u.Binary + ".exe"
	}
	return u.Binary
}

func (u *Updater) stagingDir(tag string) (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(cacheDir, u.Binary, "staged", tag)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func (u *Updater) archiveName(tag string) string {
	version := strings.TrimPrefix(tag, "v")
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	if goos == "windows" {
		return fmt.Sprintf("%s_%s_%s_%s.zip", u.Binary, version, goos, goarch)
	}
	return fmt.Sprintf("%s_%s_%s_%s.tar.gz", u.Binary, version, goos, goarch)
}

func (u *Updater) stageRelease(tag, stageDir, stagedBin string) error {
	archive := u.archiveName(tag)
	baseURL := u.githubDLBase + tag + "/"

	body, err := u.downloadBytes(baseURL + archive)
	if err != nil {
		return nil
	}

	expectedHash, err := u.fetchChecksum(baseURL+"checksums.txt", archive)
	if err != nil {
		return nil
	}

	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) != expectedHash {
		return fmt.Errorf("SHA256 mismatch for %s", archive)
	}

	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return nil
	}

	if runtime.GOOS == "windows" {
		return u.extractFromZip(body, stagedBin)
	}
	return u.extractFromTarGz(body, stagedBin)
}

func (u *Updater) downloadBytes(url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil) //nolint:gosec
	if err != nil {
		return nil, err
	}
	resp, err := u.httpClient.Do(req) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (u *Updater) fetchChecksum(url, archive string) (string, error) {
	body, err := u.downloadBytes(url)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(body), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == archive {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("checksum not found for %s", archive)
}

func (u *Updater) extractFromTarGz(data []byte, dest string) error {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer func() { _ = gr.Close() }()

	target := u.binaryName()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		if filepath.Base(hdr.Name) != target {
			continue
		}
		return hashfile.WriteAtomic(dest, tr, 0o755)
	}
	return fmt.Errorf("binary %s not found in archive", target)
}

func (u *Updater) extractFromZip(data []byte, dest string) error {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("zip: %w", err)
	}
	target := u.binaryName()
	for _, f := range r.File {
		if filepath.Base(f.Name) != target {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() { _ = rc.Close() }()
		return hashfile.WriteAtomic(dest, rc, 0o755)
	}
	return fmt.Errorf("binary %s not found in zip", target)
}

// isNewer reports whether candidate is strictly newer than current using semver.
func isNewer(candidate, current string) bool {
	a, aOK := parseSemver(candidate)
	b, bOK := parseSemver(current)
	if !aOK || !bOK {
		return false
	}
	if a[0] != b[0] {
		return a[0] > b[0]
	}
	if a[1] != b[1] {
		return a[1] > b[1]
	}
	return a[2] > b[2]
}

func parseSemver(v string) ([3]int, bool) {
	v = strings.TrimPrefix(v, "v")
	// Strip pre-release and build metadata (e.g. "-alpha", "+build").
	if idx := strings.IndexAny(v, "-+"); idx != -1 {
		v = v[:idx]
	}
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, false
	}
	var nums [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return [3]int{}, false
		}
		nums[i] = n
	}
	return nums, true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
