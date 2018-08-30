package config

import (
	"os"
	"testing"

	"github.com/newvoicemedia/chef-waiter/logs"
)

func generateFileContent() []*ValuesContainer {
	return []*ValuesContainer{
		&ValuesContainer{
			InternalStateTableSize:    10,
			InternalControlChefRun:    false,
			InternalPeriodicTimer:     600,
			InternalDebug:             true,
			InternalLogLocation:       "potatoes",
			InternalStateFileLocation: "carrot",
			InternalListenPort:        1234,
			InternalListenAddress:     "1.2.3.4",
			InternalCertPath:          "./cert.pem",
			InternalKeyPath:           "./key.key",
			InternalTLSEnabled:        false,
		},
		&ValuesContainer{
			InternalStateTableSize:    100000,
			InternalControlChefRun:    true,
			InternalPeriodicTimer:     600324234,
			InternalDebug:             false,
			InternalLogLocation:       "ajfhoaskfmlkasmdoaijmcapsjpaowsmcpam",
			InternalStateFileLocation: "carasdasdsadsadasdwewsarot",
			InternalListenPort:        14521452145214,
			InternalListenAddress:     "0.0.0.0",
			InternalCertPath:          "./cert.pem",
			InternalKeyPath:           "./key.key",
			InternalTLSEnabled:        true,
		},
	}
}

func TestReadConfigFile(t *testing.T) {
	for _, fileContents := range generateFileContent() {

		f, err := CreateMockFile(fileContents)
		if err != nil {
			t.Fatalf("Creating a fake configuration file failed. Error: %s", err)
		}

		// Load the configuration for testing.
		values, err := New(f.Name(), logs.NewFakeLogger(false))
		// Do the tests
		// InternalStateTableSize
		if values.StateTableSize() != fileContents.InternalStateTableSize {
			t.Errorf("StateTableSize is incorrect, Wanted: %v, Got; %v", fileContents.InternalStateTableSize, values.StateTableSize())
		}
		// InternalControlChefRun
		if values.ControlChefRun() != fileContents.InternalControlChefRun {
			t.Errorf("ControlChefRun is incorrect, Wanted: %v, Got: %v", fileContents.InternalControlChefRun, values.ControlChefRun())
		}
		// InternalPeriodicTimer
		if values.PeriodicTimer() != fileContents.InternalPeriodicTimer {
			t.Errorf("PeriodicTimer is incorrect, Wanted: %v, Got: %v", fileContents.InternalPeriodicTimer, values.PeriodicTimer())
		}
		// InternalDebug
		if values.Debug() != fileContents.InternalDebug {
			t.Errorf("Debug is incorrect, Wanted: %v, Got: %v", fileContents.InternalDebug, values.Debug())
		}
		// InternalLogLocation
		if values.LogLocation() != fileContents.InternalLogLocation {
			t.Errorf("LogLocation is incorrect, Wanted: %v, Got: %v", fileContents.InternalLogLocation, values.LogLocation())
		}
		// InternalStateFileLocation
		if values.StateFileLocation() != fileContents.InternalStateFileLocation {
			t.Errorf("StateFileLocation is incorrect, Wanted: %v, Got %v", fileContents.InternalStateFileLocation, values.StateFileLocation())
		}
		// InternalListenPort
		if values.ListenPort() != fileContents.InternalListenPort {
			t.Errorf("ListenPort is incorrect, Wanted: %v, Got: %v", fileContents.InternalListenPort, values.ListenPort())
		}
		// InternalListenAddress
		if values.ListenAddress() != fileContents.InternalListenAddress {
			t.Errorf("ListenAddress is incorrect, Wanted: %v, Got: %v", fileContents.InternalListenAddress, values.ListenAddress())
		}
		// TLS Tests
		if values.TLSEnabled() != fileContents.InternalTLSEnabled {
			t.Errorf("InternalTLSEnabled is incorrect. Wanted: %v, Got: %v", fileContents.InternalTLSEnabled, values.TLSEnabled())
		}
		if values.CertPath() != fileContents.InternalCertPath {
			t.Errorf("InternalCertPath is incorrect. Wanted: %v, Got: %v", fileContents.InternalCertPath, values.CertPath())
		}
		if values.KeyPath() != fileContents.InternalKeyPath {
			t.Errorf("InternalKeyPath is incorrect. Wanted: %v, Got: %v", fileContents.InternalKeyPath, values.KeyPath())
		}

		err = os.Remove(f.Name())
		if err != nil {
			t.Errorf("Failed to remove the test file: %s", err)
		}
	}
}
