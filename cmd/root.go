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
	logOffsetArg = arg{
		cliShort: "o",
		cliLong:  "log-offset",
		config:   "wander_log_offset",
	}
	copySavePathArg = arg{
		cliShort: "s",
		cliLong:  "copy-save-path",
		config:   "wander_copy_save_path",
	}
	eventTopicsArg = arg{
		cliLong: "event-topics",
		config:  "wander_event_topics",
	}
	eventNamespaceArg = arg{
		cliLong: "event-namespace",
		config:  "wander_event_namespace",
	}
	logoColorArg = arg{
		config: "wander_logo_color",
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
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", `Config file path. Default "$HOME/.wander.yaml"`)
	rootCmd.PersistentFlags().BoolP("help", "", false, "Print usage")

	// NOTE: default values here are unused even if default exists as they break the desired priority of cli args > env vars > config file > default if exists
	rootCmd.PersistentFlags().StringP(addrArg.cliLong, addrArg.cliShort, "", `Nomad address. Default "http://localhost:4646"`)
	viper.BindPFlag(addrArg.cliLong, rootCmd.PersistentFlags().Lookup(addrArg.config))
	rootCmd.PersistentFlags().StringP(tokenArg.cliLong, tokenArg.cliShort, "", `Nomad token. Default ""`)
	viper.BindPFlag(tokenArg.cliLong, rootCmd.PersistentFlags().Lookup(tokenArg.config))
	rootCmd.PersistentFlags().StringP(updateSecondsArg.cliLong, updateSecondsArg.cliShort, "", `Seconds between updates for job & allocation pages. Disable with "-1". Default "2"`)
	viper.BindPFlag(updateSecondsArg.cliLong, rootCmd.PersistentFlags().Lookup(updateSecondsArg.config))
	rootCmd.PersistentFlags().StringP(logOffsetArg.cliLong, logOffsetArg.cliShort, "", `Log byte offset from which logs start. Default "1000000"`)
	viper.BindPFlag(logOffsetArg.cliLong, rootCmd.PersistentFlags().Lookup(logOffsetArg.config))
	rootCmd.PersistentFlags().StringP(copySavePathArg.cliLong, copySavePathArg.cliShort, "", `If "true", copy the full path to file after save. Default "false"`)
	viper.BindPFlag(copySavePathArg.cliLong, rootCmd.PersistentFlags().Lookup(copySavePathArg.config))
	rootCmd.PersistentFlags().StringP(eventTopicsArg.cliLong, eventTopicsArg.cliShort, "", `Topics to follow in event streams, comma-separated. Default "Job,Allocation,Deployment,Evaluation"`)
	viper.BindPFlag(eventTopicsArg.cliLong, rootCmd.PersistentFlags().Lookup(eventTopicsArg.config))
	rootCmd.PersistentFlags().StringP(eventNamespaceArg.cliLong, eventNamespaceArg.cliShort, "", `Namespace used in stream for all events. "*" for all namespaces. Default "default"`)
	viper.BindPFlag(eventNamespaceArg.cliLong, rootCmd.PersistentFlags().Lookup(eventNamespaceArg.config))

	// colors
	viper.BindPFlag(logoColorArg.cliLong, rootCmd.PersistentFlags().Lookup(logoColorArg.config))

	// serve
	serveCmd.PersistentFlags().StringP(hostArg.cliLong, hostArg.cliShort, "", `Host for wander ssh server. Default "localhost"`)
	viper.BindPFlag(hostArg.cliLong, serveCmd.PersistentFlags().Lookup(hostArg.config))
	serveCmd.PersistentFlags().StringP(portArg.cliLong, portArg.cliShort, "", `Port for wander ssh server. Default "21324"`)
	viper.BindPFlag(portArg.cliLong, serveCmd.PersistentFlags().Lookup(portArg.config))
	serveCmd.PersistentFlags().StringP(hostKeyPathArg.cliLong, hostKeyPathArg.cliShort, "", `Host key path for wander ssh server. Default none, i.e. ""`)
	viper.BindPFlag(hostKeyPathArg.cliLong, serveCmd.PersistentFlags().Lookup(hostKeyPathArg.config))
	serveCmd.PersistentFlags().StringP(hostKeyPEMArg.cliLong, hostKeyPEMArg.cliShort, "", `Host key PEM block for wander ssh server. Default none, i.e. ""`)
	viper.BindPFlag(hostKeyPEMArg.cliLong, serveCmd.PersistentFlags().Lookup(hostKeyPEMArg.config))

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
	initialModel, options := setup(cmd, "")
	program := tea.NewProgram(initialModel, options...)

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v", err)
		os.Exit(1)
	}
}
