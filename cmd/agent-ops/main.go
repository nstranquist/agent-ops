// Command agent-ops is the local-first agent operations toolkit: everything
// that produces a RECEIPT about how an AI agent ran, in one binary.
//
// It consolidates seven single-purpose CLIs that were each shipped as their own
// repository. Individually they were ~200-750 line scaffolds with no shared
// story; together they are one coherent claim — you can trace what an agent
// did, budget what it saw, gate what it may call, and verify what it asserted,
// all offline, with durable evidence at each step.
//
//	trace     what the agent DID      — parse/index/search Claude+Codex sessions
//	context   what the agent SAW      — budgeted hydration with an inclusion receipt
//	mcp       what the agent MAY DO   — MCP tool-call policy gate + audit log
//	claims    what the agent SAID     — fail-closed verified-claims gate
//	evidence  what was PROVEN         — evidence packs with SHA-256 digests
//	bench     how good the REVIEW is  — fixture-based review eval with F1
//	fleet     what is RUNNING now     — multi-agent parallel work board
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const version = "v0.3.3"

// subcommands maps the public verb to the consolidated tool entry point and its
// one-line description.
var subcommands = []struct {
	name string
	desc string
	run  func()
}{
	{"trace", "parse, index, and search local Claude/Codex agent sessions", agentTraceMain},
	{"context", "budgeted progressive hydration with an inclusion/omission receipt", contextPackMain},
	{"mcp", "local MCP tool-call policy gate with a durable audit log", mcpGateMain},
	{"claims", "reject unproven numbers in AI-generated text", claimGuardMain},
	{"evidence", "capture and verify evidence packs (required paths + digests)", evidenceVaultMain},
	{"bench", "fixture-based code-review eval with an F1 leaderboard", prBenchMain},
	{"fleet", "multi-agent parallel work board", fleetBoardMain},
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help" {
		usage()
		if len(os.Args) < 2 {
			os.Exit(2)
		}
		return
	}
	verb := os.Args[1]
	if verb == "version" || verb == "--version" {
		fmt.Println(version)
		return
	}
	for _, sub := range subcommands {
		if sub.name != verb {
			continue
		}
		// Each consolidated tool still reads os.Args directly, exactly as it did
		// when it was its own binary. Reshaping argv here means their bodies did
		// not have to be rewritten to be merged — the merge is provably
		// behavior-preserving rather than a rewrite that happens to compile.
		os.Args = append([]string{"agent-ops " + verb}, os.Args[2:]...)
		sub.run()
		return
	}
	fmt.Fprintf(os.Stderr, "agent-ops: unknown subcommand %q\n\n", verb)
	usage()
	os.Exit(2)
}

func usage() {
	w := os.Stderr
	fmt.Fprintf(w, "agent-ops %s — local records of what an AI agent did\n\nUSAGE\n  agent-ops <command> [args]\n\nCOMMANDS\n", version)
	for _, sub := range subcommands {
		fmt.Fprintf(w, "  %-9s %s\n", sub.name, sub.desc)
	}
	fmt.Fprintf(w, "  %-9s %s\n", "version", "print the version")
	fmt.Fprintln(w, "\nEvery command runs offline against local files. Run `agent-ops <command>`")
	fmt.Fprintln(w, "with no arguments for that command's own usage.")
}

// encode is the shared JSON stdout writer. It was duplicated verbatim in the
// agenttrace and prbench binaries; consolidating removed one copy.
func encode(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
