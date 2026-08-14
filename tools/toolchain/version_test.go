package toolchain

import "testing"

func TestIsSupportedGoVersion(t *testing.T) {
	t.Parallel()

	if !IsSupportedGoVersion("go1.26.6") {
		t.Fatal("expected Go 1.26 patch release to be supported")
	}
	if IsSupportedGoVersion("go1.25.9") {
		t.Fatal("expected a different Go minor release to be unsupported")
	}
}
