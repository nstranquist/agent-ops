# Agent Ops case study

## Problem

Seven local-first command-line tools covered related agent-operation tasks.
Each tool worked, but each had its own install path, command surface, tests,
and small public repository. The fragmentation made the system harder to
evaluate and easier to misread as seven unfinished experiments.

The goal was one coherent tool without a behavior rewrite. The public result
also had to exclude private session data, personal paths, and private hosts.

## Design

Agent Ops uses one Go module and one top-level command. Its dispatcher reshapes
the argument list before it calls each existing tool package. Each package sees
the same arguments that its former standalone binary received.

```text
agent-ops
  ├── trace      session discovery and SQLite FTS5 search
  ├── context    token-budgeted context packs and receipts
  ├── mcp        local allow-and-deny policy with an audit log
  ├── claims     fail-closed checks for quantified claims
  ├── evidence   SHA-256 evidence manifests
  ├── bench      fixture-based review scoring
  └── fleet      multi-agent work views
```

This boundary kept the merge small. It also made rollback simple because the
tool logic did not change during consolidation.

## Verification strategy

Unit tests alone missed several public command defects. Four README examples
named commands that did not exist after consolidation. Sixteen usage strings
still named old binaries.

The repository now uses three layers of verification:

1. `go test ./...` and `go vet ./...` check the Go packages.
2. `make smoke` runs every README command against committed synthetic fixtures.
3. `make leak-scan-selftest` plants a synthetic leak and proves that the scan
   rejects it. The command then removes the probe and proves that the clean
   tree passes.

The smoke suite also proves three fail-closed paths. It rejects an unlisted MCP
call, an explicit deny rule, and an unverified quantified claim. The claims
command rejects malformed flags with a usage error instead of a panic.

## Public evidence

- Release [`v0.3.2`](https://github.com/nstranquist/agent-ops/releases/tag/v0.3.2)
  contains the current quantified-claim fixes.
- The [main CI run](https://github.com/nstranquist/agent-ops/actions/runs/31748340999)
  passed at commit `ada599c`.
- The [tag CI run](https://github.com/nstranquist/agent-ops/actions/runs/31748438953)
  passed at the same commit.
- `make publish-ready` runs the package checks, public-command smoke tests,
  leak scan, leak-scan self-test, build, and version smoke.

## Result

The public surface changed from seven repositories and seven installs to one
repository and one install. The combined command keeps seven focused lanes and
uses one CI contract. The retired repositories remain historical context, not
active products.

This result is engineering and publication evidence. It is not evidence of
external adoption, users, revenue, or hiring impact. The repository contains
synthetic fixtures only. Agent Ops does not execute remote agent work or send
local artifacts to a hosted service.

## Interview discussion

The useful engineering questions are about boundaries, not tool count:

- Why preserve package behavior during consolidation?
- Why run documented commands in CI?
- How does a fail-closed claim check avoid authorizing a different number with
  the same digits?
- Why must a leak gate prove that it can fail?
- Which evidence belongs in a public repository, and which evidence must stay
  private?
