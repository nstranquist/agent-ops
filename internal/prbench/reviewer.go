package prbench

import (
	"regexp"
	"strings"
)

// Reviewer predicts findings from a unified diff / patch text.
type Reviewer interface {
	Name() string
	Review(diff string) []Finding
}

// RuleReviewer is a deterministic adapter (MVP baseline).
type RuleReviewer struct{}

func (RuleReviewer) Name() string { return "rules-v1" }

var (
	reNil    = regexp.MustCompile(`(?i)\b(unwrap|!\.|panic\(|TODO\(security\))`)
	reSQL    = regexp.MustCompile(`(?i)(fmt\.Sprintf\([^\)]*SELECT|\"SELECT .*\"\s*\+|query\s*=\s*\".*\+)`)
	reRace   = regexp.MustCompile(`(?i)(map\[.*\].*=.*go func|go func.*map\[)`)
	reSQLInj = regexp.MustCompile(`(?i)(execute\([^\)]*\+|Query\([^\)]*\+)`)
)

func (RuleReviewer) Review(diff string) []Finding {
	var out []Finding
	seen := map[string]struct{}{}
	add := func(id, sev, msg string) {
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, Finding{ID: id, Severity: sev, Message: msg})
	}
	// only look at added lines
	var added []string
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			added = append(added, line)
		}
	}
	body := strings.Join(added, "\n")
	if reNil.MatchString(body) {
		add("nil-deref", "high", "possible panic/unwrap on added lines")
	}
	if reSQL.MatchString(body) || reSQLInj.MatchString(body) {
		add("sql-inject", "critical", "string-built SQL on added lines")
	}
	if reRace.MatchString(body) {
		add("race", "medium", "concurrent map write pattern")
	}
	return out
}

// ReviewAndScore runs a reviewer against gold labels.
func ReviewAndScore(r Reviewer, diff string, gold []Finding) (ScoreReport, []Finding) {
	pred := r.Review(diff)
	return Score(gold, pred), pred
}
