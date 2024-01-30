/* Admin Actions for tasks
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
		RestartTaskAction: "Restart",
		//StopTaskAction:    "Stop",
	}
)

type AllocAdminActionCompleteMsg struct {
	TaskName, AllocName, AllocID string
}

type AllocAdminActionFailedMsg struct {
	Err error
	TaskName, AllocName, AllocID string
}

func (e AllocAdminActionFailedMsg) Error() string { return e.Err.Error() }

func GetAllocAdminText(
	adminAction AdminAction, taskName, allocName, allocID string) string {
	return fmt.Sprintf(
		"%s task %s in %s (%s)",
		AllocAdminActions[adminAction],
		taskName, allocName, formatter.ShortAllocID(allocID))
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
				Err: err,
				TaskName: taskName, AllocName: allocName, AllocID: allocID}
		}

		err = client.Allocations().Restart(alloc, taskName, nil)
		if err != nil {
			return AllocAdminActionFailedMsg{
				Err: err,
				TaskName: taskName, AllocName: allocName, AllocID: allocID}
		}

		return AllocAdminActionCompleteMsg{
			TaskName: taskName, AllocName: allocName, AllocID: allocID}
	}
}
