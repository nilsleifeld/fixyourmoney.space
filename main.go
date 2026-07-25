package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/nilsleifeld/fixyourmoney.space/internal/consolelog"
	"github.com/nilsleifeld/fixyourmoney.space/internal/devserver"
	"github.com/nilsleifeld/fixyourmoney.space/internal/generator"
)

var verboseLogging bool

func main() {
	slog.SetDefault(consolelog.New(os.Stderr, false))
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "build":
		err = build(os.Args[2:])
	case "serve":
		err = serve(os.Args[2:])
	case "help", "-h", "--help":
		usage()
		return
	default:
		usage()
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		slog.Error("Command failed", "command", os.Args[1], "error", err)
		if !verboseLogging && (os.Args[1] == "build" || os.Args[1] == "serve") {
			slog.Info("Run again with -verbose for detailed build phases", "status", "hint")
		}
		os.Exit(1)
	}
}

func build(arguments []string) error {
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	source := flags.String("source", ".", "site source directory")
	output := flags.String("output", "dist", "output directory")
	var verbose bool
	flags.BoolVar(&verbose, "verbose", false, "show detailed build diagnostics")
	flags.BoolVar(&verbose, "v", false, "show detailed build diagnostics")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("build: unexpected arguments: %v", flags.Args())
	}
	verboseLogging = verbose
	slog.SetDefault(consolelog.New(os.Stderr, verbose))
	return generator.BuildWithLogger(*source, *output, slog.Default())
}

func serve(arguments []string) error {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	source := flags.String("source", ".", "site source directory")
	output := flags.String("output", "dist", "output directory")
	address := flags.String("addr", ":8080", "HTTP listen address")
	var verbose bool
	flags.BoolVar(&verbose, "verbose", false, "show source changes and build diagnostics")
	flags.BoolVar(&verbose, "v", false, "show source changes and build diagnostics")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("serve: unexpected arguments: %v", flags.Args())
	}
	verboseLogging = verbose
	slog.SetDefault(consolelog.New(os.Stderr, verbose))
	return devserver.Serve(*source, *output, *address)
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:
  go run . build [-source .] [-output dist] [-verbose]
  go run . serve [-source .] [-output dist] [-addr :8080] [-verbose]

Options:
  -v, -verbose  Show source changes and detailed build diagnostics.

serve watches source files, rebuilds the site, and reloads browsers automatically.`)
}
