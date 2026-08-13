package claimguard

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// ClaimsFile is the allowlist of verified quantified claims.
// Accepts claimguard `claims:` or jobkit `allowed:` keys (dogfood bridge).
type ClaimsFile struct {
	Claims  []string `yaml:"claims" json:"claims"`
	Allowed []string `yaml:"allowed" json:"allowed"` // jobkit claims.yaml
}

// Entries returns claims from either schema.
func (c ClaimsFile) Entries() []string {
	if len(c.Claims) > 0 {
		return c.Claims
	}
	return c.Allowed
}

// Violation is an untraceable claim found in text.
type Violation struct {
	Token   string `json:"token"`
	Context string `json:"context"`
}

// Report is the primary product observable.
type Report struct {
	OK          bool        `json:"ok"`
	Violations  []Violation `json:"violations"`
	CheckedText string      `json:"checked_text_preview"`
}

var (
	// Percentages, money, grouped or bare N+, magnitude N+, and bare numbers
	// with 3+ digits. Grouped numbers must be one token: extracting "000+"
	// from "1,000+" makes a correct allowlist fail.
	claimRe = regexp.MustCompile(`(?i)(\$\d[\d,]*(?:\.\d+)?(?:k|m|b)?|\d+(?:\.\d+)?%|\b\d[\d,]*(?:\.\d+)?(?:k|m|b)?\+|\b(?:\d{1,3}(?:,\d{3})+|\d{3,})\b)`)
)

// ExtractClaims finds quantified claim tokens.
func ExtractClaims(text string) []string {
	matches := claimRe.FindAllString(text, -1)
	var out []string
	seen := map[string]struct{}{}
	for _, m := range matches {
		// skip bare calendar-looking 20xx years when alone? keep years as claims unless allowlisted
		key := strings.ToLower(m)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, m)
	}
	return out
}

// Check fails closed when any extracted claim is not in the allowlist.
func normalizeClaim(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, ",", "")
	return s
}

func Check(text string, allow ClaimsFile) Report {
	allowed := allowedClaimTokens(allow.Entries())
	var viol []Violation
	for _, tok := range ExtractClaims(text) {
		if _, ok := allowed[normalizeClaim(tok)]; ok {
			continue
		}
		viol = append(viol, Violation{Token: tok, Context: window(text, tok)})
	}
	prev := text
	if len(prev) > 160 {
		prev = prev[:160] + "…"
	}
	return Report{OK: len(viol) == 0, Violations: viol, CheckedText: prev}
}

// allowedClaimTokens extracts the same quantified shapes from reviewed
// sentences that Check extracts from generated text. Exact token comparison is
// intentional: "$100", "100%", "100+", and bare "100" are different claims.
func allowedClaimTokens(entries []string) map[string]struct{} {
	allowed := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		for _, token := range ExtractClaims(entry) {
			allowed[normalizeClaim(token)] = struct{}{}
		}
	}
	return allowed
}

func window(text, tok string) string {
	i := strings.Index(strings.ToLower(text), strings.ToLower(tok))
	if i < 0 {
		return tok
	}
	start := i - 24
	if start < 0 {
		start = 0
	}
	end := i + len(tok) + 24
	if end > len(text) {
		end = len(text)
	}
	return strings.TrimSpace(text[start:end])
}

// LoadClaimsYAML loads allowlist from path.
func LoadClaimsYAML(path string) (ClaimsFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return ClaimsFile{}, err
	}
	var c ClaimsFile
	if err := yaml.Unmarshal(b, &c); err != nil {
		return ClaimsFile{}, fmt.Errorf("claims yaml: %w", err)
	}
	return c, nil
}
