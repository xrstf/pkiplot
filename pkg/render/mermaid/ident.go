// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package mermaid

import (
	"fmt"
	"strings"

	"go.xrstf.de/pkiplot/pkg/pkigraph"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func nodeID(node pkigraph.Node) string {
	obj := node.Object()
	ident := objectName(obj)

	if ns := obj.GetNamespace(); ns != "" {
		ident = ns + "/" + ident
	}

	ident = strings.ReplaceAll(ident, "-", "_")

	return fmt.Sprintf("%s_%s", node.ObjectKind(), ident)
}

func objectID(obj metav1.Object) string {
	ident := objectName(obj)

	if ns := obj.GetNamespace(); ns != "" {
		ident = ns + "/" + ident
	}

	return strings.ReplaceAll(ident, "-", "_")
}

func objectName(obj metav1.Object) string {
	base := obj.GetName()
	if base != "" {
		return base
	}

	base = obj.GetGenerateName()
	if base != "" {
		return base
	}

	panic("object has neither name nor generateName")
}

func nodeClass(n pkigraph.Node) string {
	class := n.ObjectKind()

	// highlight CAs in the graph
	if n.Certificate != nil && n.Certificate.Spec.IsCA {
		class = "ca"
	}

	if n.Synthetic {
		class += "_synthetic"
	}

	return class
}
