package nomad

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
	"sort"
	"time"
)

type taskRowEntry struct {
	FullAllocationAsJSON                        string
	JobID, ID, TaskGroup, Name, TaskName, State string
	StartedAt, FinishedAt                       time.Time
}

func FetchAllTasks(client api.Client) tea.Cmd {
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

		tableHeader, allPageData := tasksAsTable(taskRowEntries)
		return PageLoadedMsg{Page: AllTasksPage, TableHeader: tableHeader, AllPageRows: allPageData}
	}
}

func tasksAsTable(taskRowEntries []taskRowEntry) ([]string, []page.Row) {
	var taskResponseRows [][]string
	var keys []string
	for _, row := range taskRowEntries {
		uptime := "-"
		if row.State == "running" {
			uptime = formatter.FormatTimeNsSinceNow(row.StartedAt.UnixNano())
		}
		taskResponseRows = append(taskResponseRows, []string{
			row.JobID,
			formatter.ShortAllocID(row.ID),
			row.TaskGroup,
			row.Name,
			row.TaskName,
			row.State,
			formatter.FormatTime(row.StartedAt),
			formatter.FormatTime(row.FinishedAt),
			uptime,
		})
		keys = append(keys, toTaskKey(row.State, row.FullAllocationAsJSON, row.TaskName))
	}

	columns := []string{"Job", "Alloc ID", "Task Group", "Alloc Name", "Task Name", "State", "Started", "Finished", "Uptime"}
	table := formatter.GetRenderedTableAsString(columns, taskResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}
