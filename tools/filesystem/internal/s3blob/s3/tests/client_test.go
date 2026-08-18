package tests

import (
	"net/http"
	"strings"
	"testing"
)

type trackingBody struct {
	*strings.Reader
	closed bool
}

func (b *trackingBody) Close() error {
	b.closed = true
	return nil
}

func TestExpectHeaders(t *testing.T) {
	headers := http.Header{"Content-Type": {"application/json"}, "X-Request-Id": {"abc-123"}}
	if !ExpectHeaders(headers, map[string]string{"Content-Type": "application/json", "X-Request-Id": "^abc-[0-9]+$"}) {
		t.Fatal("ExpectHeaders rejected matching literal and regexp headers")
	}
	if ExpectHeaders(headers, map[string]string{"Content-Type": "text/plain"}) {
		t.Fatal("ExpectHeaders accepted a mismatched header")
	}
}

func TestClientConsumesMatchingStub(t *testing.T) {
	client := NewClient(&RequestStub{Method: http.MethodGet, URL: "https://example.com/items"})
	req, _ := http.NewRequest(http.MethodGet, "https://example.com/items", nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Request != req || resp.Body != http.NoBody || resp.Header == nil {
		t.Fatal("Client did not normalize the stub response")
	}
	if err := client.AssertNoRemaining(); err != nil {
		t.Fatal(err)
	}
}

func TestClientReportsUnmatchedRequest(t *testing.T) {
	client := NewClient()
	body := &trackingBody{Reader: strings.NewReader("body")}
	req, _ := http.NewRequest(http.MethodPost, "https://example.com/missing", body)
	_, err := client.Do(req)
	if err == nil || !strings.Contains(err.Error(), `Body: "body"`) {
		t.Fatalf("unexpected unmatched request error: %v", err)
	}
	if !body.closed {
		t.Fatal("unmatched request body was not closed")
	}
}
