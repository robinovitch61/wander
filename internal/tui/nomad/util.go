package nomad

import (
	"errors"
	"io/ioutil"
	"net/http"
)

func get(url, token string, params map[string]string) ([]byte, error) {
	if len(token) != 36 {
		return nil, errors.New("token must be 36 characters")
	}

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
	if string(body) == "ACL token not found" {
		return nil, errors.New("token not authorized")
	}
	return body, nil
}
