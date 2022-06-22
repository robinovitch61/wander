package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

func retrieveAssertExists(cmd *cobra.Command, short, long string) string {
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
