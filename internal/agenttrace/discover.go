package agenttrace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SessionFile is a discovered transcript on disk.
type SessionFile struct {
	Provider string    `json:"provider"` // claude|codex
	Path     string    `json:"path"`
	ModTime  time.Time `json:"mod_time"`
	Size     int64     `json:"size"`
}

// ClaudeRoot resolves ~/.claude/projects or AGENTTRACE_CLAUDE_ROOT / AGENTCLIP_CLAUDE_ROOT.
func ClaudeRoot() string {
	for _, k := range []string{"AGENTTRACE_CLAUDE_ROOT", "AGENTCLIP_CLAUDE_ROOT"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "projects")
}

// CodexRoot resolves ~/.codex/sessions or AGENTTRACE_CODEX_ROOT / AGENTCLIP_CODEX_ROOT.
func CodexRoot() string {
	for _, k := range []string{"AGENTTRACE_CODEX_ROOT", "AGENTCLIP_CODEX_ROOT"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codex", "sessions")
}

// DiscoverSessions walks Claude + Codex trees for *.jsonl files.
// limit<=0 means no limit; results newest-first.
func DiscoverSessions(limit int) ([]SessionFile, error) {
	var out []SessionFile
	if err := walkJSONL(ClaudeRoot(), "claude", &out); err != nil && !os.IsNotExist(err) {
		// continue even if missing
	}
	if err := walkJSONL(CodexRoot(), "codex", &out); err != nil && !os.IsNotExist(err) {
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ModTime.After(out[j].ModTime)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func walkJSONL(root, provider string, out *[]SessionFile) error {
	st, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return nil
	}
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		if !strings.HasSuffix(name, ".jsonl") {
			return nil
		}
		// codex rollouts are rollout-*.jsonl; claude are uuid.jsonl
		*out = append(*out, SessionFile{
			Provider: provider,
			Path:     path,
			ModTime:  info.ModTime(),
			Size:     info.Size(),
		})
		return nil
	})
}

// IndexDiscovered indexes up to limit newest sessions into dbPath.
func IndexDiscovered(dbPath string, limit int) (map[string]any, error) {
	sessions, err := DiscoverSessions(limit)
	if err != nil {
		return nil, err
	}
	var indexed, failed []string
	var totalMsgs int
	for _, s := range sessions {
		res, err := IndexSession(dbPath, s.Path)
		if err != nil {
			failed = append(failed, s.Path+": "+err.Error())
			continue
		}
		indexed = append(indexed, s.Path)
		totalMsgs += res.MessageCount
	}
	return map[string]any{
		"db":              dbPath,
		"discovered":      len(sessions),
		"indexed":         len(indexed),
		"failed":          len(failed),
		"message_count":   totalMsgs,
		"claude_root":     ClaudeRoot(),
		"codex_root":      CodexRoot(),
		"indexed_paths":   indexed,
		"failed_paths":    failed,
	}, nil
}
