package cheflogs

import "os"

type ChefLogsTest struct {
	FakeLogPath string
}

func (c *ChefLogsTest) IsLogAvailable(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return nil
	}
	return err
}

func (c *ChefLogsTest) GetLogPath(path string) string {
	return c.FakeLogPath
}

func dummyChefLogContent() string {
	return `
This is a test chef waiter log.
chef exited or something.	
`
}

func (c ChefLogsTest) RequestDelete(map[string]int64) {}

// NewFakeChefLogWorker will return a thing that represents a chef log worker.
// It would be able to read a single log. You can supply the text you want in
// the log as content.
func NewFakeChefLogWorker(content string) *ChefLogsTest {
	var c string
	if len(content) > 0 {
		c = content
	} else {
		c = dummyChefLogContent()
	}
	return &ChefLogsTest{FakeLogPath: c}
}
