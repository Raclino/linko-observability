package helpers

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// closeFunc releases the resources held by a logger. Callers must invoke it
// before the program exits, and must not log through the logger afterwards.
type closeFunc func() error

// InitializeLogger returns a logger writing to STDERR, and additionally to the
// file named by LINKO_LOG_FILE when that variable is set.
//
// Writes to the file go through a buffer, so the returned CloseFunc must be
// called before exiting: pending log lines are lost otherwise. The file itself
// stays open for the lifetime of the process, since the logger writes to it on
// every request.
func InitializeLogger(logFile string) (*slog.Logger, closeFunc, error) {
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}
		bufferedFile := bufio.NewWriterSize(file, 8192)
		multiWriter := io.MultiWriter(os.Stderr, bufferedFile)
		close := func() error {
			if err := bufferedFile.Flush(); err != nil {
				return fmt.Errorf("failed to flush log file: %w", err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("failed to close log file: %w", err)
			}
			return nil
		}

		return slog.New(slog.NewTextHandler(multiWriter, nil)), close, nil
	}
	close := func() error {
		return nil
	}
	return slog.New(slog.NewTextHandler(os.Stderr, nil)), close, nil
}
