package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nstranquist/agent-ops/internal/claimguard"
)

func claimGuardMain() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: agent-ops claims check|version ...")
		os.Exit(2)
	}
	if os.Args[1] == "version" {
		fmt.Println(version)
		return
	}
	if os.Args[1] != "check" {
		fmt.Fprintln(os.Stderr, "usage: agent-ops claims check --claims claims.yaml --file draft.md")
		os.Exit(2)
	}
	claimsPath, filePath, err := parseClaimsCheckArgs(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
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

func parseClaimsCheckArgs(args []string) (claimsPath, filePath string, err error) {
	for index := 0; index < len(args); index++ {
		flag := args[index]
		switch flag {
		case "--claims", "--file":
			if index+1 >= len(args) {
				return "", "", fmt.Errorf("agent-ops claims check: %s requires a value", flag)
			}
			value := args[index+1]
			index++
			if flag == "--claims" {
				claimsPath = value
			} else {
				filePath = value
			}
		default:
			return "", "", fmt.Errorf("agent-ops claims check: unknown argument %q", flag)
		}
	}
	if claimsPath == "" || filePath == "" {
		return "", "", fmt.Errorf("agent-ops claims check: need --claims and --file")
	}
	return claimsPath, filePath, nil
}
