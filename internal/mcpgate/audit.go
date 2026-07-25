package mcpgate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// AuditEntry is one durable policy decision record.
type AuditEntry struct {
	Timestamp string   `json:"ts"`
	Decision  Decision `json:"decision"`
}

// AppendAudit appends one JSONL decision line to path (creates parent dirs as needed).
func AppendAudit(path string, d Decision) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("audit path empty")
	}
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return err
	}
	entry := AuditEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Decision:  d,
	}
	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(b, '\n'))
	return err
}

// ReadAudit loads all audit entries from a JSONL file.
func ReadAudit(path string) ([]AuditEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []AuditEntry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var e AuditEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("audit line: %w", err)
		}
		out = append(out, e)
	}
	return out, sc.Err()
}

func dirOf(path string) string {
	i := strings.LastIndex(path, "/")
	if i <= 0 {
		return "."
	}
	return path[:i]
}
