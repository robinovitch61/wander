package nomad

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"github.com/robinovitch61/wander/internal/tui/message"
)

var (
	// TaskAdminActions maps task-specific AdminActions to their display text
	TaskAdminActions = map[AdminAction]string{
		RestartTaskAction: "Restart",
		//StopTaskAction:    "Stop",
	}
)

type TaskAdminActionCompleteMsg struct {
	TaskName, AllocName, AllocID string
}

func GetTaskAdminText(adminAction AdminAction, taskName, allocName, allocID string) string {
	return fmt.Sprintf("%s task %s in %s (%s)", TaskAdminActions[adminAction], taskName, allocName, formatter.ShortAllocID(allocID))
}

func GetCmdForTaskAdminAction(client api.Client, adminAction AdminAction, taskName, allocName, allocID string) tea.Cmd {
	switch adminAction {
	case RestartTaskAction:
		return RestartTask(client, taskName, allocName, allocID)
	//case StopTaskAction:
	//	return StopTask(client, taskName, allocName, allocID)
	default:
		return nil
	}
}

func RestartTask(client api.Client, taskName, allocName, allocID string) tea.Cmd {
	return func() tea.Msg {
		alloc, _, err := client.Allocations().Info(allocID, nil)
		if err != nil {
			// TODO LEO: we could return a TaskAdminActionFailedMsg here and display it in the toast
			return message.ErrMsg{Err: err}
		}

		err = client.Allocations().Restart(alloc, taskName, nil)
		if err != nil {
			// TODO LEO: we could return a TaskAdminActionFailedMsg here and display it in the toast
			return message.ErrMsg{Err: err}
		}

		return TaskAdminActionCompleteMsg{TaskName: taskName, AllocName: allocName, AllocID: allocID}
	}
}
