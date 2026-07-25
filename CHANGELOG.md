# Changelog

All notable changes to `agent-ops` are recorded here.

## [0.3.0] — 2026-07-25

First release as a single binary, and the first public release.

### Added

- **`agent-ops`, one binary, seven lanes.** Consolidates seven separately
  shipped CLIs — `agenttrace`, `contextpack`, `mcpgate`, `claimguard`,
  `evidence-vault`, `prbench`, `fleetboard` — into `trace`, `context`, `mcp`,
  `claims`, `evidence`, `bench`, `fleet`. The merge is behaviour-preserving:
  each lane still reads `os.Args` exactly as it did as its own binary, and the
  dispatcher reshapes argv so no tool body was rewritten to be merged.
- Release boundary: `CHANGELOG.md`, `SECURITY.md`, `CONTRIBUTING.md`,
  `DISTRIBUTION.md`, `Makefile`, and CI.
- `make smoke` — runs every command documented in the README against the
  committed fixtures, including the `claims` fail-closed assertion.

### Fixed

- **Every lane's usage string named its pre-merge binary.** `agent-ops context`
  printed `usage: contextpack pack|version ...`, instructing the user to run a
  binary that no longer exists. All 16 usage strings across the seven lanes now
  name `agent-ops <verb>`.
- **Four of seven README examples named verbs that do not exist.** `context
  select` (is `pack`, and the flag is `--fixture`, not `--chunks`), `mcp check`
  (is `decide`), `fleet board` (is `show`), and both `evidence` examples (which
  omitted the required `--root`). Every documented command is now executed by
  `make smoke`, so this class of drift fails CI instead of shipping.

### Notes

The `0.3.0` version is inherited from the consolidated lanes rather than
restarting at `0.1.0`; the binary reported `v0.3.0` before this release and
renumbering it would have made every prior local install ambiguous.

## Prior history

The seven predecessor repositories were created on 2026-07-21 and were marked
`deprecated, superseded by agent-ops` on 2026-07-25. Their histories remain in
those local repositories and are not replayed here.
