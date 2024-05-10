package simplehttp

type Logger interface {
	LogMessage(string)
}

// Default logger for Server that discards all logs
type nilLogger struct{}

func (nl nilLogger) LogMessage(string) {}
