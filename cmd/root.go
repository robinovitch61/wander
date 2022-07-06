package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type arg struct {
	cliShort, cliLong, config string
}

var (
	version, sha string

	// Used for flags.
	cfgFile string

	// TODO LEO: remove this in v1.0.0
	oldAddrArg = arg{
		cliShort: "a",
		cliLong:  "addr",
		config:   "wander_addr",
	}
	addrArg = arg{
		cliShort: "a",
		cliLong:  "addr",
		config:   "nomad_addr",
	}
	// TODO LEO: remove this in v1.0.0
	oldTokenArg = arg{
		cliShort: "t",
		cliLong:  "token",
		config:   "wander_token",
	}
	tokenArg = arg{
		cliShort: "t",
		cliLong:  "token",
		config:   "nomad_token",
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
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Config file (default is $HOME/.wander.yaml)")
	rootCmd.PersistentFlags().BoolP("help", "", false, "Print usage")

	// NOTE: default values here are unused even if default exists as they break the desired priority of cli args > env vars > config file > default if exists
	rootCmd.PersistentFlags().StringP(tokenArg.cliLong, tokenArg.cliShort, "", "Nomad token for HTTP API auth (default '')")
	viper.BindPFlag(tokenArg.cliLong, rootCmd.PersistentFlags().Lookup(tokenArg.config))
	rootCmd.PersistentFlags().StringP(addrArg.cliLong, addrArg.cliShort, "", "Nomad address, e.g. http://localhost:4646")
	viper.BindPFlag(addrArg.cliLong, rootCmd.PersistentFlags().Lookup(addrArg.config))
	rootCmd.PersistentFlags().StringP(updateSecondsArg.cliLong, updateSecondsArg.cliShort, "", "Number of seconds between page updates (-1 to disable)")
	viper.BindPFlag(updateSecondsArg.cliLong, rootCmd.PersistentFlags().Lookup(updateSecondsArg.config))

	// serve
	serveCmd.PersistentFlags().StringP(hostArg.cliLong, hostArg.cliShort, "", "Host for wander ssh server")
	viper.BindPFlag(hostArg.cliLong, serveCmd.PersistentFlags().Lookup(hostArg.config))
	serveCmd.PersistentFlags().StringP(portArg.cliLong, portArg.cliShort, "", "Port for wander ssh server")
	viper.BindPFlag(portArg.cliLong, serveCmd.PersistentFlags().Lookup(portArg.config))

	rootCmd.AddCommand(serveCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		_, err := os.Stat(cfgFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		extension := filepath.Ext(cfgFile)
		if extension != ".yaml" && extension != ".yml" {
			fmt.Println("error: config file must be .yaml or .yml")
			os.Exit(1)
		}

		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

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
	nomadAddr := retrieveAddress(cmd)
	nomadToken := retrieveToken(cmd)
	updateSeconds := retrieveUpdateSeconds(cmd)
	program := tea.NewProgram(initialModel(nomadAddr, nomadToken, updateSeconds), tea.WithAltScreen())

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v", err)
		os.Exit(1)
	}
}
