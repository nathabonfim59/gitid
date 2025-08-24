package main

import (
	"fmt"
	"os"
)

func main() {
	setupCompletion()

	if len(os.Args) == 1 {
		runTUI()
		return
	}

	if err := handleCLICommand(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
