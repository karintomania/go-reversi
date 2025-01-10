package main

import (
	"log/slog"
	"os"
	"path/filepath"
)

var logger *slog.Logger

func NewLogger(level slog.Level) *slog.Logger {
	// only show source in debug
	addSource := level == slog.LevelDebug

	logger := slog.New(slog.NewTextHandler(&stdOutMod{}, &slog.HandlerOptions{
		AddSource: addSource,
		Level:     level, // Set the desired log level
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source, _ := a.Value.Any().(*slog.Source)
				if source != nil {
					source.File = filepath.Base(source.File)
				}
			}
			return a
		}}))

	return logger
}

// append \r to all writing, otherwise, the RawMode makes logs hard to read
type stdOutMod struct {
}

func (s *stdOutMod) Write(p []byte) (n int, err error) {
	p = append([]byte("\r"), p...)
	return os.Stdout.Write(p)
}
