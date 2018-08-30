package chefrunner

import (
	"fmt"

	"github.com/newvoicemedia/chef-waiter/cmd"
	"github.com/newvoicemedia/chef-waiter/logs"
)

// runChef - will run the command based on the OS
func (r *RunRequest) runChef(guid string) (exitCode int) {
	logs.DebugMessage(fmt.Sprintf("runChef(%s)", guid))
	_, _, exitCode = cmd.RunCommand("chef-client", "-L", r.chefLogWorker.GetLogPath(guid))
	return
}
