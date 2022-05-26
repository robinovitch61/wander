package formatter

import (
	"bytes"
	"encoding/json"
	"time"
)

func ShortAllocID(allocID string) string {
	firstN := 8
	if len(allocID) < firstN {
		return ""
	}
	return allocID[:firstN]
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	local, err := time.LoadLocation("Local")
	if err != nil {
		return "-"
	}
	tLocal := t.In(local)
	return tLocal.Format("2006-01-02T15:04:05")
}

func FormatTimeNs(t int64) string {
	tm := time.Unix(0, t).UTC()
	return FormatTime(tm)
}

func prettyPrint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func PrettyJsonStringAsLines(logline string) []string {
	pretty, err := prettyPrint([]byte(logline))
	if err != nil {
		return []string{logline}
	}

	var splitLines []string
	for _, row := range bytes.Split(pretty, []byte("\n")) {
		splitLines = append(splitLines, string(row))
	}

	return splitLines
}
