package nomad

import "fmt"

var ApiPaths = map[string]string{
	"jobs":        "/v1/jobs",
	"allocations": "/v1/job/%s/allocations",
	"logs":        "/v1/client/fs/logs/%s",
}

func GetJobs(url, token string) ([]byte, error) {
	path, err := urlWithPathFor(url, "jobs")
	if err != nil {
		return nil, err
	}
	return get(path, token, nil)
}

func GetAllocations(url, token, jobId string) ([]byte, error) {
	path, err := urlWithPathFor(url, "allocations")
	if err != nil {
		return nil, err
	}
	pathWithAllocId := fmt.Sprintf(path, jobId)
	return get(pathWithAllocId, token, nil)
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
		"offset": "100000",
		"plain":  "true",
	}
	pathWithAllocId := fmt.Sprintf(path, allocId)
	return get(pathWithAllocId, token, params)
}
