package cmd

import (
	"fmt"
	"os"
)

func Execute(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 2
	}

	switch args[0] {
	case "serve-api":
		return runServeAPI(args[1:])
	case "ingest":
		return runIngest(args[1:])
	case "help", "-h", "--help":
		printUsage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  ask-my-cv serve-api")
	fmt.Fprintln(os.Stderr, "  ask-my-cv ingest ./data")
}
