package webengine

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/newvoicemedia/chef-waiter/cheflogs"
	"github.com/newvoicemedia/chef-waiter/chefrunner"
	"github.com/newvoicemedia/chef-waiter/internalstate"
	"github.com/newvoicemedia/chef-waiter/logs"

	"github.com/gorilla/mux"
)

//var logger = logs.Logger

// HTTPEngine holds all the requires types and functions for the API to work.
type HTTPEngine struct {
	router         *mux.Router
	logger         logs.SysLogger
	state          internalstate.StateTableReadWriter
	appState       internalstate.AppStatusReader
	worker         chefrunner.Worker
	chefLogsWorker cheflogs.WorkerReader
	server         *http.Server
}

// New returns a struct that holds the required details for the API engine.
// You still need to start it with StartHTTPEngine()
func New(
	state internalstate.StateTableReadWriter,
	appState internalstate.AppStatusReader,
	worker chefrunner.Worker,
	chefLogsWorker cheflogs.WorkerReader,
	logger logs.SysLogger,
) (e *HTTPEngine) {
	httpEngine := &HTTPEngine{
		logger:         logger,
		state:          state,
		appState:       appState,
		worker:         worker,
		chefLogsWorker: chefLogsWorker,
		router:         mux.NewRouter(),
	}

	httpEngine.router.HandleFunc("/chefclient", httpEngine.registerChefRun).Methods("Get")
	httpEngine.router.HandleFunc("/chefclient/{guid}", httpEngine.getChefStatus).Methods("Get")
	httpEngine.router.HandleFunc("/cheflogs/{guid}", httpEngine.getChefLogs).Methods("Get")
	httpEngine.router.HandleFunc("/chef/nextrun", httpEngine.getNextChefRun).Methods("Get")
	httpEngine.router.HandleFunc("/chef/interval", httpEngine.getChefRunInterval).Methods("Get")
	httpEngine.router.HandleFunc("/chef/interval/{i}", httpEngine.setChefRunInterval).Methods("Get")
	httpEngine.router.HandleFunc("/chef/on", httpEngine.setChefRunEnabled).Methods("Get")
	httpEngine.router.HandleFunc("/chef/off", httpEngine.setChefRunDisabled).Methods("Get")
	httpEngine.router.HandleFunc("/chef/lastrun", httpEngine.getLastRunGUID).Methods("Get")
	httpEngine.router.HandleFunc("/chef/enabled", httpEngine.getChefPeridoicRunStatus).Methods("Get")
	httpEngine.router.HandleFunc("/chef/maintenance", httpEngine.getChefMaintenance).Methods("Get")
	httpEngine.router.HandleFunc("/chef/maintenance/start/{i}", httpEngine.setChefMaintenance).Methods("Get")
	httpEngine.router.HandleFunc("/chef/maintenance/end", httpEngine.removeChefMaintenance).Methods("Get")
	httpEngine.router.HandleFunc("/chef/lock", httpEngine.getChefLock).Methods("Get")
	httpEngine.router.HandleFunc("/chef/lock/set", httpEngine.setChefLock).Methods("Get")
	httpEngine.router.HandleFunc("/chef/lock/remove", httpEngine.removeChefLock).Methods("Get")
	httpEngine.router.HandleFunc("/status", httpEngine.getStatus).Methods("Get")
	httpEngine.router.HandleFunc("/_status", httpEngine.getStatus).Methods("Get")
	httpEngine.router.HandleFunc("/healthcheck", httpEngine.healthCheck).Methods("Get")

	return httpEngine
}

// StartHTTPEngine will start the web server in a nonTLS mode.
// It also requires that the listening address be passes in as a string.
// Should be used in a go routine.
func (e *HTTPEngine) StartHTTPEngine(listenerAddress string) error {
	// Start the HTTP Engine
	e.server = &http.Server{Addr: listenerAddress, Handler: e.router}
	return e.server.ListenAndServe()
}

// StartHTTPSEngine will start the web server with TLS support using the given cert and key values.
// It also requires that the listening address be passes in as a string.
// Should be used in a go routine.
func (e *HTTPEngine) StartHTTPSEngine(listenerAddress, certPath, keyPath string) error {
	// Start the HTTP Engine
	e.server = &http.Server{Addr: listenerAddress, Handler: e.router}
	return e.server.ListenAndServeTLS(certPath, keyPath)
}

// StopHTTPEngine will stop the web server grafefully.
// It will give the server 5 seconds before just terminating it.
func (e *HTTPEngine) StopHTTPEngine() error {
	// Stop the HTTP Engine
	ctx, cancleFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancleFunc()
	return e.server.Shutdown(ctx)
}

// ServeHTTP is used to allow the router to start accepting requests before the start is started up. This will help with testing.
func (e *HTTPEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.router.ServeHTTP(w, r)
}

func setContentJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}

// RegisterChefRun is called to run chef on the server.
func (e *HTTPEngine) registerChefRun(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)
	if e.state.ReadRunLock() {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "{\"Error\":\"Chefwaiter is locked\"}\n")
		return
	}
	guid := e.worker.OnDemandRun()
	logs.DebugMessage(fmt.Sprintf("registerChefRun() - %s", guid))
	setContentJSON(w)
	json.NewEncoder(w).Encode(e.state.Read(guid))
}

// GetChefStatus - writes the state of the requested guid.
func (e *HTTPEngine) getChefStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	logs.DebugMessage(fmt.Sprintf("getChefStatus() - %s", vars["guid"]))
	setContentJSON(w)
	json.NewEncoder(w).Encode(e.state.Read(vars["guid"]))
}

// GetStatus - Writes the applications internal status in json to the http writer.
func (e *HTTPEngine) getStatus(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)
	state, err := e.appState.JSONEncoded()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.Write(state)
}

// HealthCheck - Writes a HealthCheck message that can be used to check the state
// of the chef waiter.
func (e *HTTPEngine) healthCheck(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)
	fmt.Fprint(w, "{\"state\": \"OK\"}")
}

// getChefLogs - is responsible for displaying the chef logs that have been created
// by a chef run.
func (e *HTTPEngine) getChefLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// Set the content type
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// We first need to look for the log file.
	// Throw a 404 if the file is not there
	if err := e.chefLogsWorker.IsLogAvailable(vars["guid"]); err != nil {
		w.WriteHeader(http.StatusNotFound)
		logs.DebugMessage(fmt.Sprintf("Unavailable: %s, %s", e.chefLogsWorker.GetLogPath(vars["guid"]), err))
		fmt.Fprintf(w, "404 - %s not found\n", vars["guid"])
		return
	}
	logs.DebugMessage(fmt.Sprintf("Found: %s", e.chefLogsWorker.GetLogPath(vars["guid"])))

	// If it is there then we need to read it out.
	file, err := os.Open(e.chefLogsWorker.GetLogPath(vars["guid"]))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e.logger.Errorf("Failed to open %s: %v", e.chefLogsWorker.GetLogPath(vars["guid"]), err)
		return
	}
	// remember to close it at the end.
	defer file.Close()

	// At this point we are about to read out the file so it is safe to
	// write the headers for OK Status.
	w.WriteHeader(http.StatusOK)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Fprintln(w, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		e.logger.Errorf("Failed to read file: %s, Error: %s", file.Name(), err)
	}
}

func (e *HTTPEngine) getNextChefRun(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)
	w.WriteHeader(http.StatusOK)
	// json string with epoch and string time
	epoch := e.state.GetlastRunStartTime() + e.state.ReadChefRunTimer()
	next := &struct {
		Epoch int64  `json:"epoch"`
		Str   string `json:"human"`
	}{
		Epoch: epoch,
		Str:   time.Unix(epoch, 0).String(),
	}
	json.NewEncoder(w).Encode(next)
}

func (e *HTTPEngine) setChefRunInterval(w http.ResponseWriter, r *http.Request) {
	// check if the string is a number and is positive
	setContentJSON(w)
	vars := mux.Vars(r)
	i, err := strconv.Atoi(vars["i"])
	if err != nil || i < 0 {
		e.logger.Errorf("/chef/interval/%s is not a postive number", vars["i"])
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "{\"error\":\"Only a positive number will be accepted.\"}\n")
		return
	}
	if i <= 0 {
		e.logger.Errorf("/chef/interval/%s is not a postive number", vars["i"])
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "{\"error\":\"Only a positive number will be accepted.\"}\n")
		return
	}

	e.state.WriteChefRunTimer(int64(i))
}

func (e *HTTPEngine) getChefRunInterval(w http.ResponseWriter, r *http.Request) {
	i := e.state.ReadChefRunTimer()
	setContentJSON(w)
	fmt.Fprintf(w, "{\"current_interval\":\"%d minutes\"}\n", i/60)
}

// setChefRunEnabled - enables periodic runs
func (e *HTTPEngine) setChefRunEnabled(w http.ResponseWriter, r *http.Request) {
	e.state.WritePeriodicRuns(true)
	setContentJSON(w)
	fmt.Fprintf(w, "{\"chef_runs_enabled\":%v}\n", e.state.ReadPeriodicRuns())
}

// setChefRunDisabled - disables periodic runs
func (e *HTTPEngine) setChefRunDisabled(w http.ResponseWriter, r *http.Request) {
	e.state.WritePeriodicRuns(false)
	setContentJSON(w)
	fmt.Fprintf(w, "{\"chef_runs_enabled\":%v}\n", e.state.ReadPeriodicRuns())
}

// getChefPeridoicRunStatus - returns details about if periodic runs are enabled.
func (e *HTTPEngine) getChefPeridoicRunStatus(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)
	fmt.Fprintf(w, "{\"chef_runs_enabled\":%v}\n", e.state.ReadPeriodicRuns())
}

func (e *HTTPEngine) getLastRunGUID(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)
	fmt.Fprintf(w, "{\"last_run_guid\":\"%s\"}\n", e.state.ReadLastRunGUID())
}

func (e *HTTPEngine) getChefMaintenance(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)
	fmt.Fprintf(w, "{\"end_time\":\"%s\", \"in_maintenance\":%v}\n", time.Unix(e.state.ReadMaintenanceTimeEnd(), 0), e.state.InMaintenceMode())
}
func (e *HTTPEngine) setChefMaintenance(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)

	vars := mux.Vars(r)
	minutes, err := strconv.Atoi(vars["i"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	endTime := time.Now().Unix() + int64(minutes*60)
	e.state.WriteMaintenanceTimeEnd(endTime)
	fmt.Fprintf(w, "{\"end_time\":\"%s\"}\n", time.Unix(endTime, 0))
}

func (e *HTTPEngine) removeChefMaintenance(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)

	e.state.WriteMaintenanceTimeEnd(0)
	fmt.Fprintf(w, "{\"end_time\":\"%s\"}\n", time.Unix(e.state.ReadMaintenanceTimeEnd(), 0))
}

func (e *HTTPEngine) getChefLock(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)
	fmt.Fprintf(w, "{\"Locked\": %t}\n", e.state.ReadRunLock())
}

func (e *HTTPEngine) setChefLock(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)
	e.state.LockRuns(true)
	fmt.Fprintf(w, "{\"Locked\": %t}\n", e.state.ReadRunLock())
}

func (e *HTTPEngine) removeChefLock(w http.ResponseWriter, r *http.Request) {
	setContentJSON(w)
	e.state.LockRuns(false)
	fmt.Fprintf(w, "{\"Locked\": %t}\n", e.state.ReadRunLock())
}
