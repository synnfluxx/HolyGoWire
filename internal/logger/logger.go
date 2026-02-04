package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	Log *logrus.Logger
)

func InitLogger() {
	Log = logrus.New()

	Log.Out = os.Stdout

	switch os.Getenv("LOG_LEVEL") {
		case "info":
		Log.SetLevel(logrus.InfoLevel)
		case "debug":
		Log.SetLevel(logrus.DebugLevel)
		case "error":
		Log.SetLevel(logrus.ErrorLevel)
		case "warn":
		Log.SetLevel(logrus.WarnLevel)
	}	
}