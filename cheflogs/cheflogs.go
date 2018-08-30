package cheflogs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/newvoicemedia/chef-waiter/config"
	"github.com/newvoicemedia/chef-waiter/logs"
)

// WorkerReadWriter is both a WorkerReader and a WorkerWriter
type WorkerReadWriter interface {
	WorkerReader
	WorkerWriter
}

// WorkerReader is used to describe the functions that are used to read data from the Worker.
type WorkerReader interface {
	IsLogAvailable(string) error
	GetLogPath(string) string
}

// WorkerWriter is used to describe the functuons that are used to write data to the Worker.
type WorkerWriter interface {
	RequestDelete(map[string]int64)
}

// Worker will hold the configuration and logger for the logs worker functions.
type Worker struct {
	LogWorkQ chan map[string]int64
	logger   logs.SysLogger
	config   config.Config
}

// New will return a new Chef logs worker. These are responsible for log clearing.
func New(config config.Config, logger logs.SysLogger) *Worker {
	return &Worker{
		logger:   logger,
		config:   config,
		LogWorkQ: make(chan map[string]int64, 10),
	}
}

// IsLogAvailable will return a indicator and an error which will tell you if the file is available on the disk.
func (w *Worker) IsLogAvailable(guid string) error {
	if _, err := os.Stat(w.GetLogPath(guid)); err != nil {
		// Bubble the error out and return to the caller.
		return err
	}
	return nil
}

// clearOldChefLogs will remove any logs that are deemed to be old
func (w *Worker) clearOldChefLogs(guidsToKeep map[string]int64) {
	allLogs, err := w.logsOnDisk()
	if err != nil {
		w.logger.Error(err)
	}
	// Delete the file if it needs to be removed.
	// Dirty code refactor later.
	// for each file in the <logs> check that it is not in the keep list
	// If not delete the file.
	for _, oldFile := range w.filesToDelete(guidsToKeep, allLogs) {
		if os.Remove(oldFile); err != nil {
			w.logger.Infof("Failed to delete %s. Error: %s", oldFile, err)
			continue
		}
		w.logger.Infof("Deleted file: %s\n", oldFile)
	}
}

func (w *Worker) logsOnDisk() ([]string, error) {
	// Get the logs that exist on the disk
	return filepath.Glob(fmt.Sprintf("%s/*", w.config.LogLocation()))
}

func (w *Worker) filesToDelete(guidsToKeep map[string]int64, allLogs []string) []string {
	oldFiles := make([]string, 0)
	for _, currentFile := range allLogs {
		del := true
		// Get check if the log is in the list of files.
		for guid := range guidsToKeep {
			if w.GetLogPath(guid) == currentFile {
				del = false
				break
			}
		}
		if del {
			oldFiles = append(oldFiles, currentFile)
		}
	}
	logs.DebugMessage(fmt.Sprintf("Files to delete: %s", strings.Join(oldFiles, ", ")))
	return oldFiles
}

// RequestDelete will add a guid map to a queue to have the chef files removed that are no
// longer required.
func (w *Worker) RequestDelete(GUIDmap map[string]int64) {
	w.LogWorkQ <- GUIDmap
}

// LogSweepEngine will invoke a run of the clearOldChefLogs function
func (w *Worker) LogSweepEngine() {
	for {
		select {
		case keepTheseGuids := <-w.LogWorkQ:
			w.clearOldChefLogs(keepTheseGuids)
		}
	}
}
