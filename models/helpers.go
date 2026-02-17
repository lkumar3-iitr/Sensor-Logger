package models

import (
	"strconv"
)

// ─── shared formatting helpers (package-private) ────────────────────────

func itoa(v int) string      { return strconv.Itoa(v) }
func itoa64(v int64) string  { return strconv.FormatInt(v, 10) }
func utoa64(v uint64) string { return strconv.FormatUint(v, 10) }
func ftoa(v float64, prec int) string {
	return strconv.FormatFloat(v, 'f', prec, 64)
}

// CSVRowWriter is the interface every loggable model must satisfy.
type CSVRowWriter interface {
	CSVHeader() []string
	CSVRow() []string
}
