package cmd

import (
	"fmt"
	"github.com/robinovitch61/wander/internal/tui/nomad"
	"github.com/spf13/cobra"
	"os"
)

var (
	execCmd = &cobra.Command{
		Use:               "exec",
		Short:             "Exec into a running task",
		Long:              `Exec into a running nomad task`,
		Run:               execEntrypoint,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
)

func execEntrypoint(cmd *cobra.Command, args []string) {
	client, err := getConfig(cmd, "").Client()
	if err != nil {
		fmt.Println(fmt.Errorf("could not get client: %v", err))
		os.Exit(1)
	}
	allocID := args[0]
	task := args[1]
	execArgs := args[2:]
	_, err = nomad.AllocExec(client, allocID, task, execArgs)
	if err != nil {
		fmt.Println(fmt.Errorf("could not exec into task: %v", err))
		os.Exit(1)
	}
}
