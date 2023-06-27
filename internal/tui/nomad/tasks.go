package nomad

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
	"time"
)

type taskRowEntry struct {
	JobID, AllocID, TaskGroup, AllocName, TaskName, State string
	StartedAt, FinishedAt                                 time.Time
}

func FetchTasks(client api.Client, jobNamespace string) tea.Cmd {
	return func() tea.Msg {
		allocs, _, err := client.Allocations().List(&api.QueryOptions{Namespace: jobNamespace})
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		var taskRowEntries []taskRowEntry
		for _, alloc := range allocs {
			for taskName, task := range alloc.TaskStates {
				taskRowEntries = append(taskRowEntries, taskRowEntry{
					JobID:      alloc.JobID,
					AllocID:    alloc.ID,
					TaskGroup:  alloc.TaskGroup,
					AllocName:  alloc.Name,
					TaskName:   taskName,
					State:      task.State,
					StartedAt:  task.StartedAt.UTC(),
					FinishedAt: task.FinishedAt.UTC(),
				})
			}
		}

		tableHeader, allPageData := tasksAsTable(taskRowEntries)
		return PageLoadedMsg{Page: AllocationsPage, TableHeader: tableHeader, AllPageRows: allPageData}
	}
}

func tasksAsTable(tasks []taskRowEntry) ([]string, []page.Row) {
	var taskResponseRows [][]string
	var keys []string
	for _, row := range tasks {
		uptime := "-"
		if row.State == "running" {
			uptime = formatter.FormatTimeNsSinceNow(row.StartedAt.UnixNano())
		}
		taskResponseRows = append(taskResponseRows, []string{
			row.JobID,
			formatter.ShortAllocID(row.AllocID),
			row.TaskGroup,
			row.AllocName,
			row.TaskName,
			row.State,
			formatter.FormatTime(row.StartedAt),
			formatter.FormatTime(row.FinishedAt),
			uptime,
		})
		keys = append(keys, row.TaskGroup) // TODO: fix
	}

	columns := []string{"Job ID", "Alloc ID", "Task Group", "Alloc Name", "Task Name", "State", "Started", "Finished", "Uptime"}
	table := formatter.GetRenderedTableAsString(columns, taskResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}
