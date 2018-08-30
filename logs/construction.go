package logs

// SysLogger writes to the system log.
type SysLogger interface {
	Error(v ...interface{}) error
	Warning(v ...interface{}) error
	Info(v ...interface{}) error

	Errorf(format string, a ...interface{}) error
	Warningf(format string, a ...interface{}) error
	Infof(format string, a ...interface{}) error
}

var (
	// Logger is anything that can handle the sysLogger interface
	debugLogging = false
)
