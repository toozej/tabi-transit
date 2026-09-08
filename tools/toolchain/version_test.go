package toolchain

import "testing"

func TestIsSupportedGoVersion(t *testing.T) {
	t.Parallel()

	if !IsSupportedGoVersion("go1.27.1") {
		t.Fatal("expected Go 1.27 patch release to be supported")
	}
	if IsSupportedGoVersion("go1.26.7") {
		t.Fatal("expected a different Go minor release to be unsupported")
	}
}
