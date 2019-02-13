package chefrunner

import (
	"fmt"
	"strings"
	"time"

	"github.com/morfien101/chef-waiter/cheflogs"
	"github.com/morfien101/chef-waiter/cmd"
	"github.com/morfien101/chef-waiter/internalstate"
	"github.com/morfien101/chef-waiter/logs"
	"github.com/morfien101/chef-waiter/metrics"
)

// Request is a RunRequest that is used to push messaged to a queue which will trigger runs.
var Request RunRequest

// Worker is what is needed to register runs of 2 types.
type Worker interface {
	OnDemandRun() string
	PeriodicRun() string
	CustomRun(string) string
}

// RunRequest holds 2 channels for on demand runs and periodic runs. It also has the functions to add jobs to the queues.
type RunRequest struct {
	onDemandWorkQ chan string
	periodicWorkQ chan string
	logger        logs.SysLogger
	state         internalstate.StateTableReadWriter
	chefLogWorker cheflogs.WorkerReader
}

// OnDemandRun will return a string guid for a on demand scheduled run.
func (r *RunRequest) OnDemandRun() string {
	ok, guid := r.state.RegisterRun(true, false, "")
	if ok {
		logs.DebugMessage(fmt.Sprintf("New GUID Generated: %s, submitting a new job for onDemand", guid))
		r.onDemandWorkQ <- guid
	}
	logs.DebugMessage(fmt.Sprintf("Returning GUID:%s from OnDemandRun()", guid))
	return guid
}

// CustomRun will return a guid of a custom run that has been scheduled.
func (r *RunRequest) CustomRun(runDetails string) string {
	ok, guid := r.state.RegisterRun(true, true, runDetails)
	if ok {
		logs.DebugMessage(fmt.Sprintf("New GUID Generated: %s, submitting a new job for CustomRun with text: %s", guid, runDetails))
		r.onDemandWorkQ <- guid
	}
	logs.DebugMessage(fmt.Sprintf("Returning GUID:%s from CustomRun()", guid))
	return guid
}

// PeriodicRun will return a string guid for a scheduled run.
func (r *RunRequest) PeriodicRun() string {
	ok, guid := r.state.RegisterRun(false, false, "")
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
	// Preamble for metrics shipping
	start := func(jobType string) {
		metrics.Incr("run_starting", 1, map[string]string{"type": jobType})
	}

	finished := func(jobType string) {
		metrics.Incr("run_finished", 1, map[string]string{"type": jobType})
	}

	timer := func(f func(string), guid, jobType string) {
		start(jobType)
		start := time.Now()
		f(guid)
		metrics.Timing("chef_run_time", int64(time.Since(start)/time.Millisecond), map[string]string{"type": jobType})
		finished(jobType)
	}

	for {
		select {
		case guid := <-r.periodicWorkQ:
			//run chef as periodic job
			if r.state.ReadPeriodicRuns() {
				timer(r.startChefRunProcess, guid, "periodic")
			}
		case guid := <-r.onDemandWorkQ:
			timer(r.startChefRunProcess, guid, "demand")
		}
	}
}

func (r *RunRequest) startChefRunProcess(guid string) {
	ondemand := r.state.IsDemandJob(guid)
	var lmsg string
	if ondemand {
		lmsg = "on demand"
	} else {
		lmsg = "periodic"
	}
	custom, arg := r.state.IsCustomJob(guid)
	if custom {
		r.logger.Infof("Starting %s chef custom run with argument '%s': %s", lmsg, arg, guid)
	} else {
		r.logger.Infof("Starting %s chef run: %s", lmsg, guid)
	}

	if ondemand == false {
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

// runChef will run the command based on the OS
func (r *RunRequest) runChef(guid string) (exitCode int) {
	command := chefClientCommand
	command = append(command, r.chefClientArguments(guid)...)
	logs.DebugMessage(fmt.Sprintf("runChef(%s): %s %s", guid, command[0], strings.Join(command[1:], " ")))
	stdout, stderr, exitCode := cmd.RunCommand(command[0], command[1:]...)
	logs.DebugMessage(fmt.Sprintf("STDOUT %s: %s", guid, stdout))
	logs.DebugMessage(fmt.Sprintf("STDERR %s: %s", guid, stderr))
	return
}

// chefClientArguments will compile the arguments and return them as a []string
func (r *RunRequest) chefClientArguments(guid string) []string {
	arguments := make([]string, 0)
	arguments = append(arguments, "-L", r.chefLogWorker.GetLogPath(guid))
	customJob, strValue := r.state.IsCustomJob(guid)
	if customJob {
		arguments = append(arguments, "-o", fmt.Sprintf(`%s`, strValue))
	}
	return arguments
}
