# Contributing to agent-ops

## The loop

```sh
make gate          # vet + test + smoke — run this before every commit
make publish-ready # gate + version, run before tagging
```

## Rules that matter here

1. **`make smoke` must exercise every command the README documents.** The
   consolidation shipped a README with four commands that did not exist, and the
   unit suite stayed green the whole time — because it tests the packages and
   never the CLI surface. If you add a documented command, add it to `smoke` in
   the same commit, or it is undefended.

2. **Keep fail-closed gates deterministic.** No LLM, no network, and no clock
   dependence in any code path that decides pass/fail. `claims` and `evidence`
   are only worth something if their verdicts are reproducible.

3. **Lane bodies read `os.Args` directly.** This is deliberate — it is what made
   the seven-way merge behaviour-preserving rather than a rewrite. If you
   refactor a lane onto `flag.FlagSet`, do it as its own commit with no
   behaviour change, so the diff stays reviewable.

4. **Never commit fixtures containing real session data.** Everything under
   `testdata/` must be synthetic. Real transcripts carry credentials and
   customer data — see `SECURITY.md`.

5. **No absolute local paths in committed files.** The release scan rejects
   machine-specific paths and private hostnames.

## Adding a lane

A new lane is a `<verb>.go` in `cmd/agent-ops/` that reshapes `os.Args` and
calls into `internal/<lane>/`, plus a row in the `subcommands` table in
`main.go`, plus fixtures and a `smoke` line. The usage string must name
`agent-ops <verb>` — never a standalone binary name.
