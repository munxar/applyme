package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		if err := runInit(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "create":
		if err := runCreate(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "generate":
		if err := runGenerate(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`applyme - stay sane while searching for a job

Usage:
  applyme <command> [arguments]

Commands:
  init [--force]           initialize a new applyme project in the current folder
  create [--force] <id>... create an application folder for one or more job ids
  generate <id>...         generate cv.pdf and cover.pdf for one or more application ids
  help                     show this help message

Flags:
  --force       overwrite existing files instead of skipping them`)
}
