package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// CreateMockFile create a mocked configuration file. It takes a valuesContainer and writes that in json to the file.
// It returns the the file pointer. The caller is responsible for removing the file.
func CreateMockFile(fileContents *ValuesContainer) (*os.File, error) {
	// Get working directory
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("Can't create the testing file for TestReadConfigFile: %s", err)
	}

	// Create a temp testing file.
	f, err := ioutil.TempFile(dir, "test_")
	if err != nil {
		return nil, fmt.Errorf("Can't create the testing file for TestReadConfigFile: %s", err)
	}

	// Encode the json string
	jsonBytes, err := json.Marshal(fileContents)
	if err != nil {
		return nil, fmt.Errorf("Failed to encode json: %s", err)
	}

	// Write to the file
	if _, err = f.Write(jsonBytes); err != nil {
		return nil, fmt.Errorf("Failed to write the test date to the config file: %s", err)
	}

	// Close the file for writting
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("Failed to close config file for writing: %s", err)
	}

	return f, nil
}

// TestConfigFile is a shortcut method that will create a default testing configuration
func TestConfigFile() (*os.File, error) {
	vc := &ValuesContainer{
		InternalStateTableSize:    10,
		InternalControlChefRun:    false,
		InternalPeriodicTimer:     600,
		InternalDebug:             true,
		InternalLogLocation:       "potatoes",
		InternalStateFileLocation: "carrot",
		InternalListenPort:        18080,
		InternalListenAddress:     "127.0.0.1",
	}
	return CreateMockFile(vc)
}
