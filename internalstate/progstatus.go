package internalstate

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/morfien101/chef-waiter/logs"
)

// AppStatusHandler - Hosts the AppStatus in a mutable struct.
type AppStatusHandler struct {
	sync.RWMutex
	state  *AppStatus
	logger logs.SysLogger
}

// AppStatus - Holds status information about the chef waiter itself.
type AppStatus struct {
	ServiceName string `json:"service_name"`
	HostName    string `json:"hostname"`
	StartTime   int64  `json:"start_time"`
	// Uptime is to be deprecated 19/03/2019
	Uptime            int64    `json:"uptime"`
	StartTimeHuman    string   `json:"start_time_human_readable"`
	Version           string   `json:"version"`
	ChefVersion       string   `json:"chef_version"`
	Healthy           bool     `json:"healthy"`
	InMaintenance     bool     `json:"in_maintenance_mode"`
	LastRunGUID       string   `json:"last_run_id"`
	Locked            bool     `json:"locked"`
	WhiteListsEnabled bool     `json:"whitelisting_enabled"`
	WhiteList         []string `json:"whitelisted_payloads"`
}

// AppStatusReader will show how to use the AppStatusHandler
type AppStatusReader interface {
	JSONEncoded() ([]byte, error)
}

// NewAppStatus - creates a new appStatusHandler struct. It requires a version
// number to be passed in. This is because the version is held outside of
// internalstate.
func NewAppStatus(version string, currentState *StateTable, logger logs.SysLogger) *AppStatusHandler {
	logs.DebugMessage("NewAppStatus()")
	hn, err := os.Hostname()
	if err != nil {
		hn = "na"
		logger.Errorf("Failed to determin the hostname. Error: %s", err)
	}
	appStatus := new(AppStatusHandler)
	appStatus.logger = logger
	appStatus.state = &AppStatus{
		ServiceName: "ChefWaiter",
		Version:     version,
		Healthy:     true,
		HostName:    hn,
	}
	appStatus.setTime()
	go appStatus.reconcileChefVersion()
	go appStatus.maintenanceMode(currentState)
	go appStatus.lastRun(currentState)
	go appStatus.locked(currentState)
	return appStatus
}

// SetWhiteListing is used to display the whitelist out to the status page.
func (as *AppStatusHandler) SetWhiteListing(enabled bool, currentList []string) {
	as.state.WhiteListsEnabled = enabled
	if enabled {
		as.state.WhiteList = currentList
	}
}

// setTime - is used to set the time of the state in AppStatusHandler
func (as *AppStatusHandler) setTime() {
	as.Lock()
	defer as.Unlock()
	timeNow := time.Now()
	// Uptime is to be deprecated 19/03/2019
	as.state.Uptime = timeNow.Unix()
	as.state.StartTime = timeNow.Unix()
	as.state.StartTimeHuman = timeNow.Format("Mon Jan 2 2006 - 15:04:05 -0700 MST")
}

// This is a looping function that will try to update chef waiter status with the version of chef.
func (as *AppStatusHandler) reconcileChefVersion() {
	// do it now and then again every 15 mins.s
	as.updateChefVersion()
	ticker := time.NewTicker(time.Minute * 15)
	for {
		select {
		case <-ticker.C:
			as.updateChefVersion()
		}
	}
}

func (as *AppStatusHandler) updateChefVersion() {
	version, err := chefVersion()
	as.Lock()
	defer as.Unlock()
	if err != nil {
		as.logger.Error("Failed to determine chef version.")
		as.state.Healthy = false
		return
	}
	as.state.ChefVersion = version
}

func (as *AppStatusHandler) maintenanceMode(cs *StateTable) {
	as.Lock()
	// Do it once then loop
	as.state.InMaintenance = cs.InMaintenceMode()
	as.Unlock()
	ticker := time.NewTicker(time.Millisecond * 750)
	for {
		select {
		case <-ticker.C:
			as.Lock()
			as.state.InMaintenance = cs.InMaintenceMode()
			as.Unlock()
		}
	}
}

func (as *AppStatusHandler) lastRun(cs *StateTable) {
	// Do it once then loop
	as.Lock()
	as.state.LastRunGUID = cs.ReadLastRunGUID()
	as.Unlock()
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			as.Lock()
			as.state.LastRunGUID = cs.ReadLastRunGUID()
			as.Unlock()
		}
	}
}

func (as *AppStatusHandler) locked(cs *StateTable) {
	// Do it once then loop
	lockedFunc := func() {
		as.Lock()
		as.state.Locked = cs.ReadRunLock()
		as.Unlock()
	}

	lockedFunc()
	ticker := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ticker.C:
			lockedFunc()
		}
	}
}

// JSONEncoded returns the JSON encoded state with an error if anything goes wrong.
func (as *AppStatusHandler) JSONEncoded() ([]byte, error) {
	as.RLock()
	defer as.RUnlock()
	return json.MarshalIndent(as.state, "", "  ")
}
