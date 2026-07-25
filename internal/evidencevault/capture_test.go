package evidencevault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureAndVerify_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "proof.txt"), []byte("hello-evidence"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.md"), []byte("# ok\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cap, err := Capture(dir, "roundtrip", nil)
	if err != nil {
		t.Fatal(err)
	}
	if cap.Assets < 2 {
		t.Fatalf("assets=%d", cap.Assets)
	}
	m, err := LoadManifest(cap.ManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	res := Verify(dir, m)
	if !res.OK {
		t.Fatalf("%+v", res)
	}
	// tamper
	if err := os.WriteFile(filepath.Join(dir, "proof.txt"), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	res2 := Verify(dir, m)
	if res2.OK {
		t.Fatal("expected fail after tamper")
	}
}

func TestInitPack(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "newpack")
	cap, err := InitPack(dir, "init-demo")
	if err != nil {
		t.Fatal(err)
	}
	if cap.Assets < 1 {
		t.Fatal(cap)
	}
	m, err := LoadManifest(cap.ManifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(dir, m).OK {
		t.Fatal("verify after init")
	}
}
