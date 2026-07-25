package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nstranquist/agent-ops/internal/evidencevault"
)

func evidenceVaultMain() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: evidence-vault verify|capture|init ...")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "verify":
		var root, man string
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--root":
				i++
				root = os.Args[i]
			case "--manifest":
				i++
				man = os.Args[i]
			}
		}
		if root == "" || man == "" {
			fmt.Fprintln(os.Stderr, "usage: evidence-vault verify --root DIR --manifest manifest.yaml")
			os.Exit(2)
		}
		m, err := evidencevault.LoadManifest(man)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		res := evidencevault.Verify(root, m)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		if !res.OK {
			os.Exit(1)
		}
	case "capture":
		var root, name string
		var paths []string
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--root":
				i++
				root = os.Args[i]
			case "--name":
				i++
				name = os.Args[i]
			case "--path":
				i++
				paths = append(paths, os.Args[i])
			}
		}
		if root == "" {
			fmt.Fprintln(os.Stderr, "usage: evidence-vault capture --root DIR [--name NAME] [--path rel]...")
			os.Exit(2)
		}
		res, err := evidencevault.Capture(root, name, paths)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
	case "init":
		var root, name string
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--root":
				i++
				root = os.Args[i]
			case "--name":
				i++
				name = os.Args[i]
			}
		}
		if root == "" {
			fmt.Fprintln(os.Stderr, "usage: evidence-vault init --root DIR [--name NAME]")
			os.Exit(2)
		}
		res, err := evidencevault.InitPack(root, name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
	default:
		fmt.Fprintf(os.Stderr, "unknown %q\n", os.Args[1])
		os.Exit(2)
	}
}
