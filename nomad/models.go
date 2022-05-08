package nomad

type JobResponseEntry struct {
	ID string
	Type string
	Priority int
	Status string
	SubmitTime int
}
