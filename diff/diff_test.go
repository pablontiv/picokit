package diff

import (
	"strings"
	"testing"
)

func TestNewFileFormatsAddedLines(t *testing.T) {
	got := NewFile("docs/file.md", "first\nsecond")
	for _, want := range []string{"--- /dev/null", "+++ b/docs/file.md", "+first\n", "+second\n"} {
		if !strings.Contains(got, want) {
			t.Fatalf("diff missing %q:\n%s", want, got)
		}
	}
}

func TestUpdateFileFormatsPreviousAndNewLines(t *testing.T) {
	got := UpdateFile("docs/file.md", "old\n", "new line\nsecond")
	for _, want := range []string{"--- a/docs/file.md", "+++ b/docs/file.md", "-old\n", "+new line\n", "+second\n"} {
		if !strings.Contains(got, want) {
			t.Fatalf("diff missing %q:\n%s", want, got)
		}
	}
}

func TestNewFileIgnoresTrailingEmptyFromSplit(t *testing.T) {
	got := NewFile("docs/file.md", "line\n")
	if strings.Count(got, "+line\n") != 1 {
		t.Fatalf("expected exactly one +line, got:\n%s", got)
	}
}

func TestNewFileEmptyContent(t *testing.T) {
	got := NewFile("empty.txt", "")
	if !strings.Contains(got, "--- /dev/null") || !strings.Contains(got, "+++ b/empty.txt") {
		t.Fatalf("empty file diff missing headers:\n%s", got)
	}
}

func TestUpdateFileMultilineContent(t *testing.T) {
	prev := "line1\nline2\nline3\n"
	curr := "line1\nmodified2\nline3\n"
	got := UpdateFile("file.txt", prev, curr)

	if !strings.Contains(got, "-line2\n") || !strings.Contains(got, "+modified2\n") {
		t.Fatalf("multiline update not formatted correctly:\n%s", got)
	}
}

func TestNewFileNoTrailingNewline(t *testing.T) {
	got := NewFile("no-newline.txt", "content")
	if !strings.Contains(got, "+content\n") {
		t.Fatalf("file without newline not handled:\n%s", got)
	}
}

func TestUpdateFileComplexMultilineChange(t *testing.T) {
	prev := "a\nb\nc\n"
	curr := "a\nx\ny\nc\n"
	got := UpdateFile("test.txt", prev, curr)

	expected := []string{
		"--- a/test.txt",
		"+++ b/test.txt",
		"-b\n",
		"+x\n",
		"+y\n",
	}

	for _, exp := range expected {
		if !strings.Contains(got, exp) {
			t.Fatalf("expected %q in:\n%s", exp, got)
		}
	}
}
