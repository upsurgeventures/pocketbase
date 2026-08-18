package templates

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
)

func TestLayoutRendersHTMLContent(t *testing.T) {
	tmpl, err := template.New("mail").Parse(Layout + HTMLBody)
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	err = tmpl.Execute(&output, map[string]any{"HTMLContent": template.HTML("<strong>Hello</strong>")})
	if err != nil {
		t.Fatal(err)
	}
	if got := output.String(); !strings.Contains(got, "<strong>Hello</strong>") || !strings.Contains(got, "<!DOCTYPE html") {
		t.Fatalf("unexpected rendered template: %q", got)
	}
}
