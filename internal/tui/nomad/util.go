package nomad

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const keySeparator = "|【=◈︿◈=】|"

type taskRowEntry struct {
	FullAllocationAsJSON                 string
	ID, TaskGroup, Name, TaskName, State string
	StartedAt, FinishedAt                time.Time
}

func tasksAsTable(taskRowEntries []taskRowEntry) ([]string, []page.Row) {
	var taskResponseRows [][]string
	var keys []string
	for _, row := range taskRowEntries {
		uptime := "-"
		if row.State == "running" {
			uptime = formatter.FormatTimeNsSinceNow(row.StartedAt.UnixNano())
		}
		taskResponseRows = append(taskResponseRows, []string{
			formatter.ShortAllocID(row.ID),
			row.TaskGroup,
			row.Name,
			row.TaskName,
			row.State,
			formatter.FormatTime(row.StartedAt),
			formatter.FormatTime(row.FinishedAt),
			uptime,
		})
		keys = append(keys, toTaskKey(row))
	}

	columns := []string{"Alloc ID", "Task Group", "Alloc Name", "Task Name", "State", "Started", "Finished", "Uptime"}
	table := formatter.GetRenderedTableAsString(columns, taskResponseRows)

	var rows []page.Row
	for idx, row := range table.ContentRows {
		rows = append(rows, page.Row{Key: keys[idx], Row: row})
	}

	return table.HeaderRows, rows
}

func toTaskKey(taskRowEntry taskRowEntry) string {
	isRunning := "false"
	if taskRowEntry.State == "running" {
		isRunning = "true"
	}
	return taskRowEntry.FullAllocationAsJSON + keySeparator + taskRowEntry.TaskName + keySeparator + isRunning
}

type TaskInfo struct {
	Alloc    api.Allocation
	TaskName string
	Running  bool
}

func TaskInfoFromKey(key string) (TaskInfo, error) {
	split := strings.Split(key, keySeparator)
	running, err := strconv.ParseBool(split[2])
	if err != nil {
		return TaskInfo{}, err
	}
	var alloc api.Allocation
	err = json.Unmarshal([]byte(split[0]), &alloc)
	if err != nil {
		return TaskInfo{}, err
	}
	return TaskInfo{Alloc: alloc, TaskName: split[1], Running: running}, nil
}

func getWebSocketConnection(secure bool, host, path, token string, params map[string]string) (*websocket.Conn, error) {
	urlParams := url.Values{}
	for k, v := range params {
		urlParams.Add(k, v)
	}

	scheme := "ws"
	if secure {
		scheme = "wss"
	}

	u := url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     path,
		RawQuery: urlParams.Encode(),
	}

	header := http.Header{}
	header.Add("X-Nomad-Token", token)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	return c, err
}

func PrettifyLine(l string, p Page) tea.Cmd {
	return func() tea.Msg {
		// nothing async actually happens here, but this fits the PageLoadedMsg pattern
		pretty := formatter.PrettyJsonStringAsLines(l)

		var rows []page.Row
		for _, row := range pretty {
			rows = append(rows, page.Row{Key: "", Row: row})
		}

		return PageLoadedMsg{
			Page:        p,
			TableHeader: []string{},
			AllPageRows: rows,
		}
	}
}
