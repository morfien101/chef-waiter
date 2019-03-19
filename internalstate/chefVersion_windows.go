package internalstate

import (
	"errors"

	"github.com/morfien101/chef-waiter/cmd"
)

func chefVersion() (string, error) {
	stdout, _, exitCode := cmd.RunCommand("chef-client", "-v")
	if exitCode != 0 {
		return "", errors.New("Could not determin chef version")
	}
	return extractVersion(stdout), nil
}
