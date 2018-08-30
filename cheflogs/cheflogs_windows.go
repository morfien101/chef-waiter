package cheflogs

import (
	"fmt"
	"strings"
)

// GetLogPath will return a string that points to the log for a guid on the disk.
func (w *Worker) GetLogPath(guid string) (logPath string) {
	return fmt.Sprintf("%s\\%s.log", w.cleanLogLocation(), guid)
}

func (w *Worker) cleanLogLocation() string {
	loglocation := w.config.LogLocation()
	return strings.Replace(loglocation, "/", `\`, -1)
}
