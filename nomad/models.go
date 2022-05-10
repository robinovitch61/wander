package nomad

import (
	"strings"
	"time"
)

type JobResponseEntry struct {
	ID         string
	Type       string
	Priority   int
	Status     string
	SubmitTime int
}

func (e JobResponseEntry) MatchesFilter(filter string) bool {
	return strings.Contains(e.ID, filter)
}

type AllocationResponseEntry struct {
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
