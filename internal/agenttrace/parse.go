package agenttrace

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// Message is a normalized agent transcript message.
type Message struct {
	Role      string    `json:"role"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// SessionSummary is the primary product observable.
type SessionSummary struct {
	Source       string    `json:"source"`
	MessageCount int       `json:"message_count"`
	Roles        []string  `json:"roles"`
	Messages     []Message `json:"messages"`
	Preview      string    `json:"preview"`
}

type rawLine struct {
	Type      string          `json:"type"`
	Role      string          `json:"role"`
	Message   json.RawMessage `json:"message"`
	Content   json.RawMessage `json:"content"`
	Text      string          `json:"text"`
	Timestamp string          `json:"timestamp"`
}

// ParseJSONL reads Claude/Codex-like JSONL and extracts user/assistant text.
func ParseJSONL(r io.Reader, source string) (SessionSummary, error) {
	sc := bufio.NewScanner(r)
	// 10MB lines for safety
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 10*1024*1024)

	var msgs []Message
	roleSet := map[string]struct{}{}
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var raw rawLine
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		role, text := extractRoleText(raw)
		if role == "" || text == "" {
			continue
		}
		if role != "user" && role != "assistant" && role != "system" {
			continue
		}
		msgs = append(msgs, Message{Role: role, Text: text})
		roleSet[role] = struct{}{}
	}
	if err := sc.Err(); err != nil {
		return SessionSummary{}, err
	}
	if len(msgs) == 0 {
		return SessionSummary{}, fmt.Errorf("no messages found in %s", source)
	}
	roles := make([]string, 0, len(roleSet))
	for r := range roleSet {
		roles = append(roles, r)
	}
	preview := msgs[0].Text
	if len(preview) > 120 {
		preview = preview[:120] + "…"
	}
	return SessionSummary{
		Source:       source,
		MessageCount: len(msgs),
		Roles:        roles,
		Messages:     msgs,
		Preview:      preview,
	}, nil
}

func extractRoleText(raw rawLine) (string, string) {
	role := strings.ToLower(strings.TrimSpace(raw.Role))
	if role == "" && raw.Type != "" {
		// some formats use type: user/assistant
		t := strings.ToLower(raw.Type)
		if t == "user" || t == "assistant" || t == "system" {
			role = t
		}
	}
	if text := strings.TrimSpace(raw.Text); text != "" {
		return role, text
	}
	// message may be string or object
	if len(raw.Message) > 0 {
		var s string
		if err := json.Unmarshal(raw.Message, &s); err == nil && strings.TrimSpace(s) != "" {
			return role, strings.TrimSpace(s)
		}
		var obj struct {
			Role    string `json:"role"`
			Content any    `json:"content"`
		}
		if err := json.Unmarshal(raw.Message, &obj); err == nil {
			if role == "" {
				role = strings.ToLower(obj.Role)
			}
			return role, contentToText(obj.Content)
		}
	}
	if len(raw.Content) > 0 {
		var s string
		if err := json.Unmarshal(raw.Content, &s); err == nil {
			return role, strings.TrimSpace(s)
		}
		return role, contentToText(raw.Content)
	}
	return role, ""
}

func contentToText(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case []any:
		var b strings.Builder
		for _, item := range t {
			if m, ok := item.(map[string]any); ok {
				if tx, ok := m["text"].(string); ok {
					b.WriteString(tx)
					b.WriteByte('\n')
				}
			}
		}
		return strings.TrimSpace(b.String())
	case map[string]any:
		if tx, ok := t["text"].(string); ok {
			return strings.TrimSpace(tx)
		}
	case json.RawMessage:
		var s string
		if err := json.Unmarshal(t, &s); err == nil {
			return strings.TrimSpace(s)
		}
		var arr []map[string]any
		if err := json.Unmarshal(t, &arr); err == nil {
			var b strings.Builder
			for _, m := range arr {
				if tx, ok := m["text"].(string); ok {
					b.WriteString(tx)
					b.WriteByte('\n')
				}
			}
			return strings.TrimSpace(b.String())
		}
	}
	return ""
}
