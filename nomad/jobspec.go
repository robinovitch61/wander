package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"wander/components/page"
	"wander/formatter"
	"wander/message"
)

func FetchJobSpec(url, token, jobID string) tea.Cmd {
	return func() tea.Msg {
		fullPath := fmt.Sprintf("%s%s%s", url, "/v1/job/", jobID)
		body, err := get(fullPath, token, nil)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		pretty := formatter.PrettyJsonStringAsLines(string(body), true)

		var jobSpecPageData []page.Row
		for _, row := range pretty {
			jobSpecPageData = append(jobSpecPageData, page.Row{Key: "", Row: row})
		}

		return PageLoadedMsg{
			Page:        JobSpecPage,
			TableHeader: []string{},
			AllPageData: jobSpecPageData,
		}
	}
}
