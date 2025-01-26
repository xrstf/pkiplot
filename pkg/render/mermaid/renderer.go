// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package mermaid

import (
	"fmt"

	"github.com/spf13/pflag"

	"go.xrstf.de/pkiplot/pkg/pkigraph"
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

func (r *renderer) ValidateFlags() error {
	// NOP
	return nil
}

func (r *renderer) RenderGraph(pki pkigraph.Graph) (string, error) {
	var buf types.StringBuilder
	buf.WriteString("graph TB\n")

	amap, err := pki.Raw().AdjacencyMap()
	if err != nil {
		return "", fmt.Errorf("invalid graph: %w", err)
	}

	// sort nodes alphabetically for stable output order
	nodeNames := sets.List(sets.KeySet(amap))

	// first print all the nodes
	for _, nodeHash := range nodeNames {
		srcNode, err := pki.Raw().Vertex(nodeHash)
		if err != nil {
			return "", fmt.Errorf("inconsistent graph: %w", err)
		}

		srcNodeID := nodeID(srcNode)
		buf.Printf("\t%s([%s]):::%s\n", srcNodeID, objectName(srcNode.Object()), nodeClass(srcNode))
	}

	buf.Printf("\n")

	// then print all the edges
	for nodeHash, edgeMap := range amap {
		srcNode, err := pki.Raw().Vertex(nodeHash)
		if err != nil {
			return "", fmt.Errorf("inconsistent graph: %w", err)
		}

		srcNodeID := nodeID(srcNode)

		for destNodeHash, edges := range edgeMap {
			destNode, err := pki.Raw().Vertex(destNodeHash)
			if err != nil {
				return "", fmt.Errorf("inconsistent graph: %w", err)
			}

			if false {
				fmt.Printf("node: %v\n", edges)
			}

			// To have the chart be readable from top to bottom, we reverse the edge direction here.
			buf.Printf("\t%s --> %s\n", nodeID(destNode), srcNodeID)
		}
	}

	buf.Printf("\n")
	buf.WriteString("\tclassDef clusterissuer color:#7F7\n")
	buf.WriteString("\tclassDef issuer color:#77F\n")
	buf.WriteString("\tclassDef ca color:#F77\n")
	buf.WriteString("\tclassDef cert color:orange\n")
	buf.WriteString("\tclassDef secret color:red")

	return buf.String(), nil
}
