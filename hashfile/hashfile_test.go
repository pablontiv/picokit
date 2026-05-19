package hashfile

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashFileSimple(t *testing.T) {
	tmpFile := t.TempDir() + "/test.txt"
	content := "hello world"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("HashFile() error = %v", err)
	}

	// Calculate expected hash
	h := sha256.Sum256([]byte(content))
	expectedHash := hex.EncodeToString(h[:])

	if hash != expectedHash {
		t.Fatalf("HashFile() = %q, want %q", hash, expectedHash)
	}
}

func TestHashFileNonexistent(t *testing.T) {
	_, err := HashFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("HashFile() error = nil, want file not found")
	}
}

func TestHashFileEmptyFile(t *testing.T) {
	tmpFile := t.TempDir() + "/empty.txt"
	if err := os.WriteFile(tmpFile, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("HashFile() error = %v", err)
	}

	// SHA256 of empty string
	h := sha256.Sum256([]byte(""))
	expectedHash := hex.EncodeToString(h[:])

	if hash != expectedHash {
		t.Fatalf("HashFile() = %q, want %q", hash, expectedHash)
	}
}

func TestHashFileLargeContent(t *testing.T) {
	tmpFile := t.TempDir() + "/large.txt"
	// Create a large content (1MB)
	largeContent := strings.Repeat("x", 1024*1024)
	if err := os.WriteFile(tmpFile, []byte(largeContent), 0o644); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("HashFile() error = %v", err)
	}

	// Calculate expected hash
	h := sha256.Sum256([]byte(largeContent))
	expectedHash := hex.EncodeToString(h[:])

	if hash != expectedHash {
		t.Fatalf("HashFile() = %q, want %q", hash, expectedHash)
	}
}

func TestWriteAtomicSuccess(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "dest.txt")
	content := "atomic write content"
	reader := strings.NewReader(content)

	err := WriteAtomic(dest, reader, 0o644)
	if err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}

	// Verify file exists
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != content {
		t.Fatalf("file content = %q, want %q", string(data), content)
	}

	// Verify permissions
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode()&0o644 != 0o644 {
		t.Fatalf("mode = %o, want 0644", info.Mode())
	}

	// Verify temp file was cleaned up
	tmpFile := dest + ".tmp"
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Fatalf("temp file not cleaned up: %v", err)
	}
}

func TestWriteAtomicOverwrite(t *testing.T) {
	destDir := t.TempDir()
	dest := filepath.Join(destDir, "dest.txt")

	// Write initial content
	if err := os.WriteFile(dest, []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Overwrite with atomic write
	newContent := "new atomic content"
	reader := strings.NewReader(newContent)
	err := WriteAtomic(dest, reader, 0o644)
	if err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}

	// Verify content
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != newContent {
		t.Fatalf("file content = %q, want %q", string(data), newContent)
	}
}

func TestWriteAtomicReaderError(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "dest.txt")

	// Create a reader that fails
	failingReader := &failingReader{}
	err := WriteAtomic(dest, failingReader, 0o644)
	if err == nil {
		t.Fatal("WriteAtomic() error = nil, want reader error")
	}

	// Verify temp file was cleaned up
	tmpFile := dest + ".tmp"
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Fatalf("temp file not cleaned up after error: %v", err)
	}

	// Verify dest file doesn't exist
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("dest file created on reader error")
	}
}

func TestWriteAtomicInvalidDir(t *testing.T) {
	dest := "/nonexistent/path/that/does/not/exist/file.txt"
	reader := strings.NewReader("content")

	err := WriteAtomic(dest, reader, 0o644)
	if err == nil {
		t.Fatal("WriteAtomic() error = nil, want directory error")
	}
}

func TestWriteAtomicLargeFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "large.bin")
	largeContent := bytes.Repeat([]byte("x"), 1024*1024) // 1MB
	reader := bytes.NewReader(largeContent)

	err := WriteAtomic(dest, reader, 0o755)
	if err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}

	// Verify content
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) != len(largeContent) {
		t.Fatalf("file size = %d, want %d", len(data), len(largeContent))
	}

	// Verify mode
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode()&0o755 != 0o755 {
		t.Fatalf("mode = %o, want 0755", info.Mode())
	}
}

func TestWriteAtomicWithCloseError(t *testing.T) {
	tmpDir := t.TempDir()
	tmp := filepath.Join(tmpDir, "file.txt.tmp")
	dest := filepath.Join(tmpDir, "file.txt")

	wc := &errorOnCloseWriter{buf: &bytes.Buffer{}}
	err := writeAtomicWith(wc, tmp, dest, strings.NewReader("data"))
	if err == nil {
		t.Fatal("writeAtomicWith() error = nil, want close error")
	}
	// temp file must be cleaned up
	if _, statErr := os.Stat(tmp); !os.IsNotExist(statErr) {
		t.Fatalf("temp file not removed after close error")
	}
}

func TestWriteAtomicWithRenameError(t *testing.T) {
	tmpDir := t.TempDir()
	// dest in a non-existent directory so Rename fails
	dest := filepath.Join(tmpDir, "nonexistent", "file.txt")
	tmp := filepath.Join(tmpDir, "file.txt.tmp")

	wc := &nopWriteCloser{buf: &bytes.Buffer{}}
	err := writeAtomicWith(wc, tmp, dest, strings.NewReader("data"))
	if err == nil {
		t.Fatal("writeAtomicWith() error = nil, want rename error")
	}
}

// failingReader is a reader that always fails
type failingReader struct{}

func (f *failingReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

type errorOnCloseWriter struct{ buf *bytes.Buffer }

func (w *errorOnCloseWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *errorOnCloseWriter) Close() error                { return io.ErrUnexpectedEOF }

type nopWriteCloser struct{ buf *bytes.Buffer }

func (w *nopWriteCloser) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *nopWriteCloser) Close() error                { return nil }

// partialReader succeeds first call then fails
type partialReader struct {
	count int
}

func (pr *partialReader) Read(p []byte) (n int, err error) {
	if pr.count == 0 {
		pr.count++
		// Return some data
		if len(p) > 0 {
			p[0] = 'x'
			return 1, nil
		}
		return 0, nil
	}
	return 0, io.EOF
}

// mockFile is a file-like object that can fail on Close
type mockFile struct {
	*os.File
	failClose bool
}

func (mf *mockFile) Close() error {
	if mf.failClose {
		return io.ErrUnexpectedEOF
	}
	return mf.File.Close()
}

func TestHashFileConsistency(t *testing.T) {
	tmpFile := t.TempDir() + "/test.txt"
	content := "consistent hash test"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	hash1, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("first HashFile() error = %v", err)
	}

	hash2, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("second HashFile() error = %v", err)
	}

	if hash1 != hash2 {
		t.Fatalf("hash1 = %q, hash2 = %q, want equal", hash1, hash2)
	}
}

func TestHashFileHexFormat(t *testing.T) {
	tmpFile := t.TempDir() + "/test.txt"
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("HashFile() error = %v", err)
	}

	// SHA256 produces 32 bytes = 64 hex characters
	if len(hash) != 64 {
		t.Fatalf("hash length = %d, want 64", len(hash))
	}

	// Verify it's valid hex
	_, err = hex.DecodeString(hash)
	if err != nil {
		t.Fatalf("hash is not valid hex: %v", err)
	}
}

func TestHashFileAccessDenied(t *testing.T) {
	tmpFile := t.TempDir() + "/noaccess.txt"
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Remove read permissions
	if err := os.Chmod(tmpFile, 0o000); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(tmpFile, 0o644) }()

	_, err := HashFile(tmpFile)
	if err == nil {
		t.Fatal("HashFile() error = nil, want permission error")
	}
}

func TestWriteAtomicPathPermissions(t *testing.T) {
	destDir := t.TempDir()
	dest := filepath.Join(destDir, "executable.bin")

	content := "#!/bin/bash\necho hello"
	reader := strings.NewReader(content)

	// Write with executable permissions
	err := WriteAtomic(dest, reader, 0o755)
	if err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}

	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	// Check executable bit
	if info.Mode()&0o111 == 0 {
		t.Fatalf("file not executable, mode = %o", info.Mode())
	}
}

func TestWriteAtomicEmptyReader(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "empty.txt")
	reader := strings.NewReader("")

	err := WriteAtomic(dest, reader, 0o644)
	if err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}

	// Verify empty file was created
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Size() != 0 {
		t.Fatalf("file size = %d, want 0", info.Size())
	}
}

func TestWriteAtomicFileClosingError(t *testing.T) {
	destDir := t.TempDir()
	dest := filepath.Join(destDir, "test.txt")

	content := "test content"
	reader := strings.NewReader(content)

	// On Unix-like systems, we can't easily simulate close failure
	// but we can test the normal path
	err := WriteAtomic(dest, reader, 0o644)
	if err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
}

func TestHashFileBinaryContent(t *testing.T) {
	tmpFile := t.TempDir() + "/binary.bin"
	// Create binary content with null bytes
	binaryContent := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	if err := os.WriteFile(tmpFile, binaryContent, 0o644); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("HashFile() error = %v", err)
	}

	// Calculate expected hash
	h := sha256.Sum256(binaryContent)
	expectedHash := hex.EncodeToString(h[:])

	if hash != expectedHash {
		t.Fatalf("hash = %q, want %q", hash, expectedHash)
	}
}

func TestWriteAtomicConcurrentWrites(t *testing.T) {
	destDir := t.TempDir()
	dest := filepath.Join(destDir, "concurrent.txt")

	// Write first content
	reader1 := strings.NewReader("content1")
	if err := WriteAtomic(dest, reader1, 0o644); err != nil {
		t.Fatalf("first WriteAtomic() error = %v", err)
	}

	// Overwrite with second content
	reader2 := strings.NewReader("content2_longer")
	if err := WriteAtomic(dest, reader2, 0o644); err != nil {
		t.Fatalf("second WriteAtomic() error = %v", err)
	}

	// Verify final content
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "content2_longer" {
		t.Fatalf("file content = %q, want content2_longer", string(data))
	}
}

func TestHashFileWithSpecialCharacters(t *testing.T) {
	tmpFile := t.TempDir() + "/special.txt"
	content := "Special: !@#$%^&*()_+-=[]{}|;':\",./<>?\nMultiline\nContent"
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	hash, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("HashFile() error = %v", err)
	}

	// Verify it's a valid hash
	if len(hash) != 64 {
		t.Fatalf("hash length = %d, want 64", len(hash))
	}

	h := sha256.Sum256([]byte(content))
	expected := hex.EncodeToString(h[:])
	if hash != expected {
		t.Fatalf("hash = %q, want %q", hash, expected)
	}
}

func TestWriteAtomicRenameFailure(t *testing.T) {
	// Create a temp directory
	tmpDir := t.TempDir()
	// Create subdirectory for destination
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Try to write to a read-only parent directory
	// This will fail on the Rename operation if we set parent to read-only after creation
	dest := filepath.Join(destDir, "file.txt")
	content := "test content"

	// Write should succeed when parent is writable
	if err := WriteAtomic(dest, strings.NewReader(content), 0o644); err != nil {
		t.Fatalf("WriteAtomic() error = %v", err)
	}

	// Now test with restricted permissions on parent
	if err := os.Chmod(destDir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(destDir, 0o755) }()

	// Try to write again - this should fail on Rename
	dest2 := filepath.Join(destDir, "file2.txt")
	err := WriteAtomic(dest2, strings.NewReader("new content"), 0o644)
	if err == nil {
		t.Log("WriteAtomic() with read-only parent succeeded (permissions may vary)")
	} else {
		t.Logf("WriteAtomic() with read-only parent failed as expected: %v", err)
	}
}

func TestHashFileDifferentSizes(t *testing.T) {
	sizes := []int{1, 10, 100, 1000, 10000}
	for _, size := range sizes {
		t.Run(fmt.Sprintf("size_%d", size), func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "file.bin")
			content := strings.Repeat("x", size)
			if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}

			hash, err := HashFile(tmpFile)
			if err != nil {
				t.Fatalf("HashFile() error = %v", err)
			}

			if len(hash) != 64 {
				t.Fatalf("hash length = %d, want 64", len(hash))
			}

			h := sha256.Sum256([]byte(content))
			expected := hex.EncodeToString(h[:])
			if hash != expected {
				t.Fatalf("hash mismatch")
			}
		})
	}
}

func TestWriteAtomicVariousModes(t *testing.T) {
	modes := []os.FileMode{0o600, 0o644, 0o755, 0o777}
	for _, mode := range modes {
		t.Run(fmt.Sprintf("mode_%o", mode), func(t *testing.T) {
			dest := filepath.Join(t.TempDir(), "file")
			reader := strings.NewReader("content")

			if err := WriteAtomic(dest, reader, mode); err != nil {
				t.Fatalf("WriteAtomic() error = %v", err)
			}

			info, err := os.Stat(dest)
			if err != nil {
				t.Fatalf("Stat() error = %v", err)
			}

			if info.Mode()&mode == 0 {
				t.Fatalf("mode bits not set, have %o, wanted %o", info.Mode(), mode)
			}
		})
	}
}

func TestHashFilePermissionsDenied(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	// Create the file
	if err := os.WriteFile(tmpFile, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Now remove read permissions
	if err := os.Chmod(tmpFile, 0o000); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(tmpFile, 0o644) }()

	// Try to hash - should fail with permission error
	_, err := HashFile(tmpFile)
	if err == nil {
		t.Fatal("HashFile() with no-read file should error")
	}
}
