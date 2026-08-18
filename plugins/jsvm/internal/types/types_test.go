package main

import (
	"strings"
	"testing"
)

func TestHooksDeclarations(t *testing.T) {
	got := hooksDeclarations()
	if got == "" {
		t.Fatal("hooksDeclarations() returned an empty declaration set")
	}
	if strings.Contains(got, "declare function onServe") {
		t.Fatal("hooksDeclarations() included the explicitly excluded OnServe hook")
	}
	if !strings.Contains(got, "declare function on") {
		t.Fatalf("hooksDeclarations() did not contain hook declarations: %q", got)
	}
}
