package main

import (
	"os"

	"github.com/immmmmmmu/plsnt/cmd/root"
)

func main() {
	if err := root.Execute(); err != nil {
		// site diff --exit-code returns a sentinel; cobra suppresses its
		// printing (SilenceErrors), and we still want exit 1 here. For
		// every other error cobra has already printed it. Both paths just
		// need a non-zero exit, so a single Exit(1) is enough.
		os.Exit(1)
	}
}
