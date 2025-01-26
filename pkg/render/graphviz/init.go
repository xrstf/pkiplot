// SPDX-FileCopyrightText: 2025 Christoph Mewes
// SPDX-License-Identifier: MIT

package graphviz

import (
	"go.xrstf.de/pkiplot/pkg/render"
)

func init() {
	render.Register("graphviz", New())
}
