// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package mermaid

import (
	"fmt"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ResourceIdentifier(res metav1.Object) string {
	ident := ResourceName(res)

	if ns := res.GetNamespace(); ns != "" {
		ident = ns + "/" + ident
	}

	return strings.ReplaceAll(ident, "-", "_")
}

func IssuerResourceIdentifier(cert certmanagerv1.Certificate) string {
	issuerKind := ""
	issuerIdent := ""

	ref := cert.Spec.IssuerRef
	switch ref.Kind {
	case "":
		fallthrough
	case "Issuer":
		issuerKind = "i"
		issuerIdent = cert.GetNamespace() + "/" + ref.Name

	case "ClusterIssuer":
		issuerKind = "ci"
		issuerIdent = ref.Name
	}

	return fmt.Sprintf("%s_%s", issuerKind, strings.ReplaceAll(issuerIdent, "-", "_"))
}

func ResourceName(res metav1.Object) string {
	base := res.GetName()
	if base != "" {
		return base
	}

	base = res.GetGenerateName()
	if base != "" {
		return base
	}

	panic("resource has neither name nor generateName")
}
