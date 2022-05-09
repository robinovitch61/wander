package nomad

import "fmt"

var ApiPaths = map[string]string{
	"jobs":        "/v1/jobs",
	"allocations": "/v1/job/%s/allocations",
}

func GetJobs(url, token string) ([]byte, error) {
	path, err := urlWithPathFor(url, "jobs")
	if err != nil {
		return nil, err
	}
	return get(path, token)
}

func GetAllocations(url, token, jobId string) ([]byte, error) {
	path, err := urlWithPathFor(url, "allocations")
	if err != nil {
		return nil, err
	}
	pathWithAllocId := fmt.Sprintf(path, jobId)
	return get(pathWithAllocId, token)
}
