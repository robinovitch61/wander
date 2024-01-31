package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/formatter"
)

var (
	// AllocAdminActions maps allocation-specific AdminActions to their display text
	AllocAdminActions = map[AdminAction]string{
		RestartTaskAction: "Restart",
		StopAllocAction:   "Stop",
	}
)

type AllocAdminActionCompleteMsg struct {
	Err                          error
	TaskName, AllocName, AllocID string
}

func GetAllocAdminText(adminAction AdminAction, taskName, allocName, allocID string) string {
	switch adminAction {
	case RestartTaskAction:
		return fmt.Sprintf(
			"%s task %s in %s (%s)",
			AllocAdminActions[adminAction],
			taskName, allocName, formatter.ShortAllocID(allocID))
	case StopAllocAction:
		return fmt.Sprintf(
			"%s allocation %s (%s)",
			AllocAdminActions[adminAction],
			allocName, formatter.ShortAllocID(allocID))
	default:
		return ""
	}
}

func GetCmdForAllocAdminAction(
	client api.Client,
	adminAction AdminAction,
	taskName,
	allocName,
	allocID string,
) tea.Cmd {
	switch adminAction {
	case RestartTaskAction:
		return RestartTask(client, taskName, allocName, allocID)
	case StopAllocAction:
		return StopAllocation(client, allocName, allocID)
	default:
		return nil
	}
}

func RestartTask(client api.Client, taskName, allocName, allocID string) tea.Cmd {
	return func() tea.Msg {
		alloc, _, err := client.Allocations().Info(allocID, nil)
		if err != nil {
			return AllocAdminActionCompleteMsg{
				Err:      err,
				TaskName: taskName, AllocName: allocName, AllocID: allocID,
			}
		}
		err = client.Allocations().Restart(alloc, taskName, nil)
		if err != nil {
			return AllocAdminActionCompleteMsg{
				Err:      err,
				TaskName: taskName, AllocName: allocName, AllocID: allocID,
			}
		}
		return AllocAdminActionCompleteMsg{TaskName: taskName, AllocName: allocName, AllocID: allocID}
	}
}

func StopAllocation(client api.Client, allocName, allocID string) tea.Cmd {
	return func() tea.Msg {
		alloc, _, err := client.Allocations().Info(allocID, nil)
		if err != nil {
			return AllocAdminActionCompleteMsg{
				Err:       err,
				AllocName: allocName, AllocID: allocID,
			}
		}
		_, err = client.Allocations().Stop(alloc, nil)
		if err != nil {
			return AllocAdminActionCompleteMsg{
				Err:       err,
				AllocName: allocName, AllocID: allocID,
			}
		}
		return AllocAdminActionCompleteMsg{AllocName: allocName, AllocID: allocID}
	}
}
