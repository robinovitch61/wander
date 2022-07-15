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

func FetchEventsStream(url, token, topics, namespace string, page Page) tea.Cmd {
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

		return PageLoadedMsg{Page: page, Connection: EventStreamConnection{Reader: reader, Body: resp.Body, Topics: topics, Namespace: namespace}}
	}
}

func checkStreamReadError(e error) tea.Msg {
	// I've seen both 'use of closed network connection' and 'response body closed' here
	if strings.Contains(e.Error(), "closed") {
		return EventsStreamMsg{Closed: true}
	}
	return message.ErrMsg{Err: e}
}

func ReadEventsStreamNextMessage(c EventStreamConnection) tea.Cmd {
	return func() tea.Msg {
		peek, err := c.Reader.Peek(17)
		if err == nil && string(peek) == "Permission denied" {
			return message.ErrMsg{Err: fmt.Errorf("token not authorized to access topic set\n%s\nin namespace %s", formatEventTopics(c.Topics), c.Namespace)}
		} else if err != nil {
			return checkStreamReadError(err)
		}

		line, err := c.Reader.ReadBytes('\n')
		if err != nil {
			return checkStreamReadError(err)
		}
		trimmed := strings.TrimSpace(string(line))
		return EventsStreamMsg{Value: trimmed}
	}
}

func formatEventTopics(topics string) string {
	noSpaces := strings.ReplaceAll(topics, " ", "")
	return strings.ReplaceAll(noSpaces, ",", ", ")
}

func topicPrefixes(topics string) string {
	var p []string
	for _, v := range strings.Split(topics, ",") {
		p = append(p, strings.Split(v, ":")[0])
	}
	return strings.Join(p, ", ")
}

func TopicsForJob(topics, job string) string {
	var t []string
	for _, v := range strings.Split(topics, ",") {
		prefix := strings.Split(v, ":")[0]
		t = append(t, fmt.Sprintf("%s:%s", prefix, job))
	}
	return strings.Join(t, ",")
}
