package toolchain

import "testing"

func TestIsSupportedGoVersion(t *testing.T) {
	t.Parallel()

	if !IsSupportedGoVersion("go1.27.6") {
		t.Fatal("expected Go 1.27 patch release to be supported")
	}
	if IsSupportedGoVersion("go1.25.9") {
		t.Fatal("expected a different Go minor release to be unsupported")
	}
}
