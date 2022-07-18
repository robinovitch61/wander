package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	"github.com/gliderlabs/ssh"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/app"
	"github.com/robinovitch61/wander/internal/tui/nomad"
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

func validateToken(token string) error {
	// TODO LEO: uncomment
	// if len(token) > 0 && len(token) != 36 {
	// 	return errors.New("token must be 36 characters")
	// }
	return nil
}

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

func retrieveWithFallback(cmd *cobra.Command, currArg, oldArg arg) (string, error) {
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

func retrieveWithDefault(cmd *cobra.Command, a arg, defaultVal string) string {
	val := cmd.Flag(a.cliLong).Value.String()
	if val == "" {
		val = viper.GetString(a.config)
	}
	if val == "" {
		return defaultVal
	}
	return val
}

func retrieveNonCLIWithDefault(a arg, defaultVal string) string {
	val := viper.GetString(a.config)
	if val == "" {
		return defaultVal
	}
	return val
}

func retrieveAddress(cmd *cobra.Command) string {
	val, err := retrieveWithFallback(cmd, addrArg, oldAddrArg)
	if err != nil {
		return "http://localhost:4646"
	}
	return val
}

func retrieveToken(cmd *cobra.Command) string {
	val, err := retrieveWithFallback(cmd, tokenArg, oldTokenArg)
	if err != nil {
		return ""
	}
	err = validateToken(val)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return val
}

func retrieveCopySavePath(cmd *cobra.Command) bool {
	v := retrieveWithDefault(cmd, copySavePathArg, "false")
	if strings.ToLower(strings.TrimSpace(v)) == "true" {
		return true
	}
	return false
}

func retrieveEventTopics(cmd *cobra.Command) nomad.Topics {
	matchTopic := func(t string) (api.Topic, error) {
		switch t {
		case "Deployment":
			return api.TopicDeployment, nil
		case "Evaluation":
			return api.TopicEvaluation, nil
		case "Allocation":
			return api.TopicAllocation, nil
		case "Job":
			return api.TopicJob, nil
		case "Node":
			return api.TopicNode, nil
		case "Service":
			return api.TopicService, nil
		case "*":
			return api.TopicAll, nil
		}
		return "", fmt.Errorf("%s cannot be parsed into topic", t)
	}

	topicString := retrieveWithDefault(cmd, eventTopicsArg, "Job,Allocation,Deployment,Evaluation")
	topics := make(nomad.Topics)
	for _, t := range strings.Split(topicString, ",") {
		split := strings.Split(strings.TrimSpace(t), ":")
		suffix := "*"
		if len(split) == 2 {
			suffix = strings.TrimSpace(split[1])
		}

		topic, err := matchTopic(strings.TrimSpace(split[0]))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		suffixes, exists := topics[topic]
		if exists {
			topics[topic] = append(suffixes, suffix)
		} else {
			topics[topic] = []string{suffix}
		}
	}

	return topics
}

func retrieveEventNamespace(cmd *cobra.Command) string {
	return retrieveWithDefault(cmd, eventNamespaceArg, "default")
}

func retrieveUpdateSeconds(cmd *cobra.Command) int {
	updateSecondsString := retrieveWithDefault(cmd, updateSecondsArg, "2")
	updateSeconds, err := strconv.Atoi(updateSecondsString)
	if err != nil {
		fmt.Println(fmt.Errorf("update value %s cannot be converted to an integer", updateSecondsString))
		os.Exit(1)
	}
	return updateSeconds
}

func retrieveLogOffset(cmd *cobra.Command) int {
	logOffsetString := retrieveWithDefault(cmd, logOffsetArg, "1000000")
	logOffset, err := strconv.Atoi(logOffsetString)
	if err != nil {
		fmt.Println(fmt.Errorf("log offset %s cannot be converted to an integer", logOffsetString))
		os.Exit(1)
	}
	return logOffset
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

func setup(cmd *cobra.Command, overrideToken string) (app.Model, []tea.ProgramOption) {
	nomadAddr := retrieveAddress(cmd)
	nomadToken := retrieveToken(cmd)
	if overrideToken != "" {
		err := validateToken(overrideToken)
		if err != nil {
			fmt.Println(err.Error())
		}
		nomadToken = overrideToken
	}
	logOffset := retrieveLogOffset(cmd)
	copySavePath := retrieveCopySavePath(cmd)
	eventTopics := retrieveEventTopics(cmd)
	eventNamespace := retrieveEventNamespace(cmd)
	updateSeconds := retrieveUpdateSeconds(cmd)
	logoColor := retrieveNonCLIWithDefault(logoColorArg, "")

	initialModel := app.InitialModel(app.Config{
		Version:        Version,
		SHA:            CommitSHA,
		URL:            nomadAddr,
		Token:          nomadToken,
		LogOffset:      logOffset,
		CopySavePath:   copySavePath,
		EventTopics:    eventTopics,
		EventNamespace: eventNamespace,
		UpdateSeconds:  time.Second * time.Duration(updateSeconds),
		LogoColor:      logoColor,
	})
	return initialModel, []tea.ProgramOption{tea.WithAltScreen()}
}

func getVersion() string {
	return Version
}
