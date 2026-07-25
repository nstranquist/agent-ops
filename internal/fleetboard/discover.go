package fleetboard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverDir loads *.json agent slot files from a directory (one slot per file or array).
func DiscoverDir(dir string) ([]AgentSlot, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var slots []AgentSlot
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		// try array first
		var many []AgentSlot
		if err := json.Unmarshal(b, &many); err == nil && len(many) > 0 {
			slots = append(slots, many...)
			continue
		}
		var one AgentSlot
		if err := json.Unmarshal(b, &one); err != nil {
			continue
		}
		if one.ID == "" {
			one.ID = strings.TrimSuffix(e.Name(), ".json")
		}
		slots = append(slots, one)
	}
	return slots, nil
}
