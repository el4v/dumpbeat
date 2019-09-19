package log

import (
	root "dumpbeat/pkg"
	"github.com/sirupsen/logrus"
	"os"
)

var logger = logrus.New()

// ConfigureLogging ...
func ConfigureLogging() error {
	cfg := root.GetConfig()
	//logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return err
	}
	logger.SetLevel(logLevel)
	return nil
}

func withHostname() *logrus.Entry {
	cfg := root.GetConfig()
	return logger.WithField("hostname", cfg.NodeName)
}

func AddFields(fields map[string]interface{}) *logrus.Entry {
	return withHostname().WithFields(fields)
}

// Info ...
func Info(args ...interface{}) {
	withHostname().Info(args...)
}

// Debug ...
func Debug(args ...interface{}) {
	withHostname().Debug(args...)
}

// Error ...
func Error(args ...interface{}) {
	withHostname().Error(args...)
}

// Fatal ...
func Fatal(args ...interface{}) {
	withHostname().Fatal(args...)
}
