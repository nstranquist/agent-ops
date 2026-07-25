.PHONY: test build install vet smoke leak-scan gate publish-ready clean

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
	@echo "--- evidence"
	@./$(BIN) evidence verify --root testdata/evidence-vault/pack --manifest testdata/evidence-vault/manifest.yaml >/dev/null
	@echo "--- bench"
	@./$(BIN) bench score --gold testdata/prbench/gold.json --pred testdata/prbench/pred.json >/dev/null
	@echo "--- fleet"
	@./$(BIN) fleet show --slots testdata/fleetboard/slots.json >/dev/null
	@echo "smoke OK — all seven lanes"

# leak-scan checks the generic leak classes only: absolute home paths and
# personal *.workers.dev hostnames. A private project-specific banned-fragment
# list is applied outside this repository before publishing, and is deliberately
# not committed here — a scan that embeds the strings it forbids publishes them.
leak-scan:
	@status=0; \
	for pat in '/Users/[a-z]' '/home/[a-z]' '[a-z0-9-]*\.workers\.dev'; do \
	  hits=$$(git ls-files -z | xargs -0 grep -rilE "$$pat" 2>/dev/null | grep -v '^Makefile$$' | wc -l | tr -d ' '); \
	  printf '%-28s hits=%s\n' "$$pat" "$$hits"; \
	  [ "$$hits" = "0" ] || status=1; \
	done; \
	[ $$status = 0 ] && echo "leak-scan OK" || { echo "leak-scan FAILED"; exit 1; }

gate: vet test smoke leak-scan
	@echo "agent-ops gate OK"

publish-ready: gate
	@./$(BIN) version
	@echo "agent-ops publish-ready OK"

clean:
	rm -rf bin
