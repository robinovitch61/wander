package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"math"
	"regexp"
	"strings"
	"time"
)

const (
	ansi  = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"
	osCmd = "[\u001B]]0.*[\a\u001B](?:\\\\)?"
)

var (
	ansiRe  = regexp.MustCompile(ansi)
	osCmdRe = regexp.MustCompile(osCmd)
)

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

type Table struct {
	HeaderRows, ContentRows []string
}

func (t *Table) isEmpty() bool {
	return len(t.HeaderRows) == 0 && len(t.ContentRows) == 0
}

type tableConfig struct {
	writer *tablewriter.Table
	string *strings.Builder
}

func createTableConfig(numCols int) tableConfig {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetAutoFormatHeaders(false)
	table.SetHeaderLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetBorder(false)
	if numCols > 1 {
		table.SetTablePadding(constants.TableSeparator)
	}
	table.SetNoWhiteSpace(true)
	table.SetAutoWrapText(false)

	return tableConfig{table, tableString}
}

func GetRenderedTableAsString(columns []string, data [][]string) Table {
	table := createTableConfig(len(columns))
	table.writer.SetHeader(columns)
	table.writer.AppendBulk(data)
	table.writer.Render()
	allRows := strings.Split(table.string.String(), "\n")
	headerRows := []string{allRows[0] + constants.TableSeparator}
	contentRows := allRows[1 : len(allRows)-1] // last row is \n
	return Table{headerRows, contentRows}
}

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

func pluralize(s string, q float64) string {
	if q > 1 {
		return s + "s"
	}
	return s
}

func FormatTimeNsSinceNow(t int64) string {
	tm := time.Unix(0, t).UTC()
	since := time.Now().Sub(tm)
	if secs := since.Seconds(); secs > 0 && secs < 60 {
		val := math.Floor(secs)
		out := fmt.Sprintf("%.0f second", val)
		return pluralize(out, val)
	}
	if mins := since.Minutes(); mins > 1 && mins < 60 {
		val := math.Floor(mins)
		out := fmt.Sprintf("%.0f minute", val)
		return pluralize(out, val)
	}
	if hrs := since.Hours(); hrs > 1 && hrs < 24 {
		val := math.Floor(hrs)
		out := fmt.Sprintf("%.0f hour", val)
		return pluralize(out, val)
	}
	if days := since.Hours() / 24; days > 1 && days < 365.25 {
		val := math.Floor(days)
		out := fmt.Sprintf("%.0f day", val)
		return pluralize(out, val)
	}
	if years := since.Hours() / 24 / 365.25; years > 1 {
		val := math.Floor(years)
		out := fmt.Sprintf("%.0f year", val)
		return pluralize(out, val)
	}
	return ""
}

func JsonEncodedTokenArray(s string) (string, error) {
	tokens := strings.Fields(s)
	tokensJson, err := json.Marshal(tokens)
	if err != nil {
		return "", err
	}
	return string(tokensJson), nil
}

func StripANSI(str string) string {
	return ansiRe.ReplaceAllString(str, "")
}

func StripOSCommandSequences(str string) string {
	// https://wezfurlong.org/wezterm/escape-sequences.html#operating-system-command-sequences
	// examples:
	// \x1b]0;me@123: /home/test\a
	// \x1b]0;title\x1b\\
	return osCmdRe.ReplaceAllString(str, "")
}

func CleanLogs(logs string) string {
	return StripANSI(strings.ReplaceAll(logs, "\t", "    "))
}
