package nomad

import (
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/robinovitch61/wander/internal/tui/components/page"
)

func TestJobSortValues(t *testing.T) {
	submitted := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC).UnixNano()
	job := &api.JobListStub{
		ID:         "web",
		Status:     "running",
		Priority:   50,
		SubmitTime: submitted,
		JobSummary: &api.JobSummary{Summary: map[string]api.TaskGroupSummary{
			"g": {Running: 2, Queued: 1},
		}},
		Meta: map[string]string{"team": "Payments", "replicas": "12"},
	}
	columns := []string{"Job", "Priority", "Count", "Submitted", "Since Submit", "team", "replicas", "absent"}
	got := getJobSortValuesFromColumns(job, columns, getJobRowFromColumns(job, columns))
	want := []page.SortValue{
		page.TextSortValue("web"),
		page.NumSortValue(50),
		page.NumSortValue(2, 3),
		page.NumSortValue(float64(submitted)),
		page.NumSortValue(-float64(submitted)),
		page.TextSortValue("payments"),
		page.TextSortValue("12"),
		page.MissingSortValue(),
	}
	if len(got) != len(want) {
		t.Fatalf("got %d values, want %d", len(got), len(want))
	}
	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("%s: got %+v, want %+v", columns[i], got[i], want[i])
		}
	}

	job.Status = "dead"
	since := []string{"Since Submit"}
	if v := getJobSortValuesFromColumns(job, since, getJobRowFromColumns(job, since))[0]; !v.Missing {
		t.Errorf("Since Submit of a dead job should be missing, got %+v", v)
	}
}

func TestSinceSubmitSortsByDuration(t *testing.T) {
	older := uptimeSortValue("running", time.Now().Add(-3*time.Hour).UnixNano())
	newer := uptimeSortValue("running", time.Now().Add(-time.Minute).UnixNano())
	if newer.Nums[0] >= older.Nums[0] {
		t.Errorf("a minute should sort before three hours")
	}
}

func TestTaskSortValues(t *testing.T) {
	started := time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)
	row := taskRowEntry{TaskName: "Main", State: "dead", StartedAt: started}
	columns := []string{"Task Name", "Started", "Finished", "Uptime"}
	got := getTaskSortValuesFromColumns(row, columns, getJobTaskRowFromColumns(row, columns))
	want := []page.SortValue{
		page.TextSortValue("main"),
		page.NumSortValue(float64(started.UnixNano())),
		page.MissingSortValue(),
		page.MissingSortValue(),
	}
	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("%s: got %+v, want %+v", columns[i], got[i], want[i])
		}
	}
}
