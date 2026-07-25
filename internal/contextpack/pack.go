package contextpack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

// Tier is progressive hydration class (must filled first).
type Tier string

const (
	TierMust  Tier = "must"
	TierShould Tier = "should"
	TierNice  Tier = "nice"
)

// Chunk is a prioritizable context unit.
type Chunk struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Text     string `json:"text"`
	Priority int    `json:"priority"` // lower = more important (within tier)
	Tier     Tier   `json:"tier,omitempty"`
}

// ChunkDecision is a per-chunk verified receipt line.
type ChunkDecision struct {
	ID        string `json:"id"`
	Tier      Tier   `json:"tier"`
	Status    string `json:"status"` // included|truncated|omitted
	Tokens    int    `json:"tokens"`
	Reason    string `json:"reason"`
}

// Receipt records packing decisions.
type Receipt struct {
	BudgetTokens  int             `json:"budget_tokens"`
	UsedTokens    int             `json:"used_tokens"`
	Included      []string        `json:"included"`
	Omitted       []string        `json:"omitted"`
	Substitutions []string        `json:"substitutions"`
	Decisions     []ChunkDecision `json:"decisions"`
	Tokenizer     string          `json:"tokenizer"`
	Hydration     string          `json:"hydration"` // progressive-v1
}

// PackResult is the primary product observable.
type PackResult struct {
	Pack    string  `json:"pack"`
	Receipt Receipt `json:"receipt"`
}

// EstimateTokens is a deterministic offline heuristic (word + rune blend).
func EstimateTokens(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	words := len(strings.Fields(s))
	runes := utf8.RuneCountInString(s)
	est := words
	if runes/4 > est {
		est = runes / 4
	}
	if est < 1 {
		est = 1
	}
	return est
}

func normalizeTier(c Chunk) Tier {
	switch strings.ToLower(string(c.Tier)) {
	case "must", "0":
		return TierMust
	case "should", "1":
		return TierShould
	case "nice", "2", "optional":
		return TierNice
	default:
		// map priority bands when tier unset
		if c.Priority <= 0 {
			return TierMust
		}
		if c.Priority == 1 {
			return TierShould
		}
		return TierNice
	}
}

func tierRank(t Tier) int {
	switch t {
	case TierMust:
		return 0
	case TierShould:
		return 1
	default:
		return 2
	}
}

func sectionText(c Chunk, body string, truncated bool) string {
	flag := ""
	if truncated {
		flag = " [truncated]"
	}
	return fmt.Sprintf("## %s (%s)%s\n%s", c.ID, c.Source, flag, body)
}

// Pack selects chunks under budget with progressive hydration: must → should → nice.
// Token accounting includes the rendered markdown section (header + body).
func Pack(chunks []Chunk, budget int) (PackResult, error) {
	if budget <= 0 {
		return PackResult{}, fmt.Errorf("budget must be positive")
	}
	ordered := append([]Chunk(nil), chunks...)
	for i := range ordered {
		ordered[i].Tier = normalizeTier(ordered[i])
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if tierRank(ordered[i].Tier) != tierRank(ordered[j].Tier) {
			return tierRank(ordered[i].Tier) < tierRank(ordered[j].Tier)
		}
		if ordered[i].Priority != ordered[j].Priority {
			return ordered[i].Priority < ordered[j].Priority
		}
		return ordered[i].ID < ordered[j].ID
	})

	var included, omitted, subs []string
	var decisions []ChunkDecision
	var parts []string
	used := 0

	for _, c := range ordered {
		full := sectionText(c, c.Text, false)
		cost := EstimateTokens(full)
		if used+cost <= budget {
			included = append(included, c.ID)
			parts = append(parts, full)
			used += cost
			decisions = append(decisions, ChunkDecision{
				ID: c.ID, Tier: c.Tier, Status: "included", Tokens: cost,
				Reason: "fits budget",
			})
			continue
		}
		// truncation only when some budget remains and body is non-empty
		remain := budget - used
		if remain >= 10 && strings.TrimSpace(c.Text) != "" {
			// binary-search-ish truncate words until section fits
			truncBody := truncateToTokens(c.Text, remain)
			for attempts := 0; attempts < 8 && truncBody != ""; attempts++ {
				sec := sectionText(c, truncBody, true)
				c2 := EstimateTokens(sec)
				if used+c2 <= budget {
					included = append(included, c.ID+":truncated")
					subs = append(subs, c.ID+":truncated")
					omitted = append(omitted, c.ID+":full")
					parts = append(parts, sec)
					used += c2
					decisions = append(decisions, ChunkDecision{
						ID: c.ID, Tier: c.Tier, Status: "truncated", Tokens: c2,
						Reason: "partial fit under remaining budget",
					})
					truncBody = ""
					break
				}
				// shrink further
				truncBody = truncateToTokens(truncBody, max(1, EstimateTokens(truncBody)/2))
			}
			if truncBody == "" && len(decisions) > 0 && decisions[len(decisions)-1].ID == c.ID && decisions[len(decisions)-1].Status == "truncated" {
				continue
			}
			// if last decision wasn't truncate for this id, fall through to omit
			if len(decisions) > 0 && decisions[len(decisions)-1].ID == c.ID && decisions[len(decisions)-1].Status == "truncated" {
				continue
			}
		}
		omitted = append(omitted, c.ID)
		decisions = append(decisions, ChunkDecision{
			ID: c.ID, Tier: c.Tier, Status: "omitted", Tokens: 0,
			Reason: fmt.Sprintf("need %d tokens, %d remaining", cost, remain),
		})
	}

	// verify used matches sum of included decision tokens
	sum := 0
	for _, d := range decisions {
		if d.Status == "included" || d.Status == "truncated" {
			sum += d.Tokens
		}
	}
	if sum != used {
		used = sum
	}
	if used > budget {
		return PackResult{}, fmt.Errorf("internal: used %d exceeds budget %d", used, budget)
	}

	return PackResult{
		Pack: strings.Join(parts, "\n\n"),
		Receipt: Receipt{
			BudgetTokens:  budget,
			UsedTokens:    used,
			Included:      included,
			Omitted:       omitted,
			Substitutions: subs,
			Decisions:     decisions,
			Tokenizer:     "heuristic-v1",
			Hydration:     "progressive-v1",
		},
	}, nil
}

// LoadChunksJSON loads a chunks fixture.
func LoadChunksJSON(path string) ([]Chunk, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var chunks []Chunk
	if err := json.Unmarshal(b, &chunks); err != nil {
		return nil, err
	}
	return chunks, nil
}

// LoadFileChunk reads a file into a chunk.
func LoadFileChunk(id, source, path string, priority int, tier Tier) (Chunk, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Chunk{}, err
	}
	if id == "" {
		id = filepath.Base(path)
	}
	if source == "" {
		source = path
	}
	return Chunk{ID: id, Source: source, Text: string(b), Priority: priority, Tier: tier}, nil
}

func truncateToTokens(s string, budget int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	var out []string
	used := 0
	for _, w := range words {
		c := EstimateTokens(w)
		if used+c > budget {
			break
		}
		out = append(out, w)
		used += c
	}
	if len(out) == 0 {
		// at least one rune-limited fragment
		r := []rune(strings.TrimSpace(s))
		if len(r) == 0 {
			return ""
		}
		n := budget * 4
		if n < 1 {
			n = 1
		}
		if n > len(r) {
			n = len(r)
		}
		return string(r[:n]) + "…"
	}
	return strings.Join(out, " ") + "…"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
