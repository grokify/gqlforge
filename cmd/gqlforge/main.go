package main

import (
	"os"

	"github.com/grokify/gqlforge/cmd/gqlforge/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
