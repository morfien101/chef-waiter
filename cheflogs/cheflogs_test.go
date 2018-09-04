package cheflogs

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/morfien101/chef-waiter/config"
	"github.com/morfien101/chef-waiter/logs"
	uuid "github.com/satori/go.uuid"
)

func TestDeleteOldLogs(t *testing.T) {
	var logsPath string
	if runtime.GOOS == "windows" {
		currentDir := `c:`
		logsPath = fmt.Sprintf(`%s/fakelogs`, currentDir)
	} else {
		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatal("Could not determine working dir to make fake logs")
		}
		logsPath = fmt.Sprintf("%s/fakelogs", currentDir)
	}

	// guid | time epoc
	filesLists := make([]map[string]int64, 0)

	err := os.Mkdir(logsPath, 0777)
	if err != nil {
		t.Fatalf("Failed to create the fake logs directory")
	}

	defer func() {
		if runtime.GOOS == "windows" {
			logsPath = strings.Replace(logsPath, "/", `\`, -1)
		}
		err := os.RemoveAll(logsPath)
		if err != nil {
			t.Logf("Failed to remove the folder. Error: %s", err)
		}
	}()

	configContainer := &config.ValuesContainer{
		InternalStateTableSize:    5,
		InternalLogLocation:       logsPath,
		InternalStateFileLocation: "statefile.db",
		InternalListenPort:        14521452145214,
	}
	for i := 0; i < 2; i++ {
		filesMap := make(map[string]int64)

		for j := 0; j <= configContainer.InternalStateTableSize; j++ {
			guid := uuid.NewV4().String()
			_, err := os.Create(fmt.Sprintf("%s/%s.log", logsPath, guid))
			if err != nil {
				t.Fatalf("Failed to create a test file. Error: %s", err)
			}
			filesMap[guid] = time.Now().Unix()
		}

		filesLists = append(filesLists, filesMap)
		time.Sleep(time.Millisecond * 250)
	}
	fakelogger := logs.NewFakeLogger(true)
	logs.TurnDebuggingOn(fakelogger, true)
	chefLogger := New(configContainer, fakelogger)
	chefLogger.clearOldChefLogs(filesLists[0])

	leftOverFiles, err := chefLogger.logsOnDisk()
	if err != nil {
		t.Fatalf("Could not determine the folder for left over files. Error: %s", err)
	}
	t.Log(filesLists)
	// get all the files left on the disk
	// does the file which should be delete exist on the disk?
	for _, lof := range leftOverFiles {
		if _, ok := filesLists[1][lof]; ok {
			t.Errorf("Found %s, it should be gone.", lof)
		}
	}
}
