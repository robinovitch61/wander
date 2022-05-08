package nomad

var ApiPaths = map[string]string{
	"jobs": "/v1/jobs",
}

func GetJobs(url, token string) ([]byte, error) {
	path, err := urlWithPathFor(url, "jobs")
	if err != nil {
		return nil, err
	}
	return get(path, token)
}
