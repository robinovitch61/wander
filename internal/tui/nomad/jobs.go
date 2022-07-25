package nomad

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
	"sort"
	"strconv"
	"strings"
)

func FetchJobs(client api.Client) tea.Cmd {
	return func() tea.Msg {
		jobResults, _, err := client.Jobs().List(nil)
		if err != nil {
			if strings.Contains(err.Error(), "UUID must be 36 characters") {
				return message.ErrMsg{Err: errors.New("token must be 36 characters")}
			} else if strings.Contains(err.Error(), "ACL token not found") {
				return message.ErrMsg{Err: errors.New("token not authorized to list jobs")}
			}
			return message.ErrMsg{Err: err}
		}

		sort.Slice(jobResults, func(x, y int) bool {
			firstJob := jobResults[x]
			secondJob := jobResults[y]
			if firstJob.Name == secondJob.Name {
				return firstJob.Namespace < secondJob.Namespace
			}
			return jobResults[x].Name < jobResults[y].Name
		})

		tableHeader, allPageData := jobResponsesAsTable(jobResults)
		return PageLoadedMsg{Page: JobsPage, TableHeader: tableHeader, AllPageRows: allPageData}
	}
}

func jobResponsesAsTable(jobResponse []*api.JobListStub) ([]string, []page.Row) {
	var jobResponseRows [][]string
	var keys []string
	for _, row := range jobResponse {
		uptime := "-"
		if row.Status == "running" {
			uptime = formatter.FormatTimeNsSinceNow(row.SubmitTime)
		}
		num, denom := 0, 0
		for _, v := range row.JobSummary.Summary {
			num += v.Running
			denom += v.Running + v.Starting + v.Queued
		}
		count := strconv.Itoa(num) + "/" + strconv.Itoa(denom)

		jobResponseRows = append(jobResponseRows, []string{
			row.ID,
			row.Type,
			row.Namespace,
			strconv.Itoa(row.Priority),
			row.Status,
			count,
			formatter.FormatTimeNs(row.SubmitTime),
			uptime,
		})
		keys = append(keys, toJobsKey(row))
	}

	columns := []string{"ID", "Type", "Namespace", "Priority", "Status", "Count", "Submitted", "Since Submit"}
	table := formatter.GetRenderedTableAsString(columns, jobResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}

func toJobsKey(jobResponseEntry *api.JobListStub) string {
	return jobResponseEntry.ID + " " + jobResponseEntry.Namespace
}

func JobIDAndNamespaceFromKey(key string) (string, string) {
	split := strings.Split(key, " ")
	return split[0], split[1]

}
