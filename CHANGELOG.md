# Changelog

All notable changes to `agent-ops` are recorded here.

## [0.3.2] — 2026-08-13

### Fixed

- The claims extractor now keeps grouped numbers such as `1,000+` and `50,000`
  as complete tokens. It no longer reports truncated values such as `000+`.
- JobKit-style `allowed:` entries can contain a reviewed claim sentence. The
  gate extracts and compares the quantified token exactly. A reviewed money,
  percentage, or lower-bound claim cannot authorize a bare count with the same
  digits.
- The claims command now rejects a flag with no value or an unknown flag with a
  usage error instead of panicking.

### Changed

- The product manifest names the ClaimGuard-compatible claims surface so the
  maintained replacement is discoverable after catalog refresh.

## [0.3.1] — 2026-07-25

### Fixed

- **`make leak-scan` was broken in both directions and shipped green twice.**
  Its pattern allowed zero leading characters, so it matched prose mentions of
  the wildcard hostname form — including its own comment and `DISTRIBUTION.md`.
  Separately it scanned `git ls-files` (tracked only), leaving it blind to
  exactly the files a release adds: `DISTRIBUTION.md` passed locally while
  untracked, then failed CI once committed. Patterns interpolated from a make
  variable were also word-split, silently degrading `\.` to `.`.
- Patterns now require a concrete leading label, the file list covers
  `--cached --others --exclude-standard`, and patterns are set via `set --`.

### Added

- **`make leak-scan-selftest`**, wired into `gate`. Plants a synthetic leak,
  asserts the scan rejects it, removes it, asserts the scan goes clean. A gate
  never observed rejecting anything is not known to work.

No change to any command's behaviour; `0.3.0` and `0.3.1` are identical as a
library and CLI.

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

The seven predecessor tools were written on 2026-07-21 and retired into this
repository on 2026-07-25. Their individual commit histories are not replayed
here; `0.3.0` is the first release of the combined tool.
