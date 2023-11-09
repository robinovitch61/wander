package cmd

import (
	"errors"
	"fmt"
	"github.com/carlmjohnson/versioninfo"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/hashicorp/nomad/api"
	"github.com/itchyny/gojq"
	"github.com/robinovitch61/wander/internal/dev"
	"github.com/robinovitch61/wander/internal/tui/components/app"
	"github.com/robinovitch61/wander/internal/tui/constants"
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
	// Version is public so users can optionally specify or override the version
	// at build time by passing in ldflags, e.g.
	//   go build -ldflags "-X github.com/robinovitch61/wander/cmd.Version=vX.Y.Z"
	// see https://github.com/robinovitch61/wander/issues/102
	Version = ""
)

func getVersion() string {
	if Version != "" {
		return Version
	}
	return versioninfo.Short()
}

func validateToken(token string) error {
	if len(token) > 0 && len(token) != 36 {
		return errors.New("token must be 36 characters")
	}
	return nil
}

func trueIfTrue(v string) bool {
	if strings.ToLower(strings.TrimSpace(v)) == "true" {
		return true
	}
	return false
}

// TODO LEO: rm
func retrieveWithDefault(cmd *cobra.Command, cliLong string, nameToArg map[string]arg, defaultVal string) string {
	val := cmd.Flag(cliLong).Value.String()
	if val == "" {
		// TODO LEO rm
		dev.Debug(fmt.Sprintf("No value for %s, checking env var %s", cliLong, nameToArg[cliLong].cfgFileEnvVar))
		val = viper.GetString(nameToArg[cliLong].cfgFileEnvVar)
	}
	if val == "" {
		return defaultVal
	}
	return val
}

func retrieveNonCLIWithDefault(a arg, defaultVal string) string {
	val := viper.GetString(a.cfgFileEnvVar)
	if val == "" {
		return defaultVal
	}
	return val
}

func retrieveAddress(cmd *cobra.Command) string {
	return cmd.Flag("addr").Value.String()
}

func retrieveToken(cmd *cobra.Command) string {
	val := cmd.Flag("token").Value.String()
	err := validateToken(val)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return val
}

func retrieveRegion(cmd *cobra.Command) string {
	return cmd.Flag("region").Value.String()
}

func retrieveNamespace(cmd *cobra.Command) string {
	return cmd.Flag("namespace").Value.String()
}

func retrieveHTTPAuth(cmd *cobra.Command) string {
	return cmd.Flag("http-auth").Value.String()
}

func retrieveCACert(cmd *cobra.Command) string {
	return cmd.Flag("cacert").Value.String()
}

func retrieveCAPath(cmd *cobra.Command) string {
	return cmd.Flag("capath").Value.String()
}

func retrieveClientCert(cmd *cobra.Command) string {
	return cmd.Flag("client-cert").Value.String()
}

func retrieveClientKey(cmd *cobra.Command) string {
	return cmd.Flag("client-key").Value.String()
}

func retrieveTLSServerName(cmd *cobra.Command) string {
	return cmd.Flag("tls-server-name").Value.String()
}

func retrieveSkipVerify(cmd *cobra.Command) bool {
	v := cmd.Flag("skip-verify").Value.String()
	return trueIfTrue(v)
}

func retrieveCopySavePath(cmd *cobra.Command) bool {
	v := cmd.Flag("copy-save-path").Value.String()
	return trueIfTrue(v)
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

	topicString := retrieveWithDefault(cmd, "event-topics", rootNameToArg, "Job,Allocation,Deployment,Evaluation")
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
	return retrieveWithDefault(cmd, "event-namespace", rootNameToArg, "default")
}

func retrieveEventJQQuery(cmd *cobra.Command) *gojq.Code {
	query := retrieveWithDefault(cmd, "event-jq-query", rootNameToArg, constants.DefaultEventJQQuery)
	parsed, err := gojq.Parse(query)
	if err != nil {
		fmt.Printf("Error parsing event jq query: %s\n", err.Error())
		os.Exit(1)
	}
	code, err := gojq.Compile(parsed)
	if err != nil {
		fmt.Printf("Error compiling event jq query: %s\n", err.Error())
		os.Exit(1)
	}
	return code
}

func retrieveUpdateSeconds(cmd *cobra.Command) int {
	updateSecondsString := retrieveWithDefault(cmd, "update", rootNameToArg, "2")
	updateSeconds, err := strconv.Atoi(updateSecondsString)
	if err != nil {
		fmt.Println(fmt.Errorf("update value %s cannot be converted to an integer", updateSecondsString))
		os.Exit(1)
	}
	return updateSeconds
}

func retrieveJobColumns(cmd *cobra.Command) []string {
	columnsString := retrieveWithDefault(cmd, "job-columns", rootNameToArg, "Job,Type,Namespace,Status,Count,Submitted,Since Submit")
	split := strings.Split(columnsString, ",")
	var trimmed []string
	for _, s := range split {
		trimmed = append(trimmed, strings.TrimSpace(s))
	}
	return trimmed
}

func retrieveAllTaskColumns(cmd *cobra.Command) []string {
	columnsString := retrieveWithDefault(cmd, "all-tasks-columns", rootNameToArg, "Job,Alloc ID,Task Group,Alloc Name,Task Name,State,Started,Finished,Uptime")
	split := strings.Split(columnsString, ",")
	var trimmed []string
	for _, s := range split {
		trimmed = append(trimmed, strings.TrimSpace(s))
	}
	return trimmed
}

func retrieveJobTaskColumns(cmd *cobra.Command) []string {
	columnsString := retrieveWithDefault(cmd, "tasks-for-job-columns", rootNameToArg, "Alloc ID,Task Group,Alloc Name,Task Name,State,Started,Finished,Uptime")
	split := strings.Split(columnsString, ",")
	var trimmed []string
	for _, s := range split {
		trimmed = append(trimmed, strings.TrimSpace(s))
	}
	return trimmed
}

func retrieveLogOffset(cmd *cobra.Command) int {
	logOffsetString := retrieveWithDefault(cmd, "log-offset", rootNameToArg, "1000000")
	logOffset, err := strconv.Atoi(logOffsetString)
	if err != nil {
		fmt.Println(fmt.Errorf("log offset %s cannot be converted to an integer", logOffsetString))
		os.Exit(1)
	}
	return logOffset
}

func retrieveLogTail(cmd *cobra.Command) bool {
	v := retrieveWithDefault(cmd, "log-tail", rootNameToArg, "true")
	return trueIfTrue(v)
}

func retrieveStartCompact(cmd *cobra.Command) bool {
	v := retrieveWithDefault(cmd, "compact-header", rootNameToArg, "false")
	return trueIfTrue(v)
}

func retrieveStartAllTasksView(cmd *cobra.Command) bool {
	v := retrieveWithDefault(cmd, "start-all-tasks", rootNameToArg, "false")
	return trueIfTrue(v)
}

func retrieveCompactTables(cmd *cobra.Command) bool {
	v := retrieveWithDefault(cmd, "compact-tables", rootNameToArg, "true")
	return trueIfTrue(v)
}

// customLoggingMiddleware provides basic connection logging. Connects are logged with the
// remote address, invoked command, TERM setting, window dimensions and if the
// auth was public key based. Disconnect will log the remote address and
// connection duration. It is custom because it excludes the ssh Command in the log.
func customLoggingMiddleware() wish.Middleware {
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
	region := retrieveRegion(cmd)
	namespace := retrieveNamespace(cmd)
	httpAuth := retrieveHTTPAuth(cmd)
	cacert := retrieveCACert(cmd)
	capath := retrieveCAPath(cmd)
	clientCert := retrieveClientCert(cmd)
	clientKey := retrieveClientKey(cmd)
	tlsServerName := retrieveTLSServerName(cmd)
	skipVerify := retrieveSkipVerify(cmd)
	logOffset := retrieveLogOffset(cmd)
	logTail := retrieveLogTail(cmd)
	copySavePath := retrieveCopySavePath(cmd)
	eventTopics := retrieveEventTopics(cmd)
	eventNamespace := retrieveEventNamespace(cmd)
	eventJQQuery := retrieveEventJQQuery(cmd)
	updateSeconds := retrieveUpdateSeconds(cmd)
	jobColumns := retrieveJobColumns(cmd)
	allTaskColumns := retrieveAllTaskColumns(cmd)
	jobTaskColumns := retrieveJobTaskColumns(cmd)
	logoColor := retrieveNonCLIWithDefault(rootNameToArg["logo-color"], "")
	startCompact := retrieveStartCompact(cmd)
	startAllTasksView := retrieveStartAllTasksView(cmd)
	compactTables := retrieveCompactTables(cmd)

	initialModel := app.InitialModel(app.Config{
		Version:   getVersion(),
		URL:       nomadAddr,
		Token:     nomadToken,
		Region:    region,
		Namespace: namespace,
		HTTPAuth:  httpAuth,
		TLS: app.TLSConfig{
			CACert:     cacert,
			CAPath:     capath,
			ClientCert: clientCert,
			ClientKey:  clientKey,
			ServerName: tlsServerName,
			SkipVerify: skipVerify,
		},
		Log: app.LogConfig{
			Offset: logOffset,
			Tail:   logTail,
		},
		CopySavePath: copySavePath,
		Event: app.EventConfig{
			Topics:    eventTopics,
			Namespace: eventNamespace,
			JQQuery:   eventJQQuery,
		},
		UpdateSeconds:     time.Second * time.Duration(updateSeconds),
		JobColumns:        jobColumns,
		AllTaskColumns:    allTaskColumns,
		JobTaskColumns:    jobTaskColumns,
		LogoColor:         logoColor,
		StartCompact:      startCompact,
		StartAllTasksView: startAllTasksView,
		CompactTables:     compactTables,
	})
	return initialModel, []tea.ProgramOption{tea.WithAltScreen()}
}
