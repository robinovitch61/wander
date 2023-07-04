package nomad

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/message"
	"sort"
)

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
