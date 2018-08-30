package chefrunner

import (
	"fmt"
	"time"

	"github.com/newvoicemedia/chef-waiter/cheflogs"
	"github.com/newvoicemedia/chef-waiter/internalstate"
	"github.com/newvoicemedia/chef-waiter/logs"
)

// Request is a RunRequest that is used to push messaged to a queue which will trigger runs.
var Request RunRequest

// Worker is what is needed to register runs of 2 types.
type Worker interface {
	OnDemandRun() string
	PeriodicRun() string
}

// RunRequest holds 2 channles for on demand runs and periodic runs. It also has the functions to add jobs to the queues.
type RunRequest struct {
	onDemandWorkQ chan string
	periodicWorkQ chan string
	logger        logs.SysLogger
	state         internalstate.StateTableReadWriter
	chefLogWorker cheflogs.WorkerReader
}

// OnDemandRun will return a string guid for a on demand scheduled run.
func (r *RunRequest) OnDemandRun() string {
	ok, guid := r.state.RegisterRun(true)
	if ok {
		logs.DebugMessage(fmt.Sprintf("New GUID Generated: %s, submitting a new job for onDemand", guid))
		r.onDemandWorkQ <- guid
	}
	logs.DebugMessage(fmt.Sprintf("Returning GUID:%s from OnDemandRun()", guid))
	return guid
}

// PeriodicRun will return a string guid for a scheduled run.
func (r *RunRequest) PeriodicRun() string {
	ok, guid := r.state.RegisterRun(true)
	if ok {
		logs.DebugMessage(fmt.Sprintf("New GUID Generated: %s, submitting a new job for periodic", guid))
		r.periodicWorkQ <- guid
	}
	logs.DebugMessage(fmt.Sprintf("Returning GUID:%s from PeriodicRun()", guid))
	return guid
}

// New - Runs the worker process that will run the commands one at a time.
func New(state *internalstate.StateTable, chefLogWorker cheflogs.WorkerReader, logger logs.SysLogger) *RunRequest {
	logs.DebugMessage("StartWorker()")
	worker := &RunRequest{
		onDemandWorkQ: make(chan string, 10),
		periodicWorkQ: make(chan string, 10),
		state:         state,
		logger:        logger,
		chefLogWorker: chefLogWorker,
	}
	go worker.supervisor()
	go worker.periodicRunEngine()
	return worker
}

func (r *RunRequest) supervisor() {
	for {
		select {
		case guid := <-r.periodicWorkQ:
			//run chef as periodic job
			if r.state.ReadPeriodicRuns() {
				r.startChefRunProcess(guid, false)
			}
		case guid := <-r.onDemandWorkQ:
			r.startChefRunProcess(guid, true)
		}
	}
}

func (r *RunRequest) startChefRunProcess(guid string, onDemand bool) {
	var lmsg string
	if onDemand {
		lmsg = "on demand"
	} else {
		lmsg = "periodic"
	}
	r.logger.Infof("Starting %s chef run: %s", lmsg, guid)

	if onDemand == false {
		r.state.UpdatelastRunStartTime(time.Now().Unix())
	}
	r.state.UpdateStatus(guid, "running")
	exitCode := r.runChef(guid)
	r.state.UpdateExitCode(guid, exitCode)
	if exitCode != 0 {
		r.state.UpdateStatus(guid, "failed")
	} else {
		r.state.UpdateStatus(guid, "complete")
	}
	r.state.WriteLastRunGUID(guid)

	r.logger.Infof("Finished %s run with guid: %s, exit code was: %d", lmsg, guid, exitCode)
}

// PeriodicRunEngine - checks if we need to run chef and sends a request to run chef on a interval of 1 minute.
func (r *RunRequest) periodicRunEngine() {
	logs.DebugMessage("periodicRunEngine()")
	trigger := time.NewTicker(time.Minute * 1)
	for _ = range trigger.C {
		if r.timeToRunChef() && r.state.ReadPeriodicRuns() {
			r.PeriodicRun()
		}
	}
}

// timeToRunChef - checks if it is time to run chef.
// True if the time now is later than the last run + the interval that we have currently.
// Also true if there is not a maintenance window active.
// We also check to see if the jobs have been locked which would stop anything further being
// registered.
func (r *RunRequest) timeToRunChef() bool {
	if r.state.ReadRunLock() {
		return false
	}
	return (time.Now().Unix() > r.state.GetlastRunStartTime()+r.state.ReadChefRunTimer()) && !r.state.InMaintenceMode()
}
