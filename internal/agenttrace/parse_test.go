package agenttrace

import (
	"strings"
	"testing"
)

func TestParseJSONL_ExtractsRoles(t *testing.T) {
	in := `{"role":"user","text":"hello agent"}
{"role":"assistant","text":"hi human"}
{"type":"meta","text":"ignore"}
`
	sum, err := ParseJSONL(strings.NewReader(in), "fixture.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if sum.MessageCount != 2 {
		t.Fatalf("message_count=%d want 2", sum.MessageCount)
	}
	if sum.Messages[0].Role != "user" || sum.Messages[1].Role != "assistant" {
		t.Fatalf("roles wrong: %+v", sum.Messages)
	}
	if sum.Preview == "" {
		t.Fatal("empty preview")
	}
}

func TestParseJSONL_NestedMessageContent(t *testing.T) {
	in := `{"message":{"role":"user","content":[{"type":"text","text":"nested ok"}]}}
`
	sum, err := ParseJSONL(strings.NewReader(in), "nested.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if sum.Messages[0].Text != "nested ok" {
		t.Fatalf("got %q", sum.Messages[0].Text)
	}
}
