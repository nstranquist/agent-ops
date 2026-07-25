package evidencevault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CaptureResult is the primary capture observable.
type CaptureResult struct {
	ManifestPath string   `json:"manifest_path"`
	Root         string   `json:"root"`
	Name         string   `json:"name"`
	Assets       int      `json:"assets"`
	Paths        []string `json:"paths"`
}

// Capture walks files under root (or explicit rel paths), hashes them, writes manifest.yaml in root.
func Capture(root, name string, relPaths []string) (CaptureResult, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return CaptureResult{}, err
	}
	st, err := os.Stat(root)
	if err != nil || !st.IsDir() {
		return CaptureResult{}, fmt.Errorf("root must be a directory: %s", root)
	}
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(root)
	}
	var paths []string
	if len(relPaths) == 0 {
		err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				base := info.Name()
				if base == ".git" || base == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			// skip the manifest we are about to write if present
			if rel == "manifest.yaml" {
				return nil
			}
			paths = append(paths, rel)
			return nil
		})
		if err != nil {
			return CaptureResult{}, err
		}
	} else {
		for _, r := range relPaths {
			r = filepath.Clean(r)
			full := filepath.Join(root, r)
			info, err := os.Stat(full)
			if err != nil || info.IsDir() {
				return CaptureResult{}, fmt.Errorf("missing file %s", r)
			}
			paths = append(paths, r)
		}
	}
	if len(paths) == 0 {
		return CaptureResult{}, fmt.Errorf("no files to capture under %s", root)
	}
	var required []RequiredAsset
	for _, rel := range paths {
		sum, err := fileSHA256(filepath.Join(root, rel))
		if err != nil {
			return CaptureResult{}, err
		}
		required = append(required, RequiredAsset{Path: rel, SHA256: sum})
	}
	m := Manifest{Name: name, Required: required}
	b, err := yaml.Marshal(&m)
	if err != nil {
		return CaptureResult{}, err
	}
	manPath := filepath.Join(root, "manifest.yaml")
	if err := os.WriteFile(manPath, b, 0o644); err != nil {
		return CaptureResult{}, err
	}
	return CaptureResult{
		ManifestPath: manPath,
		Root:         root,
		Name:         name,
		Assets:       len(required),
		Paths:        paths,
	}, nil
}

// InitPack creates root dir with a placeholder note and captures it.
func InitPack(root, name string) (CaptureResult, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return CaptureResult{}, err
	}
	note := filepath.Join(root, "README.md")
	if _, err := os.Stat(note); os.IsNotExist(err) {
		body := "# Evidence pack\n\nCaptured by evidence-vault init.\n"
		if err := os.WriteFile(note, []byte(body), 0o644); err != nil {
			return CaptureResult{}, err
		}
	}
	return Capture(root, name, nil)
}
