package generated

import (
	"strings"
	"testing"
)

func TestEmbeddedTypes(t *testing.T) {
	data, err := Types.ReadFile("types.d.ts")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 || !strings.Contains(string(data), "PocketBase") {
		t.Fatal("embedded types.d.ts is empty or missing PocketBase declarations")
	}
}
