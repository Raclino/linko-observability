package helpers

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"

	pkgerr "github.com/pkg/errors"
)

// closeFunc releases the resources held by a logger. Callers must invoke it
// before the program exits, and must not log through the logger afterwards.
type closeFunc func() error

type stackTracer interface {
	error
	StackTrace() pkgerr.StackTrace
}

func handlerOptions(level slog.Level) *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level:       level,
		ReplaceAttr: replaceAttr,
	}
}

// InitializeLogger returns a logger writing everything from DEBUG up to STDERR,
// and, when logFile is set, INFO and above to that file as well.
//
// Writes to the file go through a buffer, so the returned closeFunc must be
// called before exiting: pending log lines are lost otherwise. The file itself
// stays open for the lifetime of the process, since the logger writes to it on
// every request.
func InitializeLogger(logFile string) (*slog.Logger, closeFunc, error) {
	stderrHandler := slog.NewTextHandler(os.Stderr, handlerOptions(slog.LevelDebug))

	if logFile == "" {
		// Nothing to release: STDERR is unbuffered, and closing it would break
		// whatever else the process writes there.
		noop := func() error { return nil }
		return slog.New(stderrHandler), noop, nil
	}

	file, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open log file: %w", err)
	}

	bufferedFile := bufio.NewWriterSize(file, 8192)
	fileHandler := slog.NewJSONHandler(bufferedFile, handlerOptions(slog.LevelInfo))

	logger := slog.New(slog.NewMultiHandler(
		stderrHandler,
		fileHandler,
	))

	closeLogger := func() error {
		if err := bufferedFile.Flush(); err != nil {
			file.Close()
			return fmt.Errorf("failed to flush log file: %w", err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
		return nil
	}

	return logger, closeLogger, nil
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key != "error" {
		return a
	}
	err, ok := a.Value.Any().(error)
	if !ok {
		return a
	}
	stackErr, ok := errors.AsType[stackTracer](err)
	if !ok {
		return a
	}
	return slog.GroupAttrs("error",
		slog.String("message", stackErr.Error()),
		slog.String("stack_trace", fmt.Sprintf("%+v", stackErr.StackTrace())),
	)
}
