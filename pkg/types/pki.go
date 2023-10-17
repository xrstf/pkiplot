// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package types

import (
	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
)

type PKI struct {
	Certificates   []certmanagerv1.Certificate
	Issuers        []certmanagerv1.Issuer
	ClusterIssuers []certmanagerv1.ClusterIssuer
}
