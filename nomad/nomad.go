package nomad

import "fmt"

var ApiPaths = map[string]string{
	"jobs":        "/v1/jobs",
	"allocations": "/v1/job/%s/allocations",
	"logs":        "/v1/client/fs/logs/%s",
}

func GetAllocations(url, token, jobId string) ([]byte, error) {
	path, err := urlWithPathFor(url, "allocations")
	if err != nil {
		return nil, err
	}
	pathWithAllocId := fmt.Sprintf(path, jobId)
	return Get(pathWithAllocId, token, nil)
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
