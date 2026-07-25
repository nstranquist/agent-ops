# Distribution readiness — agent-ops

**Version:** v0.3.0
**License:** MIT
**Module:** `github.com/nstranquist/agent-ops`

## Install

```sh
go install github.com/nstranquist/agent-ops/cmd/agent-ops@latest
agent-ops version
```

Requires Go 1.26+. Pure Go, no cgo — the FTS5 index uses `modernc.org/sqlite`,
so there is no build toolchain requirement beyond Go itself.

Local build:

```sh
make install   # → ~/.local/bin/agent-ops
```

## Verify before publishing

```sh
make publish-ready
```

That runs `go vet`, the full unit suite, and `make smoke` — which executes every
command documented in `README.md` against committed fixtures, including the
`claims` fail-closed assertion.

## Release scan

No release may carry a machine-specific path or a private hostname. CI enforces
the generic classes (`make leak-scan`): absolute home paths and personal
`*.workers.dev` hostnames.

```sh
make leak-scan
```

A private, project-specific banned-fragment list is maintained outside this
repository and is applied before publishing. **That list is deliberately not
committed here** — a scan that embeds the strings it forbids publishes them, and
the fragments are private by definition. Last run 2026-07-25 over 55 tracked
files: clean on all fragments.

## Boundary

The local-first CLI is MIT and lives here. Hosted or multi-tenant control planes
built on top of these lanes stay proprietary and out of this repository.
