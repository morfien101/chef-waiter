package cheflogs

import (
	"fmt"
)

// GetLogPath will return a string that points to the log for a guid on the disk.
func (w *Worker) GetLogPath(guid string) (logPath string) {
	return fmt.Sprintf("%s/%s.log", w.config.LogLocation(), guid)
}
