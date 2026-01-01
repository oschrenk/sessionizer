package util

import (
	"log"
	"os"
)

// DebugLog writes debug messages to /tmp/sessionizer-debug.log
func DebugLog(format string, args ...interface{}) {
	f, err := os.OpenFile("/tmp/sessionizer-debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	logger := log.New(f, "", log.LstdFlags)
	logger.Printf(format, args...)
}
