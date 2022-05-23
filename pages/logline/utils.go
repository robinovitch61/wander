package logline

import (
	tea "github.com/charmbracelet/bubbletea"
	"wander/components/page"
	"wander/formatter"
	"wander/message"
	"wander/pages"
)

func FetchLogLine(logline string) tea.Cmd {
	return func() tea.Msg {
		// nothing actually async happens here, but this fits the PageLoadedMsg pattern
		var loglinePageData []page.Row
		pretty := formatter.PrettyJsonStringAsLines(logline)
		for _, row := range pretty {
			loglinePageData = append(loglinePageData, page.Row{Key: "", Row: row})
		}
		return func() tea.Msg {
			return message.PageLoadedMsg{
				Page:        pages.Logline,
				TableHeader: []string{},
				AllPageData: loglinePageData,
			}
		}
	}
}
