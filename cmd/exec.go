package cmd

import (
	"fmt"
	"github.com/robinovitch61/wander/internal/tui/nomad"
	"github.com/spf13/cobra"
	"os"
)

var (
	execCmd = &cobra.Command{
		Use:   "exec",
		Short: "Exec into a running task",
		Long:  `Exec into a running nomad task`,
		Example: `
  # specify job and task, assuming single allocation
  wander exec alright_stop --task redis echo "hi" 

  # specify allocation, assuming single task
  wander exec 3dca0982 echo "hi" 

  # use prefixes of jobs or allocation ids
  wander exec al echo "hi"
  wander exec 3d echo "hi"

  # specify flags for the exec command with --
  wander exec alright_stop --task redis -- echo -n "hi"
`,
		Run:               execEntrypoint,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
)

func execEntrypoint(cmd *cobra.Command, args []string) {
	task := cmd.Flags().Lookup("task").Value.String()
	client, err := getConfig(cmd, "").Client()
	if err != nil {
		fmt.Println(fmt.Errorf("could not get client: %v", err))
		os.Exit(1)
	}
	allocID := args[0]
	execArgs := args[1:]
	if len(execArgs) == 0 {
		fmt.Println("no command specified")
		os.Exit(1)
	}
	_, err = nomad.AllocExec(client, allocID, task, execArgs)
	if err != nil {
		fmt.Println(fmt.Errorf("could not exec into task: %v", err))
		os.Exit(1)
	}
}
