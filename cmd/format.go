package cmd

import (
	"github.com/thaitrn/binman/internal/human"
	"github.com/thaitrn/binman/internal/scan"
)

// humanBytes formats a byte count in binary units (KiB, MiB, ...).
func humanBytes(b int64) string {
	return human.Bytes(b)
}

// totalSize sums the Size of all matches.
func totalSize(ms []scan.Match) int64 {
	var s int64
	for _, m := range ms {
		s += m.Size
	}
	return s
}
