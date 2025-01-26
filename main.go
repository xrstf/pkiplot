// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/dominikbraun/graph/draw"
	"github.com/spf13/pflag"

	"go.xrstf.de/pkiplot/pkg/loader"
	"go.xrstf.de/pkiplot/pkg/pkigraph"
	"go.xrstf.de/pkiplot/pkg/render"
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
	namespace    string
	graphOptions pkigraph.Options
	format       string
	verbose      bool
	version      bool
}

func (o *globalOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.namespace, "namespace", "n", o.namespace, "Only include namespace-scoped resources in this namespace (also the default namespace for resources without namespace set)")
	fs.StringVarP(&o.format, "format", "f", o.format, fmt.Sprintf("Output format (one of %v)", render.All()))
	fs.BoolVarP(&o.verbose, "verbose", "v", o.verbose, "Enable more verbose output")
	fs.BoolVarP(&o.version, "version", "V", o.version, "Show version info and exit immediately")

	fs.StringVarP(&o.graphOptions.ClusterResourceNamespace, "cluster-resource-namespace", "", o.graphOptions.ClusterResourceNamespace, "cert-manager's cluster resource namespace, used to find secrets referenced by cluster-scoped objects")
	fs.BoolVarP(&o.graphOptions.ShowSecrets, "show-secrets", "", o.graphOptions.ShowSecrets, "Include Kubernetes Secrets in the graph")
	fs.BoolVarP(&o.graphOptions.ShowSynthetics, "show-synthetics", "", o.graphOptions.ShowSynthetics, "Include objects in the graph that are only referenced, but not included in the YAML files (e.g. missing Secrets or Issuers)")
}

func main() {
	allRenderers := render.All()
	for _, name := range allRenderers {
		r, _ := render.Get(name)
		r.AddFlags(pflag.CommandLine)
	}

	opts := globalOptions{
		format: "mermaid",
		graphOptions: pkigraph.Options{
			ClusterResourceNamespace: "cert-manager",
		},
	}

	opts.AddFlags(pflag.CommandLine)
	pflag.Parse()

	if opts.version {
		printVersion()
		return
	}

	args := pflag.Args()
	if len(args) == 0 {
		log.Fatal("No input file(s) provided.")
	}

	renderer, exists := render.Get(opts.format)
	if !exists {
		log.Fatalf("Invalid output format %q, must be one of %v.", opts.format, render.All())
	}

	if err := renderer.ValidateFlags(); err != nil {
		log.Fatalf("Invalid command line flags: %v.", err)
	}

	loaderOpts := loader.NewDefaultOptions()
	loaderOpts.Namespace = opts.namespace

	pki, err := loader.LoadPKI(args, loaderOpts)
	if err != nil {
		log.Fatalf("Failed to load all sources: %v.", err)
	}

	g := pkigraph.NewFromPKI(pki, opts.graphOptions)
	file, _ := os.Create("./simple.gv")
	_ = draw.DOT(g.Raw(), file)

	rendered, err := renderer.RenderGraph(g)
	if err != nil {
		log.Fatalf("Failed rendering PKI: %v.", err)
	}

	fmt.Println(rendered)
}
