# agent-ops

**Local-first agent operations, with receipts.**

When an AI agent does work on your machine, seven questions matter and none of
them are answerable by the agent's own claim about itself:

| Question | Command |
|---|---|
| What did it **do**? | `agent-ops trace` |
| What did it **see**? | `agent-ops context` |
| What may it **call**? | `agent-ops mcp` |
| What did it **say**? | `agent-ops claims` |
| What was **proven**? | `agent-ops evidence` |
| How good is the **review**? | `agent-ops bench` |
| What is **running** now? | `agent-ops fleet` |

Every answer is a durable artifact — an index, a receipt, an audit line, a
digest — produced offline from local files. Nothing phones home, and nothing
asks the model to grade itself.

## Install

```sh
go install github.com/nstranquist/agent-ops/cmd/agent-ops@latest
```

Requires Go 1.26+. Pure Go, no cgo: the SQLite FTS5 index is `modernc.org/sqlite`.

## The commands

### `trace` — what the agent did
Parses Claude Code and Codex session JSONL, discovers real local session roots,
and indexes them into SQLite FTS5 so you can search across every session you
have ever run.

```sh
agent-ops trace discover --limit 20
agent-ops trace index-discovered
agent-ops trace search "pressure coordinator"
```

### `context` — what the agent saw
Budgeted progressive hydration. Given a chunk set and a token budget, it emits
the selection **plus a receipt** recording what was included, what was omitted,
and what was truncated — so a bad answer can be traced to a missing input rather
than guessed at.

```sh
agent-ops context pack --fixture chunks.json --budget 8000
```

### `mcp` — what the agent may call
A local policy gate for MCP tool calls. Evaluates allow/deny for
`server + tool` pairs against a YAML policy and appends a durable JSONL audit
log of every decision.

```sh
agent-ops mcp decide --policy policy.yaml --server docs-puller --tool search   # allow → exit 0
agent-ops mcp decide --policy policy.yaml --server github --tool create_pr     # deny  → exit 1
```

**A deny exits non-zero**, so the gate composes directly into a shell guard
without parsing its JSON. The audit line is appended either way.

### `claims` — what the agent said
A fail-closed verified-claims gate for AI-generated text. Any quantified claim
that is not in the allowlist is a violation. This is the thing that stops a
draft asserting "3× faster" when nobody measured it.

```sh
agent-ops claims check --claims claims.yaml --file draft.md
```

### `evidence` — what was proven
Capture and verify evidence packs: a manifest of required paths plus SHA-256
digests, so "the evidence exists" is a checkable statement rather than a
sentence in a status doc.

```sh
agent-ops evidence init    --root ./pack --name release-proof
agent-ops evidence capture --root ./pack --path proof.txt
agent-ops evidence verify  --root ./pack --manifest manifest.yaml
```

### `bench` — how good the review is
Fixture-based code-review evaluation with a deterministic rule reviewer and an
F1 leaderboard. Scores a reviewer against gold labels so "the review got better"
is measured, not asserted.

```sh
agent-ops bench score --gold gold.json --pred pred.json
```

### `fleet` — what is running now
A board for multi-agent parallel work. Merges agent slots from a JSON file or a
directory of slot files into one view of who is doing what.

```sh
agent-ops fleet show --slots slots.json
```

## Why one binary

These began as seven separate repositories. Each was a real, working, tested
tool — and each was 200–750 lines with a sub-1KB README, which is to say each
one read as an abandoned weekend project no matter how well it worked.

The consolidation is not a rewrite. Every tool's logic moved across unchanged;
the merge reshapes `argv` so each embedded command sees exactly the arguments it
saw as its own binary. What changed is that there is now one story, one install,
one test suite, and one place to look — and six fewer half-finished-looking
repositories.

## License

MIT.
