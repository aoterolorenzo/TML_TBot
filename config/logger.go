package config

import (
	log "github.com/sirupsen/logrus"
)

var Log *log.Logger

func init() {
	Log = log.New()
	switch Settings.LogLevel {
	case "debug":
		Log.Level = log.DebugLevel
	case "info":
		Log.Level = log.InfoLevel
	case "fatal":
		Log.Level = log.FatalLevel
	case "trace":
		Log.Level = log.TraceLevel
	case "warn":
		Log.Level = log.WarnLevel
	case "error":
		Log.Level = log.ErrorLevel
	case "panic":
		Log.Level = log.PanicLevel
	}
}
