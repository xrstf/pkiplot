// SPDX-FileCopyrightText: 2025 Christoph Mewes
// SPDX-License-Identifier: MIT

package mermaid

import (
	"go.xrstf.de/pkiplot/pkg/render"
)

func init() {
	render.Register("mermaid", New())
}
