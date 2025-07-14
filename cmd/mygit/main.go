package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/hgsgtk/mygit/commands"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "init":
		if err := commands.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "add":
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "Error: add command requires file path(s)\n")
			os.Exit(1)
		}
		if err := commands.Add(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "commit":
		commitCmd := flag.NewFlagSet("commit", flag.ExitOnError)
		message := commitCmd.String("m", "", "commit message")
		commitCmd.Parse(args)
		
		if *message == "" {
			fmt.Fprintf(os.Stderr, "Error: commit message is required (-m flag)\n")
			os.Exit(1)
		}
		
		if err := commands.Commit(*message); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "log":
		if err := commands.Log(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: mygit <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init                    Initialize a new repository")
	fmt.Println("  add <file>...           Add file(s) to staging area")
	fmt.Println("  commit -m <message>     Commit staged changes")
	fmt.Println("  log                     Show commit history")
	fmt.Println("  help                    Show this help message")
} 