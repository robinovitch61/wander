package cmd

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
	lm "github.com/charmbracelet/wish/logging"
	"github.com/gliderlabs/ssh"
	"github.com/robinovitch61/wander/internal/tui/components/app"
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
		config:   "wander_ssh_host",
	}
	portArg = arg{
		cliShort: "p",
		cliLong:  "port",
		config:   "wander_ssh_port",
	}

	sshDescription = `wander ssh starts an ssh server that serves wander.`

	sshCmd = &cobra.Command{
		Use:   "ssh",
		Short: "Start ssh server for wander.",
		Long:  sshDescription,
		Run:   sshEntrypoint,
	}
)

func sshEntrypoint(cmd *cobra.Command, args []string) {
	host := retrieveAssertExists(cmd, hostArg.cliLong, hostArg.config)
	portStr := retrieveAssertExists(cmd, portArg.cliLong, portArg.config)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	s, err := wish.NewServer(
		wish.WithAddress(fmt.Sprintf("%s:%d", host, port)),
		wish.WithMiddleware(
			bm.Middleware(generateTeaHandler(cmd)),
			lm.Middleware(),
		),
	)
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
		nomadAddr := retrieveAssertExists(cmd, addrArg.cliLong, addrArg.config)
		nomadToken := retrieveAssertExists(cmd, tokenArg.cliLong, tokenArg.config)
		if sshCommands := s.Command(); len(sshCommands) == 1 {
			nomadToken = strings.TrimSpace(sshCommands[0])
		}
		fmt.Printf("cmd %s addr %s token %s", s.Command(), nomadAddr, nomadToken)

		return app.InitialModel(nomadAddr, nomadToken), []tea.ProgramOption{tea.WithAltScreen()}
	}
}
