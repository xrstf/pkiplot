// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package mermaid

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/spf13/pflag"

	"go.xrstf.de/pkiplot/pkg/render"
	"go.xrstf.de/pkiplot/pkg/types"

	"k8s.io/apimachinery/pkg/util/sets"
)

type renderer struct{}

var _ render.Renderer = &renderer{}

func New() *renderer {
	return &renderer{}
}

func (r *renderer) AddFlags(fs *pflag.FlagSet) {
	// NOP
}

func (r *renderer) Render(pki *types.PKI) (string, error) {
	var buf types.StringBuilder
	buf.WriteString("graph TB\n")

	// define all nodes

	allIssuers := sets.New[string]()
	for _, clusterIssuer := range pki.ClusterIssuers {
		id := "ci_" + ResourceIdentifier(&clusterIssuer)
		allIssuers.Insert(id)

		buf.Printf("\t%s([%s]):::clusterissuer\n", id, clusterIssuer.GetName())
	}

	for _, issuer := range pki.Issuers {
		id := "i_" + ResourceIdentifier(&issuer)
		allIssuers.Insert(id)

		buf.Printf("\t%s([%s]):::issuer\n", id, issuer.GetName())
	}

	for _, cert := range pki.Certificates {
		id := "c_" + ResourceIdentifier(&cert)

		class := "cert"
		if cert.Spec.IsCA {
			class = "ca"
		}

		buf.Printf("\t%s(%s):::%s\n", id, cert.GetName(), class)
	}

	buf.WriteString("\n")

	// separate certificates into those for which we know the (cluster)issuer
	// and those that seem like orphans
	orphanedCertificates := []certmanagerv1.Certificate{}

	// map of (cluster)issuer => list of regular (non-CA) certs;
	// we only separate non-CA's out in order to serialize those, as they
	// cannot have any further edges leading from the certs
	issuedRegularCertificates := map[string][]certmanagerv1.Certificate{}

	for _, cert := range pki.Certificates {
		issuerID := IssuerResourceIdentifier(cert)

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
		buf.Printf("\t%s", issuerID)

		for idx, cert := range certs {
			arrow := "---"
			if idx == len(certs)-1 {
				arrow = "-->"
			}

			certID := "c_" + ResourceIdentifier(&cert)
			buf.Printf(" %s %s", arrow, certID)
		}

		buf.WriteString("\n")
	}

	// show links for the orphaned certs
	for _, cert := range orphanedCertificates {
		id := "c_" + ResourceIdentifier(&cert)
		issuerID := IssuerResourceIdentifier(cert)

		buf.Printf("\t%s --> %s\n", issuerID, id)
	}

	// link CA-based issuers to their certs
	// (try to deduce this based on Secret names)

	for _, issuer := range pki.Issuers {
		if issuer.Spec.CA == nil {
			continue
		}

		secretName := issuer.Spec.CA.SecretName
		var cert *certmanagerv1.Certificate
		for idx, certificate := range pki.Certificates {
			if certificate.Spec.SecretName == secretName {
				cert = &pki.Certificates[idx]
				break
			}
		}

		if cert == nil {
			continue
		}

		id := "i_" + ResourceIdentifier(&issuer)
		certID := "c_" + ResourceIdentifier(cert)

		buf.Printf("\t%s --> %s\n", certID, id)
	}

	buf.WriteString("\n")
	buf.WriteString("\tclassDef clusterissuer color:#7F7\n")
	buf.WriteString("\tclassDef issuer color:#77F\n")
	buf.WriteString("\tclassDef ca color:#F77\n")
	buf.WriteString("\tclassDef cert color:orange")

	return buf.String(), nil
}

func (r *renderer) ValidateFlags() error {
	// NOP
	return nil
}
