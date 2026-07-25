package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nstranquist/agent-ops/internal/fleetboard"
)

func fleetBoardMain() {
	if len(os.Args) < 2 || os.Args[1] != "show" {
		fmt.Fprintln(os.Stderr, "usage: fleetboard show --slots file.json | --dir agent-dir")
		os.Exit(2)
	}
	var slotsPath, dir string
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--slots":
			i++
			slotsPath = os.Args[i]
		case "--dir":
			i++
			dir = os.Args[i]
		}
	}
	var slots []fleetboard.AgentSlot
	var err error
	switch {
	case dir != "":
		slots, err = fleetboard.DiscoverDir(dir)
	case slotsPath != "":
		slots, err = fleetboard.LoadSlots(slotsPath)
	default:
		slots, err = fleetboard.LoadSlots("testdata/slots.json")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	board := fleetboard.MergeBoard(slots)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(board)
}
