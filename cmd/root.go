package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

type arg struct {
	cliShort, cliLong, config string
}

var (
	version, sha string

	// Used for flags.
	cfgFile string

	addrArg = arg{
		cliShort: "a",
		cliLong:  "addr",
		config:   "wander_addr",
	}
	tokenArg = arg{
		cliShort: "t",
		cliLong:  "token",
		config:   "wander_token",
	}
	updateSecondsArg = arg{
		cliShort: "u",
		cliLong:  "update",
		config:   "wander_update_seconds",
	}

	description = `wander is a terminal application for Nomad by HashiCorp. It is used to
view jobs, allocations, tasks, logs, and more, all from the terminal
in a productivity-focused UI.`

	rootCmd = &cobra.Command{
		Use:     "wander",
		Short:   "A terminal application for Nomad by HashiCorp",
		Long:    description,
		Run:     mainEntrypoint,
		Version: getVersion(),
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// root
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.wander.yaml)")
	rootCmd.PersistentFlags().BoolP("help", "", false, "Print usage")

	// NOTE: default values here are unused even if default exists as they break the desired priority of cli args > env vars > config file > default if exists
	rootCmd.PersistentFlags().StringP(tokenArg.cliLong, tokenArg.cliShort, "", "nomad token for HTTP API auth")
	viper.BindPFlag(tokenArg.cliLong, rootCmd.PersistentFlags().Lookup(tokenArg.config))
	rootCmd.PersistentFlags().StringP(addrArg.cliLong, addrArg.cliShort, "", "nomad address, e.g. http://localhost:4646")
	viper.BindPFlag(addrArg.cliLong, rootCmd.PersistentFlags().Lookup(addrArg.config))
	rootCmd.PersistentFlags().StringP(updateSecondsArg.cliLong, updateSecondsArg.cliShort, "", "number of seconds between page updates (-1 to disable)")
	viper.BindPFlag(updateSecondsArg.cliLong, rootCmd.PersistentFlags().Lookup(updateSecondsArg.config))

	// serve
	serveCmd.PersistentFlags().StringP(hostArg.cliLong, hostArg.cliShort, "", "host for wander ssh server")
	viper.BindPFlag(hostArg.cliLong, serveCmd.PersistentFlags().Lookup(hostArg.config))
	serveCmd.PersistentFlags().StringP(portArg.cliLong, portArg.cliShort, "", "port for wander ssh server")
	viper.BindPFlag(portArg.cliLong, serveCmd.PersistentFlags().Lookup(portArg.config))

	rootCmd.AddCommand(serveCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".wander")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func mainEntrypoint(cmd *cobra.Command, args []string) {
	nomadAddr := retrieveAssertExists(cmd, addrArg.cliLong, addrArg.config)
	nomadToken := retrieveAssertExists(cmd, tokenArg.cliLong, tokenArg.config)
	updateSeconds := retrieveUpdateSeconds(cmd)
	program := tea.NewProgram(initialModel(nomadAddr, nomadToken, updateSeconds), tea.WithAltScreen())

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v", err)
		os.Exit(1)
	}
}
