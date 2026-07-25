package fleetboard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverDirAndMerge(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a1.json"), []byte(`{"id":"a1","host":"claude","status":"running","goal":"x","worktree":"/w/a"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "more.json"), []byte(`[{"id":"a2","host":"codex","status":"queued","goal":"y"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	slots, err := DiscoverDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	b := MergeBoard(slots)
	if b.Count < 2 {
		t.Fatalf("%+v", b)
	}
}
