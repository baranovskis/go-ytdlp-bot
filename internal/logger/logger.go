package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func GetLogger() zerolog.Logger {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().Msg("success setup zerolog logger")

	return log.Logger
}

// GetLoggerWithDB creates a logger that writes to both console and a DBWriter.
func GetLoggerWithDB(dbWriter *DBWriter) zerolog.Logger {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	multi := io.MultiWriter(consoleWriter, dbWriter)
	logger := zerolog.New(multi).With().Timestamp().Logger()

	logger.Info().Msg("success setup zerolog logger with database writer")

	return logger
}
