package main

import (
	"fmt"
	"os"

	"github.com/kardianos/service"

	"github.com/morfien101/chef-waiter/cheflogs"
	"github.com/morfien101/chef-waiter/chefrunner"
	"github.com/morfien101/chef-waiter/config"
	"github.com/morfien101/chef-waiter/internalstate"
	"github.com/morfien101/chef-waiter/logs"
	"github.com/morfien101/chef-waiter/metrics"
	"github.com/morfien101/chef-waiter/webengine"
)

type program struct {
	exit    chan interface{}
	finshed chan interface{}
}

func (p *program) Start(s service.Service) error {
	// Errors would relate to looking for config files.
	// To be added later.

	// This channel is used in the run section to block.
	p.exit = make(chan interface{})
	p.finshed = make(chan interface{})

	// Start the service in a async go routine
	go p.run()

	// return any errors.
	return nil
}

func (p *program) Stop(s service.Service) error {
	// This section is to shutdown the app gracefully.
	// return any errors relating to the above.
	// For now we just exit

	// This channel is used in the running section to block.
	// It can later be used to save the state of the API
	p.exit <- true
	close(p.exit)
	<-p.finshed
	return nil
}

func (p *program) run() error {
	errChan := make(chan error, 20)
	// read the file from the standard locations
	//   - We could use an environment variable here as there is no other way we know anything yet.
	logger.Infof("Starting Chefwaiter with version: %s", VERSION)
	// read the file from the standard locations
	// Use an environment variable here as there is no other way we know anything yet.
	runningConfig, err := config.New(os.Getenv("CHEFWAITER_CONFIG"), logger)
	if err != nil {
		logger.Error(err)
		terminate(2)
	}
	logs.TurnDebuggingOn(logger, runningConfig.Debug())
	// This is the first place that we can actually send a metric because we now know
	// if we need to.
	if runningConfig.MetricsEnabled {
		logs.DebugMessage("Starting metrics client.")
		if runningConfig.MetricsDefaultTags["host"] == "" {
			hostname, err := os.Hostname()
			if err != nil {
				hostname = "not_available"
			}
			runningConfig.MetricsDefaultTags["host"] = hostname
		}
		metrics.Setup(runningConfig.MetricsHost, runningConfig.MetricsDefaultTags)
	}
	metrics.Incr("starting", 1, map[string]string{"version": VERSION})
	logs.DebugMessage("Starting Service run() function.")
	// Create the directory for logs
	if err := os.MkdirAll(runningConfig.LogLocation(), 0755); err != nil {
		logger.Errorf("Failed to make directories for logs. Error: %s", err)
		terminate(1)
	}

	// Create the directory for stateFile
	if err := os.MkdirAll(runningConfig.StateFileLocation(), 0755); err != nil {
		logger.Errorf("Failed to make directory for statefile. Error: %s", err)
		terminate(1)
	}

	// Start the log sweeper engine
	chefLogWorker := cheflogs.New(runningConfig, logger)
	go chefLogWorker.LogSweepEngine()
	// Initialize a new state tables
	state := internalstate.New(runningConfig, chefLogWorker, logger)
	appState := internalstate.NewAppStatus(VERSION, state, logger)
	appState.SetWhiteListing(runningConfig.InternalWhiteListCustomRuns, runningConfig.InternalAllowedCustomRuns)
	// start the job engine that runs the commands.
	workers := chefrunner.New(state, chefLogWorker, logger)

	// Start the sweeper process to keep state tables clean.
	go state.ClearOldRuns()
	// Start the state file keeper
	go state.PersistState()

	// Start the HTTP Engine
	httpEngine := webengine.New(state, appState, workers, chefLogWorker, logger)
	if runningConfig.WhiteListCustomRuns() {
		if len(runningConfig.AllowedCustomRuns()) > 0 {
			httpEngine.SetWhitelist(runningConfig.AllowedCustomRuns())
		}
	}
	listenString := fmt.Sprintf("%s:%d", runningConfig.ListenAddress(), runningConfig.ListenPort())
	if runningConfig.TLSEnabled() {
		logs.DebugMessage("Starting Web Server with TLS Supported StartHTTPSEngine() function.")
		go func() {
			errChan <- httpEngine.StartHTTPSEngine(listenString, runningConfig.CertPath(), runningConfig.KeyPath())
		}()
	} else {
		logs.DebugMessage("Starting Web Server with StartHTTPEngine() function.")
		go func() {
			errChan <- httpEngine.StartHTTPEngine(listenString)
		}()
	}

	// We need to gather errors and return them to the service
	// controller. We will implement this later.
	// return errors

	// We hold the run function waiting for an exit signal.
	select {
	case err := <-errChan:
		logger.Errorf("We got a critical error. Stopping application. Error: %s", err)
		// This is a hack because the service wrapper doesn't stop the application
		// When we return an error.
		// Really rhe other application should run with context and we cancel them also.
		terminate(1)
		return nil
	case <-p.exit:
		// This case statement can be used to tear down the service and save
		// any state the needs it.
		logs.DebugMessage("Got exit message. Shutting down.")
		err := httpEngine.StopHTTPEngine()
		if err != nil {
			logger.Errorf("Failed to shutdown HTTP service. Error: %s", err)
		}
		err = state.SaveStateToDisk()
		if err != nil {
			logger.Error(err)
		}
		metrics.Incr("shutting_down", 1, map[string]string{"exitCode": fmt.Sprintf("%d", 0), "version": VERSION})
		metrics.Shutdown()
		p.finshed <- true
		return nil
	}
}

func terminate(exitcode int) {
	metrics.Incr("shutting_down", 1, map[string]string{"exitCode": fmt.Sprintf("%d", exitcode), "version": VERSION})
	os.Exit(exitcode)
}
