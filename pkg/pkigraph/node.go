// SPDX-FileCopyrightText: 2025 Christoph Mewes
// SPDX-License-Identifier: MIT

package pkigraph

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	corev1 "k8s.io/api/core/v1"
)

type Node struct {
	Name string
}

func nodeHash(c Node) string {
	return c.Name
}

func secretNode(secret corev1.Secret) Node {
	return Node{
		Name: secret.Name,
	}
}

func certificateNode(cert certmanagerv1.Certificate) Node {
	return Node{
		Name: cert.Name,
	}
}

func issuerNode(issuer certmanagerv1.Issuer) Node {
	return Node{
		Name: issuer.Name,
	}
}

func clusterIssuerNode(clusterIssuer certmanagerv1.ClusterIssuer) Node {
	return Node{
		Name: clusterIssuer.Name,
	}
}
