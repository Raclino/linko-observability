package helpers

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

// CloseFunc releases the resources held by a logger. Callers must invoke it
// before the program exits, and must not log through the logger afterwards.
type CloseFunc func() error

// InitializeLogger returns a logger writing to STDERR, and additionally to the
// file named by LINKO_LOG_FILE when that variable is set.
//
// Writes to the file go through a buffer, so the returned CloseFunc must be
// called before exiting: pending log lines are lost otherwise. The file itself
// stays open for the lifetime of the process, since the logger writes to it on
// every request.
func InitializeLogger() (*log.Logger, CloseFunc, error) {
	linkoLogFile := os.Getenv("LINKO_LOG_FILE")
	if linkoLogFile == "" {
		noop := func() error { return nil }
		return log.New(os.Stderr, "", log.LstdFlags), noop, nil
	}

	file, err := os.OpenFile(linkoLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open log file %s: %w", linkoLogFile, err)
	}

	bufferedFile := bufio.NewWriterSize(file, 8192)

	closeLogger := func() error {
		if err := bufferedFile.Flush(); err != nil {
			file.Close()
			return fmt.Errorf("failed to flush log buffer: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
		return nil
	}

	multiWriter := io.MultiWriter(os.Stderr, bufferedFile)

	return log.New(multiWriter, "", log.LstdFlags), closeLogger, nil
}
