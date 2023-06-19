package nomad

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
	"sort"
)

func FetchJobMeta(client api.Client, jobID, jobNamespace string) tea.Cmd {
	return func() tea.Msg {
		jobSpec, _, err := client.Jobs().Info(jobID, &api.QueryOptions{Namespace: jobNamespace})
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		tableHeader, allPageData := metaAsTable(jobSpec.Meta)
		return PageLoadedMsg{
			Page:        JobMetaPage,
			TableHeader: tableHeader,
			AllPageRows: allPageData,
		}
	}
}

func metaAsTable(meta map[string]string) ([]string, []page.Row) {
	var metaRows [][]string
	var keys []string
	for k, v := range meta {
		metaRows = append(metaRows, []string{k, v})
		keys = append(keys, k)
	}

	columns := []string{"Key", "Value"}
	table := formatter.GetRenderedTableAsString(columns, metaRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Key < rows[j].Key
	})

	return table.HeaderRows, rows
}
