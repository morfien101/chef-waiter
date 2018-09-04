package internalstate

import (
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/morfien101/chef-waiter/logs"
)

const statefile = "stateTable.db"

// GetOldStates - returns all the old state uuids.
func (st *StateTable) GetOldStates(orignalMap map[string]int64) (del []string) {
	logs.DebugMessage("GetOldStates()")
	var states = []Run{}
	for k, v := range orignalMap {
		states = append(states, Run{k, v})
	}

	// Sort the states by time in a decending order.
	times := func(run1, run2 *Run) bool {
		return run1.time > run2.time
	}
	// Do a inline sort.
	By(times).Sort(states)

	// Make a list of the guids to be returned
	guidsSlice := []string{}
	for i := 0; i <= len(states)-1; i++ {
		guidsSlice = append(guidsSlice, states[i].guid)
	}
	// return the from position 10
	// This would give us the 11th+ guids
	logs.DebugMessage(fmt.Sprintf("GetOldStates() returned: %v", guidsSlice[st.readStateTableSize():]))
	return guidsSlice[st.readStateTableSize():]
}

// ClearOldRuns - Is used to prevent memory leaking by deleting unneeded states.
func (st *StateTable) ClearOldRuns() {
	ticker := time.Tick(1 * time.Minute)
	for _ = range ticker {
		if st.len() > st.readStateTableSize() {
			logs.DebugMessage(fmt.Sprintf("State Table too large. currently: %d/%d", st.len(), st.readStateTableSize()))
			oldstates := st.GetOldStates(st.GetAllStateTimes())
			for _, v := range oldstates {
				st.RemoveState(v)
			}
			// Trigger a log sweep up now that we have removed old states
			// Should this be passed in to the function rather than be a global
			st.chefLogsWorker.RequestDelete(st.GetAllStateTimes())
		} else {
			logs.DebugMessage(fmt.Sprintf("State Table size: %d/%d", st.len(), st.readStateTableSize()))
		}
	}
}

// PersistState - will call the SaveStateToDisk at a time interval.
// This is designed to be run as a go func
func (st *StateTable) PersistState() {
	ticker := time.Tick(1 * time.Minute)
	for _ = range ticker {
		err := st.SaveStateToDisk()
		if err != nil {
			st.logger.Errorf("SaveStateToDisk error: %s", err)
		}
	}
}

// SaveStateToDisk - will save the CurrentState to a file on disk.
func (st *StateTable) SaveStateToDisk() error {
	logs.DebugMessage(fmt.Sprintf("SaveStateToDisk(%s)", st.readStateFilePath()))
	f, err := os.Create(st.readStateFilePath())
	if err != nil {
		st.logger.Errorf("Failed to create the statefile. Error was: %s", err)
		return err
	}
	defer f.Close()
	err = st.flushToDisk(f)
	if err != nil {
		st.logger.Error(err)
		return err
	}
	return nil
}

// readStateFromDisk - Will read the state from the disk if the file is there.
// It will then pass it to the linter and then put the state in the StateTable.
// It will be a copy of the current state from the reboot.
func readStateFromDisk(stateFile string, logger logs.SysLogger) (*StateTable, error) {
	// Open the file and check if it exists.
	f, err := os.Open(stateFile)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	// Decode the file and check if the decodeing works.
	dec := gob.NewDecoder(f)
	var data *StateTable
	err = dec.Decode(&data)
	if err != nil {
		return nil, err
	}
	// Pass the data to the linter to check for running jobs.
	data.Status = lintState(data.Status)
	// We need to inject a mutex into it as it is not exported when we encode it to disk
	data.mutexLock = sync.RWMutex{}
	// We need to tell the state where it is, it could have changed.
	data.StateFilePath = stateFile
	logs.DebugMessage(fmt.Sprintf("State file from disk: %v", data))
	// return a cleaned disk copy of the stateTable
	return data, nil
}

func lintState(statusList map[string]*StatusDetails) map[string]*StatusDetails {
	for k := range statusList {
		if statusList[k].Status == "running" {
			statusList[k].Status = "unknown"
		}
		if statusList[k].Status == "registered" {
			statusList[k].Status = "abandoned"
		}
	}
	return statusList
}
