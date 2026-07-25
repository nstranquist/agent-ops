package fleetboard

import "testing"

func TestMergeBoard_SlotsAndDedupe(t *testing.T) {
	b := MergeBoard([]AgentSlot{
		{ID: "a1", Host: "claude", Worktree: "wt-a", Status: "running", Goal: "scaffold"},
		{ID: "a2", Host: "codex", Worktree: "wt-b", Status: "blocked", Goal: "tests"},
		{ID: "a1", Host: "claude", Worktree: "wt-a", Status: "done", Goal: "scaffold done"},
	})
	if b.Count != 2 {
		t.Fatalf("count=%d", b.Count)
	}
	if b.Agents[0].ID != "a1" || b.Agents[0].Status != "done" {
		t.Fatalf("%+v", b.Agents[0])
	}
	if len(b.Agents) == 0 {
		t.Fatal("empty agents")
	}
}
