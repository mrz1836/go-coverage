package providers

import (
	"fmt"
	"os"
)

// DefaultLogger provides a simple implementation of the Logger interface
type DefaultLogger struct {
	debug   bool
	verbose bool
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger(debug, verbose bool) *DefaultLogger {
	return &DefaultLogger{
		debug:   debug,
		verbose: verbose,
	}
}

// Debug logs debug messages (only if debug mode is enabled)
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	if l.debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}

// Info logs informational messages
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	if l.verbose || l.debug {
		fmt.Fprintf(os.Stderr, "[INFO] "+format+"\n", args...)
	}
}

// Warn logs warning messages
func (l *DefaultLogger) Warn(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[WARN] "+format+"\n", args...)
}

// Error logs error messages
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}
