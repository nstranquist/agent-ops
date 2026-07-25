package evidencevault

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Name     string          `yaml:"name" json:"name"`
	Required []RequiredAsset `yaml:"required" json:"required"`
}

type RequiredAsset struct {
	Path   string `yaml:"path" json:"path"`
	SHA256 string `yaml:"sha256" json:"sha256"`
}

type Result struct {
	OK      bool     `json:"ok"`
	Name    string   `json:"name"`
	Passed  []string `json:"passed"`
	Failed  []string `json:"failed"`
	Message string   `json:"message"`
}

func LoadManifest(path string) (Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := yaml.Unmarshal(b, &m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

func Verify(root string, m Manifest) Result {
	var passed, failed []string
	for _, a := range m.Required {
		full := filepath.Join(root, a.Path)
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			failed = append(failed, a.Path+": missing")
			continue
		}
		if a.SHA256 == "" {
			passed = append(passed, a.Path+": exists")
			continue
		}
		sum, err := fileSHA256(full)
		if err != nil {
			failed = append(failed, a.Path+": hash error")
			continue
		}
		if !equalFoldHex(sum, a.SHA256) {
			failed = append(failed, a.Path+": sha256 mismatch")
			continue
		}
		passed = append(passed, a.Path+": sha256 ok")
	}
	ok := len(failed) == 0 && len(m.Required) > 0
	msg := "all required evidence verified"
	if !ok {
		msg = fmt.Sprintf("%d failed, %d passed", len(failed), len(passed))
	}
	return Result{OK: ok, Name: m.Name, Passed: passed, Failed: failed, Message: msg}
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func equalFoldHex(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'F' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'F' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
