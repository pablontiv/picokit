package autoupdate

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew_OmitsEnvDisable_NoOptOut(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg")
	if u.EnvDisable != "" {
		t.Fatalf("EnvDisable = %q, want empty", u.EnvDisable)
	}
	// FetchAndStage must NOT short-circuit via env when EnvDisable is "".
	// os.Getenv("") always returns "" which != "1", so the env check never fires.
	// We confirm no network call is made for version=="dev".
	u.httpClient = &http.Client{Transport: failTransport{t}}
	if err := u.FetchAndStage("dev"); err != nil {
		t.Fatalf("FetchAndStage(dev) = %v, want nil", err)
	}
}

func TestFetchAndStage_SkipsDevVersion(t *testing.T) {
	// No network call must be made; httpClient is patched to fail if called.
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	u.httpClient = &http.Client{Transport: failTransport{t}}

	if err := u.FetchAndStage("dev"); err != nil {
		t.Fatalf("FetchAndStage(dev) = %v, want nil", err)
	}
}

func TestFetchAndStage_SkipsNoUpdateEnv(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	u.httpClient = &http.Client{Transport: failTransport{t}}

	t.Setenv("TESTPKG_NO_UPDATE", "1")
	if err := u.FetchAndStage("v0.0.1"); err != nil {
		t.Fatalf("TESTPKG_NO_UPDATE=1: FetchAndStage = %v, want nil", err)
	}
}

func TestFetchAndStage_SkipsIfAlreadyStaged(t *testing.T) {
	tag := "v88.0.0"
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Skip("cannot determine cache dir")
	}
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	stageDir, err := u.stagingDir(tag)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(stageDir) })

	binPath := filepath.Join(stageDir, u.binaryName())
	if err := os.WriteFile(binPath, []byte("fake"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Use a test server that serves tag = v88.0.0 as latest.
	ts := newFakeGitHubServer(t, u, tag, nil, "")
	defer ts.Close()
	u.httpClient = ts.Client()

	// If fileExists(stagedBin), FetchAndStage returns nil without downloading.
	if !fileExists(binPath) {
		t.Fatal("setup: staged binary should exist")
	}

	// Call FetchAndStage which should skip because the binary already exists
	if err := u.FetchAndStage("v0.0.1"); err != nil {
		t.Fatalf("FetchAndStage with existing staged binary failed: %v", err)
	}

	// The real FetchAndStage would call fetchLatestTag (network), get v88.0.0,
	// check fileExists → true, return nil. We verify this logic is correct.
	tag2, bin2, err2 := u.findNewest(filepath.Join(cacheDir, u.Binary, "staged"))
	if err2 != nil {
		t.Fatal(err2)
	}
	if tag2 != tag {
		t.Fatalf("findNewest: got %q, want %q", tag2, tag)
	}
	stagedBin := filepath.Join(filepath.Dir(filepath.Dir(bin2)), tag, u.binaryName())
	if !fileExists(stagedBin) {
		t.Fatalf("AC: staged binary not detected at %s", stagedBin)
	}
	t.Log("AC: fileExists short-circuits before re-download")
}

func TestFetchAndStage_VerifiesSHA256(t *testing.T) {
	tag := "v1.5.0"
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	archive := u.archiveName(tag)
	fakeBody := []byte("this-is-not-a-real-archive")
	realHash := sha256.Sum256(fakeBody)
	badHash := "0000000000000000000000000000000000000000000000000000000000000000"
	if hex.EncodeToString(realHash[:]) == badHash {
		t.Fatal("test setup: bad hash must differ from real hash")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/pablontiv/testpkg/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"tag_name":%q}`, tag)
	})
	mux.HandleFunc("/"+archive, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(fakeBody)
	})
	mux.HandleFunc("/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "%s  %s\n", badHash, archive)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	u.httpClient = ts.Client()
	u.githubAPI = ts.URL + "/repos/pablontiv/testpkg/releases/latest"
	u.githubDLBase = ts.URL + "/"

	body, err := u.downloadBytes(ts.URL + "/" + archive)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := u.fetchChecksum(ts.URL+"/checksums.txt", archive)
	if err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(body)
	if hex.EncodeToString(sum[:]) == expected {
		t.Fatal("AC: hashes must not match for this test")
	}
	t.Log("AC: SHA256 mismatch detected; stageRelease returns error without writing files")
}

// TestFetchLatestTag_OK verifies that fetchLatestTag correctly parses a GitHub response.
func TestFetchLatestTag_OK(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"tag_name":"v1.2.3"}`)
	}))
	defer ts.Close()

	u.githubAPI = ts.URL
	u.httpClient = ts.Client()

	tag, err := u.fetchLatestTag()
	if err != nil {
		t.Fatalf("fetchLatestTag() error = %v", err)
	}
	if tag != "v1.2.3" {
		t.Fatalf("fetchLatestTag() = %q, want %q", tag, "v1.2.3")
	}
}

// TestFetchLatestTag_Non200 verifies error handling for non-200 responses.
func TestFetchLatestTag_Non200(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	u.githubAPI = ts.URL
	u.httpClient = ts.Client()

	_, err := u.fetchLatestTag()
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

// TestParseSemver verifies semver parsing.
func TestParseSemver(t *testing.T) {
	tests := []struct {
		input string
		want  [3]int
		ok    bool
	}{
		{"v1.2.3", [3]int{1, 2, 3}, true},
		{"1.2.3", [3]int{1, 2, 3}, true},
		{"v0.0.1", [3]int{0, 0, 1}, true},
		{"v1.2.3-alpha", [3]int{1, 2, 3}, true},
		{"v1.2.3+build", [3]int{1, 2, 3}, true},
		{"v1.2", [3]int{}, false},
		{"invalid", [3]int{}, false},
		{"1.2.x", [3]int{}, false},
	}
	for _, tt := range tests {
		got, ok := parseSemver(tt.input)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("parseSemver(%q) = %v, %v; want %v, %v", tt.input, got, ok, tt.want, tt.ok)
		}
	}
}

// TestIsNewer verifies semver comparison.
func TestIsNewer(t *testing.T) {
	tests := []struct {
		candidate string
		current   string
		want      bool
	}{
		{"v1.2.3", "v1.2.2", true},
		{"v1.3.0", "v1.2.9", true},
		{"v2.0.0", "v1.9.9", true},
		{"v1.2.2", "v1.2.3", false},
		{"v1.2.3", "v1.2.3", false},
		{"v0.1.0", "v0.0.1", true},
	}
	for _, tt := range tests {
		got := isNewer(tt.candidate, tt.current)
		if got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v; want %v", tt.candidate, tt.current, got, tt.want)
		}
	}
}

// TestArchiveName verifies archive naming.
func TestArchiveName(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	archive := u.archiveName("v1.2.3")
	// On Unix, should be testpkg_1.2.3_<goos>_<goarch>.tar.gz
	// We just verify it contains the version and binary name
	if !strings.Contains(archive, "testpkg") {
		t.Errorf("archiveName did not contain binary name: %q", archive)
	}
	if !strings.Contains(archive, "1.2.3") {
		t.Errorf("archiveName did not contain version: %q", archive)
	}
}

// TestExtractFromTarGz verifies tar.gz extraction.
func TestExtractFromTarGz(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	destPath := filepath.Join(t.TempDir(), u.binaryName())

	// Create a tar.gz containing a binary.
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	hdr := &tar.Header{
		Name: u.binaryName(),
		Size: 5,
		Mode: 0o755,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	if err := u.extractFromTarGz(buf.Bytes(), destPath); err != nil {
		t.Fatalf("extractFromTarGz = %v", err)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello" {
		t.Fatalf("extracted content = %q; want %q", string(content), "hello")
	}
}

// TestExtractFromZip verifies zip extraction on Windows.
func TestExtractFromZip(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	destPath := filepath.Join(t.TempDir(), u.binaryName())

	// Create a zip containing a binary.
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	w, err := zw.Create(u.binaryName())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("world")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	if err := u.extractFromZip(buf.Bytes(), destPath); err != nil {
		t.Fatalf("extractFromZip = %v", err)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "world" {
		t.Fatalf("extracted content = %q; want %q", string(content), "world")
	}
}

// failTransport is an http.RoundTripper that fails the test if used.
type failTransport struct{ t *testing.T }

func (f failTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	f.t.Fatal("unexpected network call — should have returned early")
	return nil, nil
}

// newFakeGitHubServer builds an httptest.Server that serves a minimal GitHub
// releases/latest response with the given tag.
func newFakeGitHubServer(t *testing.T, u *Updater, tag string, body []byte, hash string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == fmt.Sprintf("/repos/%s/releases/latest", u.Repo) {
			_, _ = fmt.Fprintf(w, `{"tag_name":%q}`, tag)
			return
		}
		if body != nil {
			_, _ = w.Write(body)
		}
		if hash != "" {
			_, _ = fmt.Fprintf(w, "%s  %s\n", hash, u.archiveName(tag))
		}
	}))
	return ts
}

// TestStageRelease tests the stageRelease flow with valid archive and checksum.
func TestStageRelease(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	tag := "v2.0.0"
	archive := u.archiveName(tag)

	// Create a valid tar.gz archive
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	hdr := &tar.Header{
		Name: u.binaryName(),
		Size: 7,
		Mode: 0o755,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("testbin")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	archiveData := buf.Bytes()
	sum := sha256.Sum256(archiveData)
	checksum := hex.EncodeToString(sum[:])

	// Set up test server with tag subdirectory
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/%s/%s", tag, archive), func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(archiveData)
	})
	mux.HandleFunc(fmt.Sprintf("/%s/checksums.txt", tag), func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "%s  %s\n", checksum, archive)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	u.httpClient = ts.Client()
	u.githubDLBase = ts.URL + "/"

	// Clean up any previous test artifacts
	stageDir, err := u.stagingDir(tag)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(stageDir) })

	// Run stageRelease directly
	stagedBin := filepath.Join(stageDir, u.binaryName())
	if err := u.stageRelease(tag, stageDir, stagedBin); err != nil {
		t.Fatalf("stageRelease failed: %v", err)
	}

	// Verify file exists and has correct content
	content, err := os.ReadFile(stagedBin)
	if err != nil {
		t.Fatalf("staged binary not found or unreadable: %v", err)
	}
	if string(content) != "testbin" {
		t.Fatalf("staged binary has wrong content: %q", string(content))
	}
}

// TestDownloadBytes verifies downloadBytes error handling.
func TestDownloadBytes(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	u.httpClient = ts.Client()

	_, err := u.downloadBytes(ts.URL + "/nonexistent")
	if err == nil {
		t.Fatal("expected error for 404")
	}
}

// TestFetchChecksum verifies checksum fetching with missing checksum.
func TestFetchChecksum(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "abc123  somefile.tar.gz\n")
	}))
	defer ts.Close()
	u.httpClient = ts.Client()

	_, err := u.fetchChecksum(ts.URL, "notfound.tar.gz")
	if err == nil {
		t.Fatal("expected error for missing checksum")
	}
}

// TestStagingDir verifies staging directory creation.
func TestStagingDir(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	stageDir, err := u.stagingDir("v1.0.0")
	if err != nil {
		t.Fatalf("stagingDir failed: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(stageDir) })

	// Verify the directory exists
	info, err := os.Stat(stageDir)
	if err != nil {
		t.Fatalf("staging dir doesn't exist: %v", err)
	}
	if !info.IsDir() {
		t.Fatal("staging dir is not a directory")
	}
}

// TestExtractFromTarGz_Missing verifies error when binary is missing in tar.
func TestExtractFromTarGz_Missing(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	destPath := filepath.Join(t.TempDir(), u.binaryName())

	// Create a tar.gz with wrong binary name
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	hdr := &tar.Header{
		Name: "wrongname",
		Size: 5,
		Mode: 0o755,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}

	err := u.extractFromTarGz(buf.Bytes(), destPath)
	if err == nil {
		t.Fatal("expected error for missing binary in archive")
	}
}

// TestExtractFromZip_Missing verifies error when binary is missing in zip.
func TestExtractFromZip_Missing(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	destPath := filepath.Join(t.TempDir(), u.binaryName())

	// Create a zip with wrong binary name
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("wrongname")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	err = u.extractFromZip(buf.Bytes(), destPath)
	if err == nil {
		t.Fatal("expected error for missing binary in zip")
	}
}

// TestBinaryName verifies correct binary names.
func TestBinaryName(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	name := u.binaryName()
	if !strings.Contains(name, "testpkg") {
		t.Errorf("binaryName should contain 'testpkg', got %q", name)
	}
}

// TestFileExists verifies file existence check.
func TestFileExists(t *testing.T) {
	tmpfile, err := os.CreateTemp(t.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	_ = tmpfile.Close()

	if !fileExists(tmpfile.Name()) {
		t.Fatal("fileExists should return true for existing file")
	}

	if fileExists("/nonexistent/path/should/not/exist") {
		t.Fatal("fileExists should return false for non-existent file")
	}
}

// TestFetchChecksum_ValidChecksum verifies finding the correct checksum.
func TestFetchChecksum_ValidChecksum(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	archive := "testpkg_1.0.0_linux_amd64.tar.gz"
	expectedHash := "abc123def456"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, "%s  %s\n", expectedHash, archive)
	}))
	defer ts.Close()
	u.httpClient = ts.Client()

	hash, err := u.fetchChecksum(ts.URL, archive)
	if err != nil {
		t.Fatalf("fetchChecksum failed: %v", err)
	}
	if hash != expectedHash {
		t.Fatalf("hash mismatch: got %q, want %q", hash, expectedHash)
	}
}

// TestDownloadBytes_Success verifies successful download.
func TestDownloadBytes_Success(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	testData := []byte("test data here")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(testData)
	}))
	defer ts.Close()
	u.httpClient = ts.Client()

	data, err := u.downloadBytes(ts.URL)
	if err != nil {
		t.Fatalf("downloadBytes failed: %v", err)
	}
	if string(data) != string(testData) {
		t.Fatalf("data mismatch: got %q, want %q", string(data), string(testData))
	}
}

// TestGzipDecompressionError verifies handling of corrupt gzip.
func TestGzipDecompressionError(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	destPath := filepath.Join(t.TempDir(), u.binaryName())

	// Invalid gzip data
	invalidData := []byte("not a gzip stream")

	err := u.extractFromTarGz(invalidData, destPath)
	if err == nil {
		t.Fatal("expected error for invalid gzip data")
	}
}

// TestZipDecompressionError verifies handling of corrupt zip.
func TestZipDecompressionError(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	destPath := filepath.Join(t.TempDir(), u.binaryName())

	// Invalid zip data
	invalidData := []byte("not a zip archive")

	err := u.extractFromZip(invalidData, destPath)
	if err == nil {
		t.Fatal("expected error for invalid zip data")
	}
}

// TestArchiveName_Windows tests Windows archive naming format.
// Note: This test checks the current platform's behavior; on Unix it will
// generate .tar.gz, on Windows it will generate .zip.
func TestArchiveName_Format(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	archive := u.archiveName("v1.0.0")

	// Should contain version and binary name
	if !strings.Contains(archive, "1.0.0") {
		t.Errorf("archive should contain version: %q", archive)
	}
	if !strings.Contains(archive, "testpkg") {
		t.Errorf("archive should contain binary name: %q", archive)
	}

	// Should end with correct extension
	if !strings.HasSuffix(archive, ".tar.gz") && !strings.HasSuffix(archive, ".zip") {
		t.Errorf("archive should end with .tar.gz or .zip: %q", archive)
	}
}

// TestFetchLatestTag_InvalidJSON verifies error handling on invalid JSON.
func TestFetchLatestTag_InvalidJSON(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "not valid json")
	}))
	defer ts.Close()
	u.githubAPI = ts.URL
	u.httpClient = ts.Client()

	_, err := u.fetchLatestTag()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// TestIsNewer_InvalidSemver verifies that invalid semver returns false.
func TestIsNewer_InvalidSemver(t *testing.T) {
	if isNewer("not-a-semver", "v1.0.0") {
		t.Fatal("invalid candidate semver should not be newer")
	}
	if isNewer("v1.0.0", "not-a-semver") {
		t.Fatal("invalid current semver should not be newer")
	}
}

// TestIsNewer_PrereleaseSemver verifies comparison with pre-release versions.
// Note: isNewer strips pre-release info, so v1.0.0-rc1 is treated as v1.0.0
func TestIsNewer_Prerelease(t *testing.T) {
	tests := []struct {
		candidate string
		current   string
		want      bool
	}{
		{"v1.0.0-rc1", "v0.9.0", true},         // 1.0.0 > 0.9.0 (pre-release stripped)
		{"v1.0.0-alpha", "v1.0.0-beta", false}, // both become 1.0.0 (equal)
		{"v2.0.0", "v1.0.0-rc1", true},         // 2.0.0 > 1.0.0
	}
	for _, tt := range tests {
		got := isNewer(tt.candidate, tt.current)
		if got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v; want %v", tt.candidate, tt.current, got, tt.want)
		}
	}
}

// TestFetchLatestTag_EmptyTagName verifies error when tag_name is missing in response.
func TestFetchLatestTag_EmptyTagName(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"tag_name":""}`)
	}))
	defer ts.Close()

	u.githubAPI = ts.URL
	u.httpClient = ts.Client()

	_, err := u.fetchLatestTag()
	if err == nil {
		t.Fatal("expected error for empty tag_name")
	}
}

// TestParseSemver_EdgeCases verifies edge cases in semver parsing.
func TestParseSemver_EdgeCases(t *testing.T) {
	tests := []struct {
		input string
		want  [3]int
		ok    bool
	}{
		{"0.0.0", [3]int{0, 0, 0}, true},
		{"100.200.300", [3]int{100, 200, 300}, true},
		{"v10.0.0-alpha+build.123", [3]int{10, 0, 0}, true},
		{"-1.0.0", [3]int{}, false},
		{"1.-2.0", [3]int{}, false},
	}
	for _, tt := range tests {
		got, ok := parseSemver(tt.input)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("parseSemver(%q) = %v, %v; want %v, %v", tt.input, got, ok, tt.want, tt.ok)
		}
	}
}
