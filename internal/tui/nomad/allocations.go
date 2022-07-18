package nomad

import (
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

const keySeparator = " "

// allocationRowEntry is an item extracted from allocationResponseEntry
type allocationRowEntry struct {
	ID, TaskGroup, Name, TaskName, State string
	StartedAt, FinishedAt                time.Time
}

func FetchAllocations(client api.Client, jobID, jobNamespace string) tea.Cmd {
	return func() tea.Msg {
		allocs, _, err := client.Jobs().Allocations(jobID, true, &api.QueryOptions{Namespace: jobNamespace})
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		var allocationRowEntries []allocationRowEntry
		for _, alloc := range allocs {
			for taskName, task := range alloc.TaskStates {
				allocationRowEntries = append(allocationRowEntries, allocationRowEntry{
					ID:         alloc.ID,
					TaskGroup:  alloc.TaskGroup,
					Name:       alloc.Name,
					TaskName:   taskName,
					State:      task.State,
					StartedAt:  task.StartedAt.UTC(),
					FinishedAt: task.FinishedAt.UTC(),
				})
			}
		}

		sort.Slice(allocationRowEntries, func(x, y int) bool {
			firstTask := allocationRowEntries[x]
			secondTask := allocationRowEntries[y]
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

		tableHeader, allPageData := allocationsAsTable(allocationRowEntries)
		return PageLoadedMsg{Page: AllocationsPage, TableHeader: tableHeader, AllPageRows: allPageData}
	}
}

func allocationsAsTable(allocations []allocationRowEntry) ([]string, []page.Row) {
	var allocationResponseRows [][]string
	var keys []string
	for _, row := range allocations {
		allocationResponseRows = append(allocationResponseRows, []string{
			formatter.ShortAllocID(row.ID),
			row.TaskGroup,
			row.Name,
			row.TaskName,
			row.State,
			formatter.FormatTime(row.StartedAt),
			formatter.FormatTime(row.FinishedAt),
		})
		keys = append(keys, toAllocationsKey(row))
	}

	columns := []string{"Alloc ID", "Task Group", "Alloc Name", "Task Name", "State", "Started", "Finished"}
	table := formatter.GetRenderedTableAsString(columns, allocationResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}

func toAllocationsKey(allocationRowEntry allocationRowEntry) string {
	isRunning := "false"
	if allocationRowEntry.State == "running" {
		isRunning = "true"
	}
	return allocationRowEntry.ID + keySeparator + allocationRowEntry.TaskName + keySeparator + isRunning
}

type AllocationInfo struct {
	AllocID, TaskName string
	Running           bool
}

func AllocationInfoFromKey(key string) (AllocationInfo, error) {
	split := strings.Split(key, keySeparator)
	running, err := strconv.ParseBool(split[2])
	if err != nil {
		return AllocationInfo{}, err
	}
	return AllocationInfo{split[0], split[1], running}, nil
}
