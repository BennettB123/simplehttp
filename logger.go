package simplehttp

// Logger is a simple interface to be given to the [Server].
type Logger interface {
	// LogMessage accepts a string to be logged.
	LogMessage(string)
}

type nilLogger struct{}

func (nl nilLogger) LogMessage(string) {}
