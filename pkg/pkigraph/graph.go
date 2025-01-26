// SPDX-FileCopyrightText: 2025 Christoph Mewes
// SPDX-License-Identifier: MIT

package pkigraph

import (
	"github.com/dominikbraun/graph"

	"go.xrstf.de/pkiplot/pkg/types"
)

type Graph struct {
	g graph.Graph[string, Node]
}

func New() Graph {
	g := graph.New(nodeHash, graph.Directed())

	return Graph{g: g}
}

func NewFromPKI(pki *types.PKI) Graph {
	grph := New()

	// add vertices for all PKI elements
	for _, cert := range pki.Certificates {
		grph.g.AddVertex(certificateNode(cert))
	}
	for _, issuer := range pki.Issuers {
		grph.g.AddVertex(issuerNode(issuer))
	}
	for _, clusterIssuer := range pki.ClusterIssuers {
		grph.g.AddVertex(clusterIssuerNode(clusterIssuer))
	}

	return grph
}

func (g *Graph) Raw() graph.Graph[string, Node] {
	return g.g
}
