package formatter

import (
	"github.com/olekukonko/tablewriter"
	"strconv"
	"strings"
	"time"
	"wander/nomad"
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

func getRenderedTableString(columnTitles []string, data [][]string) Table {
	table := createTableConfig(len(columnTitles))
	table.writer.SetHeader(columnTitles)
	table.writer.AppendBulk(data)
	table.writer.Render()
	allRows := strings.Split(table.string.String(), "\n")
	headerRows := []string{allRows[0]}
	contentRows := allRows[1 : len(allRows)-1] // last row is \n
	return Table{headerRows, contentRows}
}

func JobResponsesAsTable(jobResponse []nomad.JobResponseEntry) Table {
	var jobResponseRows [][]string
	for _, row := range jobResponse {
		jobResponseRows = append(jobResponseRows, []string{
			row.ID,
			row.Type,
			strconv.Itoa(row.Priority),
			row.Status,
			strconv.Itoa(row.SubmitTime),
		})
	}

	return getRenderedTableString(
		[]string{"ID", "Type", "Priority", "Status", "Submit Time"},
		jobResponseRows,
	)
}

// TODO LEO: move to utils
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02T15:04:05")
}

func AllocationsAsTable(allocations []nomad.AllocationRowEntry) Table {
	var allocationResponseRows [][]string
	for _, alloc := range allocations {
		allocationResponseRows = append(allocationResponseRows, []string{
			alloc.ID,
			alloc.Name,
			alloc.TaskName,
			alloc.State,
			formatTime(alloc.StartedAt),
			formatTime(alloc.FinishedAt),
		})
	}

	return getRenderedTableString(
		[]string{"Alloc ID", "Alloc Name", "Task Name", "State", "Started", "Finished"},
		allocationResponseRows,
	)
}

func LogsAsTable(logs []nomad.LogRow) Table {
	var logRows [][]string
	// ignore the first log line because it may be truncated due to offset
	// TODO LEO: check if there's actually a truncated line based on the offset size and log char length^
	//for _, row := range logs[1:] {
	for _, row := range logs {
		if stripped := strings.TrimSpace(string(row)); stripped != "" {
			logRows = append(logRows, []string{stripped})
		}
	}

	return getRenderedTableString(
		[]string{"Stdout"},
		logRows,
	)
}
