package cmd

import (
	"fmt"
	"github.com/charmbracelet/wish"
	"github.com/gliderlabs/ssh"
	"github.com/robinovitch61/wander/internal/tui/components/app"
	"github.com/robinovitch61/wander/internal/tui/constants"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	// Version contains the application version number. It's set via ldflags
	// in the .goreleaser.yaml file when building
	Version = ""

	// CommitSHA contains the SHA of the commit that this application was built
	// against. It's set via ldflags in the .goreleaser.yaml file when building
	CommitSHA = ""
)

func retrieve(cmd *cobra.Command, a arg) (string, error) {
	val := cmd.Flag(a.cliLong).Value.String()
	if val == "" {
		val = viper.GetString(a.config)
	}
	if val == "" {
		return "", fmt.Errorf("error: set %s env variable, %s in config file, or --%s argument", strings.ToUpper(a.config), a.config, a.cliLong)
	}
	return val, nil
}

func retrieveAssertExistsWithFallback(cmd *cobra.Command, currArg, oldArg arg) (string, error) {
	val, err := retrieve(cmd, currArg)
	if err != nil {
		val, _ = retrieve(cmd, oldArg)
		if val == "" {
			return "", err
		}
		fmt.Printf("\nwarning: use of %s env variable or %s in config file will be removed in a future release\n", strings.ToUpper(oldArg.config), oldArg.config)
		fmt.Printf("use %s env variable or %s in config file instead\n", strings.ToUpper(currArg.config), currArg.config)
	}
	return val, nil
}

func retrieveWithDefault(cmd *cobra.Command, short, long, defaultVal string) string {
	val := cmd.Flag(short).Value.String()
	if val == "" {
		val = viper.GetString(long)
	}
	if val == "" {
		return defaultVal
	}
	return val
}

func retrieveAddress(cmd *cobra.Command) string {
	val, err := retrieveAssertExistsWithFallback(cmd, addrArg, oldAddrArg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return val
}

func retrieveToken(cmd *cobra.Command) string {
	val, err := retrieveAssertExistsWithFallback(cmd, tokenArg, oldTokenArg)
	if err != nil {
		return ""
	}
	if len(val) > 0 && len(val) != 36 {
		fmt.Println("token must be 36 characters")
		os.Exit(1)
	}
	return val
}

func retrieveUpdateSeconds(cmd *cobra.Command) int {
	updateSecondsString := retrieveWithDefault(cmd, updateSecondsArg.cliLong, updateSecondsArg.config, constants.DefaultUpdateSeconds)
	updateSeconds, err := strconv.Atoi(updateSecondsString)
	if err != nil {
		fmt.Println(fmt.Errorf("update value %s cannot be converted to an integer", updateSecondsString))
		os.Exit(1)
	}
	return updateSeconds
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

func initialModel(addr, token string, updateSeconds int) app.Model {
	return app.InitialModel(Version, CommitSHA, addr, token, updateSeconds)
}

func getVersion() string {
	return Version
}
