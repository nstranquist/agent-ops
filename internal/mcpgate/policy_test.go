package mcpgate

import "testing"

func TestEvaluate_DenyByDefault(t *testing.T) {
	p := Policy{Default: "deny", Rules: []PolicyRule{
		{Server: "docs", Tool: "search", Action: "allow"},
	}}
	d := Evaluate(p, "docs", "search")
	if d.Action != "allow" {
		t.Fatalf("want allow got %s", d.Action)
	}
	d2 := Evaluate(p, "shell", "exec")
	if d2.Action != "deny" {
		t.Fatalf("want deny got %s", d2.Action)
	}
}

func TestEvaluate_ExplicitDenyWins(t *testing.T) {
	p := Policy{Default: "allow", Rules: []PolicyRule{
		{Server: "fs", Tool: "write", Action: "deny"},
	}}
	d := Evaluate(p, "fs", "write")
	if d.Action != "deny" {
		t.Fatal(d.Action)
	}
}
