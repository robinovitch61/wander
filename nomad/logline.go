package nomad

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/components/page"
	"github.com/robinovitch61/wander/formatter"
)

func FetchLogLine(logline string) tea.Cmd {
	return func() tea.Msg {
		// nothing actually async happens here, but this fits the PageLoadedMsg pattern
		pretty := formatter.PrettyJsonStringAsLines(logline)

		var loglinePageData []page.Row
		for _, row := range pretty {
			loglinePageData = append(loglinePageData, page.Row{Key: "", Row: row})
		}

		return PageLoadedMsg{
			Page:        LoglinePage,
			TableHeader: []string{},
			AllPageData: loglinePageData,
		}
	}
}
