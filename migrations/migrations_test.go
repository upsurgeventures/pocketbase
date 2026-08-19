package migrations

import (
	"reflect"
	"testing"
)

func TestCollectionIdChecksum(t *testing.T) {
	tests := []struct {
		typ  string
		name string
		want string
	}{
		{"auth", "_superusers", "pbc_3142635823"},
		{"", "", "pbc_0"},
	}

	for _, test := range tests {
		if got := collectionIdChecksum(test.typ, test.name); got != test.want {
			t.Errorf("collectionIdChecksum(%q, %q) = %q, want %q", test.typ, test.name, got, test.want)
		}
	}
}

func TestFieldIdChecksum(t *testing.T) {
	tests := []struct {
		typ  string
		name string
		want string
	}{
		{"email", "email", "email3885137012"},
		{"text", "id", "text3208210256"},
	}

	for _, test := range tests {
		if got := fieldIdChecksum(test.typ, test.name); got != test.want {
			t.Errorf("fieldIdChecksum(%q, %q) = %q, want %q", test.typ, test.name, got, test.want)
		}
	}
}

func TestMigrateRule(t *testing.T) {
	if got := migrateRule(nil); got != nil {
		t.Fatalf("migrateRule(nil) = %q, want nil", *got)
	}

	original := "@request.data.title = 'x' && @request.data.owner = @request.auth.id"
	want := "@request.body.title = 'x' && @request.body.owner = @request.auth.id"
	got := migrateRule(&original)
	if got == nil || *got != want {
		t.Fatalf("migrateRule() = %v, want %q", got, want)
	}
	if original != "@request.data.title = 'x' && @request.data.owner = @request.auth.id" {
		t.Fatalf("migrateRule modified its input: %q", original)
	}
}

func TestGetMapVal(t *testing.T) {
	data := map[string]any{
		"meta": map[string]any{
			"mail": map[string]any{"enabled": false},
			"name": "PocketBase",
		},
	}

	tests := []struct {
		name string
		keys []string
		want any
	}{
		{"nested value", []string{"meta", "mail", "enabled"}, false},
		{"direct value", []string{"meta", "name"}, "PocketBase"},
		{"missing key", []string{"meta", "missing"}, nil},
		{"non-map intermediate", []string{"meta", "name", "part"}, nil},
		{"no keys", nil, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := getMapVal(data, test.keys...); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("getMapVal(%v) = %#v, want %#v", test.keys, got, test.want)
			}
		})
	}
}

func TestZeroFallback(t *testing.T) {
	if got := zeroFallback("", "fallback"); got != "fallback" {
		t.Errorf("zeroFallback empty string = %q, want fallback", got)
	}
	if got := zeroFallback("value", "fallback"); got != "value" {
		t.Errorf("zeroFallback nonempty string = %q, want value", got)
	}
	if got := zeroFallback(0, 42); got != 42 {
		t.Errorf("zeroFallback zero integer = %d, want 42", got)
	}
	if got := zeroFallback(false, true); !got {
		t.Error("zeroFallback false boolean = false, want true")
	}
}

func TestToRelationField(t *testing.T) {
	base := map[string]any{
		"id":          "relation123",
		"name":        "owner",
		"system":      true,
		"required":    true,
		"presentable": true,
		"options": map[string]any{
			"collectionId":  "users",
			"cascadeDelete": true,
			"minSelect":     1,
			"maxSelect":     0,
		},
	}

	got := toRelationField(base)
	if got["maxSelect"] != 2147483647 {
		t.Errorf("maxSelect = %#v, want unlimited legacy fallback", got["maxSelect"])
	}
	if got["collectionId"] != "users" || got["cascadeDelete"] != true || got["minSelect"] != 1 {
		t.Errorf("relation options were not preserved: %#v", got)
	}

	base["options"].(map[string]any)["maxSelect"] = 3
	if got := toRelationField(base)["maxSelect"]; got != 3 {
		t.Errorf("positive maxSelect = %#v, want 3", got)
	}
}
