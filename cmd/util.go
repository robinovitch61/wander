package cmd

import (
	"fmt"
	"github.com/charmbracelet/wish"
	"github.com/gliderlabs/ssh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
	"time"
)

func retrieveAssertExists(cmd *cobra.Command, short, long string) string {
	val := cmd.Flag(short).Value.String()
	if val == "" {
		val = viper.GetString(long)
	}
	if val == "" {
		fmt.Println(fmt.Errorf("error: set %s env variable, %s in config file, or --%s argument", strings.ToUpper(long), long, short))
		os.Exit(1)
	}
	return val
}

// CustomLoggingMiddleware provides basic connection logging. Connects are logged with the
// remote address, invoked command, TERM setting, window dimensions and if the
// auth was public key based. Disconnect will log the remote address and
// connection duration. It is custom because it excludes the ssh Command in the log.
func CustomLoggingMiddleware() wish.Middleware {
	return func(sh ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			ct := time.Now()
			hpk := s.PublicKey() != nil
			pty, _, _ := s.Pty()
			log.Printf("%s connect %s %v %v %v %v\n", s.User(), s.RemoteAddr().String(), hpk, pty.Term, pty.Window.Width, pty.Window.Height)
			sh(s)
			log.Printf("%s disconnect %s\n", s.RemoteAddr().String(), time.Since(ct))
		}
	}
}
