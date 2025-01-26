// SPDX-FileCopyrightText: 2025 Christoph Mewes
// SPDX-License-Identifier: MIT

package graphviz

import (
	"bytes"

	"github.com/dominikbraun/graph/draw"
	"github.com/spf13/pflag"

	"go.xrstf.de/pkiplot/pkg/pkigraph"
	"go.xrstf.de/pkiplot/pkg/render"
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
	return nil
}

func (r *renderer) RenderGraph(pki pkigraph.Graph) (string, error) {
	var buf bytes.Buffer
	err := draw.DOT(pki.Raw(), &buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
