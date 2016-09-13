package config

import (
	"fmt"
	"io"
	"strings"
)

func dumpBlock(out io.Writer, block Block, level int) {
	indent := strings.Repeat("  ", level)
	for _, property := range block {
		props := "\"" + strings.Join(property.Values, "\" \"") + "\""
		fmt.Fprintf(out, "%s%s: %s\r\n", indent, property.Key, props)
		if property.Block != nil {
			dumpBlock(out, property.Block, level+1)
		}
	}
}

func (cfg Config) Dump(out io.Writer) {
	dumpBlock(out, cfg.Data, 0)
}
