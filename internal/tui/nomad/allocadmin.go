/*
	Admin Actions for tasks

Restart, Stop, etc.
*/
package nomad

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/formatter"
)

var (
	// AllocAdminActions maps task-specific AdminActions to their display text
	AllocAdminActions = map[AdminAction]string{
		RestartAllocAction: "Restart",
		//StopAllocAction:    "Stop",
	}
)

type AllocAdminActionCompleteMsg struct {
	TaskName, AllocName, AllocID string
}

type AllocAdminActionFailedMsg struct {
	Err                          error
	TaskName, AllocName, AllocID string
}

func (e AllocAdminActionFailedMsg) Error() string { return e.Err.Error() }

func GetAllocAdminText(
	adminAction AdminAction, allocName, allocID string) string {
	return fmt.Sprintf(
		"%s allocation %s (%s)",
		AllocAdminActions[adminAction],
		allocName, 
		formatter.ShortID(allocID))
}

func GetCmdForAllocAdminAction(
	client api.Client,
	adminAction AdminAction,
	allocName,
	allocID string,
) tea.Cmd {
	switch adminAction {
	case RestartAllocAction:
		return RestartAlloc(client, allocName, allocID)
	//case StopTaskAction:
	//	return StopTask(client, taskName, allocName, allocID)
	default:
		return nil
	}
}

func RestartAlloc(client api.Client, allocName, allocID string) tea.Cmd {

	taskName := ""

	return func() tea.Msg {
		alloc, _, err := client.Allocations().Info(allocID, nil)

		if err != nil {
			return AllocAdminActionFailedMsg{
				Err:      err,
				TaskName: taskName, AllocName: allocName, AllocID: allocID}
		}

		err = client.Allocations().Restart(alloc, taskName, nil)
		if err != nil {
			return AllocAdminActionFailedMsg{
				Err:      err,
				TaskName: taskName, AllocName: allocName, AllocID: allocID}
		}

		return AllocAdminActionCompleteMsg{
			TaskName: taskName, AllocName: allocName, AllocID: allocID}
	}
}
