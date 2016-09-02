package hookfs

import (
	log "github.com/Sirupsen/logrus"
	"os"
)

// minimum log level
const LogLevelMin = 0

// maximum log level
const LogLevelMax = 2

var logLevel int

func initLog() {
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stderr)
}

// Getter for log level
func LogLevel() int {
	return logLevel
}

// Setter for log level. newLevel must be >= LogLevelMin, and <= LogLevelMax.
func SetLogLevel(newLevel int) {
	if newLevel < LogLevelMin || newLevel > LogLevelMax {
		log.Fatalf("Bad log level: %d (must be %d..%d)", newLevel, LogLevelMin, LogLevelMax)
	}

	logLevel = newLevel
	if logLevel > 0 {
		log.SetLevel(log.DebugLevel)
	}
}

func init() {
	initLog()
	SetLogLevel(0)
}
