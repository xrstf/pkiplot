// SPDX-FileCopyrightText: 2025 Christoph Mewes
// SPDX-License-Identifier: MIT

package pkigraph

import (
	"fmt"
	"reflect"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Node struct {
	Secret        *corev1.Secret
	Certificate   *certmanagerv1.Certificate
	Issuer        *certmanagerv1.Issuer
	ClusterIssuer *certmanagerv1.ClusterIssuer

	// Synthetic signal whether the object was actually found in the provided
	// YAML manifests or if it was created based on reference names (e.g. a
	// Certificate creating a Secret, but that Secret was not loaded in).
	Synthetic bool
}

func (n Node) Object() metav1.Object {
	switch {
	case n.Secret != nil:
		return n.Secret
	case n.Certificate != nil:
		return n.Certificate
	case n.Issuer != nil:
		return n.Issuer
	case n.ClusterIssuer != nil:
		return n.ClusterIssuer
	default:
		panic("Invalid node: None of the four possible fields are set.")
	}
}

func (n Node) Hash() string {
	return objectHash(n.Object())
}

func nodeHash(n Node) string {
	return objectHash(n.Object())
}

func objectHash(obj metav1.Object) string {
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	kind := strings.ToLower(t.Name())

	if ns := obj.GetNamespace(); ns != "" {
		return fmt.Sprintf("%s:%s:%s", kind, ns, obj.GetName())
	} else {
		return fmt.Sprintf("%s:%s", kind, obj.GetName())
	}
}

func secretNode(secret corev1.Secret) Node {
	return Node{Secret: &secret}
}

func secretHash(secret corev1.Secret) string {
	return objectHash(&secret)
}

func certificateNode(cert certmanagerv1.Certificate) Node {
	return Node{Certificate: &cert}
}

func certificateHash(cert certmanagerv1.Certificate) string {
	return objectHash(&cert)
}

func issuerNode(issuer certmanagerv1.Issuer) Node {
	return Node{Issuer: &issuer}
}

func issuerHash(issuer certmanagerv1.Issuer) string {
	return objectHash(&issuer)
}

func clusterIssuerNode(clusterIssuer certmanagerv1.ClusterIssuer) Node {
	return Node{ClusterIssuer: &clusterIssuer}
}

func clusterIssuerHash(clusterIssuer certmanagerv1.ClusterIssuer) string {
	return objectHash(&clusterIssuer)
}
