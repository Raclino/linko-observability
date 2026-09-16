package helpers

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

type closeFunc func() error

// InitializeLogger returns a logger writing to STDERR, and additionally to the
// file named by LINKO_LOG_FILE when that variable is set.
//
// The file stays open for the lifetime of the process: the logger writes to it
// on every request, so it must not be closed when this function returns.
func InitializeLogger() (*log.Logger, closeFunc, error) {
	linkoLogFile := os.Getenv("LINKO_LOG_FILE")
	if linkoLogFile == "" {
		return log.New(os.Stderr, "", log.LstdFlags), nil, nil
	}

	file, err := os.OpenFile(linkoLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open log file %s: %w", linkoLogFile, err)
	}

	bufferedFile := bufio.NewWriterSize(file, 8192)

	closeFunc := func() error {
		err := bufferedFile.Flush()
		if err != nil {
			return err
		}
		file.Close()
		return nil
	}

	multiWriter := io.MultiWriter(os.Stderr, bufferedFile)

	return log.New(multiWriter, "", log.LstdFlags), closeFunc, nil
}
