package nomad

import (
	"bufio"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/robinovitch61/wander/internal/tui/message"
	"strings"
)

type EventsStreamMsg struct {
	Value  string
	Closed bool
}

func FetchEventsStream(url, token, topics, namespace string) tea.Cmd {
	return func() tea.Msg {
		params := [][2]string{
			{"namespace", namespace},
		}
		for _, t := range strings.Split(topics, ",") {
			params = append(params, [2]string{"topic", strings.TrimSpace(t)})
		}
		fullPath := fmt.Sprintf("%s%s", url, "/v1/event/stream")
		resp, err := doQuery(fullPath, token, params)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		reader := bufio.NewReader(resp.Body)

		peek, err := reader.Peek(17)
		if err == nil && string(peek) == "Permission denied" {
			return message.ErrMsg{Err: fmt.Errorf("token not authorized to access topic set\n%s\nin namespace %s", formatEventTopics(topics), namespace)}
		}

		return PageLoadedMsg{Page: EventsPage, Connection: PersistentConnection{Reader: reader, Body: resp.Body}}
	}
}

func ReadEventsStreamNextMessage(r *bufio.Reader) tea.Cmd {
	return func() tea.Msg {
		line, err := r.ReadBytes('\n')
		if err != nil {
			// I've seen both 'use of closed network connection' and 'response body closed' here
			if strings.Contains(err.Error(), "closed") {
				return EventsStreamMsg{Closed: true}
			}
			return message.ErrMsg{Err: err}
		}
		trimmed := strings.TrimSpace(string(line))
		return EventsStreamMsg{Value: trimmed}
	}
}

func formatEventTopics(topics string) string {
	noSpaces := strings.ReplaceAll(topics, " ", "")
	return strings.ReplaceAll(noSpaces, ",", ", ")
}
