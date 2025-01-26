// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package types

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	corev1 "k8s.io/api/core/v1"
)

type PKI struct {
	Secrets        []corev1.Secret
	Certificates   []certmanagerv1.Certificate
	Issuers        []certmanagerv1.Issuer
	ClusterIssuers []certmanagerv1.ClusterIssuer
}
