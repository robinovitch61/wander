package nomad

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
)

func FetchAllocSpec(client api.Client, allocID string) tea.Cmd {
	return func() tea.Msg {
		alloc, _, err := client.Allocations().Info(allocID, nil)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		allocBytes, err := json.Marshal(alloc)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		pretty := formatter.PrettyJsonStringAsLines(string(allocBytes))

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
