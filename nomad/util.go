package nomad

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"net/url"
	"wander/dev"
)

func get(fullPath, token string, params map[string]string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fullPath, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Nomad-Token", token)

	query := req.URL.Query()
	for key, val := range params {
		query.Add(key, val)
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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

	dev.Debug(u.String())
	c, resp, err := websocket.DefaultDialer.Dial(u.String(), header)
	dev.Debug(fmt.Sprintf("STATUS CODE %d", resp.StatusCode))
	return c, err
}
