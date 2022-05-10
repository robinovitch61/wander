package command

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"wander/formatter"
	"wander/message"
	"wander/nomad"
)

func simulateLoading() {
	for i := 0; i < 1e9; i++ {

	}
}

func FetchJobs(url, token string) tea.Cmd {
	return func() tea.Msg {
		// TODO LEO: error handling
		body, _ := nomad.GetJobs(url, token)
		//simulateLoading()
		//body := MockJobsResponse
		var jobResponse []nomad.JobResponseEntry
		if err := json.Unmarshal(body, &jobResponse); err != nil {
			// TODO LEO: error handling
			fmt.Println("Failed to decode job response")
		}

		return message.NomadJobsMsg(jobResponse)
	}
}

func FetchAllocations(url, token, jobId string) tea.Cmd {
	return func() tea.Msg {
		// TODO LEO: error handling
		body, _ := nomad.GetAllocations(url, token, jobId)
		//simulateLoading()
		//body := MockAllocationResponse
		var allocationResponse []nomad.AllocationResponseEntry
		if err := json.Unmarshal(body, &allocationResponse); err != nil {
			// TODO LEO: error handling
			fmt.Println("Failed to decode allocation response")
		}

		return message.NomadAllocationMsg{Table: formatter.AllocationResponseAsTable(allocationResponse)}
	}
}
