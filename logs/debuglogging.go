package logs

type debugLogger struct {
	logger SysLogger
	debug  bool
}

var debuglogger debugLogger

// TurnDebuggingOn will tell the logger to log debug messages.
// They appear as info messages due to limits in the logging engine
// used to run the service.
func TurnDebuggingOn(logger SysLogger, debugging bool) {
	debuglogger = debugLogger{
		logger: logger,
		debug:  debugging,
	}
}

// DebugMessage send a debug message to the systems logger.
func DebugMessage(msg string) {
	if debuglogger.debug {
		debuglogger.logger.Info("[DEBUG]", msg)
	}
}
