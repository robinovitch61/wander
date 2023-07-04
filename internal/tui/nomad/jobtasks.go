package nomad

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
	"sort"
	"strconv"
	"strings"
	"time"
)

const keySeparator = "|【=◈︿◈=】|"

// taskRowEntry is an item extracted from allocationResponseEntry
type taskRowEntry struct {
	FullAllocationAsJSON                 string
	ID, TaskGroup, Name, TaskName, State string
	StartedAt, FinishedAt                time.Time
}

func FetchTasksForJob(client api.Client, jobID, jobNamespace string) tea.Cmd {
	return func() tea.Msg {
		allocationsForJob, _, err := client.Jobs().Allocations(jobID, true, &api.QueryOptions{Namespace: jobNamespace})
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		var taskRowEntries []taskRowEntry
		for _, alloc := range allocationsForJob {
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
		return PageLoadedMsg{Page: JobTasksPage, TableHeader: tableHeader, AllPageRows: allPageData}
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
			formatter.ShortAllocID(row.ID),
			row.TaskGroup,
			row.Name,
			row.TaskName,
			row.State,
			formatter.FormatTime(row.StartedAt),
			formatter.FormatTime(row.FinishedAt),
			uptime,
		})
		keys = append(keys, toTaskKey(row))
	}

	columns := []string{"Alloc ID", "Task Group", "Alloc Name", "Task Name", "State", "Started", "Finished", "Uptime"}
	table := formatter.GetRenderedTableAsString(columns, taskResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}

func toTaskKey(taskRowEntry taskRowEntry) string {
	isRunning := "false"
	if taskRowEntry.State == "running" {
		isRunning = "true"
	}
	return taskRowEntry.FullAllocationAsJSON + keySeparator + taskRowEntry.TaskName + keySeparator + isRunning
}

type TaskInfo struct {
	Alloc    api.Allocation
	TaskName string
	Running  bool
}

func TaskInfoFromKey(key string) (TaskInfo, error) {
	split := strings.Split(key, keySeparator)
	running, err := strconv.ParseBool(split[2])
	if err != nil {
		return TaskInfo{}, err
	}
	var alloc api.Allocation
	err = json.Unmarshal([]byte(split[0]), &alloc)
	if err != nil {
		return TaskInfo{}, err
	}
	return TaskInfo{Alloc: alloc, TaskName: split[1], Running: running}, nil
}
