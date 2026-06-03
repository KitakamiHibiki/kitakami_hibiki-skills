package cmd

import (
	"flag"
	"fmt"
	"os"
)

const appName = "skills-cli"

var (
	verbose bool
)

func Execute() {
	// Pre-scan: check for -v / --verbose anywhere in args.
	for _, a := range os.Args {
		if a == "-v" || a == "--verbose" {
			verbose = true
			break
		}
	}

	root := flag.NewFlagSet(appName, flag.ExitOnError)
	root.BoolVar(&verbose, "verbose", false, " enable verbose output")
	root.BoolVar(&verbose, "v", false, " enable verbose output (shorthand)")

	// Parse root flags from os.Args[1:]; stops at the first non-flag argument.
	root.Parse(os.Args[1:])

	// The first non-flag arg is the command name.
	cmd := root.Arg(0)
	args := root.Args()[1:]

	if cmd == "" {
		printUsage(root)
		os.Exit(0)
	}

	switch cmd {
	case "help", "-h", "--help":
		printUsage(root)
	case "version", "--version":
		printVersion()
	case "pixiv":
		runPixiv(args)
	default:
		if err := run(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}

func run(subcommand string) error {
	fmt.Printf("skills-cli: running %q (verbose=%v)\n", subcommand, verbose)
	return nil
}

func printUsage(fs *flag.FlagSet) {
	fmt.Printf("Usage: %s <command> [flags]\n\n", appName)
	fmt.Println("Commands:")
	fmt.Println("  help        show this help")
	fmt.Println("  version     print version info")
	fmt.Println("  pixiv       pixiv related operations")
	fmt.Println()
	fmt.Println("Flags:")
	fs.PrintDefaults()
}

func printVersion() {
	fmt.Printf("%s 0.1.0\n", appName)
}
