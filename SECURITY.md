# Security — agent-ops

## Reporting

Report suspected vulnerabilities privately to the maintainer via a GitHub
security advisory on this repository. Please do not open a public issue for an
unpatched vulnerability.

## Threat model

`agent-ops` reads agent session transcripts, context corpora, MCP policies, and
evidence packs from the local filesystem. **Every one of those inputs is
routinely sensitive**, and the tool is designed around that:

- **Offline by default.** No command sends data to a network service. There is
  no telemetry, no update check, and no remote index.
- **Local artifacts only.** The FTS5 index, the MCP audit log, and evidence
  digests are written to paths you name, on your machine.

## What you should treat as sensitive

- **Session transcripts** (`trace`) contain everything an agent read and wrote —
  frequently including credentials pasted into a prompt, private source, and
  customer data. The generated SQLite index inherits that sensitivity. Do not
  commit it; `*.db` is gitignored for this reason.
- **Context corpora** (`context`) are excerpts of your source and documents.
- **MCP audit logs** (`mcp`) record which tools an agent attempted to call, which
  can disclose infrastructure topology.
- **Claims allowlists** (`claims`) may embed unpublished metrics.

## Fail-closed behaviour

`claims check` exits non-zero when a quantified claim is not in the allowlist.
That is the intended behaviour and the reason the command is useful — do not
paper over it in CI with `|| true`. `make smoke` asserts this exit code, so a
change that silently makes the gate permissive fails the build.
