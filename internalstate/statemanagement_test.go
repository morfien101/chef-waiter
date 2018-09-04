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
