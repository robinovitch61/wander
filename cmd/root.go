package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

type arg struct {
	cliShort, cliLong, config string
}

var (
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

	description = `wander is a terminal application for Nomad by HashiCorp. It is used to
view jobs, allocations, tasks, logs, and more, all from the terminal
in a productivity-focused UI.`

	rootCmd = &cobra.Command{
		Use:   "wander",
		Short: "A terminal application for Nomad by HashiCorp",
		Long:  description,
		Run:   mainEntrypoint,
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
	rootCmd.PersistentFlags().StringP(tokenArg.cliLong, tokenArg.cliShort, "", "nomad token for HTTP API auth")
	viper.BindPFlag(tokenArg.cliLong, rootCmd.PersistentFlags().Lookup(tokenArg.config))
	rootCmd.PersistentFlags().StringP(addrArg.cliLong, addrArg.cliShort, "", "nomad address, e.g. http://localhost:4646")
	viper.BindPFlag(addrArg.cliLong, rootCmd.PersistentFlags().Lookup(addrArg.config))

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
	program := tea.NewProgram(app.InitialModel(nomadAddr, nomadToken), tea.WithAltScreen())

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v", err)
		os.Exit(1)
	}
}
