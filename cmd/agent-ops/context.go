package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nstranquist/agent-ops/internal/contextpack"
)

func contextPackMain() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: contextpack pack|version ...")
		os.Exit(2)
	}
	if os.Args[1] == "version" {
		fmt.Println(version)
		return
	}
	if os.Args[1] != "pack" {
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		os.Exit(2)
	}
	budget := 256
	var chunks []contextpack.Chunk
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--budget":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "missing budget")
				os.Exit(2)
			}
			n, err := strconv.Atoi(args[i])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			budget = n
		case "--fixture":
			i++
			loaded, err := contextpack.LoadChunksJSON(args[i])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			chunks = append(chunks, loaded...)
		case "--file":
			i++
			// id:path:priority:tier
			parts := strings.SplitN(args[i], ":", 4)
			if len(parts) < 2 {
				fmt.Fprintln(os.Stderr, "file format id:path[:priority[:tier]]")
				os.Exit(2)
			}
			id, path := parts[0], parts[1]
			pri := 1
			tier := contextpack.TierShould
			if len(parts) >= 3 {
				pri, _ = strconv.Atoi(parts[2])
			}
			if len(parts) >= 4 {
				tier = contextpack.Tier(parts[3])
			}
			c, err := contextpack.LoadFileChunk(id, path, path, pri, tier)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			chunks = append(chunks, c)
		case "--chunk":
			i++
			// id|source|priority|text  (optional 5th |tier)
			parts := splitN(args[i], "|", 5)
			if len(parts) < 4 {
				fmt.Fprintln(os.Stderr, "chunk format id|source|priority|text[|tier]")
				os.Exit(2)
			}
			p, _ := strconv.Atoi(parts[2])
			tier := contextpack.Tier("")
			if len(parts) >= 5 {
				tier = contextpack.Tier(parts[4])
			}
			chunks = append(chunks, contextpack.Chunk{
				ID: parts[0], Source: parts[1], Priority: p, Text: parts[3], Tier: tier,
			})
		}
	}
	if len(chunks) == 0 {
		fmt.Fprintln(os.Stderr, "no chunks")
		os.Exit(2)
	}
	res, err := contextpack.Pack(chunks, budget)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(res)
}

func splitN(s, sep string, n int) []string {
	out := make([]string, 0, n)
	for len(out) < n-1 {
		i := strings.Index(s, sep)
		if i < 0 {
			break
		}
		out = append(out, s[:i])
		s = s[i+len(sep):]
	}
	out = append(out, s)
	return out
}
