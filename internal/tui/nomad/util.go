package nomad

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
	"github.com/robinovitch61/wander/internal/tui/components/page"
	"github.com/robinovitch61/wander/internal/tui/formatter"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const keySeparator = "|【=◈︿◈=】|"

func doQuery(url, token string, params [][2]string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Nomad-Token", token)

	query := req.URL.Query()
	for _, p := range params {
		query.Add(p[0], p[1])
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func get(url, token string, params [][2]string) ([]byte, error) {
	resp, err := doQuery(url, token, params)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if string(body) == "ACL token not found" {
		return nil, errors.New("token not authorized")
	}
	return body, nil
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

func toIDNamespaceKey(id, namespace string) string {
	return id + keySeparator + namespace
}

func JobIDAndNamespaceFromKey(key string) (string, string) {
	split := strings.Split(key, keySeparator)
	return split[0], split[1]
}
