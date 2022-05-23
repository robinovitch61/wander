package jobs

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"sort"
	"strconv"
	"strings"
	"wander/components/page"
	"wander/formatter"
	"wander/message"
	"wander/nomad"
)

var Columns = []string{"ID", "Type", "Namespace", "Priority", "Status", "Submit Time"}

type NomadJobsMsg []JobResponseEntry

type JobResponseEntry struct {
	ID                string      `json:"ID"`
	ParentID          string      `json:"ParentID"`
	Name              string      `json:"Name"`
	Namespace         string      `json:"Namespace"`
	Datacenters       []string    `json:"Datacenters"`
	Multiregion       interface{} `json:"Multiregion"`
	Type              string      `json:"Type"`
	Priority          int         `json:"Priority"`
	Periodic          bool        `json:"Periodic"`
	ParameterizedJob  bool        `json:"ParameterizedJob"`
	Stop              bool        `json:"Stop"`
	Status            string      `json:"Status"`
	StatusDescription string      `json:"StatusDescription"`
	JobSummary        struct {
		JobID     string `json:"JobID"`
		Namespace string `json:"Namespace"`
		Summary   struct {
			YourProjectName struct {
				Queued   int `json:"Queued"`
				Complete int `json:"Complete"`
				Failed   int `json:"Failed"`
				Running  int `json:"Running"`
				Starting int `json:"Starting"`
				Lost     int `json:"Lost"`
			} `json:"your_project_name"`
		} `json:"Summary"`
		Children struct {
			Pending int `json:"Pending"`
			Running int `json:"Running"`
			Dead    int `json:"Dead"`
		} `json:"Children"`
		CreateIndex int `json:"CreateIndex"`
		ModifyIndex int `json:"ModifyIndex"`
	} `json:"JobSummary"`
	CreateIndex    int   `json:"CreateIndex"`
	ModifyIndex    int   `json:"ModifyIndex"`
	JobModifyIndex int   `json:"JobModifyIndex"`
	SubmitTime     int64 `json:"SubmitTime"`
}

func (e JobResponseEntry) MatchesFilter(filter string) bool {
	return strings.Contains(e.ID, filter)
}

func FetchJobs(url, token string) tea.Cmd {
	return func() tea.Msg {
		params := map[string]string{
			"namespace": "*",
		}
		fullPath := fmt.Sprintf("%s%s", url, "/v1/jobs")
		body, err := nomad.Get(fullPath, token, params)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		var jobResponse []JobResponseEntry
		if err := json.Unmarshal(body, &jobResponse); err != nil {
			return message.ErrMsg{Err: err}
		}

		sort.Slice(jobResponse, func(x, y int) bool {
			firstJob := jobResponse[x]
			secondJob := jobResponse[y]
			if firstJob.Name == secondJob.Name {
				return firstJob.Namespace < secondJob.Namespace
			}
			return jobResponse[x].Name < jobResponse[y].Name
		})

		return NomadJobsMsg(jobResponse)
	}
}

func JobResponsesAsTable(jobResponse []JobResponseEntry) ([]string, []page.Row) {
	var jobResponseRows [][]string
	var keys []string
	for _, row := range jobResponse {
		jobResponseRows = append(jobResponseRows, []string{
			row.ID,
			row.Type,
			row.Namespace,
			strconv.Itoa(row.Priority),
			row.Status,
			formatter.FormatTimeNs(row.SubmitTime),
		})
		keys = append(keys, row.ID)
	}

	table := formatter.GetRenderedTableAsString(Columns, jobResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}
