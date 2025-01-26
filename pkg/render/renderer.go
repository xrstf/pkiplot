// SPDX-FileCopyrightText: 2025 Christoph Mewes
// SPDX-License-Identifier: MIT

package render

import (
	"slices"

	"github.com/spf13/pflag"

	"go.xrstf.de/pkiplot/pkg/pkigraph"
)

type Renderer interface {
	RenderGraph(pki pkigraph.Graph) (string, error)
	AddFlags(fs *pflag.FlagSet)
	ValidateFlags() error
}

var renderers = map[string]Renderer{}

func Register(name string, r Renderer) {
	renderers[name] = r
}

func All() []string {
	names := make([]string, 0, len(renderers))
	for name := range renderers {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func Get(name string) (Renderer, bool) {
	r, ok := renderers[name]
	return r, ok
}
