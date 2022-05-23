package formatter

import (
	"github.com/olekukonko/tablewriter"
	"strings"
)

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
