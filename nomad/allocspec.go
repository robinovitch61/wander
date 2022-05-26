package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"wander/components/page"
	"wander/formatter"
	"wander/message"
)

func FetchAllocSpec(url, token, allocID string) tea.Cmd {
	return func() tea.Msg {
		fullPath := fmt.Sprintf("%s%s%s", url, "/v1/allocation/", allocID)
		body, err := get(fullPath, token, nil)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		pretty := formatter.PrettyJsonStringAsLines(string(body), true)

		var allocSpecPageData []page.Row
		for _, row := range pretty {
			allocSpecPageData = append(allocSpecPageData, page.Row{Key: "", Row: row})
		}

		return PageLoadedMsg{
			Page:        AllocSpecPage,
			TableHeader: []string{},
			AllPageData: allocSpecPageData,
		}
	}
}
