package config

import (
	log "github.com/sirupsen/logrus"
	"os"
)

var Log *log.Logger

func init() {
	Log = log.New()
	Log.Out = os.Stdout
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
