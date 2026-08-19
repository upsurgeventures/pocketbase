package tests

import (
	"testing"

	"github.com/pocketbase/pocketbase/tools/mailer"
)

func TestApiScenarioNormalizedName(t *testing.T) {
	if got := (&ApiScenario{Name: "custom"}).normalizedName(); got != "custom" {
		t.Fatalf("normalizedName() = %q, want custom", got)
	}
	if got := (&ApiScenario{Method: "GET", URL: "/health"}).normalizedName(); got != "GET:/health" {
		t.Fatalf("normalizedName() = %q, want GET:/health", got)
	}
}

func TestTestMailer(t *testing.T) {
	tm := &TestMailer{}
	if tm.TotalSend() != 0 || tm.FirstMessage().Subject != "" || tm.LastMessage().Subject != "" {
		t.Fatal("new TestMailer is not empty")
	}

	first := &mailer.Message{Subject: "first"}
	second := &mailer.Message{Subject: "second"}
	if err := tm.Send(first); err != nil {
		t.Fatal(err)
	}
	if err := tm.Send(second); err != nil {
		t.Fatal(err)
	}
	if tm.TotalSend() != 2 || tm.FirstMessage().Subject != "first" || tm.LastMessage().Subject != "second" {
		t.Fatal("TestMailer did not retain sent messages in order")
	}

	messages := tm.Messages()
	messages[0] = nil
	if tm.Messages()[0] == nil {
		t.Fatal("Messages() did not return a slice copy")
	}
	tm.Reset()
	if tm.TotalSend() != 0 {
		t.Fatal("Reset() did not clear messages")
	}
}
