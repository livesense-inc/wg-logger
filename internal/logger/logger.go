package logger

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/livesense-inc/wg-logger/internal/config"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func getLogLevel(config *config.Config) (loglevel zerolog.Level) {
	switch strings.ToLower(config.LogLevel) {
	case "debug":
		loglevel = zerolog.DebugLevel
	case "info":
		loglevel = zerolog.InfoLevel
	case "warn":
		loglevel = zerolog.WarnLevel
	case "error":
		loglevel = zerolog.ErrorLevel
	default:
		loglevel = zerolog.InfoLevel
	}
	return
}

func NewFileLogger(config *config.Config) (*zerolog.Logger, *zerolog.Logger) {
	stdLogLumberjack := &lumberjack.Logger{
		Filename:  config.EventLogPath,
		MaxSize:   config.LogMaxMB,   // megabytes
		MaxAge:    config.LogMaxDays, // days
		LocalTime: true,
	}

	errLogLumberjack := &lumberjack.Logger{
		Filename:  config.DaemonLogPath,
		MaxSize:   config.LogMaxMB,   // megabytes
		MaxAge:    config.LogMaxDays, // days
		LocalTime: true,
	}

	zerolog.SetGlobalLevel(getLogLevel(config))

	loggerStd := zerolog.New(stdLogLumberjack).With().Timestamp().Logger()
	loggerErr := zerolog.New(errLogLumberjack).With().Timestamp().Logger()

	// rotate logs when SIGHUP received
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	go func() {
		for {
			<-c
			_ = stdLogLumberjack.Rotate()
			_ = errLogLumberjack.Rotate()
		}
	}()

	return &loggerStd, &loggerErr
}

func NewConsoleLogger(config *config.Config) (*zerolog.Logger, *zerolog.Logger) {
	zerolog.SetGlobalLevel(getLogLevel(config))

	outputStd := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	loggerStd := zerolog.New(outputStd).With().Timestamp().Logger()
	outputErr := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	loggerErr := zerolog.New(outputErr).With().Timestamp().Logger()

	return &loggerStd, &loggerErr
}
