package command

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"sort"
	"strings"
	"wander/message"
	"wander/nomad"
)

func simulateLoading() {
	//for i := 0; i < 1e9; i++ {
	//
	//}
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
		sort.Slice(jobResponse, func(x, y int) bool {
			firstJob := jobResponse[x]
			secondJob := jobResponse[y]
			if firstJob.Name == secondJob.Name {
				return firstJob.Namespace < secondJob.Namespace
			}
			return jobResponse[x].Name < jobResponse[y].Name
		})
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
		var allocationRowEntries []nomad.AllocationRowEntry
		for _, alloc := range allocationResponse {
			for taskName, task := range alloc.TaskStates {
				allocationRowEntries = append(allocationRowEntries, nomad.AllocationRowEntry{
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
				return firstTask.Name < secondTask.Name
			}
			return allocationRowEntries[x].TaskName < allocationRowEntries[y].TaskName
		})
		return message.NomadAllocationMsg(allocationRowEntries)
	}
}

func FetchLogs(url, token, allocId, taskName string, logType nomad.LogType) tea.Cmd {
	return func() tea.Msg {
		// TODO LEO: error handling
		logTypeString := "stdout"
		switch logType {
		case nomad.StdOut:
			logTypeString = "stdout"
		case nomad.StdErr:
			logTypeString = "stderr"
		}

		body, _ := nomad.GetLogs(url, token, allocId, taskName, logTypeString)
		//simulateLoading()
		//body := MockLogsResponse
		var logRows []nomad.LogRow
		for _, log := range strings.Split(string(body), "\n") {
			logRows = append(logRows, nomad.LogRow(log))
		}
		return message.NomadLogsMsg{LogType: logType, Data: logRows}
	}
}
