package allocations

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"sort"
	"strings"
	"time"
	"wander/dev"
	"wander/formatter"
	"wander/message"
	"wander/nomad"
)

type allocationsData struct {
	allData, filteredData []allocationRowEntry
}

type nomadAllocationMsg []allocationRowEntry

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
	ID, Name, TaskName, State string
	StartedAt, FinishedAt     time.Time
}

func (e allocationRowEntry) MatchesFilter(filter string) bool {
	return strings.Contains(e.TaskName, filter)
}

func FetchAllocations(url, token, jobID string) tea.Cmd {
	return func() tea.Msg {
		dev.Debug(fmt.Sprintf("jobID %s", jobID))
		fullPath := fmt.Sprintf("%s%s%s%s", url, "/v1/job/", jobID, "/allocations")
		body, err := nomad.Get(fullPath, token, nil)
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
					Name:       alloc.Name,
					TaskName:   taskName,
					State:      task.State,
					StartedAt:  task.StartedAt,
					FinishedAt: task.FinishedAt,
				})
			}
		}

		sort.Slice(allocationRowEntries, func(x, y int) bool {
			firstTask := allocationRowEntries[x]
			secondTask := allocationRowEntries[y]
			if firstTask.TaskName == secondTask.TaskName {
				if firstTask.Name == secondTask.Name {
					return firstTask.State > secondTask.State
				}
				return firstTask.Name < secondTask.Name
			}
			return firstTask.TaskName < secondTask.TaskName
		})

		return nomadAllocationMsg(allocationRowEntries)
	}
}

func allocationsAsTable(allocations []allocationRowEntry) formatter.Table {
	var allocationResponseRows [][]string
	for _, alloc := range allocations {
		allocationResponseRows = append(allocationResponseRows, []string{
			alloc.ID,
			alloc.Name,
			alloc.TaskName,
			alloc.State,
			formatter.FormatTime(alloc.StartedAt),
			formatter.FormatTime(alloc.FinishedAt),
		})
	}

	return formatter.GetRenderedTableAsString(
		[]string{"Alloc ID", "Alloc Name", "Task Name", "State", "Started", "Finished"},
		allocationResponseRows,
	)
}
