package nomad

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/message"
	"time"
)

type taskRowEntry struct {
	FullTaskAsJSON                                string
	ID, AllocID, TaskGroup, Name, TaskName, State string
	StartedAt, FinishedAt                         time.Time
}

func FetchTasks(client api.Client, jobNamespace string) tea.Cmd {
	return func() tea.Msg {
		allocs, _, err := client.Allocations().List(&api.QueryOptions{Namespace: jobNamespace})
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		var taskRowEntries []taskRowEntry
		for _, alloc := range allocs {
			allocAsJSON, err := json.Marshal(alloc)
			if err != nil {
				return message.ErrMsg{Err: err}
			}
			dev.Debug(fmt.Sprintf("%+v", alloc.TaskStates))

			for taskName, task := range alloc.TaskStates {
				taskRowEntries = append(taskRowEntries, taskRowEntry{
					FullTaskAsJSON: string(allocAsJSON),
					ID:             alloc.ID,
					TaskGroup:      alloc.TaskGroup,
					Name:           alloc.Name,
					TaskName:       taskName,
					State:          task.State,
					StartedAt:      task.StartedAt.UTC(),
					FinishedAt:     task.FinishedAt.UTC(),
				})
			}
		}

		//tableHeader, allPageData := allocationsAsTable([])
		return PageLoadedMsg{Page: AllocationsPage, TableHeader: []string{}, AllPageRows: []page.Row{}}
	}
}

//func tasksAsTable(tasks []taskRowEntry) ([]string, []page.Row) {
//	var allocationResponseRows [][]string
//	var keys []string
//	for _, row := range tasks {
//		uptime := "-"
//		if row.State == "running" {
//			uptime = formatter.FormatTimeNsSinceNow(row.StartedAt.UnixNano())
//		}
//		allocationResponseRows = append(allocationResponseRows, []string{
//			formatter.ShortAllocID(row.ID),
//			row.TaskGroup,
//			row.Name,
//			row.TaskName,
//			row.State,
//			formatter.FormatTime(row.StartedAt),
//			formatter.FormatTime(row.FinishedAt),
//			uptime,
//		})
//		keys = append(keys, toAllocationsKey(row))
//	}
//
//	columns := []string{"Alloc ID", "Task Group", "Alloc Name", "Task Name", "State", "Started", "Finished", "Uptime"}
//	table := formatter.GetRenderedTableAsString(columns, allocationResponseRows)
//
//	var rows []page.Row
//	for idx, row := range table.ContentRows {
//		rows = append(rows, page.Row{Key: keys[idx], Row: row})
//	}
//
//	return table.HeaderRows, rows
//}
