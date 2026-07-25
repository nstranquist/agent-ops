package evidencevault

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestVerify_PassAndFail(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "proof.txt")
	content := []byte("evidence-ok")
	if err := os.WriteFile(p, content, 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	hexSum := hex.EncodeToString(sum[:])
	m := Manifest{Name: "fixture", Required: []RequiredAsset{
		{Path: "proof.txt", SHA256: hexSum},
	}}
	r := Verify(dir, m)
	if !r.OK {
		t.Fatalf("want pass: %+v", r)
	}
	m2 := Manifest{Name: "bad", Required: []RequiredAsset{
		{Path: "proof.txt", SHA256: "deadbeef"},
	}}
	r2 := Verify(dir, m2)
	if r2.OK {
		t.Fatal("want fail")
	}
}
