package internalstate

import (
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/morfien101/chef-waiter/cheflogs"
	"github.com/morfien101/chef-waiter/config"
	"github.com/morfien101/chef-waiter/logs"
)

// JobDetails - Holds data about individual runs.
// Status can be one of the following: registered, running, complete, unknown, abandoned
// unknown: is set if the data is read from a static state file on start up and the
// job was previously set to running.
// abandoned: is set if the data is read from a static state file on start up and the
// job was previously set to registered.
type JobDetails struct {
	Status          string `json:"status"`
	ExitCode        int    `json:"exitcode"`
	RegisteredTime  int64  `json:"starttime"`
	OnDemand        bool   `json:"ondemand"`
	CustomRun       bool   `json:"custom_run"`
	CustomRunString string `json:"custom_run_string"`
}

// TODO - Switch to using this for status of runs.
//var regState = map[string]string{
//	"reg": "registered",
//	"run": "running",
//	"com": "complete",
//	"unk": "unknown",
//	"aba": "abandoned",
//}

// StateTable - holds the state map and sync functions.
type StateTable struct {
	mutexLock sync.RWMutex
	Status    map[string]*JobDetails
	// Used to hold the epoch time when chef last run and completed good or bad.
	LastRunStartTime int64
	LastRunGUID      string
	ChefRunTimer     int64
	PeriodicRuns     bool
	// This should be changed to StateTableMaxSize
	StateTableSize     int
	MaintenanceTimeEnd int64
	Locked             bool
	StateFilePath      string

	chefLogsWorker cheflogs.WorkerWriter
	logger         logs.SysLogger
}

// StateTableReadWriter describes functions that both read and write on the statetable
type StateTableReadWriter interface {
	StateTableReader
	StateTableWriter
}

// StateTableReader describes the functions required to read data from the state table.
type StateTableReader interface {
	Read(string) map[string]*JobDetails
	ReadAll() map[string]*JobDetails
	IsDemandJob(string) bool
	IsCustomJob(string) (bool, string)
	GetAllStateTimes() map[string]int64
	GetlastRunStartTime() int64
	ReadChefRunTimer() int64
	ReadPeriodicRuns() bool
	ReadLastRunGUID() string
	ReadAllJobs() map[string]JobDetails
	ReadRunLock() bool
	InMaintenceMode() bool
	ReadMaintenanceTimeEnd() int64
}

// StateTableWriter describes the functions to write data to the state table.
type StateTableWriter interface {
	Add(string, bool)
	RegisterRun(bool, bool, string) (bool, string)
	UpdateStatus(string, string)
	UpdateExitCode(string, int)
	RemoveState(string)
	UpdatelastRunStartTime(int64)
	WriteChefRunTimer(int64)
	WritePeriodicRuns(bool)
	WriteLastRunGUID(string)
	WriteMaintenanceTimeEnd(int64)
	LockRuns(bool)
}

// New will initialize a new state table either empty or with the saved state if found.
func New(
	config config.Config,
	chefLogsWorker cheflogs.WorkerWriter,
	logger logs.SysLogger,
) *StateTable {
	diskState, err := readStateFromDisk(getStatePath(config.StateFileLocation(), statefile), logger)
	if err != nil {
		logger.Warningf("There was an error reading the state from disk. Creating a new internal state. The error was: %s", err)
		// initialize the globals that we need.
		return defaultStateTable(config, chefLogsWorker, logger)
	}
	// We need to set the values to what the configuration file states if we have one.
	// If it is not there then the values would be the default ones.
	// If we don't do this then new values in configuration files are not read in when we find a statefile on disk.
	diskState.resetStateTable(config, chefLogsWorker, logger)
	return diskState
}

// newStateTable - Constructs a new state table with Zero values.
func defaultStateTable(config config.Config, chefLogsWorker cheflogs.WorkerWriter, logger logs.SysLogger) (st *StateTable) {
	logs.DebugMessage("run newStateTable()")
	return &StateTable{
		Status:             make(map[string]*JobDetails),
		LastRunStartTime:   int64(1257894000),
		ChefRunTimer:       config.PeriodicTimer() * 60,
		PeriodicRuns:       config.ControlChefRun(),
		StateTableSize:     config.StateTableSize(),
		MaintenanceTimeEnd: 0,
		Locked:             false,
		StateFilePath:      getStatePath(config.StateFileLocation(), statefile),
		chefLogsWorker:     chefLogsWorker,
		logger:             logger,
	}
}

// resetStateTable is used to reset the values stored in the State Table to those
// specified in the configuration files. If we didn't do this values would be read
// from the state file on the disk. This would mean it never gets updated unless
// we remove the file first but we also loose the run details.
func (st *StateTable) resetStateTable(config config.Config, chefLogsWorker cheflogs.WorkerWriter, logger logs.SysLogger) {
	st.ChefRunTimer = config.PeriodicTimer() * 60
	st.PeriodicRuns = config.ControlChefRun()
	st.StateTableSize = config.StateTableSize()
	st.chefLogsWorker = chefLogsWorker
	st.logger = logger
}

// Lock - locks the mutex for writing to the state table.
func (st *StateTable) lock() {
	st.mutexLock.Lock()
}

// Unlock - releases the mutex for writing to the state table.
func (st *StateTable) unlock() {
	st.mutexLock.Unlock()
}

// RLock - locks the mutex for reading from the state table.
func (st *StateTable) rLock() {
	st.mutexLock.RLock()
}

// RUnlock - releases the mutex for reading from the state table.
func (st *StateTable) rUnlock() {
	st.mutexLock.RUnlock()
}

// flushToDisk - Will copy the current state table as it is to disk.
// Used for reboots and service restarts.
func (st *StateTable) flushToDisk(sf io.Writer) error {
	logs.DebugMessage("flushToDisk()")
	st.lock()
	defer st.unlock()
	enc := gob.NewEncoder(sf)
	err := enc.Encode(st)
	if err != nil {
		st.logger.Error(err)
		return err
	}
	return nil
}

// Add - Allows us to add a guid to the state table with default values.
func (st *StateTable) Add(id string, ondemand bool) {
	st.lock()
	defer st.unlock()
	st.Status[id] = &JobDetails{
		Status:         "registered",
		ExitCode:       99,
		RegisteredTime: time.Now().Unix(),
		OnDemand:       ondemand,
	}
}

// AddCustom - Allows the caller to add a guid to the state table with details of a
// custom job.
func (st *StateTable) AddCustom(id string, customString string) {
	st.lock()
	defer st.unlock()
	st.Status[id] = &JobDetails{
		Status:          "registered",
		ExitCode:        99,
		RegisteredTime:  time.Now().Unix(),
		OnDemand:        true,
		CustomRun:       true,
		CustomRunString: customString,
	}
}

// RegisterRun - Allows us to check if a on demand run is registered and to register one
// if there is not. It will return a bool true to signal that a new run was created and also
// return a string of the guid that this run is associated with. The run could be a copy
// of a previos run that is still queuing to run.
func (st *StateTable) RegisterRun(onDemand, customRun bool, customString string) (ok bool, guid string) {
	// check if there is a on demand chef run already waiting.
	// if so collect the guid
	// else create a run and make a guid

	st.rLock()
	for id := range st.Status {
		i := st.Status[id]
		if i.Status == "registered" {
			// Determin if i is also a custom run if so match the strings.
			if customRun && i.CustomRun && i.CustomRunString == customString {
				guid = id
			} else {
				// If its not a custom run then it can either be onDemand or periodic.
				// Either way if the values match then return a guid.
				if i.OnDemand == onDemand {
					guid = id
				}
			}
		}
	}
	st.rUnlock()

	// If the guid has not been set then get one.
	if len(guid) < 1 {
		guid = uuid.Must(uuid.NewV4()).String()
		if customRun {
			st.AddCustom(guid, customString)
		} else {
			st.Add(guid, onDemand)
		}
		return true, guid
	}
	logs.DebugMessage(fmt.Sprintf("Return a queued guid: %s", guid))
	return false, guid
}

// UpdateStatus - Updates the states of an ID with the given status string
func (st *StateTable) UpdateStatus(guid string, state string) {
	logs.DebugMessage(fmt.Sprintf("UpdateStatus(%s,%s)", guid, state))
	st.lock()
	defer st.unlock()
	st.Status[guid].Status = state
}

// UpdateExitCode - Updates the ExitCode of an ID with the given int.
func (st *StateTable) UpdateExitCode(guid string, code int) {
	logs.DebugMessage(fmt.Sprintf("UpdateExitCode(%s,%d)", guid, code))
	st.lock()
	defer st.unlock()
	st.Status[guid].ExitCode = code
}

// IsDemandJob will return the value of a JobDetails OnDemand value. This
// will let the caller know if it is a on demand job.
func (st *StateTable) IsDemandJob(guid string) bool {
	st.rLock()
	defer st.rUnlock()
	value, ok := st.Status[guid]
	if !ok {
		return false
	}
	return value.OnDemand
}

// IsCustomJob will return true or false detailing if the job is a custom run
// and also the value of the custom run string.
func (st *StateTable) IsCustomJob(guid string) (bool, string) {
	st.rLock()
	defer st.rUnlock()
	value, ok := st.Status[guid]
	if !ok {
		return false, ""
	}
	return value.CustomRun, value.CustomRunString
}

// Read - Creates a copy of the current state and returns it. This makes it thread safe.
func (st *StateTable) Read(guid string) (status map[string]*JobDetails) {
	status = make(map[string]*JobDetails)
	st.rLock()
	status[guid] = st.Status[guid]
	st.rUnlock()
	return status
}

// ReadAll - returns all the state table entries.
// Can be used for saving the state
func (st *StateTable) ReadAll() (status map[string]*JobDetails) {
	st.rLock()
	defer st.rUnlock()
	return st.Status
}

// RemoveState - removes a guid from the Statetable.
func (st *StateTable) RemoveState(guid string) {
	st.lock()
	defer st.unlock()
	if st.Status[guid].Status == "complete" {
		delete(st.Status, guid)
	}
}

// GetAllStateTimes - Returns all the status guids and times
func (st *StateTable) GetAllStateTimes() (statusMap map[string]int64) {
	st.rLock()
	defer st.rUnlock()
	statusMap = make(map[string]int64)
	for k, v := range st.Status {
		statusMap[k] = v.RegisteredTime
	}
	return statusMap
}

func (st *StateTable) len() (length int) {
	st.rLock()
	defer st.rUnlock()
	return len(st.Status)
}

// GetlastRunStartTime will return the last time that chef started a run in the form of a epoch time.
func (st *StateTable) GetlastRunStartTime() int64 {
	st.rLock()
	defer st.rUnlock()
	return st.LastRunStartTime
}

// UpdatelastRunStartTime will set the last time that chef started a run to the supplied int64.
// This should be a epoch time.
func (st *StateTable) UpdatelastRunStartTime(t int64) {
	st.lock()
	defer st.unlock()
	st.LastRunStartTime = t
}

// ReadChefRunTimer will return an int64 with represents in minutes how often we run chef.
func (st *StateTable) ReadChefRunTimer() int64 {
	st.rLock()
	defer st.rUnlock()
	return st.ChefRunTimer
}

// WriteChefRunTimer will update the chef runner trigger timer to be the supplied int64 * 60
// to represent minutes.
func (st *StateTable) WriteChefRunTimer(i int64) {
	st.lock()
	defer st.unlock()
	st.ChefRunTimer = i * 60
	st.logger.Infof("Chef periodic interval changed to every %d minutes.", i)
}

// ReadPeriodicRuns will return the value of PeriodicRuns.
func (st *StateTable) ReadPeriodicRuns() bool {
	st.rLock()
	defer st.rUnlock()
	return st.PeriodicRuns
}

// WritePeriodicRuns will set the value of Periodic runs to the bool that is passed in.
func (st *StateTable) WritePeriodicRuns(enable bool) {
	st.lock()
	defer st.unlock()
	if enable {
		logs.DebugMessage("chef runs enabled.")
	} else {
		logs.DebugMessage("chef run disabled.")
	}
	st.PeriodicRuns = enable
}

func (st *StateTable) readStateTableSize() int {
	st.rLock()
	defer st.rUnlock()
	return st.StateTableSize
}

// ReadLastRunGUID will return the last guid that was linked to a chef run.
func (st *StateTable) ReadLastRunGUID() string {
	st.rLock()
	defer st.rUnlock()
	return st.LastRunGUID
}

// ReadAllJobs will create a copy of the jobs as they are and return them to the caller.
func (st *StateTable) ReadAllJobs() map[string]JobDetails {
	st.rLock()
	defer st.rUnlock()
	retVal := make(map[string]JobDetails)
	for guid, job := range st.Status {
		retVal[guid] = *job
	}
	return retVal
}

// WriteLastRunGUID will write to the state table the guid passed in.
func (st *StateTable) WriteLastRunGUID(guid string) {
	st.lock()
	defer st.unlock()
	st.LastRunGUID = guid
}

// WriteMaintenanceTimeEnd will write when Maintenance must end. It takes an int64 as and assumes this is an epoch
func (st *StateTable) WriteMaintenanceTimeEnd(epoch int64) {
	st.lock()
	defer st.unlock()
	st.MaintenanceTimeEnd = epoch
}

// ReadMaintenanceTimeEnd will return the value of MaintenanceTimeEnd. It is an epoch represented as an int64
func (st *StateTable) ReadMaintenanceTimeEnd() int64 {
	st.rLock()
	defer st.rUnlock()
	return st.MaintenanceTimeEnd
}

// InMaintenceMode will return true or false based on if the periodic run engine
// is in maintenance mode.
func (st *StateTable) InMaintenceMode() bool {
	return time.Now().Unix() < st.ReadMaintenanceTimeEnd()
}

func (st *StateTable) readStateFilePath() string {
	st.rLock()
	defer st.rUnlock()
	return st.StateFilePath
}

// LockRuns will lock the chef waiter to stop accepting runs
func (st *StateTable) LockRuns(lock bool) {
	st.lock()
	defer st.unlock()
	if lock {
		st.logger.Info("Chefwaiter has just been locked. No new runs can be scheduled.")
		st.Locked = true
	} else {
		st.logger.Info("Chefwaiter has just been unlocked. New runs can now be scheduled.")
		st.Locked = false
	}
}

// ReadRunLock will return the value of the state tables Lock value.
func (st *StateTable) ReadRunLock() bool {
	st.rLock()
	defer st.rUnlock()
	return st.Locked
}
