package nomad

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
)

func FetchJobSpec(client api.Client, jobID, jobNamespace string) tea.Cmd {
	return func() tea.Msg {
		jobSpec, _, err := client.Jobs().Info(jobID, &api.QueryOptions{Namespace: jobNamespace})
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		jobBytes, err := json.Marshal(jobSpec)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		pretty := formatter.PrettyJsonStringAsLines(string(jobBytes))

		var jobSpecPageData []page.Row
		for _, row := range pretty {
			jobSpecPageData = append(jobSpecPageData, page.Row{Key: "", Row: row})
		}

		return PageLoadedMsg{
			Page:        JobSpecPage,
			TableHeader: []string{},
			AllPageRows: jobSpecPageData,
		}
	}
}
