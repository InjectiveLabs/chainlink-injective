package logging

import (
	"os"

	"github.com/smartcontractkit/libocr/commontypes"
	log "github.com/xlab/suplog"
	bugsnagHook "github.com/xlab/suplog/hooks/bugsnag"
	debugHook "github.com/xlab/suplog/hooks/debug"
)

func Level(s string) log.Level {
	switch s {
	case "1", "error":
		return log.ErrorLevel
	case "2", "warn":
		return log.WarnLevel
	case "3", "info":
		return log.InfoLevel
	case "4", "debug":
		return log.DebugLevel
	default:
		return log.FatalLevel
	}
}

func NewSuplog(minLevel log.Level, useJSON bool) log.Logger {
	var formatter log.Formatter

	if useJSON {
		formatter = new(log.JSONFormatter)
	} else {
		formatter = new(log.TextFormatter)
	}

	appLogger := log.NewLogger(os.Stderr, formatter,
		debugHook.NewHook(log.DefaultLogger, &debugHook.HookOptions{
			StackTraceOffset: 1,
		}),
		bugsnagHook.NewHook(log.DefaultLogger, &bugsnagHook.HookOptions{
			Levels: []log.Level{
				log.PanicLevel,
				log.FatalLevel,
				log.ErrorLevel,
				// do not report anything below Error
			},
			StackTraceOffset: 1,
		}),
	)

	appLogger.(log.LoggerConfigurator).SetLevel(minLevel)
	appLogger.(log.LoggerConfigurator).SetStackTraceOffset(1)

	return appLogger
}

func WrapCommonLogger(l log.Logger) commontypes.Logger {
	return &logWrapper{
		l: l,
	}
}

type logWrapper struct {
	l log.Logger
}

func (ll *logWrapper) Trace(msg string, fields commontypes.LogFields) {
	ll.l.WithFields(mapFields(fields)).Trace(msg)
}

func (ll *logWrapper) Debug(msg string, fields commontypes.LogFields) {
	ll.l.WithFields(mapFields(fields)).Debug(msg)
}

func (ll *logWrapper) Info(msg string, fields commontypes.LogFields) {
	ll.l.WithFields(mapFields(fields)).Info(msg)
}

func (ll *logWrapper) Warn(msg string, fields commontypes.LogFields) {
	ll.l.WithFields(mapFields(fields)).Warning(msg)
}

func (ll *logWrapper) Error(msg string, fields commontypes.LogFields) {
	ll.l.WithFields(mapFields(fields)).Error(msg)
}

func mapFields(logFields commontypes.LogFields) log.Fields {
	return log.Fields(logFields)
}
