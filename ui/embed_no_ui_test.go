//go:build no_ui

package ui

import "testing"

func TestNoUIBuildHasNoEmbeddedFilesystem(t *testing.T) {
	if DistDirFS != nil {
		t.Fatal("DistDirFS must be nil for no_ui builds")
	}
}
