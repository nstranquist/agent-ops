package prbench

import (
	"fmt"
	"sort"
	"strings"
)

// RunnerResult is one leaderboard row.
type RunnerResult struct {
	Name      string      `json:"name"`
	Report    ScoreReport `json:"report"`
	Predicted int         `json:"predicted"`
}

// Leaderboard ranks runners by F1 then precision.
type Leaderboard struct {
	Rows []RunnerResult `json:"rows"`
}

func BuildLeaderboard(rows []RunnerResult) Leaderboard {
	out := append([]RunnerResult(nil), rows...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Report.F1 != out[j].Report.F1 {
			return out[i].Report.F1 > out[j].Report.F1
		}
		return out[i].Report.Precision > out[j].Report.Precision
	})
	return Leaderboard{Rows: out}
}

// HTML renders a minimal leaderboard table.
func (l Leaderboard) HTML() string {
	var b strings.Builder
	b.WriteString("<!doctype html><meta charset=utf-8><title>PR Bench</title>")
	b.WriteString("<h1>PR Bench leaderboard</h1><table border=1 cellpadding=6>")
	b.WriteString("<tr><th>rank</th><th>runner</th><th>precision</th><th>recall</th><th>f1</th><th>predicted</th></tr>")
	for i, r := range l.Rows {
		b.WriteString(fmt.Sprintf(
			"<tr><td>%d</td><td>%s</td><td>%.3f</td><td>%.3f</td><td>%.3f</td><td>%d</td></tr>",
			i+1, r.Name, r.Report.Precision, r.Report.Recall, r.Report.F1, r.Predicted,
		))
	}
	b.WriteString("</table>")
	return b.String()
}
