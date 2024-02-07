package nomad

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
)

var (
	JobAdminActions = map[AdminAction]string{
		StopJobAction:   "Stop",
		StopAndPurgeJobAction:   "Stop and purge",
	}
)

type JobAdminActionCompleteMsg struct {
	Err     error
	JobID string
}

func GetJobAdminText(adminAction AdminAction, jobID string) string {
	switch adminAction {
	case StopJobAction, StopAndPurgeJobAction:
		return fmt.Sprintf(
			"%s job %s",
			JobAdminActions[adminAction], jobID)
	default:
		return ""
	}
}

func GetCmdForJobAdminAction(
	client api.Client,
	adminAction AdminAction,
	jobID, jobNamespace string,
) tea.Cmd {
	switch adminAction {
	case StopJobAction:
		return StopJob(client, jobID, jobNamespace, false)
	case StopAndPurgeJobAction:
		return StopJob(client, jobID, jobNamespace, true)
	default:
		return nil
	}
}

func StopJob(client api.Client, jobID, jobNamespace string, purge bool) tea.Cmd {
	return func() tea.Msg {
		opts := &api.WriteOptions{Namespace: jobNamespace}
		_, _, err := client.Jobs().Deregister(jobID, purge, opts)
		if err != nil {
			return JobAdminActionCompleteMsg{
				Err: err,
				JobID: jobID,
			}
		}

		return JobAdminActionCompleteMsg{JobID: jobID}
	}
}
