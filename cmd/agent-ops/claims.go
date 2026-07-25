package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nstranquist/agent-ops/internal/claimguard"
)

func claimGuardMain() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: claimguard check|version ...")
		os.Exit(2)
	}
	if os.Args[1] == "version" {
		fmt.Println(version)
		return
	}
	if os.Args[1] != "check" {
		fmt.Fprintln(os.Stderr, "usage: claimguard check --claims claims.yaml --file draft.md")
		os.Exit(2)
	}
	var claimsPath, filePath string
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--claims":
			i++
			claimsPath = os.Args[i]
		case "--file":
			i++
			filePath = os.Args[i]
		}
	}
	if claimsPath == "" || filePath == "" {
		fmt.Fprintln(os.Stderr, "need --claims and --file")
		os.Exit(2)
	}
	allow, err := claimguard.LoadClaimsYAML(claimsPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	body, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	rep := claimguard.Check(string(body), allow)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(rep)
	if !rep.OK {
		os.Exit(1)
	}
}
