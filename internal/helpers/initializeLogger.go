package helpers

import (
	"fmt"
	"io"
	"log"
	"os"
)

// InitializeLogger returns a logger writing to STDERR, and additionally to the
// file named by LINKO_LOG_FILE when that variable is set.
//
// The file stays open for the lifetime of the process: the logger writes to it
// on every request, so it must not be closed when this function returns.
func InitializeLogger() (*log.Logger, error) {
	linkoLogFile := os.Getenv("LINKO_LOG_FILE")
	if linkoLogFile == "" {
		return log.New(os.Stderr, "", log.LstdFlags), nil
	}

	file, err := os.OpenFile(linkoLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", linkoLogFile, err)
	}

	multiWriter := io.MultiWriter(os.Stderr, file)
	return log.New(multiWriter, "", log.LstdFlags), nil
}
