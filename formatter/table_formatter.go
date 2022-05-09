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

func createTableConfig() tableConfig {
	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)

	table.SetAutoFormatHeaders(false)
	table.SetHeaderLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetBorder(false)
	table.SetTablePadding("    ")
	table.SetNoWhiteSpace(true)

	return tableConfig{table, tableString}
}

func getRenderedTableString(header []string, data [][]string) Table {
	table := createTableConfig()
	table.writer.SetHeader(header)
	table.writer.AppendBulk(data)
	table.writer.Render()
	allRows := strings.Split(table.string.String(), "\n")
	headerRows := []string{allRows[0]}
	contentRows := allRows[1 : len(allRows)-1]
	return Table{headerRows, contentRows}
}

func JobResponseAsTable(jobResponse []nomad.JobResponseEntry) Table {
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

func AllocationResponseAsTable(allocationResponse []nomad.AllocationResponseEntry) Table {
	var allocationResponseRows [][]string
	for _, alloc := range allocationResponse {
		for taskName, taskData := range alloc.TaskStates {
			allocationResponseRows = append(allocationResponseRows, []string{
				alloc.ID,
				alloc.Name,
				taskName,
				taskData.State,
				formatTime(taskData.StartedAt),
				formatTime(taskData.FinishedAt),
			})
		}
	}

	return getRenderedTableString(
		[]string{"Alloc ID", "Alloc Name", "Task Name", "State", "Started", "Finished"},
		allocationResponseRows,
	)
}
