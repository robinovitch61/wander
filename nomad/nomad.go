package nomad

import "fmt"

var ApiPaths = map[string]string{
	"jobs":       "/v1/jobs",
	"allocation": "/v1/allocation/%s",
}

func GetJobs(url, token string) ([]byte, error) {
	path, err := urlWithPathFor(url, "jobs")
	if err != nil {
		return nil, err
	}
	return get(path, token)
}

func GetAllocation(url, token, allocId string) ([]byte, error) {
	path, err := urlWithPathFor(url, "allocation")
	if err != nil {
		return nil, err
	}
	pathWithAllocId := fmt.Sprintf(path, allocId)
	return get(pathWithAllocId, token)
}
