package logs

import "fmt"

// FakeLogger can be used in tests to satify the logs.SysLogger interface.
// It can be set to print messages to the console or to ignore messages sent in.
type FakeLogger struct {
	output bool
}

func NewFakeLogger(outputEnabled bool) SysLogger {
	return &FakeLogger{output: outputEnabled}
}

// Output turns on or off the printing to STDOUT of this logger.s
func (fl *FakeLogger) Output(value bool) {
	fl.output = value
}

func (fl *FakeLogger) Error(v ...interface{}) error {
	if fl.output {
		fmt.Println(v...)
	}
	return nil
}
func (fl *FakeLogger) Warning(v ...interface{}) error {
	if fl.output {
		fmt.Println(v...)
	}
	return nil
}
func (fl *FakeLogger) Info(v ...interface{}) error {
	if fl.output {
		fmt.Println(v...)
	}
	return nil
}
func (fl *FakeLogger) Errorf(format string, a ...interface{}) error {
	if fl.output {
		fmt.Printf(format, a...)
	}
	return nil
}
func (fl *FakeLogger) Warningf(format string, a ...interface{}) error {
	if fl.output {
		fmt.Printf(format, a...)
	}
	return nil
}
func (fl *FakeLogger) Infof(format string, a ...interface{}) error {
	if fl.output {
		fmt.Printf(format, a...)
	}
	return nil
}
