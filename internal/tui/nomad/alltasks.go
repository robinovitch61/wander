package nomad

import (
	"encoding/json"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
)

func FetchAllTasks(client api.Client, columns []string) tea.Cmd {
	return func() tea.Msg {
		allocations, _, err := client.Allocations().List(&api.QueryOptions{})
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		var taskRowEntries []taskRowEntry
		for _, alloc := range allocations {
			allocAsJSON, err := json.Marshal(alloc)
			if err != nil {
				return message.ErrMsg{Err: err}
			}

			for taskName, task := range alloc.TaskStates {
				taskRowEntries = append(taskRowEntries, taskRowEntry{
					FullAllocationAsJSON: string(allocAsJSON),
					NodeID:               alloc.NodeID,
					JobID:                alloc.JobID,
					ID:                   alloc.ID,
					TaskGroup:            alloc.TaskGroup,
					Name:                 alloc.Name,
					TaskName:             taskName,
					State:                task.State,
					StartedAt:            task.StartedAt.UTC(),
					FinishedAt:           task.FinishedAt.UTC(),
				})
			}
		}

		sort.Slice(taskRowEntries, func(x, y int) bool {
			firstTask := taskRowEntries[x]
			secondTask := taskRowEntries[y]
			if firstTask.JobID == secondTask.JobID {
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
			}
			return firstTask.JobID < secondTask.JobID
		})

		tableHeader, allPageData := tasksAsTable(taskRowEntries, columns)
		return PageLoadedMsg{Page: AllTasksPage, TableHeader: tableHeader, AllPageRows: allPageData}
	}
}

func getTaskRowFromColumns(row taskRowEntry, columns []string) []string {
	knownColMap := map[string]string{
		"Node ID":    formatter.ShortID(row.NodeID),
		"Job":        row.JobID,
		"Alloc ID":   formatter.ShortID(row.ID),
		"Task Group": row.TaskGroup,
		"Alloc Name": row.Name,
		"Task Name":  row.TaskName,
		"State":      row.State,
		"Started":    formatter.FormatTime(row.StartedAt),
		"Finished":   formatter.FormatTime(row.FinishedAt),
		"Uptime":     getUptime(row.State, row.StartedAt.UnixNano()),
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

func tasksAsTable(taskRowEntries []taskRowEntry, columns []string) ([]string, []page.Row) {
	var taskResponseRows [][]string
	var keys []string
	for _, row := range taskRowEntries {
		taskResponseRows = append(taskResponseRows, getTaskRowFromColumns(row, columns))
		keys = append(keys, toTaskKey(row.State, row.FullAllocationAsJSON, row.TaskName))
	}

	table := formatter.GetRenderedTableAsString(columns, taskResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}
