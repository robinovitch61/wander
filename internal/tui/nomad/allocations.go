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

type allocationRowEntry struct {
	FullAllocationAsJSON                 string
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
			allocAsJSON, err := json.Marshal(alloc)
			if err != nil {
				return message.ErrMsg{Err: err}
			}

			for taskName, task := range alloc.TaskStates {
				allocationRowEntries = append(allocationRowEntries, allocationRowEntry{
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
		uptime := "-"
		if row.State == "running" {
			uptime = formatter.FormatTimeNsSinceNow(row.StartedAt.UnixNano())
		}
		allocationResponseRows = append(allocationResponseRows, []string{
			formatter.ShortAllocID(row.ID),
			row.TaskGroup,
			row.Name,
			row.TaskName,
			row.State,
			formatter.FormatTime(row.StartedAt),
			formatter.FormatTime(row.FinishedAt),
			uptime,
		})
		keys = append(keys, toAllocationsKey(row))
	}

	columns := []string{"Alloc ID", "Task Group", "Alloc Name", "Task Name", "State", "Started", "Finished", "Uptime"}
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
	return allocationRowEntry.FullAllocationAsJSON + keySeparator + allocationRowEntry.TaskName + keySeparator + isRunning
}

type AllocationInfo struct {
	Alloc    api.Allocation
	TaskName string
	Running  bool
}

func AllocationInfoFromKey(key string) (AllocationInfo, error) {
	split := strings.Split(key, keySeparator)
	running, err := strconv.ParseBool(split[2])
	if err != nil {
		return AllocationInfo{}, err
	}
	var alloc api.Allocation
	err = json.Unmarshal([]byte(split[0]), &alloc)
	if err != nil {
		return AllocationInfo{}, err
	}
	return AllocationInfo{Alloc: alloc, TaskName: split[1], Running: running}, nil
}
