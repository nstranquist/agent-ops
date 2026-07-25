package main

import (
	"fmt"
	"os"

	"github.com/nstranquist/agent-ops/internal/prbench"
)

func prBenchMain() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: agent-ops bench score|review|leaderboard ...")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "score":
		var goldPath, predPath string
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--gold":
				i++
				goldPath = os.Args[i]
			case "--pred":
				i++
				predPath = os.Args[i]
			}
		}
		gold, err := prbench.LoadFindings(goldPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		pred, err := prbench.LoadFindings(predPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		encode(prbench.Score(gold, pred))
	case "review":
		var diffPath, goldPath string
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--diff":
				i++
				diffPath = os.Args[i]
			case "--gold":
				i++
				goldPath = os.Args[i]
			}
		}
		if diffPath == "" {
			fmt.Fprintln(os.Stderr, "usage: agent-ops bench review --diff patch.diff [--gold gold.json]")
			os.Exit(2)
		}
		b, err := os.ReadFile(diffPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		r := prbench.RuleReviewer{}
		pred := r.Review(string(b))
		out := map[string]any{"runner": r.Name(), "predicted": pred}
		if goldPath != "" {
			gold, err := prbench.LoadFindings(goldPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			out["report"] = prbench.Score(gold, pred)
		}
		encode(out)
	case "leaderboard":
		var diffPath, goldPath, htmlOut string
		for i := 2; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "--diff":
				i++
				diffPath = os.Args[i]
			case "--gold":
				i++
				goldPath = os.Args[i]
			case "--html":
				i++
				htmlOut = os.Args[i]
			}
		}
		if diffPath == "" || goldPath == "" {
			fmt.Fprintln(os.Stderr, "usage: agent-ops bench leaderboard --diff p.diff --gold g.json [--html out.html]")
			os.Exit(2)
		}
		diff, err := os.ReadFile(diffPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		gold, err := prbench.LoadFindings(goldPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		r := prbench.RuleReviewer{}
		rep, pred := prbench.ReviewAndScore(r, string(diff), gold)
		emptyPred := []prbench.Finding{}
		lb := prbench.BuildLeaderboard([]prbench.RunnerResult{
			{Name: "empty", Report: prbench.Score(gold, emptyPred), Predicted: 0},
			{Name: r.Name(), Report: rep, Predicted: len(pred)},
		})
		if htmlOut != "" {
			if err := os.WriteFile(htmlOut, []byte(lb.HTML()), 0o644); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		encode(lb)
	default:
		fmt.Fprintf(os.Stderr, "unknown %q\n", os.Args[1])
		os.Exit(2)
	}
}

