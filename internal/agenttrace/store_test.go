package agenttrace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndexAndSearch_FTSHit(t *testing.T) {
	dir := t.TempDir()
	session := filepath.Join(dir, "session.jsonl")
	body := `{"role":"user","text":"Summarize the portfolio ranking work"}
{"role":"assistant","text":"I ranked docs-puller first and proposed agenttrace next."}
`
	if err := os.WriteFile(session, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(dir, "sessions.db")
	res, err := IndexSession(dbPath, session)
	if err != nil {
		t.Fatal(err)
	}
	if res.MessageCount != 2 {
		t.Fatalf("count=%d", res.MessageCount)
	}
	hits, err := Search(dbPath, "portfolio", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) < 1 {
		t.Fatal("expected FTS hit for portfolio")
	}
	found := false
	for _, h := range hits {
		if strings.Contains(strings.ToLower(h.Text), "portfolio") {
			found = true
			if h.Role == "" {
				t.Fatal("empty role")
			}
		}
	}
	if !found {
		t.Fatalf("hits missing portfolio token: %+v", hits)
	}
}
