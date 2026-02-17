package utils

import (
	"fmt"
	"time"
)

// NowNano returns the current time as nanoseconds since Unix epoch.
// Uses monotonic-aware time internally but returns wall-clock nanos
// so that timestamps are portable across processes.
func NowNano() int64 {
	return time.Now().UnixNano()
}

// NanoToTime converts a nanosecond Unix timestamp back to time.Time.
func NanoToTime(ns int64) time.Time {
	return time.Unix(0, ns)
}

// FormatTimestamp converts ns-epoch to a human-friendly string.
func FormatTimestamp(ns int64) string {
	return NanoToTime(ns).Format("2006-01-02_15-04-05.000000000")
}

// SessionName returns a unique session directory name:
//
//	<prefix>_YYYYMMDD_HHMMSS
func SessionName(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, time.Now().Format("20060102_150405"))
}
