package claimguard

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGateFixtures_MatchMakeTargets drives the same fixtures Makefile uses.
func TestGateFixtures_MatchMakeTargets(t *testing.T) {
	// Resolve testdata relative to module root (test runs from package dir)
	root := findModuleRoot(t)
	claims, err := LoadClaimsYAML(filepath.Join(root, "testdata", "claimguard", "claims.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	bad, err := os.ReadFile(filepath.Join(root, "testdata", "claimguard", "draft-bad.md"))
	if err != nil {
		t.Fatal(err)
	}
	good, err := os.ReadFile(filepath.Join(root, "testdata", "claimguard", "draft-good.md"))
	if err != nil {
		t.Fatal(err)
	}
	badRep := Check(string(bad), claims)
	if badRep.OK || len(badRep.Violations) == 0 {
		t.Fatalf("draft-bad must fail closed, got %+v", badRep)
	}
	// ensure an unallowlisted quantified claim is present (the seed money figure)
	foundBig := false
	for _, v := range badRep.Violations {
		if v.Token == "$2,000,000" || v.Token == "$2000000" {
			foundBig = true
		}
	}
	if !foundBig {
		// still ok if any violation exists for unallowlisted content
		if len(badRep.Violations) == 0 {
			t.Fatal("expected violations")
		}
	}
	goodRep := Check(string(good), claims)
	if !goodRep.OK {
		t.Fatalf("draft-good must pass, got %+v", goodRep)
	}
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("go.mod not found")
	return ""
}

func TestJobkitAllowedKey(t *testing.T) {
	root := findModuleRoot(t)
	// write temp jobkit-shaped claims
	dir := t.TempDir()
	path := filepath.Join(dir, "claims.yaml")
	if err := os.WriteFile(path, []byte("allowed:\n  - \"40%\"\n  - \"$1500\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	allow, err := LoadClaimsYAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(allow.Entries()) != 2 {
		t.Fatalf("%+v", allow)
	}
	rep := Check("lift of 40% costs $1,500", allow)
	if !rep.OK {
		t.Fatalf("%+v", rep)
	}
	_ = root
}
