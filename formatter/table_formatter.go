package formatter

import (
	"github.com/olekukonko/tablewriter"
	"strconv"
	"strings"
	"wander/nomad"
)

func JobResponsesAsTable(jobResponse []nomad.JobResponseEntry) []string {
	jobsTableString := &strings.Builder{}
	table := tablewriter.NewWriter(jobsTableString)
	table.SetHeader([]string{"ID", "Type", "Priority", "Status", "Submit Time"})
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
	table.AppendBulk(jobResponseRows)
	table.SetHeaderLine(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	table.Render()
	allRows := strings.Split(jobsTableString.String(), "\n")
	//allRows[0] = strings.Title(strings.ToLower(allRows[0]))
	return allRows[:len(allRows)-1] // last entry is blank line
}
