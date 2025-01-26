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

func NewFromPKI(pki *types.PKI, clusterResourceNamespace string) Graph {
	pg := New()

	// add vertices for all PKI elements
	for _, secret := range pki.Secrets {
		pg.g.AddVertex(secretNode(secret))
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

		// create an edge between a cert and the secret it produces
		if secretName := cert.Spec.SecretName; secretName != "" {
			secretNode := pg.ensureSecret(cert.Namespace, secretName)
			pg.g.AddEdge(hash, secretNode.Hash())
		}

		// create an edge between a cert and its issuer
		ref := cert.Spec.IssuerRef
		switch ref.Kind {
		case "", "Issuer":
			issuerNode := pg.ensureIssuer(cert.Namespace, ref.Name)
			pg.g.AddEdge(hash, issuerNode.Hash())
		case "ClusterIssuer":
			clusterIssuerNode := pg.ensureClusterIssuer(ref.Name)
			pg.g.AddEdge(hash, clusterIssuerNode.Hash())
		}
	}

	for _, issuer := range pki.Issuers {
		hash := issuerHash(issuer)

		// connect the secret that a CA issuer uses to sign new certs
		if caConfig := issuer.Spec.CA; caConfig != nil && caConfig.SecretName != "" {
			secretNode := pg.ensureSecret(issuer.Namespace, caConfig.SecretName)
			pg.g.AddEdge(hash, secretNode.Hash())
		}
	}

	for _, clusterIssuer := range pki.ClusterIssuers {
		hash := clusterIssuerHash(clusterIssuer)

		// connect the secret that a CA cluster issuer uses to sign new certs;
		// note that since CI's are cluster-scoped, the special cert-manager resources namespace
		// is used to find the secrets.
		if caConfig := clusterIssuer.Spec.CA; caConfig != nil && caConfig.SecretName != "" {
			secretNode := pg.ensureSecret(clusterResourceNamespace, caConfig.SecretName)
			pg.g.AddEdge(hash, secretNode.Hash())
		}
	}

	return pg
}

func (g *Graph) ensureNode(n Node) Node {
	n.Synthetic = true

	vertex, err := g.g.Vertex(nodeHash(n))
	if err != nil {
		g.g.AddVertex(n)
		return n
	}

	return vertex
}

func (g *Graph) ensureSecret(namespace, name string) Node {
	return g.ensureNode(secretNode(corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}))
}

func (g *Graph) ensureIssuer(namespace, name string) Node {
	return g.ensureNode(issuerNode(certmanagerv1.Issuer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}))
}

func (g *Graph) ensureClusterIssuer(name string) Node {
	return g.ensureNode(clusterIssuerNode(certmanagerv1.ClusterIssuer{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}))
}

func (g *Graph) Raw() graph.Graph[string, Node] {
	return g.g
}
