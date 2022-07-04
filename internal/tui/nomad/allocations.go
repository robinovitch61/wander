package nomad

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
	"sort"
	"strconv"
	"strings"
	"time"
)

const keySeparator = " "

// allocationResponseEntry is returned from GET /v1/job/:job_id/allocations
// https://www.nomadproject.io/api-docs/jobs#list-job-allocations
type allocationResponseEntry struct {
	ID                 string `json:"ID"`
	EvalID             string `json:"EvalID"`
	Name               string `json:"Name"`
	NodeID             string `json:"NodeID"`
	PreviousAllocation string `json:"PreviousAllocation"`
	NextAllocation     string `json:"NextAllocation"`
	RescheduleTracker  struct {
		Events []struct {
			PrevAllocID    string `json:"PrevAllocID"`
			PrevNodeID     string `json:"PrevNodeID"`
			RescheduleTime int64  `json:"RescheduleTime"`
			Delay          int64  `json:"Delay"`
		} `json:"Events"`
	} `json:"RescheduleTracker"`
	JobID              string `json:"JobID"`
	TaskGroup          string `json:"TaskGroup"`
	DesiredStatus      string `json:"DesiredStatus"`
	DesiredDescription string `json:"DesiredDescription"`
	ClientStatus       string `json:"ClientStatus"`
	ClientDescription  string `json:"ClientDescription"`
	TaskStates         map[string]struct {
		State      string    `json:"State"`
		Failed     bool      `json:"Failed"`
		StartedAt  time.Time `json:"StartedAt"`
		FinishedAt time.Time `json:"FinishedAt"`
		Events     []struct {
			Type             string `json:"Type"`
			Time             int64  `json:"Time"`
			FailsTask        bool   `json:"FailsTask"`
			RestartReason    string `json:"RestartReason"`
			SetupError       string `json:"SetupError"`
			DriverError      string `json:"DriverError"`
			ExitCode         int    `json:"ExitCode"`
			Signal           int    `json:"Signal"`
			Message          string `json:"Message"`
			KillTimeout      int    `json:"KillTimeout"`
			KillError        string `json:"KillError"`
			KillReason       string `json:"KillReason"`
			StartDelay       int    `json:"StartDelay"`
			DownloadError    string `json:"DownloadError"`
			ValidationError  string `json:"ValidationError"`
			DiskLimit        int    `json:"DiskLimit"`
			FailedSibling    string `json:"FailedSibling"`
			VaultError       string `json:"VaultError"`
			TaskSignalReason string `json:"TaskSignalReason"`
			TaskSignal       string `json:"TaskSignal"`
			DriverMessage    string `json:"DriverMessage"`
		} `json:"Events"`
	} `json:"TaskStates"`
	CreateIndex int   `json:"CreateIndex"`
	ModifyIndex int   `json:"ModifyIndex"`
	CreateTime  int64 `json:"CreateTime"`
	ModifyTime  int64 `json:"ModifyTime"`
}

// allocationRowEntry is an item extracted from allocationResponseEntry
type allocationRowEntry struct {
	ID, TaskGroup, Name, TaskName, State string
	StartedAt, FinishedAt                time.Time
}

func FetchAllocations(url, token, jobID, jobNamespace string) tea.Cmd {
	return func() tea.Msg {
		params := map[string]string{
			"namespace": jobNamespace,
		}
		fullPath := fmt.Sprintf("%s%s%s%s", url, "/v1/job/", jobID, "/allocations")
		body, err := get(fullPath, token, params)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		var allocationResponse []allocationResponseEntry
		if err := json.Unmarshal(body, &allocationResponse); err != nil {
			return message.ErrMsg{Err: err}
		}

		var allocationRowEntries []allocationRowEntry
		for _, alloc := range allocationResponse {
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
		return PageLoadedMsg{Page: AllocationsPage, TableHeader: tableHeader, AllPageData: allPageData}
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
