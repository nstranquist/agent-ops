package main

import (
	"strings"
	"testing"
)

func TestParseClaimsCheckArgs(t *testing.T) {
	claimsPath, filePath, err := parseClaimsCheckArgs([]string{"--claims", "claims.yaml", "--file", "draft.md"})
	if err != nil {
		t.Fatal(err)
	}
	if claimsPath != "claims.yaml" || filePath != "draft.md" {
		t.Fatalf("paths = %q, %q", claimsPath, filePath)
	}
}

func TestParseClaimsCheckArgsRejectsMissingValues(t *testing.T) {
	for _, args := range [][]string{{"--claims"}, {"--file"}, {"--claims", "claims.yaml", "--file"}} {
		_, _, err := parseClaimsCheckArgs(args)
		if err == nil || !strings.Contains(err.Error(), "requires a value") {
			t.Fatalf("args %q: error = %v, want missing-value error", args, err)
		}
	}
}

func TestParseClaimsCheckArgsRejectsUnknownFlags(t *testing.T) {
	_, _, err := parseClaimsCheckArgs([]string{"--claims", "claims.yaml", "--file", "draft.md", "--wat"})
	if err == nil || !strings.Contains(err.Error(), "unknown argument") {
		t.Fatalf("error = %v, want unknown-argument error", err)
	}
}
