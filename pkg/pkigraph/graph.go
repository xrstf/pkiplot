// SPDX-FileCopyrightText: 2025 Christoph Mewes
// SPDX-License-Identifier: MIT

package pkigraph

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/dominikbraun/graph"

	"go.xrstf.de/pkiplot/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Graph struct {
	g graph.Graph[string, Node]
}

func New() Graph {
	g := graph.New(nodeHash, graph.Directed())

	return Graph{g: g}
}

type Options struct {
	ClusterResourceNamespace string
	ShowSecrets              bool
	ShowSynthetics           bool
}

func NewFromPKI(pki *types.PKI, opt Options) Graph {
	pg := New()

	// add vertices for all PKI elements
	if opt.ShowSecrets {
		for _, secret := range pki.Secrets {
			pg.g.AddVertex(secretNode(secret))
		}
	}
	for _, cert := range pki.Certificates {
		pg.g.AddVertex(certificateNode(cert))
	}
	for _, issuer := range pki.Issuers {
		pg.g.AddVertex(issuerNode(issuer))
	}
	for _, clusterIssuer := range pki.ClusterIssuers {
		pg.g.AddVertex(clusterIssuerNode(clusterIssuer))
	}

	for _, cert := range pki.Certificates {
		hash := certificateHash(cert)

		if opt.ShowSecrets {
			// create an edge between a cert and the secret it produces
			if secretName := cert.Spec.SecretName; secretName != "" {
				if secretNode, ok := pg.ensureSecret(opt, cert.Namespace, secretName); ok {
					pg.g.AddEdge(secretNode.Hash(), hash)
				}
			}
		}

		// create an edge between a cert and its issuer
		ref := cert.Spec.IssuerRef
		switch ref.Kind {
		case "", "Issuer":
			if issuerNode, ok := pg.ensureIssuer(opt, cert.Namespace, ref.Name); ok {
				pg.g.AddEdge(hash, issuerNode.Hash())
			}
		case "ClusterIssuer":
			if clusterIssuerNode, ok := pg.ensureClusterIssuer(opt, ref.Name); ok {
				pg.g.AddEdge(hash, clusterIssuerNode.Hash())
			}
		}
	}

	if opt.ShowSecrets {
		for _, issuer := range pki.Issuers {
			hash := issuerHash(issuer)

			// connect the secret that a CA issuer uses to sign new certs
			if caConfig := issuer.Spec.CA; caConfig != nil && caConfig.SecretName != "" {
				if secretNode, ok := pg.ensureSecret(opt, issuer.Namespace, caConfig.SecretName); ok {
					pg.g.AddEdge(hash, secretNode.Hash())
				}
			}
		}

		for _, clusterIssuer := range pki.ClusterIssuers {
			hash := clusterIssuerHash(clusterIssuer)

			// connect the secret that a CA cluster issuer uses to sign new certs;
			// note that since CI's are cluster-scoped, the special cert-manager resources namespace
			// is used to find the secrets.
			if caConfig := clusterIssuer.Spec.CA; caConfig != nil && caConfig.SecretName != "" {
				if secretNode, ok := pg.ensureSecret(opt, opt.ClusterResourceNamespace, caConfig.SecretName); ok {
					pg.g.AddEdge(hash, secretNode.Hash())
				}
			}
		}
	} else {
		// If secrets are not included, we can still link (cluster)issuers (CAs only) to their
		// secrets based on the secretName ref present in both the issuers and the certificates.
		for _, issuer := range pki.Issuers {
			caConfig := issuer.Spec.CA
			if caConfig != nil && caConfig.SecretName != "" {
				pg.spanSecretEdge(pki, issuerHash(issuer), caConfig.SecretName)
			}
		}

		for _, clusterIssuer := range pki.ClusterIssuers {
			caConfig := clusterIssuer.Spec.CA
			if caConfig != nil && caConfig.SecretName != "" {
				pg.spanSecretEdge(pki, clusterIssuerHash(clusterIssuer), caConfig.SecretName)
			}
		}
	}

	return pg
}

// spanSecretEdge creates and edge between a node that references a secret, and
// certificate(s) (usually one) that create that secret. This is used whenever
// secrets are not included in the graph for brevity.
func (g *Graph) spanSecretEdge(pki *types.PKI, sourceHash string, secretName string) {
	for _, cert := range pki.Certificates {
		if cert.Spec.SecretName != secretName {
			continue
		}

		// create an edge between a cert and the issuer that will make use of it
		cHash := certificateHash(cert)
		g.g.AddEdge(sourceHash, cHash)
	}
}

func (g *Graph) ensureNode(opt Options, n Node) (Node, bool) {
	vertex, err := g.g.Vertex(nodeHash(n))
	if err != nil {
		// Node does not exist in the given PKI data.
		if opt.ShowSynthetics {
			n.Synthetic = true
			g.g.AddVertex(n)
			return n, true
		}

		return Node{}, false
	}

	// Node exists already.
	return vertex, true
}

func (g *Graph) ensureSecret(opt Options, namespace, name string) (Node, bool) {
	return g.ensureNode(opt, secretNode(corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}))
}

func (g *Graph) ensureIssuer(opt Options, namespace, name string) (Node, bool) {
	return g.ensureNode(opt, issuerNode(certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}))
}

func (g *Graph) ensureClusterIssuer(opt Options, name string) (Node, bool) {
	return g.ensureNode(opt, clusterIssuerNode(certmanagerv1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}))
}

func (g *Graph) Raw() graph.Graph[string, Node] {
	return g.g
}
