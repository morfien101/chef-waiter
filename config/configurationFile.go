package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/morfien101/chef-waiter/logs"
)

// Config is used to read out the vales of the configuration file or default values used to run the program.
type Config interface {
	StateTableSize() int
	StateFileLocation() string
	ControlChefRun() bool
	PeriodicTimer() int64
	Debug() bool
	LogLocation() string
	ListenPort() int
	ListenAddress() string
	TLSEnabled() bool
	CertPath() string
	KeyPath() string
}

func (vc *ValuesContainer) StateTableSize() int {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalStateTableSize
}

func (vc *ValuesContainer) StateFileLocation() string {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalStateFileLocation
}

func (vc *ValuesContainer) ControlChefRun() bool {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalControlChefRun
}

func (vc *ValuesContainer) PeriodicTimer() int64 {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalPeriodicTimer
}

func (vc *ValuesContainer) Debug() bool {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalDebug
}

func (vc *ValuesContainer) LogLocation() string {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalLogLocation
}

func (vc *ValuesContainer) ListenPort() int {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalListenPort
}

func (vc *ValuesContainer) ListenAddress() string {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalListenAddress
}

func (vc *ValuesContainer) TLSEnabled() bool {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalTLSEnabled
}

func (vc *ValuesContainer) CertPath() string {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalCertPath
}

func (vc *ValuesContainer) KeyPath() string {
	vc.RLock()
	defer vc.RUnlock()
	return vc.InternalKeyPath
}

// ValuesContainer is a struct that holds the values of the configuration file.
type ValuesContainer struct {
	InternalStateTableSize    int    `json:"state_table_size"`
	InternalControlChefRun    bool   `json:"periodic_chef_runs"`
	InternalPeriodicTimer     int64  `json:"run_interval"`
	InternalDebug             bool   `json:"debug"`
	InternalLogLocation       string `json:"logs_location"`
	InternalStateFileLocation string `json:"state_location"`
	InternalListenPort        int    `json:"listen_port"`
	InternalListenAddress     string `json:"listen_address"`
	InternalTLSEnabled        bool   `json:"enable_tls"`
	InternalCertPath          string `json:"certificate_path"`
	InternalKeyPath           string `json:"key_path"`
	sync.RWMutex
}

// New creates a configuration container and returns it. It will return an error if something goes wrong while reading the configuration.
func New(fileLocation string, logger logs.SysLogger) (*ValuesContainer, error) {
	// Create a new config container
	// setup defaults
	nc := &ValuesContainer{
		InternalStateTableSize: 20,
		InternalControlChefRun: true,
		InternalPeriodicTimer:  30,
		InternalDebug:          false,
		InternalListenPort:     8901,
		InternalListenAddress:  "0.0.0.0",
		InternalCertPath:       "./cert.crt",
		InternalKeyPath:        "./key.key",
	}
	// Call OS_default for config files
	nc.writeConfigFileOSDefaults()

	// Read in the configuration found if any.
	err := nc.loadConfigFile(fileLocation, logger)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

// loadConfigFile reads the configuration file from the disk if it is there.
// If the file is not there then we just return nil and use the default values.
// If the file is there but in valid we return an error.
// If the file is good, we update the Values with values.
func (vc *ValuesContainer) loadConfigFile(fileLocation string, logger logs.SysLogger) error {
	// Load the struct with default values to start with.
	// This way we don't require every value to be available in the configuration file.
	if fileLocation == "" {
		fileLocation = defaultFileLocation
	}
	cf, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		logger.Info("Config file not found. Using default values.")
		return nil
	}

	// Set the Values struct to the value of the configuration that we
	// have obtained.
	vc.Lock()
	defer vc.Unlock()
	err = json.Unmarshal(cf, vc)
	if err != nil {
		// Create and return an error here.
		return fmt.Errorf("Config file found but not valid. Error was: %s", err)
	}

	return nil
}
