.PHONY: test build install vet smoke leak-scan leak-scan-selftest gate publish-ready clean

BIN := bin/agent-ops

test:
	go test ./...

vet:
	go vet ./...

build:
	go build -o $(BIN) ./cmd/agent-ops

install: build
	mkdir -p $(HOME)/.local/bin
	cp -f $(BIN) $(HOME)/.local/bin/agent-ops

# smoke exercises every command documented in README.md against the committed
# fixtures. This target exists because the consolidation shipped a README whose
# `context`, `mcp`, `evidence`, and `fleet` examples named verbs that did not
# exist — the unit tests were green throughout, because they test the packages
# and never the CLI surface the README promises. Any command in the README that
# is not exercised here is undefended.
smoke: build
	@echo "--- trace"
	@./$(BIN) trace summarize testdata/agenttrace/session.jsonl >/dev/null
	@echo "--- context"
	@./$(BIN) context pack --fixture testdata/contextpack/chunks.json --budget 8000 >/dev/null
	@echo "--- mcp (allow)"
	@./$(BIN) mcp decide --policy testdata/mcpgate/policy.yaml --server docs-puller --tool search >/dev/null
	@echo "--- mcp (fail-closed: unlisted server must be denied)"
	@! ./$(BIN) mcp decide --policy testdata/mcpgate/policy.yaml --server github --tool create_pr >/dev/null 2>&1
	@echo "--- mcp (fail-closed: explicit deny rule)"
	@! ./$(BIN) mcp decide --policy testdata/mcpgate/policy.yaml --server shell --tool exec >/dev/null 2>&1
	@echo "--- claims (allow)"
	@./$(BIN) claims check --claims testdata/claimguard/claims.yaml --file testdata/claimguard/draft-good.md >/dev/null
	@echo "--- claims (fail-closed: unverified claim must be rejected)"
	@! ./$(BIN) claims check --claims testdata/claimguard/claims.yaml --file testdata/claimguard/draft-bad.md >/dev/null 2>&1
	@echo "--- claims (malformed flag: usage error, no panic)"
	@status=0; output=$$(./$(BIN) claims check --claims 2>&1) || status=$$?; \
		test "$$status" -eq 2; \
		! printf '%s\n' "$$output" | grep -q 'panic:'
	@echo "--- evidence"
	@./$(BIN) evidence verify --root testdata/evidence-vault/pack --manifest testdata/evidence-vault/manifest.yaml >/dev/null
	@echo "--- bench"
	@./$(BIN) bench score --gold testdata/prbench/gold.json --pred testdata/prbench/pred.json >/dev/null
	@echo "--- fleet"
	@./$(BIN) fleet show --slots testdata/fleetboard/slots.json >/dev/null
	@echo "smoke OK — all seven lanes"

# leak-scan checks the generic leak classes: absolute home paths and personal
# workers.dev hostnames. A private project-specific banned-fragment list is
# applied outside this repository before publishing, and is deliberately not
# committed here — a scan that embeds the strings it forbids publishes them.
#
# Two things this got wrong once, both worth keeping fixed:
#
#  1. Each pattern requires a concrete leading label, so it matches a real
#     hostname and not a prose mention of the wildcard form. The earlier pattern
#     allowed zero leading characters and so matched its own documentation.
#     Corollary: do not write an example hostname anywhere in this file.
#  2. The file list includes untracked-but-not-ignored files. Scanning only
#     tracked files made the local run blind to exactly the new files a release
#     adds — they pass locally while untracked, then fail CI once committed.
#  3. Patterns are set with `set --` so single quotes survive; interpolating a
#     make variable into the loop word-splits it and eats the regex backslashes.
#  4. The self-test builds its probe hostname from fragments. Writing the
#     hostname literally here would plant a match in this very file, which is
#     how the second CI failure happened.
LEAK_FILES = git ls-files --cached --others --exclude-standard -z

leak-scan:
	@status=0; \
	set -- '/Users/[a-z]' '/home/[a-z]' '[a-z0-9][a-z0-9-]*\.workers\.dev'; \
	for pat in "$$@"; do \
	  found=$$($(LEAK_FILES) | xargs -0 grep -rilE -- "$$pat" 2>/dev/null | sort -u); \
	  hits=$$(printf '%s' "$$found" | grep -c . || true); \
	  printf '%-34s hits=%s\n' "$$pat" "$$hits"; \
	  if [ "$$hits" != "0" ]; then printf '%s\n' "$$found" | sed 's/^/    /'; status=1; fi; \
	done; \
	[ $$status = 0 ] && echo "leak-scan OK" || { echo "leak-scan FAILED"; exit 1; }

# leak-scan-selftest proves the scan can actually fail. A gate never observed
# rejecting anything is not known to work — this one shipped green twice while
# broken in opposite directions.
leak-scan-selftest:
	@w=workers; d=dev; probe=./.leak-probe.md; \
	printf 'contact https://example-private.%s.%s/admin\n' "$$w" "$$d" > "$$probe"; \
	if $(MAKE) --no-print-directory leak-scan >/dev/null 2>&1; then \
	  rm -f "$$probe"; echo "leak-scan-selftest FAILED: scan accepted a planted leak"; exit 1; \
	fi; \
	rm -f "$$probe"; \
	$(MAKE) --no-print-directory leak-scan >/dev/null 2>&1 || { \
	  echo "leak-scan-selftest FAILED: scan still red after probe removal"; exit 1; }; \
	echo "leak-scan-selftest OK — planted leak rejected, clean tree accepted"

gate: vet test smoke leak-scan leak-scan-selftest
	@echo "agent-ops gate OK"

publish-ready: gate
	@./$(BIN) version
	@echo "agent-ops publish-ready OK"

clean:
	rm -rf bin
