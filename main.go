// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log"
	"runtime"
	"slices"

	"github.com/spf13/pflag"

	"go.xrstf.de/pkiplot/pkg/loader"
	"go.xrstf.de/pkiplot/pkg/render/mermaid"
)

// These variables get set by ldflags during compilation.
var (
	BuildTag    string
	BuildCommit string
	BuildDate   string // RFC3339 format ("2006-01-02T15:04:05Z07:00")
)

func printVersion() {
	// handle empty values in case `go install` was used
	if BuildCommit == "" {
		fmt.Printf("pkiplot dev, built with %s\n",
			runtime.Version(),
		)
	} else {
		fmt.Printf("pkiplot %s (%s), built with %s on %s\n",
			BuildTag,
			BuildCommit[:10],
			runtime.Version(),
			BuildDate,
		)
	}
}

type globalOptions struct {
	namespace string
	format    string
	verbose   bool
	version   bool
}

var outputFormats = []string{"mermaid"}

func (o *globalOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.namespace, "namespace", "n", o.namespace, "Only include namespace-scoped resources in this namespace (also the default namespace for resources without namespace set)")
	fs.StringVarP(&o.format, "format", "f", o.format, fmt.Sprintf("Output format (one of %v)", outputFormats))
	fs.BoolVarP(&o.verbose, "verbose", "v", o.verbose, "Enable more verbose output")
	fs.BoolVarP(&o.version, "version", "V", o.version, "Show version info and exit immediately")
}

func main() {
	opts := globalOptions{
		format: "mermaid",
	}

	opts.AddFlags(pflag.CommandLine)
	pflag.Parse()

	if opts.version {
		printVersion()
		return
	}

	if !slices.Contains(outputFormats, opts.format) {
		log.Fatalf("Invalid output format %q, must be one of %v.", opts.format, outputFormats)
	}

	args := pflag.Args()
	if len(args) == 0 {
		log.Fatal("No input file(s) provided.")
	}

	loaderOpts := loader.NewDefaultOptions()
	loaderOpts.Namespace = opts.namespace

	manifests, err := loader.LoadPKI(args, loaderOpts)
	if err != nil {
		log.Fatalf("Failed to load all sources: %v.", err)
	}

	var rendered string
	switch opts.format {
	case "mermaid":
		rendered, err = mermaid.Render(manifests)
	default:
		panic("Unknown format, $outputFormats is out of sync with codebase.")
	}

	if err != nil {
		log.Fatalf("Failed rendering PKI: %v.", err)
	}

	fmt.Println(rendered)
}
