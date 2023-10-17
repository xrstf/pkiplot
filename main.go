// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"log"
	"runtime"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"

	"go.xrstf.de/pkiplot/pkg/loader"
	"go.xrstf.de/pkiplot/pkg/mermaid"
	"go.xrstf.de/pkiplot/pkg/types"
)

// These variables get set by ldflags during compilation.
var (
	BuildTag    string
	BuildCommit string
	BuildDate   string // RFC3339 format ("2006-01-02T15:04:05Z07:00")
)

func printVersion() {
	fmt.Printf(
		"pkiplot %s (%s), built with %s on %s\n",
		BuildTag,
		BuildCommit[:10],
		runtime.Version(),
		BuildDate,
	)
}

type globalOptions struct {
	namespace string
	verbose   bool
	version   bool
}

func (o *globalOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.namespace, "namespace", "n", o.namespace, "Only include namespace-scoped resources in this namespace (also the default namespace for resources without namespace set)")
	fs.BoolVarP(&o.verbose, "verbose", "v", o.verbose, "Enable more verbose output")
	fs.BoolVarP(&o.version, "version", "V", o.version, "Show version info and exit immediately")
}

func main() {
	opts := globalOptions{}

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

	loaderOpts := loader.NewDefaultOptions()
	loaderOpts.Namespace = opts.namespace

	manifests, err := loader.LoadPKI(args, loaderOpts)
	if err != nil {
		log.Fatalf("Failed to load all sources: %v", err)
	}

	render(manifests)
}

func render(m *types.PKI) {
	fmt.Println("graph TB")

	// define all nodes

	allIssuers := sets.New[string]()
	for _, clusterIssuer := range m.ClusterIssuers {
		id := "ci_" + mermaid.ResourceIdentifier(&clusterIssuer)
		allIssuers.Insert(id)

		fmt.Printf("\t%s([%s]):::clusterissuer\n", id, clusterIssuer.GetName())
	}

	for _, issuer := range m.Issuers {
		id := "i_" + mermaid.ResourceIdentifier(&issuer)
		allIssuers.Insert(id)

		fmt.Printf("\t%s([%s]):::issuer\n", id, issuer.GetName())
	}

	for _, cert := range m.Certificates {
		id := "c_" + mermaid.ResourceIdentifier(&cert)

		class := "cert"
		if cert.Spec.IsCA {
			class = "ca"
		}

		fmt.Printf("\t%s(%s):::%s\n", id, cert.GetName(), class)
	}

	fmt.Println("")

	// separate certificates into those for which we know the (cluster)issuer
	// and those that seem like orphans
	orphanedCertificates := []certmanagerv1.Certificate{}

	// map of (cluster)issuer => list of regular (non-CA) certs;
	// we only separate non-CA's out in order to serialize those, as they
	// cannot have any further edges leading from the certs
	issuedRegularCertificates := map[string][]certmanagerv1.Certificate{}

	for _, cert := range m.Certificates {
		issuerID := mermaid.IssuerResourceIdentifier(cert)

		if !cert.Spec.IsCA && allIssuers.Has(issuerID) {
			issuedRegularCertificates[issuerID] = append(issuedRegularCertificates[issuerID], cert)
		} else {
			orphanedCertificates = append(orphanedCertificates, cert)
		}
	}

	// now we can display a serialized list of certs for a CA,
	// e.g. "i_my_ca --- c_a --- c_b ---> c_c", which simply makes
	// the diagram look nicer
	for issuerID, certs := range issuedRegularCertificates {
		fmt.Printf("\t%s", issuerID)

		for idx, cert := range certs {
			arrow := "---"
			if idx == len(certs)-1 {
				arrow = "-->"
			}

			certID := "c_" + mermaid.ResourceIdentifier(&cert)
			fmt.Printf(" %s %s", arrow, certID)
		}

		fmt.Println("")
	}

	// show links for the orphaned certs
	for _, cert := range orphanedCertificates {
		id := "c_" + mermaid.ResourceIdentifier(&cert)
		issuerID := mermaid.IssuerResourceIdentifier(cert)

		fmt.Printf("\t%s --> %s\n", issuerID, id)
	}

	// link CA-based issuers to their certs
	// (try to deduce this based on Secret names)

	for _, issuer := range m.Issuers {
		if issuer.Spec.CA == nil {
			continue
		}

		secretName := issuer.Spec.CA.SecretName
		var cert *certmanagerv1.Certificate
		for idx, certificate := range m.Certificates {
			if certificate.Spec.SecretName == secretName {
				cert = &m.Certificates[idx]
				break
			}
		}

		if cert == nil {
			continue
		}

		id := "i_" + mermaid.ResourceIdentifier(&issuer)
		certID := "c_" + mermaid.ResourceIdentifier(cert)

		fmt.Printf("\t%s --> %s\n", certID, id)
	}

	fmt.Println("")
	fmt.Println("\tclassDef clusterissuer color:#7F7")
	fmt.Println("\tclassDef issuer color:#77F")
	fmt.Println("\tclassDef ca color:#F77")
	fmt.Println("\tclassDef cert color:orange")
}
