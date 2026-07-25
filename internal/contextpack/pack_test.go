package contextpack

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPack_ProgressiveMustFirstAndBudget(t *testing.T) {
	chunks := []Chunk{
		{ID: "nice-long", Source: "x", Text: strings.Repeat("word ", 80), Priority: 5, Tier: TierNice},
		{ID: "must-core", Source: "y", Text: "critical agent instruction", Priority: 0, Tier: TierMust},
		{ID: "should-mid", Source: "z", Text: "useful secondary context about packing", Priority: 1, Tier: TierShould},
	}
	res, err := Pack(chunks, 30)
	if err != nil {
		t.Fatal(err)
	}
	if res.Receipt.UsedTokens > res.Receipt.BudgetTokens {
		t.Fatalf("used %d > budget %d", res.Receipt.UsedTokens, res.Receipt.BudgetTokens)
	}
	if res.Receipt.Hydration != "progressive-v1" {
		t.Fatal(res.Receipt.Hydration)
	}
	// must should be included first among decisions that are included
	foundMust := false
	for _, d := range res.Receipt.Decisions {
		if d.ID == "must-core" && (d.Status == "included" || d.Status == "truncated") {
			foundMust = true
		}
	}
	if !foundMust {
		t.Fatalf("must-core missing: %+v", res.Receipt.Decisions)
	}
	// must appears in pack before nice if both present
	if strings.Contains(res.Pack, "must-core") && strings.Contains(res.Pack, "nice-long") {
		if strings.Index(res.Pack, "must-core") > strings.Index(res.Pack, "nice-long") {
			t.Fatal("must should appear before nice in pack")
		}
	}
}

func TestPack_CountsHeaderTokens(t *testing.T) {
	// tiny budget where body alone would fit but full section might not without truncation
	chunks := []Chunk{
		{ID: "a", Source: "s", Text: "alpha beta gamma delta epsilon zeta", Priority: 0, Tier: TierMust},
	}
	res, err := Pack(chunks, 8)
	if err != nil {
		t.Fatal(err)
	}
	if res.Receipt.UsedTokens > 8 {
		t.Fatalf("used %d", res.Receipt.UsedTokens)
	}
	if len(res.Receipt.Decisions) != 1 {
		t.Fatalf("%+v", res.Receipt.Decisions)
	}
}

func TestLoadFileChunk(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "note.md")
	if err := os.WriteFile(p, []byte("hello from file source"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := LoadFileChunk("note", "fs", p, 0, TierMust)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(c.Text, "hello") {
		t.Fatal(c.Text)
	}
	res, err := Pack([]Chunk{c}, 50)
	if err != nil {
		t.Fatal(err)
	}
	if res.Receipt.UsedTokens == 0 || res.Pack == "" {
		t.Fatal(res)
	}
}
