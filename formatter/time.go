package formatter

import "time"

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02T15:04:05")
}

func formatTimeNs(t int64) string {
	tm := time.Unix(0, t)
	return formatTime(tm)
}
