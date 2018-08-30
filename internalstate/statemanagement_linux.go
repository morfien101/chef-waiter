package internalstate

import (
	"fmt"
)

// getStatePath will return a string that represents the location of the state file on disk
func getStatePath(prefix, statedb string) string {
	return fmt.Sprintf("%s/%s", prefix, statedb)
}
