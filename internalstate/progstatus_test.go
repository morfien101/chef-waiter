package internalstate

import (
	"regexp"
	"testing"

	"github.com/morfien101/chef-waiter/logs"
)

func TestWhiteList(t *testing.T) {
	type fakeConfig struct {
		whitelist      bool
		whitelistItems []string
	}

	fc := &fakeConfig{
		whitelist:      true,
		whitelistItems: []string{"recipe[test]", "role[test::something]"},
	}
	stateTableMock := &StateTable{
		Status: make(map[string]*JobDetails),
	}
	logger := logs.NewFakeLogger(false)
	appState := NewAppStatus("0.0.1", stateTableMock, logger)
	appState.SetWhiteListing(fc.whitelist, fc.whitelistItems)
	b, err := appState.JSONEncoded()
	if err != nil {
		t.Logf("Failed to JSON encode app state, Error: %s", err)
		t.FailNow()
	}

	matchers := []string{
		`whitelisting_enabled": true`,
		`"recipe\[test\]"`,
		`"role\[test::something\]"`,
	}
	ok := true
	for _, match := range matchers {
		re := regexp.MustCompile(match)
		isMatch := re.MatchString(string(b))
		if !isMatch {
			ok = false
			t.Logf("matcher '%s' did not find anything", match)
		}
	}
	if !ok {
		t.Logf("InternalState did not match the regex looking for the whitelisting. Current State:\n%s", string(b))
		t.Fail()
	}
}
