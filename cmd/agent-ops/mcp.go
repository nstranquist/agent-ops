package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nstranquist/agent-ops/internal/mcpgate"
)

func mcpGateMain() {
	if len(os.Args) < 2 || os.Args[1] != "decide" {
		fmt.Fprintln(os.Stderr, "usage: agent-ops mcp decide --policy policy.yaml --server S --tool T [--audit path.jsonl]")
		os.Exit(2)
	}
	var policyPath, server, tool, auditPath string
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--policy":
			i++
			policyPath = os.Args[i]
		case "--server":
			i++
			server = os.Args[i]
		case "--tool":
			i++
			tool = os.Args[i]
		case "--audit":
			i++
			auditPath = os.Args[i]
		}
	}
	if policyPath == "" || server == "" || tool == "" {
		fmt.Fprintln(os.Stderr, "need --policy --server --tool")
		os.Exit(2)
	}
	p, err := mcpgate.LoadPolicy(policyPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	d := mcpgate.Evaluate(p, server, tool)
	if auditPath != "" {
		if err := mcpgate.AppendAudit(auditPath, d); err != nil {
			fmt.Fprintln(os.Stderr, "audit:", err)
			os.Exit(1)
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(d)
	if d.Action == "deny" {
		os.Exit(1)
	}
}
