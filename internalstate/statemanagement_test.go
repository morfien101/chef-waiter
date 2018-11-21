package internalstate

import (
	"testing"

	"github.com/morfien101/chef-waiter/logs"
	uuid "github.com/satori/go.uuid"
)

func TestGetOldStates(t *testing.T) {
	st := &StateTable{
		StateTableSize: 10,
		logger:         logs.NewFakeLogger(true),
	}

	runs := make(map[string]int64)
	var theLarge int64 = 11
	var i int64

	for i = 1; i <= theLarge; i++ {
		runs[uuid.NewV4().String()] = int64(i)
	}
	delSlice := st.GetOldStates(runs)

	foundTheLarge := false

	for _, guid := range delSlice {
		if _, ok := runs[guid]; ok {
			if runs[guid] == theLarge {
				foundTheLarge = true
			}

			t.Logf("%s: %d\n", guid, runs[guid])
		}
	}
	if foundTheLarge {
		t.Fail()
	}
}

func TestReadAllJobs(t *testing.T) {
	st := &StateTable{
		Status: map[string]*JobDetails{
			"1": &JobDetails{
				Status:          "running",
				ExitCode:        0,
				RegisteredTime:  1,
				OnDemand:        false,
				CustomRun:       false,
				CustomRunString: "",
			},
			"2": &JobDetails{
				Status:          "registered",
				ExitCode:        0,
				RegisteredTime:  1,
				OnDemand:        true,
				CustomRun:       false,
				CustomRunString: "",
			},
			"3": &JobDetails{
				Status:          "registered",
				ExitCode:        0,
				RegisteredTime:  1,
				OnDemand:        true,
				CustomRun:       true,
				CustomRunString: "test::run",
			},
		},
	}

	jobs := st.ReadAllJobs()

	if len(jobs) != len(st.Status) {
		t.Logf("Failed to copy all jobs. Wanted: %d, got %d", len(st.Status), len(jobs))
		t.Fail()
	}
}
