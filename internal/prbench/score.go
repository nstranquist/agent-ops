package prbench

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Finding is a review finding id (stable label).
type Finding struct {
	ID       string `json:"id"`
	Severity string `json:"severity,omitempty"`
	Message  string `json:"message,omitempty"`
}

// ScoreReport is the primary product observable.
type ScoreReport struct {
	TruePositives  int     `json:"true_positives"`
	FalsePositives int     `json:"false_positives"`
	FalseNegatives int     `json:"false_negatives"`
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
	F1             float64 `json:"f1"`
	Matched        []string `json:"matched"`
	Extra          []string `json:"extra"`
	Missing        []string `json:"missing"`
}

// Score compares predicted finding IDs to gold labels.
func Score(gold, predicted []Finding) ScoreReport {
	g := map[string]struct{}{}
	for _, f := range gold {
		id := strings.TrimSpace(f.ID)
		if id != "" {
			g[id] = struct{}{}
		}
	}
	p := map[string]struct{}{}
	for _, f := range predicted {
		id := strings.TrimSpace(f.ID)
		if id != "" {
			p[id] = struct{}{}
		}
	}
	var matched, extra, missing []string
	tp := 0
	for id := range p {
		if _, ok := g[id]; ok {
			tp++
			matched = append(matched, id)
		} else {
			extra = append(extra, id)
		}
	}
	fn := 0
	for id := range g {
		if _, ok := p[id]; !ok {
			fn++
			missing = append(missing, id)
		}
	}
	fp := len(extra)
	var precision, recall, f1 float64
	if tp+fp > 0 {
		precision = float64(tp) / float64(tp+fp)
	}
	if tp+fn > 0 {
		recall = float64(tp) / float64(tp+fn)
	}
	if precision+recall > 0 {
		f1 = 2 * precision * recall / (precision + recall)
	}
	return ScoreReport{
		TruePositives:  tp,
		FalsePositives: fp,
		FalseNegatives: fn,
		Precision:      precision,
		Recall:         recall,
		F1:             f1,
		Matched:        matched,
		Extra:          extra,
		Missing:        missing,
	}
}

func LoadFindings(path string) ([]Finding, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var fs []Finding
	if err := json.Unmarshal(b, &fs); err != nil {
		return nil, fmt.Errorf("findings: %w", err)
	}
	return fs, nil
}
