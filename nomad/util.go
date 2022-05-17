package nomad

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func Get(url, token string, params map[string]string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
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

func urlWithPathFor(url, key string) (string, error) {
	if val, exists := ApiPaths[key]; exists {
		return fmt.Sprintf("%s/%s", url, val), nil
	}
	return "", fmt.Errorf("key '%s' has no associated path", key)
}
