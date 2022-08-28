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
	Value  string
	Topics Topics
}

func FetchEventsStream(client api.Client, topics Topics, namespace string, page Page) tea.Cmd {
	return func() tea.Msg {
		eventsChan, err := client.EventStream().Stream(context.Background(), topics, 0, &api.QueryOptions{Namespace: namespace})
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		return PageLoadedMsg{Page: page, Connection: EventsStream{Chan: eventsChan, Topics: topics, Namespace: namespace}}
	}
}

func ReadEventsStreamNextMessage(c EventsStream) tea.Cmd {
	return func() tea.Msg {
		line := <-c.Chan
		lineBytes, err := json.Marshal(line)
		if err != nil {
			return message.ErrMsg{Err: err}
		}
		trimmed := strings.TrimSpace(string(lineBytes))
		return EventsStreamMsg{Value: formatEvent(trimmed), Topics: c.Topics}
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

func formatEvent(event string) string {
	query, err := gojq.Parse(`.Events[] | {"1:Index": .Index, "2:Topic": .Topic, "3:Type": .Type, "4:Name": .Payload | (.Job // .Allocation // .Deployment // .Evaluation) | (.JobID // .ID), "5:AllocID": .Payload | (.Allocation // .Deployment // .Evaluation).ID[:8]}`)
	if err != nil {
		return event
	}
	code, err := gojq.Compile(query)
	if err != nil {
		return event
	}
	result := make(map[string]interface{})
	json.Unmarshal([]byte(event), &result)
	iter := code.Run(result)
	v, ok := iter.Next()
	if !ok {
		return event
	}
	if _, ok := v.(error); ok {
		return event
	}
	j, _ := json.Marshal(v)
	return fmt.Sprintf("%s", j)
}
