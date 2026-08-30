package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/cshaiku/loom/internal/loom"
)

func main() {
	if err := loom.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, loom.ErrCommandFailed) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "[fatal] %v\n", err)
		fmt.Fprintln(os.Stderr, "[hint] Run `loom help` or `loom list`.")
		os.Exit(2)
	}
}
