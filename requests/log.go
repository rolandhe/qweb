package requests

type LogLevel int

const (
	LOG_LEVEL_NONE   LogLevel = 0
	LOG_LEVEL_PARAM  LogLevel = 1
	LOG_LEVEL_RETURN LogLevel = 2
	LOG_LEVEL_ALL    LogLevel = 3
)
