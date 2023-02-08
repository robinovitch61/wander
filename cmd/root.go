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
	cliShort, cliLong, cfgFileEnvVar, description string
}

var (
	version, sha string

	// Used for flags.
	cfgFile string

	cfgArg = arg{
		cliShort:    "c",
		cliLong:     "config",
		description: `Config file path. Default "$HOME/.wander.yaml"`,
	}
	helpArg = arg{
		cliLong:     "help",
		description: `Print usage`,
	}
	// TODO LEO: remove this in v1.0.0
	oldAddrArg = arg{
		cliShort:      "a",
		cliLong:       "addr",
		cfgFileEnvVar: "wander_addr",
	}
	addrArg = arg{
		cliShort:      "a",
		cliLong:       "addr",
		cfgFileEnvVar: "nomad_addr",
		description:   `Nomad address. Default "http://localhost:4646"`,
	}
	// TODO LEO: remove this in v1.0.0
	oldTokenArg = arg{
		cliShort:      "t",
		cliLong:       "token",
		cfgFileEnvVar: "wander_token",
	}
	tokenArg = arg{
		cliShort:      "t",
		cliLong:       "token",
		cfgFileEnvVar: "nomad_token",
		description:   `Nomad token. Default ""`,
	}
	regionArg = arg{
		cliShort:      "r",
		cliLong:       "region",
		cfgFileEnvVar: "nomad_region",
		description:   `Nomad region. Default ""`,
	}
	namespaceArg = arg{
		cliShort:      "n",
		cliLong:       "namespace",
		cfgFileEnvVar: "nomad_namespace",
		description:   `Nomad namespace. Default "*"`,
	}
	httpAuthArg = arg{
		cliLong:       "http-auth",
		cfgFileEnvVar: "nomad_http_auth",
		description:   `Nomad http auth, in the form of "user" or "user:pass". Default ""`,
	}
	cacertArg = arg{
		cliLong:       "cacert",
		cfgFileEnvVar: "nomad_cacert",
		description:   `Path to a PEM encoded CA cert file to use to verify the Nomad server SSL certificate. Default ""`,
	}
	capathArg = arg{
		cliLong:       "capath",
		cfgFileEnvVar: "nomad_capath",
		description:   `Path to a directory of PEM encoded CA cert files to verify the Nomad server SSL certificate. If both cacert and capath are specified, cacert is used. Default ""`,
	}
	clientCertArg = arg{
		cliLong:       "client-cert",
		cfgFileEnvVar: "nomad_client_cert",
		description:   `Path to a PEM encoded client cert for TLS authentication to the Nomad server. Must also specify client key. Default ""`,
	}
	clientKeyArg = arg{
		cliLong:       "client-key",
		cfgFileEnvVar: "nomad_client_key",
		description:   `Path to an unencrypted PEM encoded private key matching the client cert. Default ""`,
	}
	tlsServerNameArg = arg{
		cliLong:       "tls-server-name",
		cfgFileEnvVar: "nomad_tls_server_name",
		description:   `The server name to use as the SNI host when connecting via TLS. Default ""`,
	}
	skipVerifyArg = arg{
		cliLong:       "skip-verify",
		cfgFileEnvVar: "nomad_skip_verify",
		description:   `If "true", do not verify TLS certificates. Default "false"`,
	}
	updateSecondsArg = arg{
		cliShort:      "u",
		cliLong:       "update",
		cfgFileEnvVar: "wander_update_seconds",
		description:   `Seconds between updates for job & allocation pages. Disable with "-1". Default "2"`,
	}
	logOffsetArg = arg{
		cliShort:      "o",
		cliLong:       "log-offset",
		cfgFileEnvVar: "wander_log_offset",
		description:   `Log byte offset relative to end from which logs start. Default "1000000"`,
	}
	logTailArg = arg{
		cliShort:      "f",
		cliLong:       "log-tail",
		cfgFileEnvVar: "wander_log_tail",
		description:   `Follow new logs as they come in rather than having to reload. Default "true"`,
	}
	copySavePathArg = arg{
		cliShort:      "s",
		cliLong:       "copy-save-path",
		cfgFileEnvVar: "wander_copy_save_path",
		description:   `If "true", copy the full path to file after save. Default "false"`,
	}
	eventTopicsArg = arg{
		cliLong:       "event-topics",
		cfgFileEnvVar: "wander_event_topics",
		description:   `Topics to follow in event streams, comma-separated. Default "Job,Allocation,Deployment,Evaluation"`,
	}
	eventNamespaceArg = arg{
		cliLong:       "event-namespace",
		cfgFileEnvVar: "wander_event_namespace",
		description:   `Namespace used in stream for all events. "*" for all namespaces. Default "default"`,
	}
	eventJQQueryArg = arg{
		cliLong:       "event-jq-query",
		cfgFileEnvVar: "wander_event_jq_query",
		description:   `jq query for events. "." for entire JSON. Default shown at https://github.com/robinovitch61/wander`,
	}
	logoColorArg = arg{
		cfgFileEnvVar: "wander_logo_color",
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
	// NOTE: default values here are unused even if default exists as they break the desired priority of cli args > env vars > config file > default if exists

	// root
	rootCmd.PersistentFlags().StringVarP(&cfgFile, cfgArg.cliLong, cfgArg.cliShort, "", cfgArg.description)
	rootCmd.PersistentFlags().BoolP(helpArg.cliLong, helpArg.cliShort, false, helpArg.description)
	for _, c := range []arg{
		addrArg,
		tokenArg,
		regionArg,
		namespaceArg,
		httpAuthArg,
		cacertArg,
		capathArg,
		clientCertArg,
		clientKeyArg,
		tlsServerNameArg,
		skipVerifyArg,
		updateSecondsArg,
		logOffsetArg,
		logTailArg,
		copySavePathArg,
		eventTopicsArg,
		eventNamespaceArg,
		eventJQQueryArg,
	} {
		rootCmd.PersistentFlags().StringP(c.cliLong, c.cliShort, "", c.description)
		viper.BindPFlag(c.cliLong, rootCmd.PersistentFlags().Lookup(c.cfgFileEnvVar))
	}

	// colors, config or env var only
	viper.BindPFlag(logoColorArg.cliLong, rootCmd.PersistentFlags().Lookup(logoColorArg.cfgFileEnvVar))

	// serve
	for _, c := range []arg{
		hostArg,
		portArg,
		hostKeyPathArg,
		hostKeyPEMArg,
	} {
		serveCmd.PersistentFlags().StringP(c.cliLong, c.cliShort, "", c.description)
		viper.BindPFlag(c.cliLong, serveCmd.PersistentFlags().Lookup(c.cfgFileEnvVar))
	}

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
