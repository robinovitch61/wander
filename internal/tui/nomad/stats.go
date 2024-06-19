package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
	"github.com/robinovitch61/wander/internal/tui/style"
)

func FetchStats(client api.Client, allocID, allocName string, styles style.Styles) tea.Cmd {
	return func() tea.Msg {
		alloc, _, err := client.Allocations().Info(allocID, nil)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		if alloc == nil {
			return message.ErrMsg{Err: fmt.Errorf("allocation %s (%s) not found", allocName, allocID)}
		}

		allocGivenResources := alloc.Resources
		if allocGivenResources == nil {
			return message.ErrMsg{Err: fmt.Errorf("allocation %s (%s) has no allocGivenResources", allocName, allocID)}
		}
		allocatedCpuMhz := allocGivenResources.CPU
		if allocatedCpuMhz == nil {
			return message.ErrMsg{Err: fmt.Errorf("allocation %s (%s) has no CPU resources", allocName, allocID)}
		}
		allocatedMemoryMB := allocGivenResources.MemoryMB
		if allocatedMemoryMB == nil {
			return message.ErrMsg{Err: fmt.Errorf("allocation %s (%s) has no memory resources", allocName, allocID)}
		}

		stats, err := client.Allocations().Stats(alloc, nil)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		var tableRows [][]string
		if stats != nil {
			allocUsage := stats.ResourceUsage
			var allocCpuTableVal, allocMemoryTableVal string
			if allocUsage != nil {
				allocMemory := allocUsage.MemoryStats
				if allocMemory != nil {
					memMiB := float64(allocMemory.Usage) / 1024 / 1024
					givenMemMiB := *allocatedMemoryMB
					perc := memMiB / float64(givenMemMiB) * 100
					stylePercent := styles.Regular
					if perc > 100 {
						stylePercent = styles.StatBad
					}
					percStr := stylePercent.Render(fmt.Sprintf("%.1f%%", perc))
					allocMemoryTableVal = fmt.Sprintf("%.0f/%d MiB (%s)", memMiB, givenMemMiB, percStr)
				}
				allocCpu := allocUsage.CpuStats
				if allocCpu != nil {
					cpuMhz := int(allocCpu.TotalTicks)
					givenCpuMhz := *allocatedCpuMhz
					perc := float64(cpuMhz) / float64(givenCpuMhz) * 100
					stylePercent := styles.Regular
					if perc > 100 {
						stylePercent = styles.StatBad
					}
					percStr := stylePercent.Render(fmt.Sprintf("%.1f%%", perc))
					allocCpuTableVal = fmt.Sprintf("%d/%d MHz (%s)", cpuMhz, givenCpuMhz, percStr)
				}
				tableRows = append(tableRows, []string{
					fmt.Sprintf("Allocation %s", allocName), allocMemoryTableVal, allocCpuTableVal,
				})
			}

			for taskName, taskResources := range stats.Tasks {
				if taskResources == nil {
					continue
				}
				taskGivenResources := alloc.TaskResources[taskName]
				if taskGivenResources == nil {
					continue
				}
				taskUsage := stats.ResourceUsage
				var taskCpuTableVal, taskMemoryTableVal string
				if taskUsage != nil {
					taskMemory := taskUsage.MemoryStats
					if taskMemory != nil {
						memMiB := float64(taskMemory.Usage) / 1024 / 1024
						givenMemMiB := *taskGivenResources.MemoryMB
						perc := memMiB / float64(givenMemMiB) * 100
						stylePercent := styles.Regular
						if perc > 100 {
							stylePercent = styles.StatBad
						}
						percStr := stylePercent.Render(fmt.Sprintf("%.1f%%", perc))
						taskMemoryTableVal = fmt.Sprintf("%.0f/%d MiB (%s)", memMiB, givenMemMiB, percStr)
					}
					taskCpu := taskUsage.CpuStats
					if taskCpu != nil {
						cpuMhz := int(taskCpu.TotalTicks)
						givenCpuMhz := *taskGivenResources.CPU
						perc := float64(cpuMhz) / float64(givenCpuMhz) * 100
						stylePercent := styles.Regular
						if perc > 100 {
							stylePercent = styles.StatBad
						}
						percStr := stylePercent.Render(fmt.Sprintf("%.1f%%", perc))
						taskCpuTableVal = fmt.Sprintf("%d/%d MHz (%s)", cpuMhz, givenCpuMhz, percStr)
					}
					tableRows = append(tableRows, []string{
						fmt.Sprintf("Task %s", taskName), taskMemoryTableVal, taskCpuTableVal,
					})
				}
			}
		}

		table := formatter.GetRenderedTableAsString([]string{"Entity", "Memory", "CPU"}, tableRows)
		var pageRows []page.Row
		for _, row := range table.ContentRows {
			pageRows = append(pageRows, page.Row{Key: "", Row: row})
		}

		return PageLoadedMsg{Page: StatsPage, TableHeader: table.HeaderRows, AllPageRows: pageRows}
	}
}
