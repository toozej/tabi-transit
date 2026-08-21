// Package toolchain contains small, dependency-free checks for repository tooling.
package toolchain

import "strings"

// IsSupportedGoVersion reports whether a Go toolchain identifies as the Phase 0
// baseline major/minor release. Patch versions are deliberately accepted.
func IsSupportedGoVersion(version string) bool {
	return strings.HasPrefix(version, "go1.27.")
}
