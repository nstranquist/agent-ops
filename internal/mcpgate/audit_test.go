package mcpgate

import (
	"path/filepath"
	"testing"
)

func TestEvaluateAndAudit_AllowDeny(t *testing.T) {
	p := Policy{Default: "deny", Rules: []PolicyRule{
		{Server: "docs", Tool: "search", Action: "allow"},
	}}
	dir := t.TempDir()
	auditPath := filepath.Join(dir, "audit.jsonl")

	allow := Evaluate(p, "docs", "search")
	if allow.Action != "allow" {
		t.Fatalf("want allow: %+v", allow)
	}
	if err := AppendAudit(auditPath, allow); err != nil {
		t.Fatal(err)
	}
	deny := Evaluate(p, "shell", "exec")
	if deny.Action != "deny" {
		t.Fatalf("want deny: %+v", deny)
	}
	if err := AppendAudit(auditPath, deny); err != nil {
		t.Fatal(err)
	}
	entries, err := ReadAudit(auditPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries=%d", len(entries))
	}
	if entries[0].Decision.Action != "allow" || entries[1].Decision.Action != "deny" {
		t.Fatalf("%+v", entries)
	}
	if entries[0].Timestamp == "" || entries[1].Timestamp == "" {
		t.Fatal("missing timestamps")
	}
}
