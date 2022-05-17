package nomad

import "fmt"

var ApiPaths = map[string]string{
	"logs": "/v1/client/fs/logs/%s",
}

func GetLogs(url, token, allocId, taskName, logType string) ([]byte, error) {
	path, err := urlWithPathFor(url, "logs")
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"task":   taskName,
		"type":   logType,
		"origin": "end",
		"offset": "1000000",
		"plain":  "true",
	}
	pathWithAllocId := fmt.Sprintf(path, allocId)
	return Get(pathWithAllocId, token, params)
}
