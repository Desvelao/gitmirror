package logger

import (
	"io"
	"os"
	"time"
	"github.com/rs/zerolog"
)

var Log zerolog.Logger

// InitLogger initializes the logger to write to a file and standard output.
func Init(level bool, logFile string) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if level {
		zerolog.SetGlobalLevel(zerolog.DebugLevel) // TODO: make this configurable zerolog.InfoLevel
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	if logFile == "" {
		Log = zerolog.New(output).With().Timestamp().Logger()
		return nil
	}

	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	writer := io.MultiWriter(output, f)
	Log = zerolog.New(writer).With().Timestamp().Logger()

	return nil
}

func Debug(msg string) {
	Log.Debug().Msg(msg)
}

func Info(msg string) {
	Log.Info().Msg(msg)
}

func Error(msg string) {
	Log.Error().Msg(msg)
}

func Warn(msg string) {
	Log.Warn().Msg(msg)
}
