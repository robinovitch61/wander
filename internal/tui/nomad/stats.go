package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
)

func FetchStats(client api.Client, allocID, allocName, taskName string) tea.Cmd {
	// TODO LEO: list all task resources, and add denominators and percentages
	return func() tea.Msg {
		alloc, _, err := client.Allocations().Info(allocID, nil)
		if err != nil {
			return message.ErrMsg{Err: err}
		}

		stats, err := client.Allocations().Stats(alloc, nil)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		var tableRows [][]string
		if stats != nil {
			allocUsage := stats.ResourceUsage
			if allocUsage != nil {
				allocMemory := allocUsage.MemoryStats
				allocCpu := allocUsage.CpuStats
				if allocMemory != nil {
					tableRows = append(tableRows, []string{fmt.Sprintf("Allocation %s Memory", allocName), fmt.Sprintf("%d KiB", int(float64(allocMemory.Usage)/1024))})
				}
				if allocCpu != nil {
					tableRows = append(tableRows, []string{fmt.Sprintf("Allocation %s CPU", allocName), fmt.Sprintf("%d MHz", int(allocCpu.TotalTicks))})
				}
			}

			if taskResources, ok := stats.Tasks[taskName]; ok && taskResources != nil {
				taskUsage := stats.ResourceUsage
				if taskUsage != nil {
					taskMemory := taskUsage.MemoryStats
					taskCpu := taskUsage.CpuStats
					if taskMemory != nil {
						tableRows = append(tableRows, []string{fmt.Sprintf("Task %s Memory", taskName), fmt.Sprintf("%d KiB", int(float64(taskMemory.Usage)/1024))})
					}
					if taskCpu != nil {
						tableRows = append(tableRows, []string{fmt.Sprintf("Task %s CPU", taskName), fmt.Sprintf("%d MHz", int(taskCpu.TotalTicks))})
					}
				}
			}
		}

		table := formatter.GetRenderedTableAsString([]string{"Stat", "Value"}, tableRows)
		var pageRows []page.Row
		for _, row := range table.ContentRows {
			pageRows = append(pageRows, page.Row{Key: "", Row: row})
		}

		return PageLoadedMsg{Page: StatsPage, TableHeader: []string{"Stats"}, AllPageRows: pageRows}
	}
}
