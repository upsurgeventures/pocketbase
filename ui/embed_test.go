//go:build !no_ui

package ui

import (
	"io/fs"
	"testing"
)

func TestEmbeddedUI(t *testing.T) {
	data, err := fs.ReadFile(DistDirFS, "index.html")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("embedded index.html is empty")
	}
}
