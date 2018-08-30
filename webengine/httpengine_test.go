package webengine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/newvoicemedia/chef-waiter/cheflogs"
	"github.com/newvoicemedia/chef-waiter/chefrunner"
	"github.com/newvoicemedia/chef-waiter/config"
	"github.com/newvoicemedia/chef-waiter/internalstate"
	"github.com/newvoicemedia/chef-waiter/logs"
)

type FakeAppStatus struct {
	jsonError bool
}

// NewFakeAppStatus will create an app status that is constant with your supplied
// values.
func NewFakeAppStatus() *FakeAppStatus {
	return &FakeAppStatus{}
}

func (fa *FakeAppStatus) JSONEncoded() ([]byte, error) {
	if fa.jsonError {
		return []byte(`{"service_name":"C`), fmt.Errorf("Mocking error")
	}
	return []byte(`{"service_name":"ChefWaiter","hostname":"randy-laptop","uptime":1520949021,"version":"17.10.200","chef_version":"13.6.4","healthy":true,"in_maintenance_mode":false,"last_run_id":"88527564-4919-4933-8c7d-0b4bdb81dc18"}`), nil
}

func cleanup(f *os.File, t *testing.T) {
	if err := os.Remove(f.Name()); err != nil {
		t.Fatalf("Deleting file %s failed, Error: %s", f.Name(), err)
	}
}

var testconfigFile *os.File

func url(uri string) string {
	serverAddres := "http://localhost:8901"
	return fmt.Sprintf("%s%s", serverAddres, uri)
}

func genNewHTTPServer(t *testing.T) *HTTPEngine {
	// HTTP Engine needs this
	// state internalstate.StateTableReadWriter,
	// appState internalstate.AppStatusReader,
	// worker chefrunner.Worker,
	// chefLogsWorker cheflogs.WorkerReader,
	// logger logs.SysLogger,

	//Internal state needs this
	// config config.Config,
	// chefLogsWorker cheflogs.WorkerWriter,
	// logger logs.SysLogger,
	logger := logs.NewFakeLogger(false)
	configFile, err := config.TestConfigFile()
	if err != nil {
		t.Fatal(err)
	}

	config, err := config.New(configFile.Name(), logger)
	cleanup(configFile, t)
	if err != nil {
		t.Fatalf("Failed to create the config handler. Error: %s", err)
	}
	cheflogsworker := cheflogs.NewFakeChefLogWorker("")
	internalstate := internalstate.New(config, cheflogsworker, logger)
	appstate := NewFakeAppStatus()
	worker := chefrunner.NewFakeChefRunnerWorker(false)
	return New(internalstate, appstate, worker, cheflogsworker, logger)
}

func TestStatus(t *testing.T) {
	webEngine := genNewHTTPServer(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, url("/_status"), nil)
	webEngine.ServeHTTP(w, r)
	result := w.Result()

	defer result.Body.Close()
	if result.StatusCode != 200 {
		t.Errorf("/_status did not return a 200. Got: %v", w.Result().StatusCode)
	}
}

func TestLock(t *testing.T) {
	webEngine := genNewHTTPServer(t)

	type returnJSON struct {
		Locked bool `json:"Locked"`
	}

	testSquecence := []struct {
		name         string
		url          string
		JSONReturn   bool
		registerChef bool
		responceCode int
	}{
		{name: "First encounter", url: "/chef/lock", JSONReturn: false},
		{name: "Set the lock", url: "/chef/lock/set", JSONReturn: true},
		{name: "Check locked", url: "/chef/lock", JSONReturn: true},
		{name: "Register Chef run", url: "/chefclient", registerChef: true, responceCode: 403},
		{name: "Remove lock", url: "/chef/lock/remove", JSONReturn: false},
		{name: "Check unlocked", url: "/chef/lock", JSONReturn: false},
	}

	for _, tc := range testSquecence {

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, url(tc.url), nil)
		webEngine.ServeHTTP(w, r)
		result := w.Result()

		body, err := ioutil.ReadAll(result.Body)
		if err != nil {
			t.Fatalf("%s: Failed to pull out the body from %s. Error: %s",
				tc.name,
				tc.url,
				err,
			)
		}
		defer result.Body.Close()
		if tc.responceCode == 0 {
			tc.responceCode = 200
		}
		if result.StatusCode != tc.responceCode {
			t.Errorf("%s: %s responce code incorrect. Want: %d. Got: %d.",
				tc.name,
				tc.url,
				tc.responceCode,
				result.StatusCode,
			)
		}
		if tc.registerChef {
			JSONResponce := &struct{ Error string }{}
			if err := json.Unmarshal(body, JSONResponce); err != nil {
				t.Errorf("%s: failed to pull the body out the request. Error: %s", tc.name, err)
			} else {
				if errMsg := "Chefwaiter is locked"; JSONResponce.Error != errMsg {
					t.Errorf("%s: Error message is not correct. Want: %s, Got: %s",
						tc.name,
						errMsg,
						JSONResponce.Error,
					)
				}
			}
		} else {
			res := &returnJSON{}
			json.Unmarshal(body, res)
			if res.Locked != tc.JSONReturn {
				t.Errorf(
					"%s: Inspecting the lock did not product what we want. Got: %t, want %t",
					tc.name,
					res.Locked,
					tc.JSONReturn,
				)
			}
		}
	}

}
