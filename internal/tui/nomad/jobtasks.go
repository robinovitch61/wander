package nomad

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
	"sort"
)

func FetchTasksForJob(client api.Client, jobID, jobNamespace string, columns []string) tea.Cmd {
	return func() tea.Msg {
		allocationsForJob, _, err := client.Jobs().Allocations(
			jobID,
			true,
			&api.QueryOptions{
				Namespace: jobNamespace,
				Resources: true,
			},
		)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		var jobTaskRowEntries []taskRowEntry
		for _, alloc := range allocationsForJob {
			allocAsJSON, err := json.Marshal(alloc)
			if err != nil {
				return message.ErrMsg{Err: err}
			}

			resources := getAllocatedResources(alloc)
			dev.Debug(fmt.Sprintf("Allocated resources for alloc %s: %v", alloc.ID, resources))
			for taskName, task := range alloc.TaskStates {
				taskResources := getTaskResources(resources, taskName)
				jobTaskRowEntries = append(jobTaskRowEntries, taskRowEntry{
					NodeID:               alloc.NodeID,
					JobID:                alloc.JobID,
					FullAllocationAsJSON: string(allocAsJSON),
					ID:                   alloc.ID,
					TaskGroup:            alloc.TaskGroup,
					Name:                 alloc.Name,
					TaskName:             taskName,
					State:                task.State,
					StartedAt:            task.StartedAt.UTC(),
					FinishedAt:           task.FinishedAt.UTC(),
					CpuShares:            taskResources.CpuShares,
					Memory:               taskResources.MemoryMB,
					MaxMemory:            taskResources.MaxMemoryMB,
				})
			}
		}

		sort.Slice(jobTaskRowEntries, func(x, y int) bool {
			firstTask := jobTaskRowEntries[x]
			secondTask := jobTaskRowEntries[y]
			if firstTask.TaskName == secondTask.TaskName {
				if firstTask.Name == secondTask.Name {
					if firstTask.State == secondTask.State {
						if firstTask.StartedAt.Equal(secondTask.StartedAt) {
							return firstTask.ID > secondTask.ID
						}
						return firstTask.StartedAt.After(secondTask.StartedAt)
					}
					return firstTask.State > secondTask.State
				}
				return firstTask.Name < secondTask.Name
			}
			return firstTask.TaskName < secondTask.TaskName
		})

		tableHeader, allPageData := jobTasksAsTable(jobTaskRowEntries, columns)
		return PageLoadedMsg{Page: JobTasksPage, TableHeader: tableHeader, AllPageRows: allPageData}
	}
}

func getJobTaskRowFromColumns(row taskRowEntry, columns []string) []string {
	knownColMap := map[string]string{
		"Node ID":    formatter.ShortAllocID(row.NodeID),
		"Job":        row.JobID,
		"Alloc ID":   formatter.ShortAllocID(row.ID),
		"Task Group": row.TaskGroup,
		"Alloc Name": row.Name,
		"Task Name":  row.TaskName,
		"State":      row.State,
		"Started":    formatter.FormatTime(row.StartedAt),
		"Finished":   formatter.FormatTime(row.FinishedAt),
		"Uptime":     getUptime(row.State, row.StartedAt.UnixNano()),
		"CPU":        row.CpuShares,
		"Memory":     row.Memory,
		"Max Memory": row.MaxMemory,
	}

	var rowEntries []string
	for _, col := range columns {
		if v, exists := knownColMap[col]; exists {
			rowEntries = append(rowEntries, v)
		} else {
			rowEntries = append(rowEntries, "")
		}
	}
	return rowEntries
}

func jobTasksAsTable(jobTaskRowEntries []taskRowEntry, columns []string) ([]string, []page.Row) {
	var taskResponseRows [][]string
	var keys []string
	for _, row := range jobTaskRowEntries {
		taskResponseRows = append(taskResponseRows, getJobTaskRowFromColumns(row, columns))
		keys = append(keys, toTaskKey(row.State, row.FullAllocationAsJSON, row.TaskName))
	}

	table := formatter.GetRenderedTableAsString(columns, taskResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}
