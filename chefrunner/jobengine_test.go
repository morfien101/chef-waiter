package chefrunner

import (
	"fmt"
	"os"
	"testing"

	"github.com/Flaque/filet"

	"github.com/morfien101/chef-waiter/cheflogs"
	"github.com/morfien101/chef-waiter/config"
	"github.com/morfien101/chef-waiter/internalstate"
	"github.com/morfien101/chef-waiter/logs"
)

type logworker struct{}

func (lw *logworker) GetLogPath(guid string) string {
	return fmt.Sprintf("/var/log/chefwaiter/%s.log", guid)
}

func TestCustomJob(t *testing.T) {
	testGUID := "1234-1234-1234"
	testRecipe := "recipe[chefwaiter::test]"
	testLogLocation := "/var/log/chefwaiter"
	testDir := filet.TmpDir(t, "")
	defer os.RemoveAll(testDir)

	configContainer := &config.ValuesContainer{
		InternalStateFileLocation: testDir,
		InternalLogLocation:       testLogLocation,
	}
	fakelogger := logs.NewFakeLogger(false)
	chefLogger := cheflogs.New(configContainer, fakelogger)

	st := internalstate.New(configContainer, chefLogger, fakelogger)
	st.Status = map[string]*internalstate.JobDetails{
		testGUID: &internalstate.JobDetails{
			CustomRun:       true,
			CustomRunString: testRecipe,
		},
	}

	rr := &RunRequest{
		state:         st,
		chefLogWorker: chefLogger,
	}

	args := rr.chefClientArguments(testGUID)
	expectedLogPath := fmt.Sprintf("%s/%s.log", testLogLocation, testGUID)
	collectedRecipe := args[len(args)-1]
	// Test the log location
	if args[1] != expectedLogPath {
		t.Logf("Log path is incorrect. Got: %s, Want:%s", args[1], expectedLogPath)
		t.Fail()
	}
	// Test the recipe ro run
	if collectedRecipe != testRecipe {
		t.Logf("Recipe is incorrect. Got: %s, Want: %s", collectedRecipe, testRecipe)
		t.Fail()
	}
}
