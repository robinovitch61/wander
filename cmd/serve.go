package cmd

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	"github.com/gliderlabs/ssh"
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
	hostArg = arg{
		cliShort: "h",
		cliLong:  "host",
		config:   "wander_host",
	}
	portArg = arg{
		cliShort: "p",
		cliLong:  "port",
		config:   "wander_port",
	}
	hostKeyPathArg = arg{
		cliShort: "k",
		cliLong:  "host-key-path",
		config:   "wander_host_key_path",
	}
	hostKeyPEMArg = arg{
		cliShort: "m",
		cliLong:  "host-key-pem",
		config:   "wander_host_key_pem",
	}

	serveDescription = `Starts an ssh server hosting wander.`

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start ssh server for wander",
		Long:  serveDescription,
		Run:   serveEntrypoint,
	}
)

func serveEntrypoint(cmd *cobra.Command, args []string) {
	host := retrieveWithDefault(cmd, hostArg, "localhost")
	portStr := retrieveWithDefault(cmd, portArg, "21324")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		fmt.Println(fmt.Errorf("could not convert %s to integer", portStr))
		os.Exit(1)
	}
	hostKeyPath := retrieveWithDefault(cmd, hostKeyPathArg, "")
	hostKeyPEM := retrieveWithDefault(cmd, hostKeyPEMArg, "")

	options := []ssh.Option{wish.WithAddress(fmt.Sprintf("%s:%d", host, port))}
	if hostKeyPath != "" {
		options = append(options, wish.WithHostKeyPath(hostKeyPath))
	}
	if hostKeyPEM != "" {
		options = append(options, wish.WithHostKeyPEM([]byte(hostKeyPEM)))
	}
	middleware := wish.WithMiddleware(
		bm.Middleware(generateTeaHandler(cmd)),
		CustomLoggingMiddleware(),
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
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		// optionally override token - MUST run with `-t` flag to force pty, e.g. ssh -p 20000 localhost -t <token>
		var overrideToken string
		if sshCommands := s.Command(); len(sshCommands) == 1 {
			overrideToken = strings.TrimSpace(sshCommands[0])
		}
		return setup(cmd, overrideToken)
	}
}
