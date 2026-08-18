package fileblob

import (
	"reflect"
	"testing"
)

func TestEscapeKeyRoundTrip(t *testing.T) {
	for _, key := range []string{"file.txt", "dir/file.txt", "dir//file", "../secret", "trailing/", "control\x01"} {
		if got := unescapeKey(escapeKey(key)); got != key {
			t.Errorf("unescapeKey(escapeKey(%q)) = %q", key, got)
		}
	}
}

func TestAttrsRoundTripAndDefault(t *testing.T) {
	path := t.TempDir() + "/object"
	want := xattrs{ContentType: "text/plain", Metadata: map[string]string{"key": "value"}, MD5: []byte{1, 2, 3}}
	if err := setAttrs(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := getAttrs(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("getAttrs() = %#v, want %#v", got, want)
	}

	missing, err := getAttrs(t.TempDir() + "/missing")
	if err != nil || missing.ContentType != "application/octet-stream" {
		t.Fatalf("missing attrs = %#v, %v", missing, err)
	}
}
