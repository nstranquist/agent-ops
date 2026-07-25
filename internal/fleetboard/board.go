package fleetboard

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// AgentSlot is one concurrent agent.
type AgentSlot struct {
	ID       string `json:"id"`
	Host     string `json:"host"`
	Worktree string `json:"worktree"`
	Status   string `json:"status"` // running|blocked|done|queued
	Goal     string `json:"goal"`
	Updated  string `json:"updated_at"`
}

// Board is the primary product observable.
type Board struct {
	GeneratedAt string      `json:"generated_at"`
	Agents      []AgentSlot `json:"agents"`
	Count       int         `json:"count"`
}

// MergeBoard builds a board from slots (last id wins).
func MergeBoard(slots []AgentSlot) Board {
	byID := map[string]AgentSlot{}
	for _, s := range slots {
		if s.ID == "" {
			continue
		}
		if s.Status == "" {
			s.Status = "queued"
		}
		if s.Updated == "" {
			s.Updated = time.Now().UTC().Format(time.RFC3339)
		}
		byID[s.ID] = s
	}
	out := make([]AgentSlot, 0, len(byID))
	for _, s := range byID {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return Board{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Agents:      out,
		Count:       len(out),
	}
}

func LoadSlots(path string) ([]AgentSlot, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var slots []AgentSlot
	if err := json.Unmarshal(b, &slots); err != nil {
		return nil, fmt.Errorf("slots json: %w", err)
	}
	return slots, nil
}
