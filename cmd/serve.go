package cmd

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	serveNameToArg = map[string]arg{
		"host": {
			cliShort:      "h",
			cfgFileEnvVar: "wander_host",
			description:   `Host for wander ssh server`,
			defaultString: "localhost",
		},
		"port": {
			cliShort:      "p",
			cfgFileEnvVar: "wander_port",
			description:   `Port for wander ssh server`,
			isInt:         true,
			defaultIfInt:  21324,
		},
		"host-key-path": {
			cliShort:      "k",
			cfgFileEnvVar: "wander_host_key_path",
			description:   `Host key path for wander ssh server`,
		},
		"host-key-pem": {
			cliShort:      "m",
			cfgFileEnvVar: "wander_host_key_pem",
			description:   `Host key PEM block for wander ssh server`,
		},
	}

	serveDescription = `Starts an ssh server hosting wander.`

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start ssh server for wander",
		Long:  serveDescription,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(cmd, serveNameToArg)
		},
		Run: serveEntrypoint,
	}
)

func serveEntrypoint(cmd *cobra.Command, args []string) {
	host := cmd.Flags().Lookup("host").Value.String()
	portStr := cmd.Flags().Lookup("port").Value.String()
	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println(fmt.Errorf("could not convert %s to integer", portStr))
		os.Exit(1)
	}
	hostKeyPath := cmd.Flags().Lookup("host-key-path").Value.String()
	hostKeyPEM := cmd.Flags().Lookup("host-key-pem").Value.String()

	options := []ssh.Option{wish.WithAddress(fmt.Sprintf("%s:%d", host, port))}
	if hostKeyPath != "" {
		options = append(options, wish.WithHostKeyPath(hostKeyPath))
	}
	if hostKeyPEM != "" {
		options = append(options, wish.WithHostKeyPEM([]byte(hostKeyPEM)))
	}
	middleware := wish.WithMiddleware(
		bm.Middleware(generateTeaHandler(cmd)),
		customLoggingMiddleware(),
	)
	options = append(options, middleware)

	s, err := wish.NewServer(options...)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s:%d", host, port)
	go func() {
		if err = s.ListenAndServe(); err != nil {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}

func generateTeaHandler(cmd *cobra.Command) func(ssh.Session) (tea.Model, []tea.ProgramOption) {
	changedOpts := getRootOpts(cmd.Parent())
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		// optionally override token - MUST run with `-t` flag to force pty, e.g. ssh -p 20000 localhost -t <token>
		var overrideToken string
		if sshCommands := s.Command(); len(sshCommands) == 1 {
			overrideToken = strings.TrimSpace(sshCommands[0])
		}
		return setup(cmd, changedOpts, overrideToken)
	}
}
