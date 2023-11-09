package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type arg struct {
	cliShort, cfgFileEnvVar, description, defaultString string
	isBool, isInt, defaultIfBool                        bool
	defaultIfInt                                        int
}

var (
	rootNameToArg = map[string]arg{
		"config": {
			cliShort:    "c",
			description: `Config file path. Will check for $HOME/.wander.yaml if not specified`,
		},
		"help": {
			description: `Print usage`,
		},
		"addr": {
			cliShort:      "a",
			cfgFileEnvVar: "nomad_addr",
			description:   `Nomad address`,
			defaultString: "http://localhost:4646",
		},
		"token": {
			cliShort:      "t",
			cfgFileEnvVar: "nomad_token",
			description:   `Nomad token`,
		},
		"region": {
			cliShort:      "r",
			cfgFileEnvVar: "nomad_region",
			description:   `Nomad region`,
		},
		"namespace": {
			cliShort:      "n",
			cfgFileEnvVar: "nomad_namespace",
			description:   `Nomad namespace`,
			defaultString: "*",
		},
		"http-auth": {
			cfgFileEnvVar: "nomad_http_auth",
			description:   `Nomad http auth, in the form of "user" or "user:pass"`,
		},
		"cacert": {
			cfgFileEnvVar: "nomad_cacert",
			description:   `Path to a PEM encoded CA cert file to use to verify the Nomad server SSL certificate`,
		},
		"capath": {
			cfgFileEnvVar: "nomad_capath",
			description:   `Path to a directory of PEM encoded CA cert files to verify the Nomad server SSL certificate. If both cacert and capath are specified, cacert is used`,
		},
		"client-cert": {
			cfgFileEnvVar: "nomad_client_cert",
			description:   `Path to a PEM encoded client cert for TLS authentication to the Nomad server. Must also specify client key`,
		},
		"client-key": {
			cfgFileEnvVar: "nomad_client_key",
			description:   `Path to an unencrypted PEM encoded private key matching the client cert`,
		},
		"tls-server-name": {
			cfgFileEnvVar: "nomad_tls_server_name",
			description:   `The server name to use as the SNI host when connecting via TLS`,
		},
		"skip-verify": {
			cfgFileEnvVar: "nomad_skip_verify",
			description:   `Do not verify TLS certificates`,
			isBool:        true,
			defaultIfBool: false,
		},
		"update": {
			cliShort:      "u",
			cfgFileEnvVar: "wander_update_seconds",
			description:   `Seconds between updates for job & allocation pages. Disable with -1`,
			isInt:         true,
			defaultIfInt:  2,
		},
		"job-columns": {
			cfgFileEnvVar: "wander_job_columns",
			description:   `Columns to display for Jobs view - can reference Meta keys`,
			defaultString: "Job,Type,Namespace,Status,Count,Submitted,Since Submit",
		},
		"all-tasks-columns": {
			cfgFileEnvVar: "wander_all_tasks_columns",
			description:   `Columns to display for All Tasks view`,
			defaultString: "Job,Alloc ID,Task Group,Alloc Name,Task Name,State,Started,Finished,Uptime",
		},
		"tasks-for-job-columns": {
			cfgFileEnvVar: "wander_tasks_for_job_columns",
			description:   `Columns to display for Tasks for Job view`,
			defaultString: "Alloc ID,Task Group,Alloc Name,Task Name,State,Started,Finished,Uptime",
		},
		"log-offset": {
			cliShort:      "o",
			cfgFileEnvVar: "wander_log_offset",
			description:   `Log byte offset from which logs start`,
			isInt:         true,
			defaultIfInt:  1000000,
		},
		"log-tail": {
			cliShort:      "f",
			cfgFileEnvVar: "wander_log_tail",
			description:   `Follow new logs as they come in rather than having to reload`,
			isBool:        true,
			defaultIfBool: true,
		},
		"copy-save-path": {
			cliShort:      "s",
			cfgFileEnvVar: "wander_copy_save_path",
			description:   `Copy the full path to file after save`,
			isBool:        true,
			defaultIfBool: false,
		},
		"event-topics": {
			cfgFileEnvVar: "wander_event_topics",
			description:   `Topics to follow in event streams, comma-separated`,
			defaultString: "Job,Allocation,Deployment,Evaluation",
		},
		"event-namespace": {
			cfgFileEnvVar: "wander_event_namespace",
			description:   `Namespace used in stream for all events. "*" for all namespaces`,
			defaultString: "default",
		},
		"event-jq-query": {
			cfgFileEnvVar: "wander_event_jq_query",
			description:   `jq query for events. "." for entire JSON`,
			defaultString: constants.DefaultEventJQQuery,
		},
		"logo-color": {
			cfgFileEnvVar: "wander_logo_color",
		},
		"compact-header": {
			cfgFileEnvVar: "wander_compact_header",
			description:   `Start with compact header`,
			isBool:        true,
			defaultIfBool: false,
		},
		"start-all-tasks": {
			cfgFileEnvVar: "wander_start_all_tasks",
			description:   `Start in All Tasks view`,
			isBool:        true,
			defaultIfBool: false,
		},
		"compact-tables": {
			cfgFileEnvVar: "wander_compact_tables",
			description:   `Remove unnecessary gaps between table columns when possible`,
			isBool:        true,
			defaultIfBool: true,
		},
	}

	description = `wander is a terminal application for Nomad by HashiCorp. It is used to
view jobs, allocations, tasks, logs, and more, all from the terminal
in a productivity-focused UI.`

	rootCmd = &cobra.Command{
		Use:   "wander",
		Short: "A terminal application for Nomad by HashiCorp",
		Long:  description,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(cmd, rootNameToArg)
		},
		Run:     mainEntrypoint,
		Version: getVersion(),
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// init is called once when the cmd package is loaded
// https://golangdocs.com/init-function-in-golang
func init() {
	cliLong := "config"
	rootCmd.PersistentFlags().StringP(cliLong, rootNameToArg[cliLong].cliShort, rootNameToArg[cliLong].defaultString, rootNameToArg[cliLong].description)

	cliLong = "help"
	rootCmd.PersistentFlags().BoolP(cliLong, rootNameToArg[cliLong].cliShort, rootNameToArg[cliLong].defaultIfBool, rootNameToArg[cliLong].description)

	// colors, config or env var only
	viper.BindPFlag("", rootCmd.PersistentFlags().Lookup(rootNameToArg["logo-color"].cfgFileEnvVar))

	for _, cliLong = range []string{
		"addr",
		"token",
		"region",
		"namespace",
		"http-auth",
		"cacert",
		"capath",
		"client-cert",
		"client-key",
		"tls-server-name",
		"skip-verify",
		"update",
		"job-columns",
		"all-tasks-columns",
		"tasks-for-job-columns",
		"log-offset",
		"log-tail",
		"copy-save-path",
		"event-topics",
		"event-namespace",
		"event-jq-query",
		"compact-header",
		"start-all-tasks",
		"compact-tables",
	} {
		c := rootNameToArg[cliLong]
		if c.isBool {
			rootCmd.PersistentFlags().BoolP(cliLong, c.cliShort, c.defaultIfBool, c.description)
		} else if c.isInt {
			rootCmd.PersistentFlags().IntP(cliLong, c.cliShort, c.defaultIfInt, c.description)
		} else {
			rootCmd.PersistentFlags().StringP(cliLong, c.cliShort, c.defaultString, c.description)
		}
		viper.BindPFlag(cliLong, rootCmd.PersistentFlags().Lookup(c.cfgFileEnvVar))
	}

	// serve
	for _, cliLong := range []string{
		"host",
		"port",
		"host-key-path",
		"host-key-pem",
	} {
		c := serveNameToArg[cliLong]
		if c.isBool {
			serveCmd.PersistentFlags().BoolP(cliLong, c.cliShort, c.defaultIfBool, c.description)
		} else if c.isInt {
			serveCmd.PersistentFlags().IntP(cliLong, c.cliShort, c.defaultIfInt, c.description)
		} else {
			serveCmd.PersistentFlags().StringP(cliLong, c.cliShort, c.defaultString, c.description)
		}
		viper.BindPFlag(cliLong, serveCmd.PersistentFlags().Lookup(c.cfgFileEnvVar))
	}

	rootCmd.AddCommand(serveCmd)
}

func initConfig(cmd *cobra.Command, nameToArg map[string]arg) error {
	cfgFile := cmd.Flags().Lookup("config").Value.String()
	if cfgFile != "" {
		// Use config file from the flag
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
		// check if default ~/.wander.yaml exists and use it if so
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".wander")
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// no config file found, that's ok
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// bind viper to env vars
	viper.AutomaticEnv()

	bindFlags(cmd, nameToArg)
	return nil
}

func bindFlags(cmd *cobra.Command, nameToArg map[string]arg) {
	v := viper.GetViper()
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		cliLong := f.Name
		viperName := nameToArg[cliLong].cfgFileEnvVar

		// Apply the viper config value to the flag when the flag is not manually specified
		// and viper has a value from the config file or env var
		if !f.Changed && v.IsSet(viperName) {
			val := v.Get(viperName)
			err := cmd.Flags().Set(cliLong, fmt.Sprintf("%v", val))
			if err != nil {
				fmt.Println(fmt.Sprintf("error setting flag %s: %v", cliLong, err))
				os.Exit(1)
			}
		}
	})
}

func mainEntrypoint(cmd *cobra.Command, args []string) {
	initialModel, options := setup(cmd, "")
	program := tea.NewProgram(initialModel, options...)

	dev.Debug("~STARTING UP~")
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error on wander startup: %v", err)
		os.Exit(1)
	}
}
