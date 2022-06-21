package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/app"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	// Used for flags.
	cfgFile string

	nomadAddrArg       = "nomad_addr"
	nomadAddrArgShort  = "addr"
	nomadTokenArg      = "nomad_token"
	nomadTokenArgShort = "token"

	rootCmd = &cobra.Command{
		Use:   "wander",
		Short: "A terminal application for Nomad by HashiCorp.",
		Long: `wander is a terminal application for Nomad by HashiCorp. It is used to
view jobs, allocations, tasks, logs, and more, all from the terminal
in a productivity-focused UI.`,
		Run: entrypoint,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.wander.yaml)")

	rootCmd.PersistentFlags().StringP(nomadTokenArgShort, "t", "", "nomad token for HTTP API auth.")
	rootCmd.MarkFlagRequired(nomadTokenArgShort)
	viper.BindPFlag(nomadTokenArgShort, rootCmd.PersistentFlags().Lookup(nomadTokenArg))

	rootCmd.PersistentFlags().StringP(nomadAddrArgShort, "a", "", "nomad address, e.g. http://localhost:4646.")
	rootCmd.MarkFlagRequired(nomadAddrArgShort)
	viper.BindPFlag(nomadAddrArgShort, rootCmd.PersistentFlags().Lookup(nomadAddrArg))
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

func postInitCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		presetRequiredFlags(cmd)
		if cmd.HasSubCommands() {
			postInitCommands(cmd.Commands())
		}
	}
}

func presetRequiredFlags(cmd *cobra.Command) {
	viper.BindPFlags(cmd.Flags())
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
			cmd.Flags().Set(f.Name, viper.GetString(f.Name))
		}
	})
}

func entrypoint(cmd *cobra.Command, args []string) {
	nomadAddr := retrieve(cmd, nomadAddrArgShort, nomadAddrArg)
	nomadToken := retrieve(cmd, nomadTokenArgShort, nomadTokenArg)
	program := tea.NewProgram(app.InitialModel(nomadAddr, nomadToken), tea.WithAltScreen())

	dev.Debug("~STARTING UP~")
	if err := program.Start(); err != nil {
		fmt.Printf("Error on wander startup: %v", err)
		os.Exit(1)
	}
}

func retrieve(cmd *cobra.Command, short, long string) string {
	val := cmd.Flag(short).Value.String()
	if val == "" {
		val = viper.GetString(long)
	}
	if val == "" {
		fmt.Println(fmt.Errorf("error: set %s env variable, %s in config file, or --%s flag", strings.ToUpper(long), long, short))
		os.Exit(1)
	}
	return val
}

// func main() {
// 	program := tea.NewProgram(app.InitialModel(), tea.WithAltScreen())
//
// 	dev.Debug("~STARTING UP~")
// 	if err := program.Start(); err != nil {
// 		fmt.Printf("Error on wander startup: %v", err)
// 		os.Exit(1)
// 	}
// }
