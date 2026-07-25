package prbench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuleReviewerAndLeaderboard(t *testing.T) {
	diff := `diff --git a/x.go b/x.go
--- a/x.go
+++ b/x.go
@@
+func f(u *User) {
+  u.Name.unwrap()
+  q := "SELECT * FROM t WHERE id=" + id
+}
`
	gold := []Finding{{ID: "nil-deref"}, {ID: "sql-inject"}, {ID: "race"}}
	r := RuleReviewer{}
	rep, pred := ReviewAndScore(r, diff, gold)
	if len(pred) < 1 {
		t.Fatal("expected predictions")
	}
	if rep.Precision <= 0 && rep.TruePositives == 0 {
		// may still have FP only; require some structure
	}
	if rep.TruePositives < 1 {
		t.Fatalf("expected at least one TP, pred=%+v report=%+v", pred, rep)
	}
	// empty reviewer baseline
	empty := RunnerResult{Name: "empty", Report: Score(gold, nil), Predicted: 0}
	rules := RunnerResult{Name: r.Name(), Report: rep, Predicted: len(pred)}
	lb := BuildLeaderboard([]RunnerResult{empty, rules})
	if lb.Rows[0].Name != "rules-v1" {
		t.Fatalf("expected rules on top: %+v", lb.Rows)
	}
	html := lb.HTML()
	if !strings.Contains(html, "precision") || !strings.Contains(html, "rules-v1") {
		t.Fatal(html)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "board.html"), []byte(html), 0o644); err != nil {
		t.Fatal(err)
	}
}
