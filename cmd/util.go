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
	"github.com/robinovitch61/wander/internal/tui/components/app"
	"github.com/robinovitch61/wander/internal/tui/nomad"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

func retrieveLogoColor() string {
	return viper.GetString(rootNameToArg["logo-color"].cfgFileEnvVar)
}

func retrieveAddress(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("addr").Value.String()
}

func retrieveToken(cmd *cobra.Command) string {
	val := cmd.Flags().Lookup("token").Value.String()
	if err := validateToken(val); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return val
}

func retrieveRegion(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("region").Value.String()
}

func retrieveNamespace(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("namespace").Value.String()
}

func retrieveHTTPAuth(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("http-auth").Value.String()
}

func retrieveCACert(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("cacert").Value.String()
}

func retrieveCAPath(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("capath").Value.String()
}

func retrieveClientCert(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("client-cert").Value.String()
}

func retrieveClientKey(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("client-key").Value.String()
}

func retrieveTLSServerName(cmd *cobra.Command) string {
	return cmd.Flags().Lookup("tls-server-name").Value.String()
}

func retrieveSkipVerify(cmd *cobra.Command) bool {
	v := cmd.Flags().Lookup("skip-verify").Value.String()
	return trueIfTrue(v)
}

func retrieveCopySavePath(cmd *cobra.Command) bool {
	v := cmd.Flags().Lookup("copy-save-path").Value.String()
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

	topicString := cmd.Flags().Lookup("event-topics").Value.String()
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
	return cmd.Flags().Lookup("event-namespace").Value.String()
}

func retrieveEventJQQuery(cmd *cobra.Command) *gojq.Code {
	query := cmd.Flags().Lookup("event-jq-query").Value.String()
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

func retrieveAllocEventJQQuery(cmd *cobra.Command) *gojq.Code {
	query := cmd.Flags().Lookup("alloc-event-jq-query").Value.String()
	parsed, err := gojq.Parse(query)
	if err != nil {
		fmt.Printf("Error parsing alloc event jq query: %s\n", err.Error())
		os.Exit(1)
	}
	code, err := gojq.Compile(parsed)
	if err != nil {
		fmt.Printf("Error compiling alloc event jq query: %s\n", err.Error())
		os.Exit(1)
	}
	return code
}

func retrieveUpdateSeconds(cmd *cobra.Command) int {
	updateSecondsString := cmd.Flags().Lookup("update").Value.String()
	updateSeconds, err := strconv.Atoi(updateSecondsString)
	if err != nil {
		fmt.Println(fmt.Errorf("update value %s cannot be converted to an integer", updateSecondsString))
		os.Exit(1)
	}
	return updateSeconds
}

func retrieveAllocColumns(cmd *cobra.Command) []string {
	columnsString := cmd.Flags().Lookup("alloc-columns").Value.String()

	split := strings.Split(columnsString, ",")
	var trimmed []string
	for _, s := range split {
		trimmed = append(trimmed, strings.TrimSpace(s))
	}
	return trimmed
}

func retrieveJobColumns(cmd *cobra.Command) []string {
	columnsString := cmd.Flags().Lookup("job-columns").Value.String()
	split := strings.Split(columnsString, ",")
	var trimmed []string
	for _, s := range split {
		trimmed = append(trimmed, strings.TrimSpace(s))
	}
	return trimmed
}

func retrieveAllTaskColumns(cmd *cobra.Command) []string {
	columnsString := cmd.Flags().Lookup("all-tasks-columns").Value.String()
	split := strings.Split(columnsString, ",")
	var trimmed []string
	for _, s := range split {
		trimmed = append(trimmed, strings.TrimSpace(s))
	}
	return trimmed
}

func retrieveJobTaskColumns(cmd *cobra.Command) []string {
	columnsString := cmd.Flags().Lookup("tasks-for-job-columns").Value.String()
	split := strings.Split(columnsString, ",")
	var trimmed []string
	for _, s := range split {
		trimmed = append(trimmed, strings.TrimSpace(s))
	}
	return trimmed
}

func retrieveLogOffset(cmd *cobra.Command) int {
	logOffsetString := cmd.Flags().Lookup("log-offset").Value.String()
	logOffset, err := strconv.Atoi(logOffsetString)
	if err != nil {
		fmt.Println(fmt.Errorf("log offset %s cannot be converted to an integer", logOffsetString))
		os.Exit(1)
	}
	return logOffset
}

func retrieveLogTail(cmd *cobra.Command) bool {
	v := cmd.Flags().Lookup("log-tail").Value.String()
	return trueIfTrue(v)
}

func retrieveStartCompact(cmd *cobra.Command) bool {
	v := cmd.Flags().Lookup("compact-header").Value.String()
	return trueIfTrue(v)
}

func retrieveStartAllTasksView(cmd *cobra.Command) bool {
	v := cmd.Flags().Lookup("start-all-tasks").Value.String()
	return trueIfTrue(v)
}

func retrieveCompactTables(cmd *cobra.Command) bool {
	v := cmd.Flags().Lookup("compact-tables").Value.String()
	return trueIfTrue(v)
}

func retrieveStartFiltering(cmd *cobra.Command) bool {
	v := cmd.Flags().Lookup("start-filtering").Value.String()
	return trueIfTrue(v)
}

func retrieveFilterWithContext(cmd *cobra.Command) bool {
	v := cmd.Flags().Lookup("filter-with-context").Value.String()
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

func getConfig(cmd *cobra.Command, rootOpts []string, overrideToken string) app.Config {
	nomadAddr := retrieveAddress(cmd)
	nomadToken := retrieveToken(cmd)
	if overrideToken != "" {
		if err := validateToken(overrideToken); err != nil {
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
	allocEventJQQuery := retrieveAllocEventJQQuery(cmd)
	updateSeconds := retrieveUpdateSeconds(cmd)
	allocColumns := retrieveAllocColumns(cmd)
	jobColumns := retrieveJobColumns(cmd)
	allTaskColumns := retrieveAllTaskColumns(cmd)
	jobTaskColumns := retrieveJobTaskColumns(cmd)
	logoColor := retrieveLogoColor()
	startCompact := retrieveStartCompact(cmd)
	startAllTasksView := retrieveStartAllTasksView(cmd)
	compactTables := retrieveCompactTables(cmd)
	startFiltering := retrieveStartFiltering(cmd)
	filterWithContext := retrieveFilterWithContext(cmd)

	return app.Config{
		RootOpts:  rootOpts,
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
			Topics:       eventTopics,
			Namespace:    eventNamespace,
			JQQuery:      eventJQQuery,
			AllocJQQuery: allocEventJQQuery,
		},
		UpdateSeconds:     time.Second * time.Duration(updateSeconds),
		JobColumns:        jobColumns,
		AllocColumns:      allocColumns,
		AllTaskColumns:    allTaskColumns,
		JobTaskColumns:    jobTaskColumns,
		LogoColor:         logoColor,
		StartCompact:      startCompact,
		StartAllTasksView: startAllTasksView,
		CompactTables:     compactTables,
		StartFiltering:    startFiltering,
		FilterWithContext: filterWithContext,
	}
}

func getRootOpts(cmd *cobra.Command) []string {
	if cmd.Name() != "wander" {
		panic("getRootOpts should only be called on the root wander command, for which both serve and exec are subcommands")
	}
	var opts []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			if f.Name != "" {
				opts = append(opts, fmt.Sprintf("--%s=%s", f.Name, f.Value))
			} else if f.Shorthand != "" {
				opts = append(opts, fmt.Sprintf("-%s=%s", f.Shorthand, f.Value))
			}
		}
	})
	return opts
}

func setup(cmd *cobra.Command, rootOpts []string, overrideToken string) (app.Model, []tea.ProgramOption) {
	initialModel := app.InitialModel(getConfig(cmd, rootOpts, overrideToken))
	return initialModel, []tea.ProgramOption{tea.WithAltScreen()}
}
