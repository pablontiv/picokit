//go:build !windows

package autoupdate

import "testing"

func TestPlatformExec_Error(t *testing.T) {
	err := platformExec("/nonexistent-binary-for-test")
	if err == nil {
		t.Fatal("expected error for nonexistent binary in platformExec")
	}
}
