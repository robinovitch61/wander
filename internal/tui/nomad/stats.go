package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/message"
)

func FetchStats(client api.Client, allocID, taskName string) tea.Cmd {
	return func() tea.Msg {
		alloc, _, err := client.Allocations().Info(allocID, nil)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		stats, err := client.Allocations().Stats(alloc, nil)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		dev.Debug(fmt.Sprintf("Stats: %+v", stats))

		return PageLoadedMsg{Page: StatsPage, TableHeader: []string{"Stats"}, AllPageRows: []page.Row{}}
	}
}
