package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
)

func FetchAllocSpec(url, token, allocID string) tea.Cmd {
	return func() tea.Msg {
		fullPath := fmt.Sprintf("%s%s%s", url, "/v1/allocation/", allocID)
		body, err := get(fullPath, token, nil)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		pretty := formatter.PrettyJsonStringAsLines(string(body))

		var allocSpecPageData []page.Row
		for _, row := range pretty {
			allocSpecPageData = append(allocSpecPageData, page.Row{Key: "", Row: row})
		}

		return PageLoadedMsg{
			Page:        AllocSpecPage,
			TableHeader: []string{},
			AllPageRows: allocSpecPageData,
		}
	}
}
