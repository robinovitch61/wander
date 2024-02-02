package nomad

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
)

// FetchAllocs fetches allocs for a given job
func FetchAllocs(client api.Client, jobID, namespace string, columns []string) tea.Cmd {
	return func() tea.Msg {

		queryOpts := &api.QueryOptions{Namespace: namespace}

		allocResults, _, err := client.Jobs().Allocations(jobID, true, queryOpts)

		if len(allocResults) == 0 {
			return message.ErrMsg{Err: errors.New("no allocs found")}
		}

		if err != nil {
			if strings.Contains(err.Error(), "UUID must be 36 characters") {
				return message.ErrMsg{Err: errors.New("token must be 36 characters")}
			} else if strings.Contains(err.Error(), "ACL token not found") {
				return message.ErrMsg{Err: errors.New("token not authorized to list allocs")}
			}
			return message.ErrMsg{Err: err}
		}

		sort.Slice(allocResults, func(x, y int) bool {
			firstAlloc := allocResults[x]
			secondAlloc := allocResults[y]
			if firstAlloc.Name == secondAlloc.Name {
				return firstAlloc.ID < secondAlloc.ID
			}
			return firstAlloc.Name < secondAlloc.Name
		})

		tableHeader, allPageData := allocResponsesAsTable(allocResults, columns)
		return PageLoadedMsg{
			Page: AllocsPage, TableHeader: tableHeader, AllPageRows: allPageData}
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func getAllocRowFromColumns(row *api.AllocationListStub, columns []string) []string {
	knownColMap := map[string]string{
		"Alloc":      formatter.ShortID(row.ID),
		"Task Group": row.TaskGroup,
		"Version":    strconv.FormatUint(row.JobVersion, 10),
		"Created":    getUptime("running", row.CreateTime),
		"Modified":   formatter.FormatTimeNs(row.ModifyTime),
		"Node Name":  row.NodeName,
		"Node ID":    formatter.ShortID(row.NodeID),
		"Namespace":  row.Namespace,
		"Status":     row.ClientStatus,
	}

	var rowEntries []string
	for _, col := range columns {
		if v, exists := knownColMap[col]; exists {
			rowEntries = append(rowEntries, v)
		} else {
			rowEntries = append(rowEntries, "-")
		}
	}
	return rowEntries
}

func toAllocsKey(allocResponseEntry *api.AllocationListStub) string {
	return allocResponseEntry.ID + " " + allocResponseEntry.Namespace
}

func allocResponsesAsTable(allocResponse []*api.AllocationListStub, columns []string) ([]string, []page.Row) {
	var allocResponseRows [][]string
	var keys []string
	for _, row := range allocResponse {
		allocResponseRows = append(allocResponseRows, getAllocRowFromColumns(row, columns))
		keys = append(keys, toAllocsKey(row))
	}
	table := formatter.GetRenderedTableAsString(columns, allocResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}

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

func getJobRowFromColumns(row *api.JobListStub, columns []string) []string {
	knownColMap := map[string]string{
		"Job":          row.ID,
		"Type":         row.Type,
		"Namespace":    row.Namespace,
		"Priority":     strconv.Itoa(row.Priority),
		"Status":       row.Status,
		"Count":        getCount(row),
		"Submitted":    formatter.FormatTimeNs(row.SubmitTime),
		"Since Submit": getUptime(row.Status, row.SubmitTime),
	}

	var rowEntries []string
	for _, col := range columns {
		// potential conflict here between "known job columns" and meta columns,
		// e.g. if meta key is "Type", it will be overwritten by the job type
		if v, exists := knownColMap[col]; exists {
			rowEntries = append(rowEntries, v)
		} else if m, inMeta := row.Meta[col]; inMeta {
			rowEntries = append(rowEntries, m)
		} else {
			rowEntries = append(rowEntries, "-")
		}
	}
	return rowEntries
}

func jobResponsesAsTable(jobResponse []*api.JobListStub, columns []string) ([]string, []page.Row) {
	var jobResponseRows [][]string
	var keys []string
	for _, row := range jobResponse {
		jobResponseRows = append(jobResponseRows, getJobRowFromColumns(row, columns))
		keys = append(keys, toJobsKey(row))
	}
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
