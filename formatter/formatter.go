package formatter

import (
	"bytes"
	"encoding/json"
	"github.com/TylerBrock/colorjson"
	"github.com/olekukonko/tablewriter"
	"strings"
	"time"
)

func prettyPrint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

func colorJSON(b []byte) ([]byte, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(b, &obj)
	if err != nil {
		return []byte{}, err
	}

	f := colorjson.NewFormatter()
	f.Indent = 2

	colored, marshallErr := f.Marshal(obj)
	if marshallErr != nil {
		return []byte{}, marshallErr
	}
	return colored, nil
}

func PrettyJsonStringAsLines(logline string, color bool) []string {
	pretty, err := prettyPrint([]byte(logline))
	if err != nil {
		return []string{logline}
	}

	if color {
		pretty, err = colorJSON(pretty)
		if err != nil {
			return []string{logline}
		}
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

func (t *Table) IsEmpty() bool {
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
		table.SetTablePadding("    ")
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
	headerRows := []string{allRows[0]}
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
