package agenttrace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverAndIndex_ClaudeLayout(t *testing.T) {
	tmp := t.TempDir()
	claude := filepath.Join(tmp, "projects", "cwd-enc")
	if err := os.MkdirAll(claude, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"role":"user","text":"dogfood discovery token portfolio-xyz"}
{"role":"assistant","text":"indexed ok"}
`
	if err := os.WriteFile(filepath.Join(claude, "sess-1.jsonl"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AGENTTRACE_CLAUDE_ROOT", filepath.Join(tmp, "projects"))
	t.Setenv("AGENTTRACE_CODEX_ROOT", filepath.Join(tmp, "empty-codex"))
	_ = os.MkdirAll(filepath.Join(tmp, "empty-codex"), 0o755)

	found, err := DiscoverSessions(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 1 || found[0].Provider != "claude" {
		t.Fatalf("%+v", found)
	}
	db := filepath.Join(tmp, "t.db")
	res, err := IndexDiscovered(db, 5)
	if err != nil {
		t.Fatal(err)
	}
	if res["indexed"].(int) != 1 {
		t.Fatalf("%+v", res)
	}
	hits, err := Search(db, "portfolio-xyz", 5)
	if err != nil || len(hits) < 1 {
		t.Fatalf("hits=%v err=%v", hits, err)
	}
}
