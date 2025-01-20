// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package types

import (
	"fmt"
	"strings"
)

type StringBuilder struct {
	strings.Builder
}

func (sb *StringBuilder) Printf(format string, args ...any) {
	sb.WriteString(fmt.Sprintf(format, args...))
}
