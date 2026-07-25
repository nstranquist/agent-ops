package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/nstranquist/agent-ops/internal/agenttrace"
)

func agentTraceMain() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: agenttrace summarize|index|search|discover|index-discovered|doctor|version ...")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "version":
		fmt.Println(version)
	case "doctor":
		sessions, _ := agenttrace.DiscoverSessions(5)
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{
			"version":     version,
			"claude_root": agenttrace.ClaudeRoot(),
			"codex_root":  agenttrace.CodexRoot(),
			"recent":      sessions,
			"ok":          true,
		})
	case "discover":
		limit := 50
		for i := 2; i < len(os.Args); i++ {
			if os.Args[i] == "--limit" {
				i++
				limit, _ = strconv.Atoi(os.Args[i])
			}
		}
		sessions, err := agenttrace.DiscoverSessions(limit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		encode(map[string]any{
			"claude_root": agenttrace.ClaudeRoot(),
			"codex_root":  agenttrace.CodexRoot(),
			"count":       len(sessions),
			"sessions":    sessions,
		})
	case "index-discovered":
		limit := 20
		dbPath := defaultDB()
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--db":
				i++
				dbPath = os.Args[i]
			case "--limit":
				i++
				limit, _ = strconv.Atoi(os.Args[i])
			}
		}
		res, err := agenttrace.IndexDiscovered(dbPath, limit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		encode(res)
	case "summarize":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: agenttrace summarize <session.jsonl>")
			os.Exit(2)
		}
		path := os.Args[2]
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
		sum, err := agenttrace.ParseJSONL(f, path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		encode(sum)
	case "index":
		var dbPath, session string
		dbPath = defaultDB()
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--db":
				i++
				dbPath = os.Args[i]
			default:
				if session == "" && len(os.Args[i]) > 0 && os.Args[i][0] != '-' {
					session = os.Args[i]
				}
			}
		}
		if session == "" {
			fmt.Fprintln(os.Stderr, "usage: agenttrace index --db path.db <session.jsonl>")
			os.Exit(2)
		}
		res, err := agenttrace.IndexSession(dbPath, session)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		encode(res)
	case "search":
		dbPath := defaultDB()
		var query string
		limit := 10
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--db":
				i++
				dbPath = os.Args[i]
			case "--limit":
				i++
				n, _ := strconv.Atoi(os.Args[i])
				if n > 0 {
					limit = n
				}
			default:
				if len(os.Args[i]) > 0 && os.Args[i][0] != '-' {
					if query != "" {
						query += " "
					}
					query += os.Args[i]
				}
			}
		}
		if query == "" {
			fmt.Fprintln(os.Stderr, "usage: agenttrace search --db path.db <query>")
			os.Exit(2)
		}
		hits, err := agenttrace.Search(dbPath, query, limit)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		encode(map[string]any{"query": query, "hits": hits, "count": len(hits)})
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(2)
	}
}


func defaultDB() string {
	if v := os.Getenv("AGENTTRACE_DB"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".agenttrace", "sessions.db")
}
