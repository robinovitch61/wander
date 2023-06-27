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

func FetchJobs(client api.Client, columns []string) tea.Cmd {
	return func() tea.Msg {
		jobListOpts := &api.JobListOptions{
			Fields: &api.JobListFields{Meta: true},
		}
		jobResults, _, err := client.Jobs().ListOptions(jobListOpts, nil)
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

		tableHeader, allPageData := jobResponsesAsTable(jobResults, columns)
		return PageLoadedMsg{Page: JobsPage, TableHeader: tableHeader, AllPageRows: allPageData}
	}
}

func getCount(row *api.JobListStub) string {
	num, denom := 0, 0
	for _, v := range row.JobSummary.Summary {
		num += v.Running
		denom += v.Running + v.Starting + v.Queued
	}
	return strconv.Itoa(num) + "/" + strconv.Itoa(denom)
}

func getUptime(row *api.JobListStub) string {
	uptime := "-"
	if row.Status == "running" {
		uptime = formatter.FormatTimeNsSinceNow(row.SubmitTime)
	}
	return uptime
}

func getRowFromColumns(row *api.JobListStub, columns []string) []string {
	knownColMap := map[string]string{
		"ID":           row.ID,
		"Type":         row.Type,
		"Namespace":    row.Namespace,
		"Priority":     strconv.Itoa(row.Priority),
		"Status":       row.Status,
		"Count":        getCount(row),
		"Submitted":    formatter.FormatTimeNs(row.SubmitTime),
		"Since Submit": getUptime(row),
	}

	var rowEntries []string
	for _, col := range columns {
		// potential conflict here between "known job columns" and meta columns,
		// e.g. if meta key is "ID", it will be overwritten by the job ID
		if v, exists := knownColMap[col]; exists {
			rowEntries = append(rowEntries, v)
		} else if m, inMeta := row.Meta[col]; inMeta {
			rowEntries = append(rowEntries, m)
		} else {
			rowEntries = append(rowEntries, "")
		}
	}
	return rowEntries
}

func jobResponsesAsTable(jobResponse []*api.JobListStub, columns []string) ([]string, []page.Row) {
	var jobResponseRows [][]string
	var keys []string
	for _, row := range jobResponse {
		jobResponseRows = append(jobResponseRows, getRowFromColumns(row, columns))
		keys = append(keys, toIDNamespaceKey(row.ID, row.Namespace))
	}
	table := formatter.GetRenderedTableAsString(columns, jobResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}
