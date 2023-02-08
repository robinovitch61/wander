package nomad

import (
	"context"
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/hashicorp/nomad/api"
	"github.com/itchyny/gojq"
	"github.com/robinovitch61/wander/internal/tui/message"
	"strings"
)

type Topics map[api.Topic][]string

type EventsStreamMsg struct {
	CompleteValue string
	JQValue       string
	Topics        Topics
}

func FetchEventsStream(client api.Client, topics Topics, namespace string, page Page) tea.Cmd {
	return func() tea.Msg {
		eventsChan, err := client.EventStream().Stream(context.Background(), topics, 0, &api.QueryOptions{Namespace: namespace})
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		return PageLoadedMsg{Page: page, EventsStream: EventsStream{Chan: eventsChan, Topics: topics, Namespace: namespace}}
	}
}

func ReadEventsStreamNextMessage(c EventsStream, code *gojq.Code) tea.Cmd {
	return func() tea.Msg {
		line := <-c.Chan
		lineBytes, err := json.Marshal(line)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		trimmed := strings.TrimSpace(string(lineBytes))
		jq, err := runJQQueryOnEvent(trimmed, code)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		return EventsStreamMsg{CompleteValue: trimmed, JQValue: jq, Topics: c.Topics}
	}
}

func formatEventTopics(topics Topics) string {
	t := ""
	for k, v := range topics {
		t += fmt.Sprintf("%s:[%s], ", string(k), strings.Join(v, ","))
	}
	return t[:len(t)-2]
}

func getTopicNames(topics Topics) string {
	t := ""
	for k := range topics {
		t += fmt.Sprintf("%s, ", k)
	}
	return t[:len(t)-2]
}

func TopicsForJob(topics Topics, job string) Topics {
	t := make(Topics)
	for k := range topics {
		t[k] = []string{job}
	}
	return t
}

func TopicsForAlloc(topics Topics, allocID string) Topics {
	t := make(Topics)
	for k := range topics {
		t[k] = []string{allocID}
	}
	return t
}

func runJQQueryOnEvent(event string, code *gojq.Code) (string, error) {
	result := make(map[string]interface{})
	err := json.Unmarshal([]byte(event), &result)
	if err != nil {
		return "", err
	}
	iter := code.Run(result)
	v, ok := iter.Next()
	if !ok {
		return event, nil
	}
	if err, ok := v.(error); ok {
		return "", err
	}
	j, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s", j), nil
}
