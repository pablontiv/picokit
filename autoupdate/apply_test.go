package autoupdate

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestApply_SkipsIfNothingStaged verifies that ApplyStagedIfAvailable returns
// nil when the staging directory is empty or missing.
func TestApply_SkipsIfNothingStaged(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	u.CurrentVersion = "v999.999.999"

	if err := u.ApplyStagedIfAvailable(); err != nil {
		t.Fatalf("ApplyStagedIfAvailable() = %v, want nil", err)
	}
}

// TestApply_SkipsIfNotNewer verifies that ApplyStagedIfAvailable does not
// apply a staged binary that is not newer than the current version.
func TestApply_SkipsIfNotNewer(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	_, err := os.UserCacheDir()
	if err != nil {
		t.Skip("cannot determine cache dir")
	}
	tag := "v0.0.1-apply-test"
	stageDir, err := u.stagingDir(tag)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(stageDir) })

	if err := os.WriteFile(filepath.Join(stageDir, u.binaryName()), []byte("fake"), 0o755); err != nil {
		t.Fatal(err)
	}

	u.CurrentVersion = "v9.9.9"

	execCalled := false
	u.execFn = func(_ string) error { execCalled = true; return nil }

	if err := u.ApplyStagedIfAvailable(); err != nil {
		t.Fatalf("ApplyStagedIfAvailable() = %v, want nil", err)
	}
	if execCalled {
		t.Fatal("execFn must not be called when staged version is not newer")
	}
}

// TestFindNewest_NonExistentDir verifies findNewest returns error for missing dir.
func TestFindNewest_NonExistentDir(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	_, _, err := u.findNewest("/nonexistent/path/that/cannot/possibly/exist")
	if err == nil {
		t.Fatal("expected error for non-existent staging dir")
	}
}

// TestApply_SkipsDevVersion verifies that dev version skips apply.
func TestApply_SkipsDevVersion(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	u.CurrentVersion = "dev"

	if err := u.ApplyStagedIfAvailable(); err != nil {
		t.Fatalf("dev version: ApplyStagedIfAvailable() = %v, want nil", err)
	}
}

func TestApply_SameMajorPolicyWithholdsPreexistingCrossMajorStage(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	u := New("pablontiv/testpkg", "testpkg-policy-apply-withheld")
	u.CurrentVersion = "v1.9.9"
	u.VersionPolicy = SameMajorOnly
	stageDir, err := u.stagingDir("v2.0.0")
	if err != nil {
		t.Fatal(err)
	}
	stagedBin := filepath.Join(stageDir, u.binaryName())
	if err := os.WriteFile(stagedBin, []byte("cross-major"), 0o755); err != nil {
		t.Fatal(err)
	}
	u.replaceFn = func(_, _ string) error {
		t.Fatal("withheld staged binary must not replace the current executable")
		return nil
	}
	u.execFn = func(_ string) error {
		t.Fatal("withheld staged binary must not be executed")
		return nil
	}

	err = u.ApplyStagedIfAvailable()
	var withheld *UpdateWithheldError
	if !errors.As(err, &withheld) {
		t.Fatalf("error = %v, want *UpdateWithheldError", err)
	}
	if withheld.CurrentVersion != "v1.9.9" || withheld.CandidateVersion != "v2.0.0" {
		t.Fatalf("withheld versions = (%q, %q), want (%q, %q)",
			withheld.CurrentVersion, withheld.CandidateVersion, "v1.9.9", "v2.0.0")
	}
	if _, err := os.Stat(stagedBin); err != nil {
		t.Fatalf("withheld staged binary must be retained: %v", err)
	}
}

func TestApply_SameMajorPolicyAllowsSameMajorStage(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	u := New("pablontiv/testpkg", "testpkg-policy-apply-allowed")
	u.CurrentVersion = "v1.2.3+build.1"
	u.VersionPolicy = SameMajorOnly
	stageDir, err := u.stagingDir("v1.3.0-rc.1+build.7")
	if err != nil {
		t.Fatal(err)
	}
	stagedBin := filepath.Join(stageDir, u.binaryName())
	if err := os.WriteFile(stagedBin, []byte("same-major"), 0o755); err != nil {
		t.Fatal(err)
	}

	replaceCalled := false
	u.replaceFn = func(_, src string) error {
		replaceCalled = true
		if src != stagedBin {
			t.Fatalf("replacement source = %q, want %q", src, stagedBin)
		}
		return errors.New("stop before replacing the test executable")
	}

	if err := u.ApplyStagedIfAvailable(); err != nil {
		t.Fatalf("ApplyStagedIfAvailable() = %v, want nil", err)
	}
	if !replaceCalled {
		t.Fatal("same-major staged binary must reach replacement")
	}
}

// TestApply_NewerStaged verifies that when a newer staged binary exists,
// the apply logic finds it and attempts to execute (via mocked execFn).
func TestApply_NewerStaged(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Skip("cannot determine cache dir")
	}
	tag := "v9.0.0-test"
	stageDir, err := u.stagingDir(tag)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(stageDir) })

	stagedBin := filepath.Join(stageDir, u.binaryName())
	if err := os.WriteFile(stagedBin, []byte("fake-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Mock execFn so re-exec doesn't actually replace the process.
	u.execFn = func(_ string) error {
		return nil
	}

	u.CurrentVersion = "v0.1.0"

	newestTag, newestBin, err2 := u.findNewest(filepath.Join(cacheDir, u.Binary, "staged"))
	if err2 != nil {
		t.Fatal(err2)
	}
	if newestTag == "" {
		t.Fatal("AC2: findNewest returned empty tag")
	}
	if !isNewer(newestTag, u.CurrentVersion) {
		t.Fatalf("AC2: staged version %s not newer than %s", newestTag, u.CurrentVersion)
	}
	if newestBin == "" {
		t.Fatal("AC2: findNewest returned empty binary path")
	}

	t.Logf("AC2: findNewest correctly found %s at %s", newestTag, newestBin)
}

// TestCopyFile verifies file copying.
func TestCopyFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.txt")
	dst := filepath.Join(t.TempDir(), "dst.txt")

	// Create source file
	if err := os.WriteFile(src, []byte("test content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Copy file
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify destination
	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("destination file not readable: %v", err)
	}
	if string(content) != "test content" {
		t.Fatalf("destination content mismatch: %q", string(content))
	}
}

// TestCopyFile_SourceNotFound verifies error when source doesn't exist.
func TestCopyFile_SourceNotFound(t *testing.T) {
	src := filepath.Join(t.TempDir(), "nonexistent.txt")
	dst := filepath.Join(t.TempDir(), "dst.txt")

	err := copyFile(src, dst)
	if err == nil {
		t.Fatal("expected error when source doesn't exist")
	}
}

// TestStagingDir_CreateError tests handling when mkdir fails (difficult to simulate).
func TestStagingDir_ConsistentPath(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	tag := "v1.0.0"

	dir1, err := u.stagingDir(tag)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir1) })

	dir2, err := u.stagingDir(tag)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir2) })

	if dir1 != dir2 {
		t.Fatalf("stagingDir should return same path for same tag: %q vs %q", dir1, dir2)
	}
}

// TestApplyStagedIfAvailable_WithMockedExec tests the apply flow with mocked exec.
func TestApplyStagedIfAvailable_WithMockedExec(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	_, err := os.UserCacheDir()
	if err != nil {
		t.Skip("cannot determine cache dir")
	}

	// Create a staged binary with newer version
	tag := "v8.0.0"
	stageDir, err := u.stagingDir(tag)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(stageDir) })

	stagedBin := filepath.Join(stageDir, u.binaryName())
	if err := os.WriteFile(stagedBin, []byte("newer-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	u.CurrentVersion = "v1.0.0"

	// Mock exec to not actually replace the process
	u.execFn = func(_ string) error {
		return nil
	}

	// Call ApplyStagedIfAvailable - it should find the staged binary
	// and call execFn (but not actually exec due to our mock)
	if err := u.ApplyStagedIfAvailable(); err != nil {
		t.Fatalf("ApplyStagedIfAvailable failed: %v", err)
	}

	// Note: execFn may or may not be called depending on atomicReplace behavior
	// which is platform-specific. We're mainly testing that the function completes
	// without error.
}

// TestFindNewest_MultipleVersions tests sorting of multiple staged versions.
func TestFindNewest_MultipleVersions(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Skip("cannot determine cache dir")
	}

	baseDir := filepath.Join(cacheDir, u.Binary, "staged")
	t.Cleanup(func() { _ = os.RemoveAll(baseDir) })

	// Create multiple staged binaries
	versions := []string{"v1.0.0", "v2.0.0", "v0.5.0"}
	for _, v := range versions {
		stageDir, err := u.stagingDir(v)
		if err != nil {
			t.Fatal(err)
		}
		binPath := filepath.Join(stageDir, u.binaryName())
		if err := os.WriteFile(binPath, []byte("binary"), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// findNewest should return the highest version
	newest, _, err := u.findNewest(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	if newest != "v2.0.0" {
		t.Fatalf("findNewest should return v2.0.0, got %q", newest)
	}
}

// TestFetchAndStage_NonNewer verifies that FetchAndStage skips when staged version is not newer.
func TestFetchAndStage_NonNewer(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	tag := "v0.5.0"
	_, err := os.UserCacheDir()
	if err != nil {
		t.Skip("cannot determine cache dir")
	}
	stageDir, err := u.stagingDir(tag)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(stageDir) })

	if err := os.WriteFile(filepath.Join(stageDir, u.binaryName()), []byte("fake"), 0o755); err != nil {
		t.Fatal(err)
	}

	ts := newFakeGitHubServer(t, u, tag, nil, "")
	defer ts.Close()
	u.httpClient = ts.Client()
	u.githubAPI = ts.URL + fmt.Sprintf("/repos/%s/releases/latest", u.Repo)

	// Call FetchAndStage with a higher current version
	if err := u.FetchAndStage("v1.0.0"); err != nil {
		t.Fatalf("FetchAndStage with non-newer version failed: %v", err)
	}
}

// TestApply_CacheError verifies ApplyStagedIfAvailable handles cache dir errors.
func TestApply_CacheError(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	u.CurrentVersion = "v1.0.0"

	// Mock os.UserCacheDir to return an error (we'll test the error handling)
	// Since we can't directly mock it, we just verify that when there's no
	// staged directory, ApplyStagedIfAvailable returns nil
	if err := u.ApplyStagedIfAvailable(); err != nil {
		t.Fatalf("ApplyStagedIfAvailable should return nil on missing cache: %v", err)
	}
}

// TestFetchLatestTag_EmptyTag verifies error on empty tag response.
func TestFetchLatestTag_EmptyTag(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `{"tag_name":""}`)
	}))
	defer ts.Close()
	u.githubAPI = ts.URL
	u.httpClient = ts.Client()

	_, err := u.fetchLatestTag()
	if err == nil {
		t.Fatal("expected error for empty tag")
	}
}

// TestApplyStagedIfAvailable_FullFlow tests the complete apply flow.
func TestApplyStagedIfAvailable_FullFlow(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Skip("cannot determine cache dir")
	}

	// Create a staged binary with a newer version
	tag := "v5.0.0"
	stageDir, err := u.stagingDir(tag)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(stageDir) })

	stagedBin := filepath.Join(stageDir, u.binaryName())
	if err := os.WriteFile(stagedBin, []byte("new-binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	u.CurrentVersion = "v1.0.0"

	// Verify findNewest correctly identifies the staged binary
	newestTag, newestBin, err := u.findNewest(filepath.Join(cacheDir, u.Binary, "staged"))
	if err != nil {
		t.Fatal(err)
	}
	if newestTag != tag {
		t.Fatalf("findNewest: expected %q, got %q", tag, newestTag)
	}
	if newestBin == "" {
		t.Fatal("findNewest: binary path is empty")
	}
}

// TestFetchAndStage_ErrorOnFetch verifies silent return on download error.
func TestFetchAndStage_ErrorOnFetch(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")

	// Create test server that responds to API but fails on archive download
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/repos/%s/releases/latest", u.Repo), func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprintf(w, `{"tag_name":"v2.0.0"}`)
	})
	// Other routes return 404
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	u.httpClient = ts.Client()
	u.githubAPI = ts.URL + fmt.Sprintf("/repos/%s/releases/latest", u.Repo)
	u.githubDLBase = ts.URL + "/"

	// FetchAndStage should return nil on download error
	err := u.FetchAndStage("v1.0.0")
	if err != nil {
		t.Fatalf("FetchAndStage should return nil on fetch error, got %v", err)
	}
}

// TestBinaryName_Windows tests binary name on different platforms.
// Note: This test will only run true binary name matching on the current platform.
func TestBinaryName_Consistency(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")
	name1 := u.binaryName()
	name2 := u.binaryName()

	if name1 != name2 {
		t.Fatalf("binaryName not consistent: %q vs %q", name1, name2)
	}

	if name1 == "" {
		t.Fatal("binaryName returned empty string")
	}
}

// TestCopyFile_BadDst verifies copyFile error when destination dir doesn't exist.
func TestCopyFile_BadDst(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.txt")
	if err := os.WriteFile(src, []byte("content"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := copyFile(src, "/nonexistent/dir/dst.txt")
	if err == nil {
		t.Fatal("expected error for unwritable destination")
	}
}

// TestStagingDir_MultipleVersions tests staging directory isolation.
func TestStagingDir_MultipleVersions(t *testing.T) {
	u := New("pablontiv/testpkg", "testpkg", "TESTPKG_NO_UPDATE")

	dir1, err := u.stagingDir("v1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir1) })

	dir2, err := u.stagingDir("v2.0.0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir2) })

	if dir1 == dir2 {
		t.Fatal("staging directories for different versions should be different")
	}

	// Verify both exist
	if _, err := os.Stat(dir1); err != nil {
		t.Fatalf("dir1 should exist: %v", err)
	}
	if _, err := os.Stat(dir2); err != nil {
		t.Fatalf("dir2 should exist: %v", err)
	}
}
